package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// ListRegions 获取区域列表
func ListRegions(c *gin.Context) {
	var regions []models.Region
	if err := database.GetDB().Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": regions})
}

// CreateRegion 创建区域
func CreateRegion(c *gin.Context) {
	var region models.Region
	if err := c.ShouldBindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 清理软删除的同名记录（避免唯一索引冲突）
	var softDeleted models.Region
	result := database.GetDB().Unscoped().Where("name = ?", region.Name).Find(&softDeleted)
	if result.Error == nil && result.RowsAffected > 0 {
		database.GetDB().Unscoped().Delete(&softDeleted)
	}

	if err := database.GetDB().Create(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "名称", NewValue: region.Name},
		{FieldName: "Description", FieldLabel: "描述", NewValue: region.Description},
	}
	services.LogOperation(username, displayName, "创建区域", "region", region.ID, region.Name, "", c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": region})
}

// UpdateRegion 更新区域
func UpdateRegion(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var region models.Region
	if err := database.GetDB().First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "区域不存在"})
		return
	}

	// 保存旧值快照
	oldRegion := region

	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	region.Name = input.Name
	region.Description = input.Description

	if err := database.GetDB().Save(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("region")
	details := services.DiffStructs(oldRegion, region, fieldLabels)
	services.LogOperation(username, displayName, "更新区域", "region", region.ID, region.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": region})
}

// DeleteRegion 删除区域
func DeleteRegion(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var region models.Region
	if err := database.GetDB().First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "区域不存在"})
		return
	}

	// 检查是否有关联资产
	var count int64
	database.GetDB().Model(&models.Asset{}).Where("region_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该区域下还有资产，无法删除"})
		return
	}

	if err := database.GetDB().Unscoped().Delete(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("region")
	details := services.DiffStructs(region, models.Region{}, fieldLabels)
	services.LogOperation(username, displayName, "删除区域", "region", region.ID, region.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
