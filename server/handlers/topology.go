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

// ListTopologies 获取拓扑图列表
func ListTopologies(c *gin.Context) {
	var topologies []models.Topology
	if err := database.GetDB().Order("created_at DESC").Find(&topologies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": topologies})
}

// CreateTopology 创建拓扑图
func CreateTopology(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "名称不能为空"})
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
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持PNG、JPG、GIF、SVG格式图片"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(config.Cfg.Upload.TopologyPath, filename)

	// 确保目录存在
	os.MkdirAll(config.Cfg.Upload.TopologyPath, 0755)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	topology := models.Topology{
		Name:        name,
		Description: description,
		FileName:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
	}

	if err := database.GetDB().Create(&topology).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "名称", NewValue: topology.Name},
		{FieldName: "Description", FieldLabel: "描述", NewValue: topology.Description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: topology.FileName},
		{FieldName: "FilePath", FieldLabel: "文件路径", NewValue: topology.FilePath},
		{FieldName: "FileSize", FieldLabel: "文件大小", NewValue: fmt.Sprintf("%d", topology.FileSize)},
	}
	services.LogOperation(username, displayName, "创建拓扑图", "topology", topology.ID, topology.Name, "", c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": topology})
}

// UpdateTopology 更新拓扑图信息
func UpdateTopology(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var topology models.Topology
	if err := database.GetDB().First(&topology, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "拓扑图不存在"})
		return
	}

	// 保存旧值快照
	oldTopology := topology

	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	topology.Name = input.Name
	topology.Description = input.Description

	if err := database.GetDB().Save(&topology).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("topology")
	details := services.DiffStructs(oldTopology, topology, fieldLabels)
	services.LogOperation(username, displayName, "更新拓扑图", "topology", topology.ID, topology.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": topology})
}

// ReplaceTopologyFile 替换拓扑图文件
func ReplaceTopologyFile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var topology models.Topology
	if err := database.GetDB().First(&topology, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "拓扑图不存在"})
		return
	}

	// 保存旧值快照
	oldTopology := topology

	// 获取新文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持PNG、JPG、GIF、SVG格式图片"})
		return
	}

	// 删除旧文件
	os.Remove(topology.FilePath)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(config.Cfg.Upload.TopologyPath, filename)

	// 保存新文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	topology.FileName = file.Filename
	topology.FilePath = filePath
	topology.FileSize = file.Size

	if err := database.GetDB().Save(&topology).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("topology")
	details := services.DiffStructs(oldTopology, topology, fieldLabels)
	services.LogOperation(username, displayName, "替换拓扑图文件", "topology", topology.ID, topology.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "文件替换成功", "data": topology})
}

// DeleteTopology 删除拓扑图
func DeleteTopology(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var topology models.Topology
	if err := database.GetDB().First(&topology, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "拓扑图不存在"})
		return
	}

	// 删除文件
	os.Remove(topology.FilePath)

	if err := database.GetDB().Delete(&topology).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("topology")
	details := services.DiffStructs(topology, models.Topology{}, fieldLabels)
	services.LogOperation(username, displayName, "删除拓扑图", "topology", topology.ID, topology.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewTopology 预览拓扑图
func PreviewTopology(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var topology models.Topology
	if err := database.GetDB().First(&topology, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "拓扑图不存在"})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(topology.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	// 根据扩展名设置Content-Type
	ext := strings.ToLower(filepath.Ext(topology.FilePath))
	contentType := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	}

	c.Header("Content-Type", contentType)
	c.File(topology.FilePath)
}

// DownloadTopology 下载拓扑图
func DownloadTopology(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var topology models.Topology
	if err := database.GetDB().First(&topology, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "拓扑图不存在"})
		return
	}

	if _, err := os.Stat(topology.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	ext := strings.ToLower(filepath.Ext(topology.FilePath))
	contentType := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+topology.FileName+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", topology.FileSize))
	c.File(topology.FilePath)
}
