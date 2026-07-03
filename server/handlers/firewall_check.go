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

// ListFirewallChecks 获取防火墙检查记录列表
func ListFirewallChecks(c *gin.Context) {
	var records []models.FirewallCheck

	query := database.GetDB().Model(&models.FirewallCheck{})

	// 按年份筛选
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", year)
		}
	}

	// 按季度筛选
	if quarterStr := c.Query("quarter"); quarterStr != "" {
		if quarter, err := strconv.Atoi(quarterStr); err == nil {
			query = query.Where("quarter = ?", quarter)
		}
	}

	// 按检查结果筛选
	if result := c.Query("check_result"); result != "" {
		query = query.Where("check_result = ?", result)
	}

	// 按关键词筛选
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("report_date LIKE ? OR check_result LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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

	if err := query.Preload("Asset").Order("year DESC, quarter DESC, report_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreateFirewallCheck 创建防火墙检查记录
func CreateFirewallCheck(c *gin.Context) {
	yearStr := c.PostForm("year")
	quarterStr := c.PostForm("quarter")
	reportDate := c.PostForm("report_date")
	assetIDStr := c.PostForm("asset_id")
	checkResult := c.PostForm("check_result")

	if yearStr == "" || quarterStr == "" || assetIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份、季度和防火墙不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "季度格式不正确"})
		return
	}

	assetID, err := strconv.Atoi(assetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "资产ID格式不正确"})
		return
	}

	if checkResult == "" {
		checkResult = "compliant"
	}

	// 验证资产是否存在
	var asset models.Asset
	if err := database.GetDB().First(&asset, assetID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的资产不存在"})
		return
	}

	record := models.FirewallCheck{
		Year:        year,
		Quarter:     quarter,
		ReportDate:  reportDate,
		AssetID:     uint(assetID),
		CheckResult: checkResult,
	}

	if err := database.GetDB().Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	database.GetDB().Preload("Asset").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", NewValue: strconv.Itoa(year)},
		{FieldName: "Quarter", FieldLabel: "季度", NewValue: fmt.Sprintf("Q%d", quarter)},
		{FieldName: "ReportDate", FieldLabel: "报告日期", NewValue: reportDate},
		{FieldName: "AssetID", FieldLabel: "防火墙", NewValue: asset.ComputerName},
		{FieldName: "CheckResult", FieldLabel: "检查结果", NewValue: checkResult},
	}

	services.LogOperation(username, displayName, "创建防火墙检查记录", "firewall_check", record.ID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": record})
}

// UpdateFirewallCheck 更新防火墙检查记录
func UpdateFirewallCheck(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.FirewallCheck
	if err := database.GetDB().Preload("Asset").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	yearStr := c.PostForm("year")
	quarterStr := c.PostForm("quarter")
	reportDate := c.PostForm("report_date")
	assetIDStr := c.PostForm("asset_id")
	checkResult := c.PostForm("check_result")

	// 记录旧值用于日志
	oldYear := record.Year
	oldQuarter := record.Quarter
	oldReportDate := record.ReportDate
	oldAssetID := record.AssetID
	oldCheckResult := record.CheckResult

	// 更新字段
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 2000 && y <= 2100 {
			record.Year = y
		}
	}
	if quarterStr != "" {
		if q, err := strconv.Atoi(quarterStr); err == nil && q >= 1 && q <= 4 {
			record.Quarter = q
		}
	}
	if reportDate != "" {
		record.ReportDate = reportDate
	}
	if assetIDStr != "" {
		if aid, err := strconv.Atoi(assetIDStr); err == nil {
			var asset models.Asset
			if err := database.GetDB().First(&asset, aid).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的资产不存在"})
				return
			}
			record.AssetID = uint(aid)
		}
	}
	if checkResult != "" {
		record.CheckResult = checkResult
	}

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	database.GetDB().Preload("Asset").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	var details []services.LogDetail
	if oldYear != record.Year {
		details = append(details, services.LogDetail{FieldName: "Year", FieldLabel: "年份", OldValue: strconv.Itoa(oldYear), NewValue: strconv.Itoa(record.Year)})
	}
	if oldQuarter != record.Quarter {
		details = append(details, services.LogDetail{FieldName: "Quarter", FieldLabel: "季度", OldValue: fmt.Sprintf("Q%d", oldQuarter), NewValue: fmt.Sprintf("Q%d", record.Quarter)})
	}
	if oldReportDate != record.ReportDate {
		details = append(details, services.LogDetail{FieldName: "ReportDate", FieldLabel: "报告日期", OldValue: oldReportDate, NewValue: record.ReportDate})
	}
	if oldAssetID != record.AssetID {
		details = append(details, services.LogDetail{FieldName: "AssetID", FieldLabel: "防火墙", OldValue: strconv.Itoa(int(oldAssetID)), NewValue: strconv.Itoa(int(record.AssetID))})
	}
	if oldCheckResult != record.CheckResult {
		details = append(details, services.LogDetail{FieldName: "CheckResult", FieldLabel: "检查结果", OldValue: oldCheckResult, NewValue: record.CheckResult})
	}

	services.LogOperation(username, displayName, "更新防火墙检查记录", "firewall_check", record.ID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeleteFirewallCheck 删除防火墙检查记录
func DeleteFirewallCheck(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.FirewallCheck
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 删除整改报告文件
	if record.RectFilePath != "" {
		os.Remove(record.RectFilePath)
	}

	// 删除记录
	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", OldValue: strconv.Itoa(record.Year), NewValue: ""},
		{FieldName: "Quarter", FieldLabel: "季度", OldValue: fmt.Sprintf("Q%d", record.Quarter), NewValue: ""},
	}

	services.LogOperation(username, displayName, "删除防火墙检查记录", "firewall_check", record.ID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// UploadFirewallRectReport 上传防火墙整改报告
func UploadFirewallRectReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.FirewallCheck
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

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 或 DOCX 格式文件"})
		return
	}

	// 构建上传路径
	yearDir := filepath.Join(config.Cfg.Upload.FirewallCheckPath, strconv.Itoa(record.Year), "rect")
	os.MkdirAll(yearDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	// 删除旧的整改报告文件
	if record.RectFilePath != "" {
		os.Remove(record.RectFilePath)
	}

	// 更新记录
	record.RectFileName = file.Filename
	record.RectFilePath = filePath
	record.RectFileSize = file.Size
	record.RectFileType = file.Header.Get("Content-Type")

	if err := database.GetDB().Save(&record).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "RectFileName", FieldLabel: "整改报告", OldValue: "", NewValue: file.Filename},
	}

	services.LogOperation(username, displayName, "上传防火墙整改报告", "firewall_check", record.ID, file.Filename, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// PreviewFirewallRectReport 预览防火墙整改报告
func PreviewFirewallRectReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.FirewallCheck
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.RectFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "整改报告不存在"})
		return
	}

	if _, err := os.Stat(record.RectFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	ext := strings.ToLower(filepath.Ext(record.RectFilePath))
	if ext == ".pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "inline; filename=\""+record.RectFileName+"\"")
	} else {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		c.Header("Content-Disposition", "inline; filename=\""+record.RectFileName+"\"")
	}
	c.File(record.RectFilePath)
}

// DownloadFirewallRectReport 下载防火墙整改报告
func DownloadFirewallRectReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.FirewallCheck
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.RectFilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "整改报告不存在"})
		return
	}

	if _, err := os.Stat(record.RectFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+record.RectFileName+"\"")
	c.File(record.RectFilePath)
}

// DeleteFirewallRectReport 删除防火墙整改报告
func DeleteFirewallRectReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.FirewallCheck
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.RectFilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "整改报告不存在"})
		return
	}

	oldFileName := record.RectFileName

	// 删除文件
	os.Remove(record.RectFilePath)

	// 更新记录
	record.RectFileName = ""
	record.RectFilePath = ""
	record.RectFileSize = 0
	record.RectFileType = ""

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "RectFileName", FieldLabel: "整改报告", OldValue: oldFileName, NewValue: ""},
	}

	services.LogOperation(username, displayName, "删除防火墙整改报告", "firewall_check", record.ID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
