package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

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

// GetPositionPermissions 获取特定岗位的权限规则
func GetPositionPermissions(c *gin.Context) {
	positionName := c.Query("position_name")
	if positionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请提供岗位名称"})
		return
	}

	var rule models.PermissionRule
	if err := database.GetDB().Where("position_name = ?", positionName).First(&rule).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "未找到该岗位的权限规则"})
		return
	}

	// 解析权限规则 - roles 是对象数组 [{"enabled":true,"name":"角色名"}]
	var allowedRules []struct {
		System string `json:"system"`
		Roles  []struct {
			Enabled bool   `json:"enabled"`
			Name    string `json:"name"`
		} `json:"roles"`
	}
	if err := json.Unmarshal([]byte(rule.RulesJSON), &allowedRules); err != nil {
		log.Printf("解析权限规则失败: %v, rules_json=%s", err, rule.RulesJSON)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解析权限规则失败"})
		return
	}

	// 构建响应格式: { system: [role1, role2] } - 只返回 enabled=true 的角色
	result := make(map[string][]string)
	for _, ar := range allowedRules {
		var roles []string
		for _, role := range ar.Roles {
			if role.Enabled {
				roles = append(roles, role.Name)
			}
		}
		if len(roles) > 0 {
			result[ar.System] = roles
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
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

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "PositionName", FieldLabel: "岗位名称", NewValue: rule.PositionName},
		{FieldName: "SortOrder", FieldLabel: "排序", NewValue: strconv.Itoa(rule.SortOrder)},
		{FieldName: "RulesJSON", FieldLabel: "权限规则", NewValue: rule.RulesJSON},
	}
	services.LogOperation(username, displayName, "创建岗位权限", "permission_rule", rule.ID, rule.PositionName, "", c.ClientIP(), details)

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

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "SystemName", FieldLabel: "系统名称", NewValue: req.SystemName},
	}
	services.LogOperation(username, displayName, "添加系统到权限", "permission_rule", 0, req.SystemName, "", c.ClientIP(), details)
}

// UpdatePermissionRule 更新岗位权限规则
func UpdatePermissionRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取旧值
	var oldRule models.PermissionRule
	if err := database.GetDB().First(&oldRule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "岗位不存在"})
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

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("permission_rule")
	details := services.DiffStructs(oldRule, rule, fieldLabels)
	services.LogOperation(username, displayName, "更新岗位权限", "permission_rule", rule.ID, rule.PositionName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": rule, "message": "更新成功"})
}

// DeletePermissionRule 删除岗位
func DeletePermissionRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 获取旧值用于日志
	var rule models.PermissionRule
	database.GetDB().First(&rule, id)

	if err := database.GetDB().Delete(&models.PermissionRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("permission_rule")
	details := services.DiffStructs(rule, models.PermissionRule{}, fieldLabels)
	services.LogOperation(username, displayName, "删除岗位权限", "permission_rule", rule.ID, rule.PositionName, approver, c.ClientIP(), details)

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

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	details := []services.LogDetail{
		{FieldName: "SystemName", FieldLabel: "系统名称", NewValue: req.SystemName},
	}
	services.LogOperation(username.(string), displayName.(string), "从权限移除系统", "permission_rule", 0, req.SystemName, "", c.ClientIP(), details)
}

// RenameSystemInPermissions 重命名系统
func RenameSystemInPermissions(c *gin.Context) {
	var req struct {
		OldName string `json:"old_name" binding:"required"`
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var allRules []models.PermissionRule
	database.GetDB().Find(&allRules)
	updatedCount := 0
	for _, rule := range allRules {
		var sysRules []map[string]interface{}
		json.Unmarshal([]byte(rule.RulesJSON), &sysRules)
		changed := false
		for _, sr := range sysRules {
			if name, ok := sr["system"].(string); ok && name == req.OldName {
				sr["system"] = req.NewName
				changed = true
				break
			}
		}
		if changed {
			newJSON, _ := json.Marshal(sysRules)
			database.GetDB().Model(&rule).Update("rules_json", string(newJSON))
			updatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "重命名成功", "data": gin.H{"updated_count": updatedCount}})

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	details := []services.LogDetail{
		{FieldName: "OldName", FieldLabel: "旧名称", OldValue: req.OldName, NewValue: req.NewName},
	}
	services.LogOperation(username.(string), displayName.(string), "重命名系统", "permission_rule", 0, req.OldName+"->"+req.NewName, "", c.ClientIP(), details)
}

// ManageRolesInSystem 管理系统内的角色（新增/重命名/删除）
func ManageRolesInSystem(c *gin.Context) {
	var req struct {
		SystemName string `json:"system_name" binding:"required"`
		Action     string `json:"action" binding:"required"` // add / rename / delete
		OldName    string `json:"old_name"`                   // rename/delete时需要
		NewName    string `json:"new_name"`                   // add/rename时需要
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var allRules []models.PermissionRule
	database.GetDB().Find(&allRules)
	updatedCount := 0

	for _, rule := range allRules {
		var sysRules []map[string]interface{}
		json.Unmarshal([]byte(rule.RulesJSON), &sysRules)
		changed := false

		for _, sr := range sysRules {
			name, _ := sr["system"].(string)
			if name != req.SystemName {
				continue
			}
			roles, _ := sr["roles"].([]interface{})

			switch req.Action {
			case "add":
				// 检查是否已存在
				exists := false
				for _, r := range roles {
					if m, ok := r.(map[string]interface{}); ok {
						if m["name"] == req.NewName {
							exists = true
							break
						}
					}
				}
				if !exists {
					roles = append(roles, map[string]interface{}{"name": req.NewName, "enabled": false})
					sr["roles"] = roles
					changed = true
				}
			case "rename":
				for i, r := range roles {
					if m, ok := r.(map[string]interface{}); ok {
						if m["name"] == req.OldName {
							m["name"] = req.NewName
							roles[i] = m
							changed = true
							break
						}
					}
				}
			case "delete":
				newRoles := []interface{}{}
				for _, r := range roles {
					if m, ok := r.(map[string]interface{}); ok {
						if m["name"] == req.OldName {
							changed = true
							continue
						}
					}
					newRoles = append(newRoles, r)
				}
				sr["roles"] = newRoles
			}
			break
		}

		if changed {
			newJSON, _ := json.Marshal(sysRules)
			database.GetDB().Model(&rule).Update("rules_json", string(newJSON))
			updatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "操作成功", "data": gin.H{"updated_count": updatedCount}})

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	details := []services.LogDetail{
		{FieldName: "SystemName", FieldLabel: "系统名称", NewValue: req.SystemName},
		{FieldName: "Action", FieldLabel: "操作", NewValue: req.Action},
	}
	services.LogOperation(username.(string), displayName.(string), "管理角色", "permission_rule", 0, req.SystemName, "", c.ClientIP(), details)
}

// ReorderPermissionRule 调整岗位排序（上移/下移）
func ReorderPermissionRule(c *gin.Context) {
	var req struct {
		ID        uint   `json:"id" binding:"required"`
		Direction string `json:"direction" binding:"required"` // "up" or "down"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var current models.PermissionRule
	if err := database.GetDB().First(&current, req.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "岗位不存在"})
		return
	}

	// 查找相邻岗位
	var adjacent models.PermissionRule
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

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	details := []services.LogDetail{
		{FieldName: "PositionName", FieldLabel: "岗位名称", NewValue: current.PositionName},
		{FieldName: "Direction", FieldLabel: "方向", NewValue: req.Direction},
	}
	services.LogOperation(username.(string), displayName.(string), "调整岗位排序", "permission_rule", current.ID, current.PositionName, "", c.ClientIP(), details)
}

// ReorderSystemInPermissions 调整系统在所有岗位中的排序
func ReorderSystemInPermissions(c *gin.Context) {
	var req struct {
		SystemName string `json:"system_name" binding:"required"`
		Direction  string `json:"direction" binding:"required"` // "up" or "down"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var allRules []models.PermissionRule
	database.GetDB().Find(&allRules)
	if len(allRules) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "没有岗位数据"})
		return
	}

	updatedCount := 0
	for _, rule := range allRules {
		var sysRules []map[string]interface{}
		json.Unmarshal([]byte(rule.RulesJSON), &sysRules)

		// 找到当前系统的索引
		currentIdx := -1
		for i, sr := range sysRules {
			if name, ok := sr["system"].(string); ok && name == req.SystemName {
				currentIdx = i
				break
			}
		}
		if currentIdx == -1 {
			continue
		}

		targetIdx := currentIdx
		if req.Direction == "up" {
			if currentIdx == 0 {
				continue // 已到顶部，跳过
			}
			targetIdx = currentIdx - 1
		} else {
			if currentIdx == len(sysRules)-1 {
				continue // 已到底部，跳过
			}
			targetIdx = currentIdx + 1
		}

		// 交换位置（保持每个岗位的角色状态不变）
		sysRules[currentIdx], sysRules[targetIdx] = sysRules[targetIdx], sysRules[currentIdx]

		newJSON, _ := json.Marshal(sysRules)
		database.GetDB().Model(&rule).Update("rules_json", string(newJSON))
		updatedCount++
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移动成功", "data": gin.H{"updated_count": updatedCount}})

	// 记录操作日志
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	details := []services.LogDetail{
		{FieldName: "SystemName", FieldLabel: "系统名称", NewValue: req.SystemName},
		{FieldName: "Direction", FieldLabel: "方向", NewValue: req.Direction},
	}
	services.LogOperation(username.(string), displayName.(string), "调整系统排序", "permission_rule", 0, req.SystemName, "", c.ClientIP(), details)
}
