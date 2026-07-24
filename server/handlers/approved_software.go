package handlers

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "image/png"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ListApprovedSoftware 获取核准软件列表
func ListApprovedSoftware(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	keyword := c.Query("keyword")
	licenseType := c.Query("license_type")
	needUpdate := c.Query("need_update")

	db := database.GetDB()
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	if licenseType != "" {
		db = db.Where("license_type = ?", licenseType)
	}
	if needUpdate != "" {
		db = db.Where("need_update = ?", needUpdate == "true")
	}

	var total int64
	db.Model(&models.ApprovedSoftware{}).Count(&total)

	var list []models.ApprovedSoftware
	offset := (page - 1) * pageSize
	if err := db.Order("id desc").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list, "total": total, "page": page, "page_size": pageSize})
}

// ListApprovedSoftwareNeedUpdate 获取需要更新的核准软件
func ListApprovedSoftwareNeedUpdate(c *gin.Context) {
	var list []models.ApprovedSoftware
	if err := database.GetDB().Where("need_update = ?", true).Order("id desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

// CreateApprovedSoftware 创建核准软件
func CreateApprovedSoftware(c *gin.Context) {
	var sw models.ApprovedSoftware
	if err := c.ShouldBindJSON(&sw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := database.GetDB().Create(&sw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "名称", NewValue: sw.Name},
		{FieldName: "Version", FieldLabel: "版本", NewValue: sw.Version},
		{FieldName: "LatestVersion", FieldLabel: "最新版本", NewValue: sw.LatestVersion},
		{FieldName: "NeedUpdate", FieldLabel: "需要更新", NewValue: fmt.Sprintf("%v", sw.NeedUpdate)},
		{FieldName: "UpdateReason", FieldLabel: "更新原因", NewValue: sw.UpdateReason},
		{FieldName: "Vendor", FieldLabel: "厂商", NewValue: sw.Vendor},
		{FieldName: "VendorWebsite", FieldLabel: "厂商网站", NewValue: sw.VendorWebsite},
		{FieldName: "LicenseType", FieldLabel: "许可证类型", NewValue: sw.LicenseType},
		{FieldName: "Purpose", FieldLabel: "用途", NewValue: sw.Purpose},
	}
	services.LogOperation(username, displayName, "创建核准软件", "approved_software", sw.ID, sw.Name, "", c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": sw})
}

// UpdateApprovedSoftware 更新核准软件
func UpdateApprovedSoftware(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var sw models.ApprovedSoftware
	if err := database.GetDB().First(&sw, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "软件不存在"})
		return
	}

	// 保存旧值快照
	oldSw := sw

	var input struct {
		Name          string `json:"name" binding:"required"`
		Version       string `json:"version"`
		LatestVersion string `json:"latest_version"`
		NeedUpdate    bool   `json:"need_update"`
		UpdateReason  string `json:"update_reason"`
		Vendor        string `json:"vendor"`
		VendorWebsite string `json:"vendor_website"`
		LicenseType   string `json:"license_type"`
		Purpose       string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	sw.Name = input.Name
	sw.Version = input.Version
	sw.LatestVersion = input.LatestVersion
	sw.NeedUpdate = input.NeedUpdate
	sw.UpdateReason = input.UpdateReason
	sw.Vendor = input.Vendor
	sw.VendorWebsite = input.VendorWebsite
	sw.LicenseType = input.LicenseType
	sw.Purpose = input.Purpose

	if err := database.GetDB().Save(&sw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("approved_software")
	details := services.DiffStructs(oldSw, sw, fieldLabels)
	services.LogOperation(username, displayName, "更新核准软件", "approved_software", sw.ID, sw.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": sw})
}

// DeleteApprovedSoftware 删除核准软件
func DeleteApprovedSoftware(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var sw models.ApprovedSoftware
	if err := database.GetDB().First(&sw, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "软件不存在"})
		return
	}
	// 删除关联记录
	database.GetDB().Where("approved_software_id = ?", id).Delete(&models.AssetSoftware{})
	if err := database.GetDB().Unscoped().Delete(&sw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("approved_software")
	details := services.DiffStructs(sw, models.ApprovedSoftware{}, fieldLabels)
	services.LogOperation(username, displayName, "删除核准软件", "approved_software", sw.ID, sw.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ListAssetSoftware 获取资产与软件关联列表（展示所有资产，附带已关联软件信息）
func ListAssetSoftware(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	search := c.Query("search")
	softwareIDsParam := c.Query("software_ids") // 逗号分隔的软件ID
	db := database.GetDB()

	var total int64
	var assets []models.Asset
	offset := (page - 1) * pageSize

	hasSearch := search != ""
	hasSoftwareIDs := softwareIDsParam != ""

	if hasSearch || hasSoftwareIDs {
		// 分别收集两种过滤条件的资产ID集合
		var searchIDSet, swIDSet map[uint]bool

		// 1. 文本搜索：按计算机名或IP地址模糊匹配
		if hasSearch {
			searchLower := strings.ToLower(search)
			searchIDSet = make(map[uint]bool)
			var directAssets []models.Asset
			db.Where("LOWER(computer_name) LIKE ? OR ip_address LIKE ?",
				"%"+searchLower+"%", "%"+search+"%").
				Select("id").Find(&directAssets)
			for _, a := range directAssets {
				searchIDSet[a.ID] = true
			}
		}

		// 2. 软件筛选：按选中的软件ID查找资产
		if hasSoftwareIDs {
			ids := parseUintList(softwareIDsParam)
			if len(ids) > 0 {
				swIDSet = make(map[uint]bool)
				type assetIDRow struct {
					AssetID uint `gorm:"column:asset_id"`
				}
				var swAssetIDs []assetIDRow
				db.Model(&models.AssetSoftware{}).
					Where("approved_software_id IN ?", ids).
					Select("DISTINCT asset_software.asset_id").
					Find(&swAssetIDs)
				for _, r := range swAssetIDs {
					swIDSet[r.AssetID] = true
				}
			}
		}

		// 取交集（AND逻辑）
		finalSet := make(map[uint]bool)
		if hasSearch && hasSoftwareIDs {
			// 两个条件都指定时取交集
			for id := range searchIDSet {
				if swIDSet[id] {
					finalSet[id] = true
				}
			}
		} else if hasSearch {
			finalSet = searchIDSet
		} else {
			finalSet = swIDSet
		}

		// 将ID集合转为有序切片（降序，与默认顺序一致）
		var allIDs []uint
		for id := range finalSet {
			allIDs = append(allIDs, id)
		}
		sort.Slice(allIDs, func(i, j int) bool { return allIDs[i] > allIDs[j] })

		total = int64(len(allIDs))

		// 分页截取ID
		var pageIDs []uint
		if offset < len(allIDs) {
			end := offset + pageSize
			if end > len(allIDs) {
				end = len(allIDs)
			}
			pageIDs = allIDs[offset:end]
		}

		if len(pageIDs) > 0 {
			db.Where("id IN ?", pageIDs).Find(&assets)
		}
	} else {
		// 无搜索条件，保持原有逻辑
		db.Model(&models.Asset{}).Count(&total)
		if err := db.Order("id desc").Offset(offset).Limit(pageSize).Find(&assets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
			return
		}
	}

	// 为每个资产查询关联的核准软件
	type AssetRow struct {
		ID           uint                      `json:"id"`
		ComputerName string                    `json:"computer_name"`
		IPAddress    string                    `json:"ip_address"`
		SoftwareList []models.ApprovedSoftware `json:"software_list"`
	}

	var result []AssetRow
	for _, a := range assets {
		row := AssetRow{
			ID:           a.ID,
			ComputerName: a.ComputerName,
			IPAddress:    a.IPAddress,
		}
		var links []models.AssetSoftware
		database.GetDB().Where("asset_id = ?", a.ID).Find(&links)
		var swList []models.ApprovedSoftware
		for _, l := range links {
			var sw models.ApprovedSoftware
			if err := database.GetDB().First(&sw, l.ApprovedSoftwareID).Error; err == nil {
				swList = append(swList, sw)
			}
		}
		if swList == nil {
			swList = []models.ApprovedSoftware{}
		}
		row.SoftwareList = swList
		result = append(result, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"data":  result,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetAssetSoftwareLinks 获取某资产的已关联软件ID列表
func GetAssetSoftwareLinks(c *gin.Context) {
	assetID, _ := strconv.Atoi(c.Param("id"))
	var links []models.AssetSoftware
	database.GetDB().Where("asset_id = ?", assetID).Find(&links)
	var ids []uint
	for _, l := range links {
		ids = append(ids, l.ApprovedSoftwareID)
	}
	if ids == nil {
		ids = []uint{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ids})
}

// ExportPatchUpdateRecord 导出第三方应用补丁更新记录表
func ExportPatchUpdateRecord(c *gin.Context) {
	// 查询所有需要更新的核准软件
	var needUpdateSoftware []models.ApprovedSoftware
	if err := database.GetDB().Where("need_update = ?", true).Find(&needUpdateSoftware).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	if len(needUpdateSoftware) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "当前没有需要更新的软件"})
		return
	}

	// 构建软件ID集合
	var swIDs []uint
	for _, sw := range needUpdateSoftware {
		swIDs = append(swIDs, sw.ID)
	}

	// 查询关联的资产软件记录
	var assetSoftwareLinks []models.AssetSoftware
	database.GetDB().Preload("Asset").Preload("ApprovedSoftware").
		Where("approved_software_id IN ?", swIDs).
		Find(&assetSoftwareLinks)

	// 构建导出数据行
	type PatchRow struct {
		ComputerName   string
		SoftwareName   string
		Version        string
		LatestVersion  string
	}
	var rows []PatchRow
	for _, link := range assetSoftwareLinks {
		if link.Asset.ID == 0 || link.ApprovedSoftware.ID == 0 {
			continue
		}
		rows = append(rows, PatchRow{
			ComputerName:  link.Asset.ComputerName,
			SoftwareName:  link.ApprovedSoftware.Name,
			Version:       link.ApprovedSoftware.Version,
			LatestVersion: link.ApprovedSoftware.LatestVersion,
		})
	}

	// 创建Excel
	f := excelize.NewFile()
	sheetName := "补丁更新记录"
	f.SetSheetName("Sheet1", sheetName)

	now := time.Now()
	yearMonth := fmt.Sprintf("%d年%d月", now.Year(), now.Month())

	// ---- 样式定义 ----
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	headerInfoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E1F2"}},
	})

	dataCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true, Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	signCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	// ---- Row 1: 标题 ----
	f.SetCellValue(sheetName, "A1", "第三方应用补丁更新记录表")
	f.MergeCell(sheetName, "A1", "I1")
	f.SetCellStyle(sheetName, "A1", "I1", titleStyle)
	f.SetRowHeight(sheetName, 1, 36)

	// ---- Row 2: 信息行 ----
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("导出日期：%s", yearMonth))
	f.MergeCell(sheetName, "A2", "C2")
	f.SetCellStyle(sheetName, "A2", "C2", headerInfoStyle)
	f.SetRowHeight(sheetName, 2, 28)

	// ---- Row 3: 表头 ----
	headers := []string{"序号", "计算机名", "软件名", "软件版本", "更新后版本", "实施人", "实施日期", "确认人", "确认日期"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"3", h)
		f.SetCellStyle(sheetName, col+"3", col+"3", tableHeaderStyle)
	}
	f.SetRowHeight(sheetName, 3, 28)

	// ---- 数据行 ----
	rowNum := 4
	for i, row := range rows {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row.ComputerName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row.SoftwareName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row.Version)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row.LatestVersion)
		// 实施人、实施日期、确认人、确认日期 留空供签名
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), "")

		// 设置样式：前5列居中，后4列左对齐（签名区）
		for j := 0; j < 5; j++ {
			col := string(rune('A' + j))
			f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), dataCellStyle)
		}
		for j := 5; j < 9; j++ {
			col := string(rune('A' + j))
			f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), signCellStyle)
		}
		// 签名行高度加大，留出足够签字空间
		f.SetRowHeight(sheetName, rowNum, 36)
		rowNum++
	}

	// 如果没有数据，显示空提示
	if len(rows) == 0 {
		f.SetCellValue(sheetName, "A4", "（暂无需要更新的软件记录）")
		f.MergeCell(sheetName, "A4", "I4")
		f.SetCellStyle(sheetName, "A4", "I4", dataCellStyle)
		rowNum = 5
	}

	// ---- 设置列宽 ----
	f.SetColWidth(sheetName, "A", "A", 8)   // 序号
	f.SetColWidth(sheetName, "B", "B", 22)  // 计算机名
	f.SetColWidth(sheetName, "C", "C", 22)  // 软件名
	f.SetColWidth(sheetName, "D", "D", 16)  // 软件版本
	f.SetColWidth(sheetName, "E", "E", 16)  // 更新后版本
	f.SetColWidth(sheetName, "F", "F", 18)  // 实施人
	f.SetColWidth(sheetName, "G", "G", 16)  // 实施日期
	f.SetColWidth(sheetName, "H", "H", 18)  // 确认人
	f.SetColWidth(sheetName, "I", "I", 16)  // 确认日期

	// 设置右侧页脚：版本号 + 信息等级（淡灰色）
	footerText := fmt.Sprintf("&R&K999999%s\n信息等级：内部公开 Info Class: Internal Disclosure", config.Cfg.Document.AssetDocumentVersion)
	f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
		OddFooter: footerText,
	})

	// 设置左侧页头：插入Logo图片到左上角（缩小尺寸，避免遮挡文字）
	logoPath := config.Cfg.Document.LogoPath
	if logoPath != "" {
		logoData, err := os.ReadFile(logoPath)
		if err != nil {
			fmt.Printf("读取Logo文件失败: %v\n", err)
		} else {
			if err := f.AddPictureFromBytes(sheetName, "A1", &excelize.Picture{
				Extension: ".png",
				File:      logoData,
				Format: &excelize.GraphicOptions{
					ScaleX:       0.08,
					ScaleY:       0.08,
				},
			}); err != nil {
				fmt.Printf("插入Logo失败: %v\n", err)
			}
		}
	}

	// 输出Excel文件
	fileName := fmt.Sprintf("第三方应用补丁更新记录表(%s).xlsx", yearMonth)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")

	if err := f.Write(c.Writer); err != nil {
		fmt.Printf("导出Excel失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出失败"})
		return
	}
}

// UpdateAssetSoftwareLinks 更新资产的软件关联（全量替换）
func UpdateAssetSoftwareLinks(c *gin.Context) {
	assetID, _ := strconv.Atoi(c.Param("id"))

	// 验证资产存在
	var asset models.Asset
	if err := database.GetDB().First(&asset, assetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}

	var input struct {
		SoftwareIDs []uint `json:"software_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 删除旧关联
	database.GetDB().Where("asset_id = ?", assetID).Delete(&models.AssetSoftware{})

	// 创建新关联
	for _, swID := range input.SoftwareIDs {
		link := models.AssetSoftware{
			AssetID:            uint(assetID),
			ApprovedSoftwareID: swID,
		}
		database.GetDB().Create(&link)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "AssetID", FieldLabel: "资产ID", NewValue: fmt.Sprintf("%d", assetID)},
		{FieldName: "SoftwareIDs", FieldLabel: "软件ID列表", NewValue: fmt.Sprintf("%v", input.SoftwareIDs)},
	}
	services.LogOperation(username, displayName, "更新资产软件关联", "asset_software", uint(assetID), asset.ComputerName, approver, c.ClientIP(), details)
}

// parseUintList 将逗号分隔的字符串解析为 uint 切片
func parseUintList(s string) []uint {
	var result []uint
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.ParseUint(part, 10, 64); err == nil {
			result = append(result, uint(n))
		}
	}
	return result
}
