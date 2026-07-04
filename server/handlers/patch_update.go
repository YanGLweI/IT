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
)

// ListPatchUpdates 获取补丁更新记录列表
func ListPatchUpdates(c *gin.Context) {
	var records []models.PatchUpdate

	query := database.GetDB().Model(&models.PatchUpdate{})

	// 按年份筛选
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", year)
		}
	}

	// 按月份筛选
	if monthStr := c.Query("month"); monthStr != "" {
		if month, err := strconv.Atoi(monthStr); err == nil {
			query = query.Where("month = ?", month)
		}
	}

	// 按合规性筛选
	if compliance := c.Query("compliance"); compliance != "" {
		query = query.Where("compliance = ?", compliance)
	}

	// 按关键词筛选文件名
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("file_name LIKE ?", "%"+keyword+"%")
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

	if err := query.Order("year DESC, month DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreatePatchUpdate 上传合规性报表
func CreatePatchUpdate(c *gin.Context) {
	yearStr := c.PostForm("year")
	monthStr := c.PostForm("month")
	totalAssetsStr := c.PostForm("total_assets")
	compliance := c.PostForm("compliance")
	nonCompliantAssetsStr := c.PostForm("non_compliant_assets")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和月份不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "月份格式不正确，应为1-12"})
		return
	}

	totalAssets := 0
	if totalAssetsStr != "" {
		totalAssets, err = strconv.Atoi(totalAssetsStr)
		if err != nil || totalAssets < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "资产总数格式不正确"})
			return
		}
	}

	if compliance == "" {
		compliance = "compliant"
	}

	nonCompliantAssets := 0
	if nonCompliantAssetsStr != "" {
		nonCompliantAssets, err = strconv.Atoi(nonCompliantAssetsStr)
		if err != nil || nonCompliantAssets < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不合规资产数格式不正确"})
			return
		}
	}

	// 不合规时必须填写不合规资产数
	if compliance == "non_compliant" && nonCompliantAssets <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不合规时必须填写不合规资产数"})
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传合规性报表"})
		return
	}

	// 检查文件类型，仅允许PDF
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持PDF格式文件"})
		return
	}

	// 构建按年份的上传路径
	yearDir := filepath.Join(config.Cfg.Upload.PatchUpdatePath, strconv.Itoa(year))
	os.MkdirAll(yearDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.PatchUpdate{
		Year:               year,
		Month:              month,
		TotalAssets:        totalAssets,
		Compliance:         compliance,
		NonCompliantAssets: nonCompliantAssets,
		FileName:           file.Filename,
		FilePath:           filePath,
		FileSize:           file.Size,
		FileType:           file.Header.Get("Content-Type"),
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
		{FieldName: "Month", FieldLabel: "月份", NewValue: strconv.Itoa(month)},
		{FieldName: "TotalAssets", FieldLabel: "资产总数", NewValue: strconv.Itoa(totalAssets)},
		{FieldName: "Compliance", FieldLabel: "合规性", NewValue: compliance},
		{FieldName: "FileName", FieldLabel: "合规性报表", NewValue: record.FileName},
	}
	if nonCompliantAssets > 0 {
		details = append(details, services.LogDetail{FieldName: "NonCompliantAssets", FieldLabel: "不合规资产数", NewValue: strconv.Itoa(nonCompliantAssets)})
	}
	services.LogOperation(username, displayName, "上传合规性报表", "patch_update", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// UpdatePatchUpdate 更新补丁更新记录
func UpdatePatchUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	yearStr := c.PostForm("year")
	monthStr := c.PostForm("month")
	totalAssetsStr := c.PostForm("total_assets")
	compliance := c.PostForm("compliance")
	nonCompliantAssetsStr := c.PostForm("non_compliant_assets")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和月份不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "月份格式不正确，应为1-12"})
		return
	}

	oldRecord := record
	oldYear := record.Year

	// 如果年份变化，需要移动文件到新目录
	if year != oldYear {
		newYearDir := filepath.Join(config.Cfg.Upload.PatchUpdatePath, strconv.Itoa(year))
		os.MkdirAll(newYearDir, 0755)

		// 迁移主报表
		if record.FilePath != "" {
			fileName := filepath.Base(record.FilePath)
			newFilePath := filepath.Join(newYearDir, fileName)
			if err := os.Rename(record.FilePath, newFilePath); err != nil {
				if copyErr := services.CopyFile(record.FilePath, newFilePath); copyErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件迁移失败"})
					return
				}
				os.Remove(record.FilePath)
			}
			record.FilePath = newFilePath
		}

		// 迁移修复报表
		if record.FixFilePath != "" {
			fixDir := filepath.Join(newYearDir, "fix")
			os.MkdirAll(fixDir, 0755)
			fixFileName := filepath.Base(record.FixFilePath)
			newFixPath := filepath.Join(fixDir, fixFileName)
			if err := os.Rename(record.FixFilePath, newFixPath); err != nil {
				if copyErr := services.CopyFile(record.FixFilePath, newFixPath); copyErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "修复报表文件迁移失败"})
					return
				}
				os.Remove(record.FixFilePath)
			}
			record.FixFilePath = newFixPath
		}

		// 清理旧目录
		oldYearDir := filepath.Join(config.Cfg.Upload.PatchUpdatePath, strconv.Itoa(oldYear))
		os.Remove(oldYearDir)
	}

	record.Year = year
	record.Month = month

	if totalAssetsStr != "" {
		if ta, err := strconv.Atoi(totalAssetsStr); err == nil {
			record.TotalAssets = ta
		}
	}
	if compliance != "" {
		if compliance != "compliant" && compliance != "non_compliant" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "合规性值无效"})
			return
		}
		record.Compliance = compliance
		// 若改为不合规且有修复报表，清空修复报表
		if compliance == "non_compliant" && record.FixFilePath != "" {
			os.Remove(record.FixFilePath)
			record.FixFileName = ""
			record.FixFilePath = ""
			record.FixFileSize = 0
			record.FixFileType = ""
		}
	}
	if nonCompliantAssetsStr != "" {
		if nca, err := strconv.Atoi(nonCompliantAssetsStr); err == nil {
			record.NonCompliantAssets = nca
		}
	}

	// 不合规时必须填写不合规资产数
	if record.Compliance == "non_compliant" && record.NonCompliantAssets <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不合规时必须填写不合规资产数"})
		return
	}

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("patch_update")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	services.LogOperation(username, displayName, "更新合规性报表", "patch_update", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeletePatchUpdate 删除补丁更新记录
func DeletePatchUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 删除主报表文件
	if record.FilePath != "" {
		os.Remove(record.FilePath)
	}
	// 删除修复报表文件
	if record.FixFilePath != "" {
		os.Remove(record.FixFilePath)
	}

	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("patch_update")
	details := services.DiffStructs(record, models.PatchUpdate{}, fieldLabels)
	services.LogOperation(username, displayName, "删除合规性报表", "patch_update", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewPatchUpdate 预览合规性报表
func PreviewPatchUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
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

// DownloadPatchUpdate 下载合规性报表
func DownloadPatchUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
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

// UploadPatchFixReport 上传修复报表
func UploadPatchFixReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
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

	// 构建上传路径
	fixDir := filepath.Join(config.Cfg.Upload.PatchUpdatePath, strconv.Itoa(record.Year), "fix")
	if err := os.MkdirAll(fixDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(fixDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	// 删除旧的修复报表文件
	if record.FixFilePath != "" {
		os.Remove(record.FixFilePath)
	}

	oldFixFileName := record.FixFileName
	oldCompliance := record.Compliance

	// 更新修复报表信息，并自动将合规性改为"合规"
	record.FixFileName = file.Filename
	record.FixFilePath = filePath
	record.FixFileSize = file.Size
	record.FixFileType = file.Header.Get("Content-Type")
	record.Compliance = "compliant"

	if err := database.GetDB().Save(&record).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "FixFileName", FieldLabel: "修复报表", OldValue: oldFixFileName, NewValue: file.Filename},
	}
	if oldCompliance != record.Compliance {
		details = append(details, services.LogDetail{FieldName: "Compliance", FieldLabel: "合规性", OldValue: oldCompliance, NewValue: record.Compliance})
	}
	services.LogOperation(username, displayName, "上传修复报表", "patch_update", record.ID, file.Filename, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// DeletePatchFixReport 删除修复报表
func DeletePatchFixReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.FixFilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "修复报表不存在"})
		return
	}

	oldFileName := record.FixFileName
	oldCompliance := record.Compliance

	// 删除文件
	os.Remove(record.FixFilePath)

	// 清空修复报表字段，并恢复为"不合规"
	record.FixFileName = ""
	record.FixFilePath = ""
	record.FixFileSize = 0
	record.FixFileType = ""
	record.Compliance = "non_compliant"

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "FixFileName", FieldLabel: "修复报表", OldValue: oldFileName, NewValue: ""},
	}
	if oldCompliance != record.Compliance {
		details = append(details, services.LogDetail{FieldName: "Compliance", FieldLabel: "合规性", OldValue: oldCompliance, NewValue: record.Compliance})
	}
	services.LogOperation(username, displayName, "删除修复报表", "patch_update", record.ID, oldFileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewPatchFixReport 预览修复报表
func PreviewPatchFixReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.FixFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "修复报表不存在"})
		return
	}

	if _, err := os.Stat(record.FixFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\""+record.FixFileName+"\"")
	c.File(record.FixFilePath)
}

// DownloadPatchFixReport 下载修复报表
func DownloadPatchFixReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PatchUpdate
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.FixFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "修复报表不存在"})
		return
	}

	if _, err := os.Stat(record.FixFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+record.FixFileName+"\"")
	c.File(record.FixFilePath)
}
