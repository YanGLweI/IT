package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListSftpServers 获取SFTP服务器列表
func ListSftpServers(c *gin.Context) {
	var servers []models.SftpServer
	if err := database.GetDB().Order("sort_order ASC, id ASC").Find(&servers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": servers})
}

// CreateSftpServer 创建SFTP服务器
func CreateSftpServer(c *gin.Context) {
	var server models.SftpServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查同名
	var count int64
	database.GetDB().Model(&models.SftpServer{}).Where("name = ?", server.Name).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "服务器名称已存在"})
		return
	}

	if err := database.GetDB().Create(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": server})
}

// UpdateSftpServer 更新SFTP服务器
func UpdateSftpServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var server models.SftpServer
	if err := database.GetDB().First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务器不存在"})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查同名（排除自己）
	var count int64
	database.GetDB().Model(&models.SftpServer{}).Where("name = ? AND id != ?", input.Name, id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "服务器名称已存在"})
		return
	}

	server.Name = input.Name
	if err := database.GetDB().Save(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": server})
}

// DeleteSftpServer 删除SFTP服务器（级联删除账号）
func DeleteSftpServer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var server models.SftpServer
	if err := database.GetDB().First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务器不存在"})
		return
	}

	// 级联软删除该服务器下的所有账号
	database.GetDB().Where("server_id = ?", id).Delete(&models.SftpAccount{})

	// 删除服务器
	if err := database.GetDB().Delete(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
