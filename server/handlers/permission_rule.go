package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListPermissionRules 获取所有岗位权限规则
func ListPermissionRules(c *gin.Context) {
	var rules []models.PermissionRule
	if err := database.GetDB().Order("sort_order asc").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rules})
}

// UpdatePermissionRule 更新岗位权限规则
func UpdatePermissionRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		PositionName string `json:"position_name"`
		RulesJSON    string `json:"rules_json"`
		SortOrder    int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	updates := map[string]interface{}{}
	if req.PositionName != "" {
		updates["position_name"] = req.PositionName
	}
	if req.RulesJSON != "" {
		updates["rules_json"] = req.RulesJSON
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}

	if err := database.GetDB().Model(&models.PermissionRule{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	var rule models.PermissionRule
	database.GetDB().First(&rule, id)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rule, "message": "更新成功"})
}
