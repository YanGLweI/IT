package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListDepartments 获取所有部门
func ListDepartments(c *gin.Context) {
	var departments []models.Department
	if err := database.GetDB().Order("sort_order asc").Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": departments})
}

// CreateDepartment 新增部门
func CreateDepartment(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部门名称不能为空"})
		return
	}

	// 检查是否已存在
	var existing models.Department
	if database.GetDB().Where("name = ?", req.Name).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部门名称已存在"})
		return
	}

	// 获取最大 sort_order
	var maxSort int64
	database.GetDB().Model(&models.Department{}).Select("COALESCE(MAX(sort_order),0)").Scan(&maxSort)

	dept := models.Department{
		Name:      req.Name,
		SortOrder: int(maxSort) + 1,
	}
	if err := database.GetDB().Create(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": dept, "message": "创建成功"})
}

// UpdateDepartment 编辑部门
func UpdateDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	var dept models.Department
	if database.GetDB().First(&dept, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "部门不存在"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部门名称不能为空"})
		return
	}

	// 检查名称是否与其他部门冲突
	var existing models.Department
	if database.GetDB().Where("name = ? AND id != ?", req.Name, id).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部门名称已存在"})
		return
	}

	// 更新部门名称，同时更新关联的用户权限表中的岗位名称（如果有用户引用了旧岗位名）
	oldName := dept.Name
	dept.Name = req.Name
	if err := database.GetDB().Save(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	_ = oldName // 部门改名不影响用户的岗位字段（岗位来自 permission_rules，不是部门名）

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": dept, "message": "更新成功"})
}

// DeleteDepartment 删除部门
func DeleteDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	// 检查是否有关联用户
	var count int64
	database.GetDB().Model(&models.UserPermission{}).Where("department_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该部门下还有用户，无法删除"})
		return
	}

	// 删除部门岗位关联
	database.GetDB().Where("department_id = ?", id).Delete(&models.DepartmentPosition{})

	// 删除部门
	if err := database.GetDB().Delete(&models.Department{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ListDepartmentPositions 获取部门岗位列表
func ListDepartmentPositions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	var positions []models.DepartmentPosition
	database.GetDB().Where("department_id = ?", id).Find(&positions)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": positions})
}

// AddDepartmentPosition 添加部门岗位
func AddDepartmentPosition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	// 检查部门是否存在
	var dept models.Department
	if database.GetDB().First(&dept, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "部门不存在"})
		return
	}

	var req struct {
		PositionName string `json:"position_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "岗位名称不能为空"})
		return
	}

	// 检查岗位是否在权限规则中存在
	var rule models.PermissionRule
	if database.GetDB().Where("position_name = ?", req.PositionName).First(&rule).Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该岗位不在权限设置规则中"})
		return
	}

	// 检查是否已存在
	var existing models.DepartmentPosition
	if database.GetDB().Where("department_id = ? AND position_name = ?", id, req.PositionName).First(&existing).Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该岗位已添加到该部门"})
		return
	}

	pos := models.DepartmentPosition{
		DepartmentID: uint(id),
		PositionName: req.PositionName,
	}
	if err := database.GetDB().Create(&pos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": pos, "message": "添加成功"})
}

// RemoveDepartmentPosition 移除部门岗位
func RemoveDepartmentPosition(c *gin.Context) {
	pid, err := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的岗位ID"})
		return
	}

	if err := database.GetDB().Delete(&models.DepartmentPosition{}, pid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "移除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移除成功"})
}
