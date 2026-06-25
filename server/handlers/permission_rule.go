package handlers

import (
	"encoding/json"
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

// CreatePermissionRule 新增岗位
func CreatePermissionRule(c *gin.Context) {
	var req struct {
		PositionName string `json:"position_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "岗位名称不能为空"})
		return
	}

	// 从第一个已有岗位复制系统结构，所有角色设为未授权
	var first models.PermissionRule
	if err := database.GetDB().Order("sort_order asc").First(&first).Error; err != nil {
		// 没有参考数据，创建一个空规则
		rule := models.PermissionRule{
			PositionName: req.PositionName,
			SortOrder:    0,
			RulesJSON:    "[]",
		}
		database.GetDB().Create(&rule)
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": rule, "message": "创建成功"})
		return
	}

	// 解析原有规则，所有角色设为未授权
	var sysRules []map[string]interface{}
	json.Unmarshal([]byte(first.RulesJSON), &sysRules)
	for _, sr := range sysRules {
		if roles, ok := sr["roles"].([]interface{}); ok {
			for _, r := range roles {
				if role, ok := r.(map[string]interface{}); ok {
					role["enabled"] = false
				}
			}
		}
	}
	newRulesJSON, _ := json.Marshal(sysRules)

	// 获取最大 sort_order
	var maxSort int64
	database.GetDB().Model(&models.PermissionRule{}).Select("COALESCE(MAX(sort_order),0)").Scan(&maxSort)

	rule := models.PermissionRule{
		PositionName: req.PositionName,
		SortOrder:    int(maxSort) + 1,
		RulesJSON:    string(newRulesJSON),
	}
	database.GetDB().Create(&rule)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rule, "message": "创建成功"})
}

// AddSystemToPermissions 向所有岗位新增系统
func AddSystemToPermissions(c *gin.Context) {
	var req struct {
		SystemName string   `json:"system_name" binding:"required"`
		Roles      []string `json:"roles"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "系统名称不能为空"})
		return
	}

	// 构建新系统的角色列表
	type roleItem struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	var newRoles []roleItem
	for _, name := range req.Roles {
		newRoles = append(newRoles, roleItem{Name: name, Enabled: false})
	}
	newSystemRule := map[string]interface{}{
		"system": req.SystemName,
		"roles":  newRoles,
	}

	// 更新所有岗位
	var allRules []models.PermissionRule
	database.GetDB().Find(&allRules)
	updatedCount := 0
	for _, rule := range allRules {
		var sysRules []map[string]interface{}
		json.Unmarshal([]byte(rule.RulesJSON), &sysRules)
		sysRules = append(sysRules, newSystemRule)
		newJSON, _ := json.Marshal(sysRules)
		database.GetDB().Model(&rule).Update("rules_json", string(newJSON))
		updatedCount++
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "添加成功", "data": gin.H{"updated_count": updatedCount}})
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

// DeletePermissionRule 删除岗位
func DeletePermissionRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := database.GetDB().Delete(&models.PermissionRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// RemoveSystemFromPermissions 从所有岗位移除系统
func RemoveSystemFromPermissions(c *gin.Context) {
	var req struct {
		SystemName string `json:"system_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "系统名称不能为空"})
		return
	}

	var allRules []models.PermissionRule
	database.GetDB().Find(&allRules)
	updatedCount := 0
	for _, rule := range allRules {
		var sysRules []map[string]interface{}
		json.Unmarshal([]byte(rule.RulesJSON), &sysRules)
		filtered := []map[string]interface{}{}
		for _, sr := range sysRules {
			if name, ok := sr["system"].(string); ok && name == req.SystemName {
				continue
			}
			filtered = append(filtered, sr)
		}
		if len(filtered) != len(sysRules) {
			newJSON, _ := json.Marshal(filtered)
			database.GetDB().Model(&rule).Update("rules_json", string(newJSON))
			updatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移除成功", "data": gin.H{"updated_count": updatedCount}})
}
