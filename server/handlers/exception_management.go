package handlers

import (
	"fmt"
	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ListExceptionManagement 查询例外管理列表（不需要双控）
func ListExceptionManagement(c *gin.Context) {
	applyDate := c.Query("apply_date")
	endDate := c.Query("end_date")
	applicant := c.Query("applicant")
	keyword := c.Query("keyword")

	var total int64
	var records []models.ExceptionManagement

	query := database.GetDB().Model(&models.ExceptionManagement{}).Where("1=1")

	if applyDate != "" {
		query = query.Where("apply_date LIKE ?", "%"+applyDate+"%")
	}
	if endDate != "" {
		query = query.Where("end_date LIKE ?", "%"+endDate+"%")
	}
	if applicant != "" {
		query = query.Where("applicant LIKE ?", "%"+applicant+"%")
	}
	if keyword != "" {
		query = query.Where("(reason LIKE ? OR applicant LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"data":     records,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateExceptionManagement 创建例外管理记录（需要双控）
func CreateExceptionManagement(c *gin.Context) {
	applyDate := c.PostForm("apply_date")
	applicant := strings.TrimSpace(c.PostForm("applicant"))
	reason := strings.TrimSpace(c.PostForm("reason"))
	endDate := c.PostForm("endDate")

	if applyDate == "" || applicant == "" || reason == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "申请日期、申请人、说明和截止时间不能为空"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil || file == nil {
		log.Printf("获取上传文件失败：%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择 PDF 文件"})
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 格式"})
		return
	}

	// XSS 防护：清理文件名中的危险字符
	cleanFileName := filepath.Base(file.Filename)
	safeFileName := cleanFileName // 使用基础文件名

	// 生成唯一文件名（使用安全文件名）
	now := time.Now()
	filename := fmt.Sprintf("%d_%s%s", now.UnixNano(), applicant, ext)
	dirPath := config.Cfg.Upload.ExceptionManagementPath
	os.MkdirAll(dirPath, 0750) // 统一使用安全权限：仅所有者和组可读可写可执行
	filePath := filepath.Join(dirPath, filename)

	// 转换为绝对路径存储，避免路径格式不一致问题
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		log.Printf("解析文件路径失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件上传失败"})
		return
	}
	filePath = absFilePath

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("保存文件失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件上传失败"})
		return
	}

	record := models.ExceptionManagement{
		ApplyDate:  applyDate,
		Applicant:  applicant,
		Reason:     reason,
		EndDate:    endDate,
		FileName:   safeFileName,
		FilePath:   filePath,
		FileSize:   file.Size,
		FileType:   "pdf",
	}

	if err := database.GetDB().Create(&record).Error; err != nil {
		os.Remove(filePath)
		log.Printf("保存数据库失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "ApplyDate", FieldLabel: "申请日期", OldValue: "-", NewValue: applyDate},
		{FieldName: "Applicant", FieldLabel: "申请人", OldValue: "-", NewValue: applicant},
		{FieldName: "Reason", FieldLabel: "例外说明", OldValue: "-", NewValue: reason},
		{FieldName: "EndDate", FieldLabel: "持续时间", OldValue: "-", NewValue: endDate},
		{FieldName: "FileName", FieldLabel: "扫描件", OldValue: "-", NewValue: file.Filename},
	}
	services.LogOperation(username, displayName, "新增例外管理", "exception_management", record.ID, file.Filename, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "添加成功"})
}

// UpdateExceptionManagement 更新例外管理记录（需要双控）
func UpdateExceptionManagement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ExceptionManagement
	
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	oldRecord := record
	applyDate := c.PostForm("apply_date")
	applicant := strings.TrimSpace(c.PostForm("applicant"))
	reason := strings.TrimSpace(c.PostForm("reason"))
	endDate := c.PostForm("endDate")

	var newFileName, newPath string
	var fileSize int64
	var fileType string

	// 处理文件替换
	newFile, err := c.FormFile("file")
	if err == nil && newFile != nil {
		// 有新文件时才验证必填字段
		if applyDate == "" || applicant == "" || reason == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "申请日期、申请人、说明和截止时间不能为空"})
			return
		}

		ext := strings.ToLower(filepath.Ext(newFile.Filename))
		if ext != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 格式"})
			return
		}

		now := time.Now()
		newFileName = fmt.Sprintf("%d_%s%s", now.UnixNano(), applicant, ext)
		dirPath := config.Cfg.Upload.ExceptionManagementPath
		os.MkdirAll(dirPath, 0750) // 改进权限：0750 比 0755 更安全
		newPath = filepath.Join(dirPath, newFileName)

		if err := c.SaveUploadedFile(newFile, newPath); err != nil {
			log.Printf("保存文件失败：%v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件上传失败"})
			return
		}

		// 转换为绝对路径存储
		absNewPath, err := filepath.Abs(newPath)
		if err != nil {
			log.Printf("解析文件路径失败：%v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件上传失败"})
			return
		}
		newPath = absNewPath

		fileSize = newFile.Size
		fileType = "pdf"
		// 注意：暂不删除旧文件，等后续数据库保存成功后再删除
	}

	// 更新数据
	record.ApplyDate = applyDate
	record.Applicant = applicant
	record.Reason = reason
	record.EndDate = endDate
	if newFile != nil {
		record.FileName = newFileName
		record.FilePath = newPath
		record.FileSize = fileSize
		record.FileType = fileType
	}

	// 先保存数据库
	if err := database.GetDB().Save(&record).Error; err != nil {
		log.Printf("更新数据库失败：%v", err)
		// 如果数据库保存失败，删除新上传的文件
		if newFile != nil {
			os.Remove(newPath)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 只有数据库保存成功后才删除旧文件
	if newFile != nil {
		if err := os.Remove(record.FilePath); err != nil {
			log.Printf("删除旧文件失败：%v", err)
		}
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("exception_management")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	
	if newFile != nil {
		details = append(details, services.LogDetail{
			FieldName: "FileName", FieldLabel: "扫描件", 
			OldValue: oldRecord.FileName, 
			NewValue: newFileName,
		})
	}

	services.LogOperation(username, displayName, "更新例外管理", "exception_management", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteExceptionManagement 删除例外管理记录（需要双控）
func DeleteExceptionManagement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ExceptionManagement
	
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	oldFileName := record.FileName

	// 先从数据库删除（使用软删除）
	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 确认数据库删除成功后再删除文件
	os.Remove(record.FilePath)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "FileName", FieldLabel: "扫描件", OldValue: oldFileName, NewValue: "-"},
	}
	services.LogOperation(username, displayName, "删除例外管理", "exception_management", record.ID, oldFileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewExceptionManagement 预览 PDF 文件（不需要双控）
func PreviewExceptionManagement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ExceptionManagement
	
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	// 验证文件路径在允许的目录内（防止路径遍历攻击）
	allowedDir := config.Cfg.Upload.ExceptionManagementPath
	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		log.Printf("解析允许目录路径失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "路径解析失败"})
		return
	}

	// 使用 filepath.Clean 清理路径并转换为绝对路径
	absFilePath := filepath.Clean(record.FilePath)
	if absFilePath[0] != '/' && absFilePath[0] != '.' {
		var err error
		absFilePath, err = filepath.Abs(absFilePath)
		if err != nil {
			log.Printf("解析文件路径失败：%v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "路径解析失败"})
			return
		}
	}

	if !strings.HasPrefix(absFilePath, absAllowedDir+string(filepath.Separator)) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "非法的文件访问"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", record.FileName))
	c.File(record.FilePath)
}

// DownloadExceptionManagement 下载 PDF 文件（不需要双控）
func DownloadExceptionManagement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ExceptionManagement
	
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	// 验证文件路径在允许的目录内（防止路径遍历攻击）
	allowedDir := config.Cfg.Upload.ExceptionManagementPath
	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		log.Printf("解析允许目录路径失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "路径解析失败"})
		return
	}

	// 使用 filepath.Clean 清理路径并转换为绝对路径
	absFilePath := filepath.Clean(record.FilePath)
	if absFilePath[0] != '/' && absFilePath[0] != '.' {
		var err error
		absFilePath, err = filepath.Abs(absFilePath)
		if err != nil {
			log.Printf("解析文件路径失败：%v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "路径解析失败"})
			return
		}
	}

	if !strings.HasPrefix(absFilePath, absAllowedDir+string(filepath.Separator)) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "非法的文件访问"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", record.FileName))
	c.Header("Content-Length", fmt.Sprintf("%d", record.FileSize))
	c.File(record.FilePath)
}
