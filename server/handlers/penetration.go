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

// ListPenetrationTests 获取渗透测试报告列表
func ListPenetrationTests(c *gin.Context) {
	var records []models.PenetrationTest

	query := database.GetDB().Model(&models.PenetrationTest{})

	// 按类型筛选（internal / external）
	if testType := c.Query("test_type"); testType != "" {
		query = query.Where("test_type = ?", testType)
	}

	// 按年份筛选
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", year)
		}
	}

	// 按关键词筛选
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("file_name LIKE ? OR report_date LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
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

	if err := query.Preload("VulnerabilityScans").Order("report_date DESC, year DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreatePenetrationTest 创建渗透测试报告
func CreatePenetrationTest(c *gin.Context) {
	testType := c.PostForm("test_type")
	if testType == "" {
		testType = "internal"
	}
	yearStr := c.PostForm("year")
	reportDate := c.PostForm("report_date")
	vulnCountStr := c.PostForm("vuln_count")
	description := c.PostForm("description")
	vulnScanIDsStr := c.PostForm("vulnerability_scan_ids")

	if yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	vulnCount, _ := strconv.Atoi(vulnCountStr)

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 检查文件类型，允许 PDF 和 DOCX
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" && ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 或 DOCX 格式文件"})
		return
	}

	// 构建按年份的上传路径
	yearDir := filepath.Join(config.Cfg.Upload.PenetrationTestPath, strconv.Itoa(year))
	os.MkdirAll(yearDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.PenetrationTest{
		TestType:    testType,
		Year:        year,
		ReportDate:  reportDate,
		VulnCount:   vulnCount,
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

	// 处理关联漏洞扫描报告
	if vulnScanIDsStr != "" {
		var vulnScans []models.VulnerabilityScan
		for _, idStr := range strings.Split(vulnScanIDsStr, ",") {
			if vid, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				vulnScans = append(vulnScans, models.VulnerabilityScan{ID: uint(vid)})
			}
		}
		if len(vulnScans) > 0 {
			database.GetDB().Model(&record).Association("VulnerabilityScans").Replace(vulnScans)
		}
	}

	database.GetDB().Preload("VulnerabilityScans").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "TestType", FieldLabel: "测试类型", NewValue: testType},
		{FieldName: "Year", FieldLabel: "年份", NewValue: strconv.Itoa(year)},
		{FieldName: "ReportDate", FieldLabel: "报告日期", NewValue: reportDate},
		{FieldName: "VulnCount", FieldLabel: "可渗透漏洞数", NewValue: strconv.Itoa(vulnCount)},
		{FieldName: "Description", FieldLabel: "结果描述", NewValue: description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: record.FileName},
	}
	services.LogOperation(username, displayName, "创建渗透测试报告", "penetration_test", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// UpdatePenetrationTest 更新渗透测试报告
func UpdatePenetrationTest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PenetrationTest
	if err := database.GetDB().Preload("VulnerabilityScans").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	testType := c.PostForm("test_type")
	if testType == "" {
		testType = record.TestType
	}
	yearStr := c.PostForm("year")
	reportDate := c.PostForm("report_date")
	vulnCountStr := c.PostForm("vuln_count")
	description := c.PostForm("description")
	vulnScanIDsStr := c.PostForm("vulnerability_scan_ids")

	year := record.Year
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 2000 && y <= 2100 {
			year = y
		}
	}

	vulnCount := record.VulnCount
	if vulnCountStr != "" {
		if v, err := strconv.Atoi(vulnCountStr); err == nil {
			vulnCount = v
		}
	}

	// 记录旧值用于日志对比
	oldTestType := record.TestType
	oldYear := record.Year
	oldReportDate := record.ReportDate
	oldVulnCount := record.VulnCount
	oldDescription := record.Description

	// 更新字段
	record.TestType = testType
	record.Year = year
	record.ReportDate = reportDate
	record.VulnCount = vulnCount
	record.Description = description

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 更新关联漏洞扫描报告
	if vulnScanIDsStr != "" {
		var vulnScans []models.VulnerabilityScan
		for _, idStr := range strings.Split(vulnScanIDsStr, ",") {
			if vid, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				vulnScans = append(vulnScans, models.VulnerabilityScan{ID: uint(vid)})
			}
		}
		database.GetDB().Model(&record).Association("VulnerabilityScans").Replace(vulnScans)
	}

	database.GetDB().Preload("VulnerabilityScans").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	var details []services.LogDetail
	if oldTestType != record.TestType {
		details = append(details, services.LogDetail{FieldName: "TestType", FieldLabel: "测试类型", OldValue: oldTestType, NewValue: record.TestType})
	}
	if oldYear != record.Year {
		details = append(details, services.LogDetail{FieldName: "Year", FieldLabel: "年份", OldValue: strconv.Itoa(oldYear), NewValue: strconv.Itoa(record.Year)})
	}
	if oldReportDate != record.ReportDate {
		details = append(details, services.LogDetail{FieldName: "ReportDate", FieldLabel: "报告日期", OldValue: oldReportDate, NewValue: record.ReportDate})
	}
	if oldVulnCount != record.VulnCount {
		details = append(details, services.LogDetail{FieldName: "VulnCount", FieldLabel: "可渗透漏洞数", OldValue: strconv.Itoa(oldVulnCount), NewValue: strconv.Itoa(record.VulnCount)})
	}
	if oldDescription != record.Description {
		details = append(details, services.LogDetail{FieldName: "Description", FieldLabel: "结果描述", OldValue: oldDescription, NewValue: record.Description})
	}
	services.LogOperation(username, displayName, "更新渗透测试报告", "penetration_test", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeletePenetrationTest 删除渗透测试报告
func DeletePenetrationTest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PenetrationTest
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 清除多对多关联
	database.GetDB().Model(&record).Association("VulnerabilityScans").Clear()

	// 删除文件
	if record.FilePath != "" {
		os.Remove(record.FilePath)
	}

	// 删除记录
	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "FileName", FieldLabel: "文件名", OldValue: record.FileName, NewValue: ""},
	}
	services.LogOperation(username, displayName, "删除渗透测试报告", "penetration_test", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewPenetrationTest 预览渗透测试报告
func PreviewPenetrationTest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PenetrationTest
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	ext := strings.ToLower(filepath.Ext(record.FilePath))
	if ext == ".pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "inline; filename=\""+record.FileName+"\"")
	} else {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		c.Header("Content-Disposition", "inline; filename=\""+record.FileName+"\"")
	}
	c.File(record.FilePath)
}

// DownloadPenetrationTest 下载渗透测试报告
func DownloadPenetrationTest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.PenetrationTest
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+record.FileName+"\"")
	c.File(record.FilePath)
}
