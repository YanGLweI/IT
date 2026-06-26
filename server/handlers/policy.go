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

// ListPolicies 获取政策列表
func ListPolicies(c *gin.Context) {
	var policies []models.Policy
	if err := database.GetDB().Order("created_at DESC").Find(&policies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": policies})
}

// CreatePolicy 创建政策
func CreatePolicy(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "标题不能为空"})
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(config.Cfg.Upload.PolicyPath, filename)

	// 确保目录存在
	os.MkdirAll(config.Cfg.Upload.PolicyPath, 0755)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	policy := models.Policy{
		Title:       title,
		Description: description,
		FileName:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
		FileType:    file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&policy).Error; err != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	details := []services.LogDetail{
		{FieldName: "Title", FieldLabel: "标题", NewValue: policy.Title},
		{FieldName: "Description", FieldLabel: "描述", NewValue: policy.Description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: policy.FileName},
		{FieldName: "FilePath", FieldLabel: "文件路径", NewValue: policy.FilePath},
		{FieldName: "FileSize", FieldLabel: "文件大小", NewValue: fmt.Sprintf("%d", policy.FileSize)},
		{FieldName: "FileType", FieldLabel: "文件类型", NewValue: policy.FileType},
	}
	services.LogOperation(username.(string), displayName.(string), "创建政策", "policy", policy.ID, policy.Title, "", c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": policy})
}

// UpdatePolicy 更新政策信息
func UpdatePolicy(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy models.Policy
	if err := database.GetDB().First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "政策不存在"})
		return
	}

	// 保存旧值快照
	oldPolicy := policy

	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	policy.Title = input.Title
	policy.Description = input.Description

	if err := database.GetDB().Save(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	approver, _ := c.Get("dual_control_verified_by")
	fieldLabels := services.GetFieldLabels("policy")
	details := services.DiffStructs(oldPolicy, policy, fieldLabels)
	services.LogOperation(username.(string), displayName.(string), "更新政策", "policy", policy.ID, policy.Title, approver.(string), c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": policy})
}

// ReplacePolicyFile 替换政策文件
func ReplacePolicyFile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy models.Policy
	if err := database.GetDB().First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "政策不存在"})
		return
	}

	// 保存旧值快照
	oldPolicy := policy

	// 获取新文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 删除旧文件
	os.Remove(policy.FilePath)

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(config.Cfg.Upload.PolicyPath, filename)

	// 保存新文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	policy.FileName = file.Filename
	policy.FilePath = filePath
	policy.FileSize = file.Size
	policy.FileType = file.Header.Get("Content-Type")

	if err := database.GetDB().Save(&policy).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	approver, _ := c.Get("dual_control_verified_by")
	fieldLabels := services.GetFieldLabels("policy")
	details := services.DiffStructs(oldPolicy, policy, fieldLabels)
	services.LogOperation(username.(string), displayName.(string), "替换政策文件", "policy", policy.ID, policy.Title, approver.(string), c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "文件替换成功", "data": policy})
}

// DeletePolicy 删除政策
func DeletePolicy(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy models.Policy
	if err := database.GetDB().First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "政策不存在"})
		return
	}

	// 删除文件
	os.Remove(policy.FilePath)

	if err := database.GetDB().Delete(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	approver, _ := c.Get("dual_control_verified_by")
	fieldLabels := services.GetFieldLabels("policy")
	details := services.DiffStructs(policy, models.Policy{}, fieldLabels)
	services.LogOperation(username.(string), displayName.(string), "删除政策", "policy", policy.ID, policy.Title, approver.(string), c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewPolicy 预览政策文件
func PreviewPolicy(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy models.Policy
	if err := database.GetDB().First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "政策不存在"})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(policy.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	// 设置Content-Type
	c.Header("Content-Type", policy.FileType)
	c.Header("Content-Disposition", "inline; filename=\""+policy.FileName+"\"")
	c.File(policy.FilePath)
}

// DownloadPolicy 下载政策文件
func DownloadPolicy(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy models.Policy
	if err := database.GetDB().First(&policy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "政策不存在"})
		return
	}

	if _, err := os.Stat(policy.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", policy.FileType)
	c.Header("Content-Disposition", "attachment; filename=\""+policy.FileName+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", policy.FileSize))
	c.File(policy.FilePath)
}
