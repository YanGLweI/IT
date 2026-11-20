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

// wirelessScanContentType 按原始文件扩展名返回 Content-Type
func wirelessScanContentType(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

// ListWirelessScans 获取无线扫描报告列表
func ListWirelessScans(c *gin.Context) {
	var records []models.WirelessScan

	query := database.GetDB().Model(&models.WirelessScan{})

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

	// 排序：先按年份降序，再按月份降序（最新的在前面）
	if err := query.Order("year DESC, month DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreateWirelessScan 上传无线扫描报告
func CreateWirelessScan(c *gin.Context) {
	yearStr := c.PostForm("year")
	monthStr := c.PostForm("month")
	description := c.PostForm("description")

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

	// 同一年月只允许存在一条记录
	var dup int64
	if err := database.GetDB().Model(&models.WirelessScan{}).
		Where("year = ? AND month = ?", year, month).
		Count(&dup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	if dup > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("%d年%d月的无线扫描报告已存在", year, month)})
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 检查文件类型，仅允许 PDF 与 DOCX
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 或 DOCX 格式文件"})
		return
	}

	// 构建按年份的上传路径: wireless_scans/{year}/
	yearDir := filepath.Join(config.Cfg.Upload.WirelessScanPath, strconv.Itoa(year))
	if err := os.MkdirAll(yearDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.WirelessScan{
		Year:        year,
		Month:       month,
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
		{FieldName: "Month", FieldLabel: "月份", NewValue: strconv.Itoa(month)},
		{FieldName: "Description", FieldLabel: "描述", NewValue: description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: record.FileName},
		{FieldName: "FileSize", FieldLabel: "文件大小", NewValue: fmt.Sprintf("%d", record.FileSize)},
	}
	services.LogOperation(username, displayName, "上传无线扫描报告", "wireless_scan", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// UpdateWirelessScan 更新无线扫描报告
func UpdateWirelessScan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.WirelessScan
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	yearStr := c.PostForm("year")
	monthStr := c.PostForm("month")
	description := c.PostForm("description")

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

	// 同一年月只允许存在一条记录（排除自身）
	var dup int64
	if err := database.GetDB().Model(&models.WirelessScan{}).
		Where("year = ? AND month = ? AND id <> ?", year, month, record.ID).
		Count(&dup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	if dup > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("%d年%d月的无线扫描报告已存在", year, month)})
		return
	}

	oldRecord := record
	oldYear := record.Year
	oldFilePath := record.FilePath

	// 如果年份变化，需要移动文件到新目录
	if year != oldYear {
		oldYearDir := filepath.Join(config.Cfg.Upload.WirelessScanPath, strconv.Itoa(oldYear))
		newYearDir := filepath.Join(config.Cfg.Upload.WirelessScanPath, strconv.Itoa(year))
		if err := os.MkdirAll(newYearDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
			return
		}

		// 移动文件
		fileName := filepath.Base(record.FilePath)
		newFilePath := filepath.Join(newYearDir, fileName)
		if err := os.Rename(oldFilePath, newFilePath); err != nil {
			// 如果rename失败（跨分区），尝试复制后删除
			if copyErr := services.CopyFile(oldFilePath, newFilePath); copyErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件移动失败"})
				return
			}
			os.Remove(oldFilePath)
			record.FilePath = newFilePath
		} else {
			record.FilePath = newFilePath
		}

		// 清理旧的空目录
		os.Remove(oldYearDir)
	}

	record.Year = year
	record.Month = month
	record.Description = description

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("wireless_scan")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	services.LogOperation(username, displayName, "更新无线扫描报告", "wireless_scan", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeleteWirelessScan 删除无线扫描报告
func DeleteWirelessScan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.WirelessScan
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录删除成功后再清理磁盘文件，避免 DB 失败导致文件已丢失
	os.Remove(record.FilePath)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("wireless_scan")
	details := services.DiffStructs(record, models.WirelessScan{}, fieldLabels)
	services.LogOperation(username, displayName, "删除无线扫描报告", "wireless_scan", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// DownloadWirelessScan 下载无线扫描报告
func DownloadWirelessScan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.WirelessScan
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", wirelessScanContentType(record.FileName))
	c.Header("Content-Disposition", "attachment; filename=\""+record.FileName+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", record.FileSize))
	c.File(record.FilePath)
}

// PreviewWirelessScan 预览无线扫描报告
func PreviewWirelessScan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.WirelessScan
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", wirelessScanContentType(record.FileName))
	c.Header("Content-Disposition", "inline; filename=\""+record.FileName+"\"")
	c.File(record.FilePath)
}
