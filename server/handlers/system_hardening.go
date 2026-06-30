package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ExportSystemHardeningChecklist 导出系统加固检查表Excel
func ExportSystemHardeningChecklist(c *gin.Context) {
	// 查询所有资产（预加载操作系统类型）
	var assets []models.Asset
	if err := database.GetDB().Preload("OSType").Order("computer_name asc").Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询资产列表失败"})
		return
	}

	// 创建Excel
	f := excelize.NewFile()
	sheetName := "系统加固检查表"
	f.SetSheetName("Sheet1", sheetName)

	// ---- 样式定义 ----
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	subTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#D9E1F2"},
		},
	})

	dataCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			WrapText:   true,
			Horizontal: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	nameCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			WrapText:   true,
			Horizontal: "left",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	// ---- Row 1: 中文标题 ----
	f.SetCellValue(sheetName, "A1", "系统加固检查表")
	f.MergeCell(sheetName, "A1", "G1")
	f.SetCellStyle(sheetName, "A1", "G1", titleStyle)
	f.SetRowHeight(sheetName, 1, 36)

	// ---- Row 2: 英文副标题 ----
	f.SetCellValue(sheetName, "A2", "System Hardening Checklist")
	f.MergeCell(sheetName, "A2", "G2")
	f.SetCellStyle(sheetName, "A2", "G2", subTitleStyle)
	f.SetRowHeight(sheetName, 2, 28)

	// ---- Row 3: 表头（中英双语）----
	headers := []string{
		"序号\nNO.",
		"主机名\nHost Name",
		"操作系统版本\nSystem Version",
		"加固检查\nHardening Check",
		"BIOS检查\nBIOS Check",
		"检查人\nExecutor",
		"检查日期\nDate",
	}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"3", h)
		f.SetCellStyle(sheetName, col+"3", col+"3", tableHeaderStyle)
	}
	f.SetRowHeight(sheetName, 3, 40)

	// ---- 数据行 ----
	rowNum := 4
	for i, asset := range assets {
		osVersion := ""
		if asset.OSType.Name != "" {
			osVersion = asset.OSType.Name
		}
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), asset.ComputerName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), osVersion)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), "")
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), "")

		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("A%d", rowNum), dataCellStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", rowNum), fmt.Sprintf("B%d", rowNum), nameCellStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("G%d", rowNum), dataCellStyle)
		f.SetRowHeight(sheetName, rowNum, 24)
		rowNum++
	}

	// ---- 设置列宽 ----
	f.SetColWidth(sheetName, "A", "A", 8)  // 序号
	f.SetColWidth(sheetName, "B", "B", 22) // 主机名
	f.SetColWidth(sheetName, "C", "C", 24) // 操作系统版本
	f.SetColWidth(sheetName, "D", "D", 16) // 加固检查
	f.SetColWidth(sheetName, "E", "E", 16) // BIOS检查
	f.SetColWidth(sheetName, "F", "F", 14) // 检查人
	f.SetColWidth(sheetName, "G", "G", 14) // 检查日期

	// 设置页面布局：A4横向
	f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
		Size:        intPtr(9),
		Orientation: stringPtr("landscape"),
	})
	f.SetPageMargins(sheetName, &excelize.PageLayoutMarginsOptions{
		Bottom: float64Ptr(0.5),
		Footer: float64Ptr(0.3),
		Header: float64Ptr(0.3),
		Left:   float64Ptr(0.5),
		Right:  float64Ptr(0.5),
		Top:    float64Ptr(0.5),
	})

	// 设置右侧页脚：版本号 + 信息等级（淡灰色）
	footerText := fmt.Sprintf("&R&K999999%s\n信息等级：内部公开 Info Class: Internal Disclosure", config.Cfg.Document.SystemHardeningDocumentVersion)
	f.SetHeaderFooter(sheetName, &excelize.HeaderFooterOptions{
		OddFooter: footerText,
	})

	// 设置左侧页头：插入Logo图片
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
					ScaleX: 0.08,
					ScaleY: 0.08,
				},
			}); err != nil {
				fmt.Printf("插入Logo失败: %v\n", err)
			}
		}
	}

	// 输出Excel文件
	fileName := "系统加固检查表.xlsx"
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")

	if err := f.Write(c.Writer); err != nil {
		fmt.Printf("导出系统加固检查表失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出失败"})
		return
	}
}

// ListSystemHardeningHistories 获取系统加固检查记录列表
func ListSystemHardeningHistories(c *gin.Context) {
	var records []models.SystemHardeningHistory

	query := database.GetDB().Model(&models.SystemHardeningHistory{})

	// 按年份筛选
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", year)
		}
	}

	// 按关键词筛选描述
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("description LIKE ?", "%"+keyword+"%")
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	query.Count(&total)

	if err := query.Order("year DESC, quarter DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreateSystemHardeningHistory 上传系统加固检查记录
func CreateSystemHardeningHistory(c *gin.Context) {
	yearStr := c.PostForm("year")
	quarterStr := c.PostForm("quarter")
	description := c.PostForm("description")

	if yearStr == "" || quarterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和季度不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "季度格式不正确，应为1-4"})
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 检查文件类型，仅允许PDF
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持PDF格式文件"})
		return
	}

	// 构建按年份的上传路径
	yearDir := filepath.Join(config.Cfg.Upload.SystemHardeningCheckPath, strconv.Itoa(year))
	os.MkdirAll(yearDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.SystemHardeningHistory{
		Year:        year,
		Quarter:     quarter,
		Description: description,
		FileName:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
		FileType:    file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&record).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", NewValue: strconv.Itoa(year)},
		{FieldName: "Quarter", FieldLabel: "季度", NewValue: fmt.Sprintf("Q%d", quarter)},
		{FieldName: "Description", FieldLabel: "描述", NewValue: description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: record.FileName},
		{FieldName: "FileSize", FieldLabel: "文件大小", NewValue: fmt.Sprintf("%d", record.FileSize)},
	}
	services.LogOperation(username, displayName, "上传系统加固检查记录", "system_hardening_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// UpdateSystemHardeningHistory 更新系统加固检查记录
func UpdateSystemHardeningHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.SystemHardeningHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	yearStr := c.PostForm("year")
	quarterStr := c.PostForm("quarter")
	description := c.PostForm("description")

	if yearStr == "" || quarterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和季度不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "季度格式不正确，应为1-4"})
		return
	}

	oldRecord := record
	oldYear := record.Year
	oldFilePath := record.FilePath

	// 如果年份变化，需要移动文件到新目录
	if year != oldYear {
		oldYearDir := filepath.Join(config.Cfg.Upload.SystemHardeningCheckPath, strconv.Itoa(oldYear))
		newYearDir := filepath.Join(config.Cfg.Upload.SystemHardeningCheckPath, strconv.Itoa(year))
		os.MkdirAll(newYearDir, 0755)

		fileName := filepath.Base(record.FilePath)
		newFilePath := filepath.Join(newYearDir, fileName)
		if err := os.Rename(oldFilePath, newFilePath); err != nil {
			if copyErr := services.CopyFile(oldFilePath, newFilePath); copyErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件移动失败"})
				return
			}
			os.Remove(oldFilePath)
			record.FilePath = newFilePath
		} else {
			record.FilePath = newFilePath
		}

		os.Remove(oldYearDir)
	}

	record.Year = year
	record.Quarter = quarter
	record.Description = description

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("system_hardening_history")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	services.LogOperation(username, displayName, "更新系统加固检查记录", "system_hardening_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeleteSystemHardeningHistory 删除系统加固检查记录
func DeleteSystemHardeningHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.SystemHardeningHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	os.Remove(record.FilePath)

	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("system_hardening_history")
	details := services.DiffStructs(record, models.SystemHardeningHistory{}, fieldLabels)
	services.LogOperation(username, displayName, "删除系统加固检查记录", "system_hardening_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// DownloadSystemHardeningHistory 下载系统加固检查文件
func DownloadSystemHardeningHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.SystemHardeningHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+record.FileName+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", record.FileSize))
	c.File(record.FilePath)
}

// PreviewSystemHardeningHistory 预览系统加固检查文件
func PreviewSystemHardeningHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.SystemHardeningHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\""+record.FileName+"\"")
	c.File(record.FilePath)
}
