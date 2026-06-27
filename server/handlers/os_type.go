package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

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

	// 清理名称前后空格
	osType.Name = strings.TrimSpace(osType.Name)

	// 清理软删除的同名记录（避免唯一索引冲突）
	var softDeleted models.OSType
	result := database.GetDB().Unscoped().Where("name = ?", osType.Name).Find(&softDeleted)
	if result.Error == nil && result.RowsAffected > 0 {
		database.GetDB().Unscoped().Delete(&softDeleted)
	}

	if err := database.GetDB().Create(&osType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "名称", NewValue: osType.Name},
	}
	services.LogOperation(username, displayName, "创建操作系统类型", "os_type", osType.ID, osType.Name, "", c.ClientIP(), details)

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

	// 保存旧值快照
	oldOsType := osType

	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	osType.Name = strings.TrimSpace(input.Name)

	if err := database.GetDB().Save(&osType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("os_type")
	details := services.DiffStructs(oldOsType, osType, fieldLabels)
	services.LogOperation(username, displayName, "更新操作系统类型", "os_type", osType.ID, osType.Name, approver, c.ClientIP(), details)

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

	// 检查是否有关联资产
	var count int64
	database.GetDB().Model(&models.Asset{}).Where("os_type_id = ?", osType.ID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该操作系统下还有资产，无法删除"})
		return
	}

	if err := database.GetDB().Unscoped().Delete(&osType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("os_type")
	details := services.DiffStructs(osType, models.OSType{}, fieldLabels)
	services.LogOperation(username, displayName, "删除操作系统类型", "os_type", osType.ID, osType.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
