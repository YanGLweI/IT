package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// normalizeSubnet 校验并规范化子网为 CIDR 格式（如 192.168.1.0/24），空串表示未设置
func normalizeSubnet(subnet string) (string, error) {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		return "", nil
	}
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil || ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("invalid subnet")
	}
	// 统一使用网络地址（如输入 192.168.1.5/24 存为 192.168.1.0/24）
	ones, _ := ipnet.Mask.Size()
	return fmt.Sprintf("%s/%d", ipnet.IP.String(), ones), nil
}

// ListRegions 获取区域列表
func ListRegions(c *gin.Context) {
	var regions []models.Region
	if err := database.GetDB().Order("sort_order ASC, id ASC").Find(&regions).Error; err != nil {
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

	// 清理名称前后空格
	region.Name = strings.TrimSpace(region.Name)
	region.Description = strings.TrimSpace(region.Description)

	// 校验并规范化子网
	subnet, err := normalizeSubnet(region.Subnet)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "子网格式不正确"})
		return
	}
	region.Subnet = subnet

	// 清理软删除的同名记录（避免唯一索引冲突）
	var softDeleted models.Region
	result := database.GetDB().Unscoped().Where("name = ?", region.Name).Find(&softDeleted)
	if result.Error == nil && result.RowsAffected > 0 {
		database.GetDB().Unscoped().Delete(&softDeleted)
	}

	// 新区域排到末尾
	var maxOrder int
	database.GetDB().Model(&models.Region{}).Select("COALESCE(MAX(sort_order), 0)").Scan(&maxOrder)
	region.SortOrder = maxOrder + 1

	if err := database.GetDB().Create(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	networkLevelText := map[int]string{0: "未设置", 1: "一级（互联网）", 2: "二级", 3: "三级", 4: "四级", 5: "五级（高安全区）"}
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "名称", NewValue: region.Name},
		{FieldName: "Description", FieldLabel: "描述", NewValue: region.Description},
		{FieldName: "NetworkLevel", FieldLabel: "网络等级", NewValue: networkLevelText[region.NetworkLevel]},
		{FieldName: "Subnet", FieldLabel: "子网", NewValue: region.Subnet},
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
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		NetworkLevel int    `json:"network_level"`
		Subnet       string `json:"subnet"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	region.Name = strings.TrimSpace(input.Name)
	region.Description = strings.TrimSpace(input.Description)
	region.NetworkLevel = input.NetworkLevel

	// 校验并规范化子网
	subnet, err := normalizeSubnet(input.Subnet)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "子网格式不正确"})
		return
	}
	region.Subnet = subnet

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

// ReorderRegion 调整区域排序（上移/下移）
func ReorderRegion(c *gin.Context) {
	var req struct {
		ID        uint   `json:"id" binding:"required"`
		Direction string `json:"direction" binding:"required"` // "up" or "down"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var current models.Region
	if err := database.GetDB().First(&current, req.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "区域不存在"})
		return
	}

	// 查找相邻区域
	var adjacent models.Region
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

	// 记录操作日志
	username, displayName, _ := services.GetUserContext(c)
	moveText := "上移"
	if req.Direction == "down" {
		moveText = "下移"
	}
	details := []services.LogDetail{
		{FieldName: "SortOrder", FieldLabel: "排序", NewValue: moveText},
	}
	services.LogOperation(username, displayName, "调整区域排序", "region", current.ID, current.Name, "", c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "移动成功"})
}
