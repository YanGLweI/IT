package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

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
