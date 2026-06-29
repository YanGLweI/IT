package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 变更类型管理
// ============================================================

// ListChangeTypes 获取变更类型列表
func ListChangeTypes(c *gin.Context) {
	var types []models.ChangeType
	if err := database.GetDB().Order("sort_order ASC, id ASC").Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": types})
}

// CreateChangeType 创建变更类型
func CreateChangeType(c *gin.Context) {
	var input struct {
		Name      string `json:"name" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查名称是否重复
	var count int64
	database.GetDB().Model(&models.ChangeType{}).Where("name = ?", input.Name).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "类型名称已存在"})
		return
	}

	ct := models.ChangeType{
		Name:      input.Name,
		SortOrder: input.SortOrder,
	}
	if err := database.GetDB().Create(&ct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "名称", NewValue: ct.Name},
		{FieldName: "SortOrder", FieldLabel: "排序", NewValue: strconv.Itoa(ct.SortOrder)},
	}
	services.LogOperation(username, displayName, "创建变更类型", "change_type", ct.ID, ct.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": ct})
}

// UpdateChangeType 更新变更类型
func UpdateChangeType(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ct models.ChangeType
	if err := database.GetDB().First(&ct, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "类型不存在"})
		return
	}

	var input struct {
		Name      string `json:"name" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查名称是否与其他记录重复
	var count int64
	database.GetDB().Model(&models.ChangeType{}).Where("name = ? AND id != ?", input.Name, id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "类型名称已存在"})
		return
	}

	oldType := ct
	ct.Name = input.Name
	ct.SortOrder = input.SortOrder

	if err := database.GetDB().Save(&ct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("change_type")
	details := services.DiffStructs(oldType, ct, fieldLabels)
	services.LogOperation(username, displayName, "更新变更类型", "change_type", ct.ID, ct.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": ct})
}

// DeleteChangeType 删除变更类型
func DeleteChangeType(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ct models.ChangeType
	if err := database.GetDB().First(&ct, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "类型不存在"})
		return
	}

	// 检查是否有关联的变更记录
	var count int64
	database.GetDB().Table("change_record_change_types").Where("change_type_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该类型已被变更记录使用，无法删除"})
		return
	}

	if err := database.GetDB().Delete(&ct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("change_type")
	details := services.DiffStructs(ct, models.ChangeType{}, fieldLabels)
	services.LogOperation(username, displayName, "删除变更类型", "change_type", ct.ID, ct.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ReorderChangeType 调整变更类型排序（上移/下移）
func ReorderChangeType(c *gin.Context) {
	var req struct {
		ID        uint   `json:"id" binding:"required"`
		Direction string `json:"direction" binding:"required"` // "up" or "down"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var current models.ChangeType
	if err := database.GetDB().First(&current, req.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "类型不存在"})
		return
	}

	// 查找相邻类型
	var adjacent models.ChangeType
	var query string
	var order string
	if req.Direction == "up" {
		query = "sort_order < ?"
		order = "sort_order desc"
	} else {
		query = "sort_order > ?"
		order = "sort_order asc"
	}

	if err := database.GetDB().Where(query, current.SortOrder).Order(order).First(&adjacent).Error; err != nil {
		msg := "已到达顶部"
		if req.Direction == "down" {
			msg = "已到达底部"
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": msg})
		return
	}

	// 交换 sort_order
	tempOrder := current.SortOrder
	database.GetDB().Model(&current).Update("sort_order", adjacent.SortOrder)
	database.GetDB().Model(&adjacent).Update("sort_order", tempOrder)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移动成功"})
}
