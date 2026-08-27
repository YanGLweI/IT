package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// ListAssets 获取资产列表（支持分页和排序）
func ListAssets(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 排序参数
	sortBy := c.DefaultQuery("sort_by", "id")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// 允许的排序字段
	allowedSort := map[string]bool{
		"id": true, "computer_name": true, "ip_address": true,
		"os_type_id": true, "purpose": true, "asset_level": true,
		"created_at": true,
	}
	if !allowedSort[sortBy] {
		sortBy = "id"
	}
	if strings.ToLower(sortOrder) != "asc" {
		sortOrder = "desc"
	}

	query := database.GetDB().Model(&models.Asset{}).Preload("Region").Preload("OSType").Preload("HostAsset")

	// 支持按区域过滤
	regionID := c.Query("region_id")
	if regionID != "" {
		query = query.Where("region_id = ?", regionID)
	}

	// 支持按计算机名或IP地址模糊搜索（计算机名不区分大小写）
	search := strings.TrimSpace(c.Query("search"))
	if search != "" {
		searchLower := strings.ToLower(search)
		query = query.Where("LOWER(computer_name) LIKE ? OR ip_address LIKE ?",
			"%"+searchLower+"%", "%"+search+"%")
	}

	// 查询总数
	var total int64
	query.Count(&total)

	// 分页查询
	var assets []models.Asset
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	if err := query.Order(orderClause).Offset(offset).Limit(pageSize).Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"data":      assets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAsset 获取单个资产
func GetAsset(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var asset models.Asset
	if err := database.GetDB().Preload("Region").Preload("OSType").Preload("HostAsset").First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": asset})
}

// ListVirtualHosts 获取虚拟主机列表（操作系统名含 ESXI 的资产）
func ListVirtualHosts(c *gin.Context) {
	var assets []models.Asset
	err := database.GetDB().
		Joins("JOIN os_types ON assets.os_type_id = os_types.id").
		Where("UPPER(os_types.name) LIKE '%ESXI%'").
		Order("computer_name ASC").
		Find(&assets).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": assets})
}

// boolText 将布尔值转为日志展示文本
func boolText(v bool) string {
	if v {
		return "是"
	}
	return "否"
}

// hostAssetName 查询虚拟主机计算机名，无则返回空字符串
func hostAssetName(hostID *uint) string {
	if hostID == nil || *hostID == 0 {
		return ""
	}
	var host models.Asset
	if err := database.GetDB().Select("computer_name").First(&host, *hostID).Error; err != nil {
		return ""
	}
	return host.ComputerName
}

// CreateAsset 创建资产
func CreateAsset(c *gin.Context) {
	var input struct {
		ComputerName     string `json:"computer_name" binding:"required"`
		RegionID         uint   `json:"region_id" binding:"required"`
		IPAddress        string `json:"ip_address"`
		OSTypeID         uint   `json:"os_type_id" binding:"required"`
		Purpose          string `json:"purpose"`
		AssetLevel       string `json:"asset_level"`
		IsVirtualMachine bool   `json:"is_virtual_machine"`
		HostAssetID      *uint  `json:"host_asset_id"`
		Remark           string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 虚拟机必须选择所属虚拟主机，非虚拟机强制清空
	if input.IsVirtualMachine {
		if input.HostAssetID == nil || *input.HostAssetID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "虚拟机必须选择所属虚拟主机"})
			return
		}
	} else {
		input.HostAssetID = nil
	}

	asset := models.Asset{
		ComputerName:     input.ComputerName,
		RegionID:         input.RegionID,
		IPAddress:        input.IPAddress,
		OSTypeID:         input.OSTypeID,
		Purpose:          input.Purpose,
		AssetLevel:       input.AssetLevel,
		IsVirtualMachine: input.IsVirtualMachine,
		HostAssetID:      input.HostAssetID,
		Remark:           input.Remark,
	}

	if err := database.GetDB().Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}

	// 重新查询以获取关联的区域和操作系统信息
	database.GetDB().Preload("Region").Preload("OSType").Preload("HostAsset").First(&asset, asset.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	osTypeName := ""
	if asset.OSType.ID > 0 {
		osTypeName = asset.OSType.Name
	}
	details := []services.LogDetail{
		{FieldName: "ComputerName", FieldLabel: "计算机名", NewValue: asset.ComputerName},
		{FieldName: "RegionID", FieldLabel: "区域ID", NewValue: fmt.Sprintf("%d", asset.RegionID)},
		{FieldName: "IPAddress", FieldLabel: "IP地址", NewValue: asset.IPAddress},
		{FieldName: "OSTypeID", FieldLabel: "操作系统", NewValue: osTypeName},
		{FieldName: "Purpose", FieldLabel: "用途", NewValue: asset.Purpose},
		{FieldName: "AssetLevel", FieldLabel: "资产等级", NewValue: asset.AssetLevel},
		{FieldName: "IsVirtualMachine", FieldLabel: "是否虚拟机", NewValue: boolText(asset.IsVirtualMachine)},
		{FieldName: "HostAssetID", FieldLabel: "所属虚拟主机", NewValue: hostAssetName(asset.HostAssetID)},
		{FieldName: "Remark", FieldLabel: "备注", NewValue: asset.Remark},
	}
	services.LogOperation(username, displayName, "创建资产", "asset", asset.ID, asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": asset})
}

// UpdateAsset 更新资产
func UpdateAsset(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var asset models.Asset
	if err := database.GetDB().Preload("OSType").First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}

	// 保存旧值快照
	oldAsset := asset

	var input struct {
		ComputerName     string `json:"computer_name" binding:"required"`
		RegionID         uint   `json:"region_id" binding:"required"`
		IPAddress        string `json:"ip_address"`
		OSTypeID         uint   `json:"os_type_id" binding:"required"`
		Purpose          string `json:"purpose"`
		AssetLevel       string `json:"asset_level"`
		IsVirtualMachine bool   `json:"is_virtual_machine"`
		HostAssetID      *uint  `json:"host_asset_id"`
		Remark           string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 虚拟机必须选择所属虚拟主机，非虚拟机强制清空
	if input.IsVirtualMachine {
		if input.HostAssetID == nil || *input.HostAssetID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "虚拟机必须选择所属虚拟主机"})
			return
		}
	} else {
		input.HostAssetID = nil
	}

	asset.ComputerName = input.ComputerName
	asset.RegionID = input.RegionID
	asset.IPAddress = input.IPAddress
	asset.OSTypeID = input.OSTypeID
	asset.OSType = models.OSType{} // 清空预加载的关联对象，避免 Save 时 GORM 将 os_type_id 覆盖回旧值
	asset.Purpose = input.Purpose
	asset.AssetLevel = input.AssetLevel
	asset.IsVirtualMachine = input.IsVirtualMachine
	asset.HostAssetID = input.HostAssetID
	asset.HostAsset = nil // 清空预加载的关联对象，避免 Save 时 GORM 覆盖 host_asset_id
	asset.Remark = input.Remark

	if err := database.GetDB().Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 重新查询以获取关联的区域和操作系统信息
	database.GetDB().Preload("Region").Preload("OSType").Preload("HostAsset").First(&asset, asset.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("asset")
	details := services.DiffStructs(oldAsset, asset, fieldLabels)
	services.LogOperation(username, displayName, "更新资产", "asset", asset.ID, asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": asset})
}

// DeleteAsset 删除资产
func DeleteAsset(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var asset models.Asset
	if err := database.GetDB().First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}

	if err := database.GetDB().Unscoped().Delete(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("asset")
	details := services.DiffStructs(asset, models.Asset{}, fieldLabels)
	services.LogOperation(username, displayName, "删除资产", "asset", asset.ID, asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
