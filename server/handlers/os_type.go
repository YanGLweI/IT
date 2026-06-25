package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListOSTypes 获取操作系统类型列表
func ListOSTypes(c *gin.Context) {
	var osTypes []models.OSType
	if err := database.GetDB().Find(&osTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": osTypes})
}

// CreateOSType 创建操作系统类型
func CreateOSType(c *gin.Context) {
	var osType models.OSType
	if err := c.ShouldBindJSON(&osType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := database.GetDB().Create(&osType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": osType})
}

// UpdateOSType 更新操作系统类型
func UpdateOSType(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var osType models.OSType
	if err := database.GetDB().First(&osType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "操作系统类型不存在"})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	osType.Name = input.Name

	if err := database.GetDB().Save(&osType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": osType})
}

// DeleteOSType 删除操作系统类型
func DeleteOSType(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var osType models.OSType
	if err := database.GetDB().First(&osType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "操作系统类型不存在"})
		return
	}

	if err := database.GetDB().Delete(&osType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
