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

// ListFormVaultItems 获取保管区所有表单列表
func ListFormVaultItems(c *gin.Context) {
	var items []models.FormVaultItem

	query := database.GetDB().Model(&models.FormVaultItem{})

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if sourceType := c.Query("source_type"); sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if published := c.Query("is_published"); published != "" {
		if published == "true" {
			query = query.Where("is_published = ?", true)
		} else if published == "false" {
			query = query.Where("is_published = ?", false)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Count(&total)

	if err := query.Order("sort_order ASC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "total": total, "page_size": pageSize})
}

// UploadFormVaultItem 上传新表单到保管区
func UploadFormVaultItem(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	category := c.PostForm("category")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "表单标题不能为空"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".docx": true, ".pdf": true, ".xlsx": true, ".xls": true, ".doc": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 DOCX、PDF、XLSX、XLS、DOC 格式文件"})
		return
	}

	// 按年份归档
	yearDir := filepath.Join(config.Cfg.Upload.FormVaultPath, time.Now().Format("2006"))
	os.MkdirAll(yearDir, 0755)

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	item := models.FormVaultItem{
		Title:       title,
		Description: description,
		Category:    category,
		SourceType:  models.SourceTypeUpload,
		FileName:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
		FileType:    file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&item).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存表单失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Title", FieldLabel: "标题", NewValue: title},
		{FieldName: "Category", FieldLabel: "分类", NewValue: category},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: item.FileName},
	}
	services.LogOperation(username, displayName, "上传表单", "form_vault", item.ID, item.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": item})
}

// UpdateFormVaultItem 编辑表单信息
func UpdateFormVaultItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在"})
		return
	}

	oldItem := item

	title := c.PostForm("title")
	description := c.PostForm("description")
	category := c.PostForm("category")

	if title != "" {
		item.Title = title
	}
	item.Description = description
	item.Category = category

	if err := database.GetDB().Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("form_vault")
	details := services.DiffStructs(oldItem, item, fieldLabels)
	services.LogOperation(username, displayName, "编辑表单", "form_vault", item.ID, item.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": item})
}

// DeleteFormVaultItem 删除表单
func DeleteFormVaultItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在"})
		return
	}

	// 删除文件
	if item.SourceType == models.SourceTypeUpload && item.FilePath != "" {
		os.Remove(item.FilePath)
	}
	if item.SourceType == models.SourceTypeStatic && item.SnapshotPath != "" {
		os.Remove(item.SnapshotPath)
	}

	if err := database.GetDB().Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("form_vault")
	details := services.DiffStructs(item, models.FormVaultItem{}, fieldLabels)
	services.LogOperation(username, displayName, "删除表单", "form_vault", item.ID, item.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PublishFormVaultItem 发布表单
func PublishFormVaultItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在"})
		return
	}

	now := time.Now()
	if err := database.GetDB().Model(&item).Updates(map[string]interface{}{
		"is_published": true,
		"published_at": &now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "发布失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "IsPublished", FieldLabel: "发布状态", OldValue: "未发布", NewValue: "已发布"},
	}
	services.LogOperation(username, displayName, "发布表单", "form_vault", item.ID, item.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "发布成功"})
}

// UnpublishFormVaultItem 取消发布
func UnpublishFormVaultItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在"})
		return
	}

	if err := database.GetDB().Model(&item).Updates(map[string]interface{}{
		"is_published": false,
		"published_at": nil,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消发布失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "IsPublished", FieldLabel: "发布状态", OldValue: "已发布", NewValue: "未发布"},
	}
	services.LogOperation(username, displayName, "取消发布表单", "form_vault", item.ID, item.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "取消发布成功"})
}

// DownloadFormVaultItem 下载表单文件（管理端）
func DownloadFormVaultItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在"})
		return
	}

	serveFormFile(c, &item)
}

// PreviewFormVaultItem 预览表单文件（管理端）
func PreviewFormVaultItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在"})
		return
	}

	serveFormFileInline(c, &item)
}

// serveFormFile 根据来源类型提供文件下载
func serveFormFile(c *gin.Context, item *models.FormVaultItem) {
	switch item.SourceType {
	case models.SourceTypeUpload:
		if _, err := os.Stat(item.FilePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
			return
		}
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename=\""+item.FileName+"\"")
		c.File(item.FilePath)

	case models.SourceTypeStatic:
		if _, err := os.Stat(item.SnapshotPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
			return
		}
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename=\""+item.FileName+"\"")
		c.File(item.SnapshotPath)

	case models.SourceTypeDynamic:
		generator, err := GetDynamicGenerator(item.RefHandler)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "动态生成器未找到: " + item.RefHandler})
			return
		}
		generator(c)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未知的来源类型"})
	}
}

// serveFormFileInline 根据来源类型提供文件预览
func serveFormFileInline(c *gin.Context, item *models.FormVaultItem) {
	switch item.SourceType {
	case models.SourceTypeUpload:
		if _, err := os.Stat(item.FilePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
			return
		}
		setInlineContentType(c, item.FileName)
		c.Header("Content-Disposition", "inline; filename=\""+item.FileName+"\"")
		c.File(item.FilePath)

	case models.SourceTypeStatic:
		if _, err := os.Stat(item.SnapshotPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
			return
		}
		setInlineContentType(c, item.FileName)
		c.Header("Content-Disposition", "inline; filename=\""+item.FileName+"\"")
		c.File(item.SnapshotPath)

	case models.SourceTypeDynamic:
		generator, err := GetDynamicGenerator(item.RefHandler)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "动态生成器未找到: " + item.RefHandler})
			return
		}
		generator(c)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未知的来源类型"})
	}
}

// setInlineContentType 根据文件扩展名设置预览 Content-Type
func setInlineContentType(c *gin.Context, fileName string) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		c.Header("Content-Type", "application/pdf")
	case ".docx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	case ".xlsx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}
}
