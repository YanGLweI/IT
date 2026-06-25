package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// validateUserRoles 验证用户选择的角色是否符合岗位权限规则
// positionNames: 多个岗位用逗号分隔，如 "IT管理员,IT配置管理员"
func validateUserRoles(positionNames string, systemRolesJSON string) error {
	if systemRolesJSON == "" || systemRolesJSON == "[]" {
		return nil // 没有选择任何角色，无需验证
	}

	// 解析多个岗位
	positionNameList := strings.Split(positionNames, ",")
	for i := range positionNameList {
		positionNameList[i] = strings.TrimSpace(positionNameList[i])
	}

	// 合并所有岗位的权限规则
	allowedMap := make(map[string]map[string]bool)
	for _, positionName := range positionNameList {
		if positionName == "" {
			continue
		}

		// 获取该岗位的权限规则
		var rule models.PermissionRule
		if err := database.GetDB().Where("position_name = ?", positionName).First(&rule).Error; err != nil {
			return fmt.Errorf("未找到岗位 '%s' 的权限规则", positionName)
		}

		// 解析岗位权限规则 - roles 是对象数组 [{"enabled":true,"name":"角色名"}]
		var allowedRules []struct {
			System string `json:"system"`
			Roles  []struct {
				Enabled bool   `json:"enabled"`
				Name    string `json:"name"`
			} `json:"roles"`
		}
		if err := json.Unmarshal([]byte(rule.RulesJSON), &allowedRules); err != nil {
			return fmt.Errorf("解析岗位 '%s' 权限规则失败: %v", positionName, err)
		}

		// 构建允许的角色映射（只包含 enabled=true 的角色）
		for _, ar := range allowedRules {
			if allowedMap[ar.System] == nil {
				allowedMap[ar.System] = make(map[string]bool)
			}
			for _, role := range ar.Roles {
				if role.Enabled {
					allowedMap[ar.System][role.Name] = true
				}
			}
		}
	}

	// 解析用户选择的角色
	var userRoles []struct {
		System string   `json:"system"`
		Roles  []string `json:"roles"`
	}
	if err := json.Unmarshal([]byte(systemRolesJSON), &userRoles); err != nil {
		return fmt.Errorf("解析用户角色失败: %v", err)
	}

	// 验证每个角色是否在任一岗位中允许
	var invalidRoles []string
	for _, ur := range userRoles {
		if allowedMap[ur.System] == nil {
			// 该系统在所有岗位中都不允许
			invalidRoles = append(invalidRoles, fmt.Sprintf("%s: 所有角色", ur.System))
			continue
		}
		for _, role := range ur.Roles {
			if !allowedMap[ur.System][role] {
				invalidRoles = append(invalidRoles, fmt.Sprintf("%s: %s", ur.System, role))
			}
		}
	}

	if len(invalidRoles) > 0 {
		return fmt.Errorf("所选岗位不允许以下权限: %s", joinStrings(invalidRoles))
	}

	return nil
}

// joinStrings 将字符串数组用逗号连接
func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

// ListUserPermissions 获取用户权限列表
func ListUserPermissions(c *gin.Context) {
	departmentID := c.Query("department_id")

	query := database.GetDB()
	if departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}

	var users []models.UserPermission
	if err := query.Order("created_at asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": users})
}

// GetUserPermission 获取单个用户权限
func GetUserPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var user models.UserPermission
	if database.GetDB().First(&user, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}

// CreateUserPermission 新增用户权限
func CreateUserPermission(c *gin.Context) {
	var req struct {
		Name            string `json:"name" binding:"required"`
		DepartmentID    uint   `json:"department_id" binding:"required"`
		PositionName    string `json:"position_name" binding:"required"`
		SystemRolesJSON string `json:"system_roles_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写必要信息", "error": err.Error()})
		return
	}

	// 调试日志
	fmt.Printf("创建用户请求: Name=%s, DepartmentID=%d, PositionName=%s\n", req.Name, req.DepartmentID, req.PositionName)

	// 检查部门是否存在
	var dept models.Department
	if err := database.GetDB().First(&dept, req.DepartmentID).Error; err != nil {
		fmt.Printf("查找部门失败: DepartmentID=%d, Error=%v\n", req.DepartmentID, err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("部门不存在 (ID=%d)", req.DepartmentID)})
		return
	}
	fmt.Printf("找到部门: ID=%d, Name=%s\n", dept.ID, dept.Name)

	// 检查岗位是否属于该部门
	var pos models.DepartmentPosition
	if database.GetDB().Where("department_id = ? AND position_name = ?", req.DepartmentID, req.PositionName).First(&pos).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该岗位不属于此部门"})
		return
	}

	// 验证用户选择的角色是否符合岗位权限规则
	if err := validateUserRoles(req.PositionName, req.SystemRolesJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 检查同部门同姓名是否已存在
	var existing models.UserPermission
	if database.GetDB().Where("name = ? AND department_id = ?", req.Name, req.DepartmentID).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该部门下已有同名用户"})
		return
	}

	user := models.UserPermission{
		Name:            req.Name,
		DepartmentID:    req.DepartmentID,
		PositionName:    req.PositionName,
		SystemRolesJSON: req.SystemRolesJSON,
	}
	if err := database.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user, "message": "创建成功"})
}

// UpdateUserPermission 编辑用户权限
func UpdateUserPermission(c *gin.Context) {
	idStr := c.Param("id")
	fmt.Printf("更新用户请求: ID参数=%s\n", idStr)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Printf("解析ID失败: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var user models.UserPermission
	if database.GetDB().First(&user, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	var req struct {
		Name            string `json:"name" binding:"required"`
		DepartmentID    uint   `json:"department_id" binding:"required"`
		PositionName    string `json:"position_name" binding:"required"`
		SystemRolesJSON string `json:"system_roles_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写必要信息"})
		return
	}

	// 检查部门是否存在
	var dept models.Department
	if database.GetDB().First(&dept, req.DepartmentID).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部门不存在"})
		return
	}

	// 检查岗位是否属于该部门
	var pos models.DepartmentPosition
	if database.GetDB().Where("department_id = ? AND position_name = ?", req.DepartmentID, req.PositionName).First(&pos).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该岗位不属于此部门"})
		return
	}

	// 验证用户选择的角色是否符合岗位权限规则
	if err := validateUserRoles(req.PositionName, req.SystemRolesJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 检查同部门同姓名是否与其他用户冲突
	var existing models.UserPermission
	if database.GetDB().Where("name = ? AND department_id = ? AND id != ?", req.Name, req.DepartmentID, id).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该部门下已有同名用户"})
		return
	}

	user.Name = req.Name
	user.DepartmentID = req.DepartmentID
	user.PositionName = req.PositionName
	user.SystemRolesJSON = req.SystemRolesJSON

	if err := database.GetDB().Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user, "message": "更新成功"})
}

// DeleteUserPermission 删除用户权限
func DeleteUserPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	if err := database.GetDB().Delete(&models.UserPermission{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
