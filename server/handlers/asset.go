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
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 10 }

	// 排序参数
	sortBy := c.DefaultQuery("sort_by", "id")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// 允许的排序字段
	allowedSort := map[string]bool{
		"id": true, "computer_name": true, "ip_address": true,
		"os_type": true, "purpose": true, "asset_level": true,
		"status": true, "created_at": true,
	}
	if !allowedSort[sortBy] {
		sortBy = "id"
	}
	if strings.ToLower(sortOrder) != "asc" {
		sortOrder = "desc"
	}

	query := database.GetDB().Model(&models.Asset{}).Preload("Region")

	// 支持按区域过滤
	regionID := c.Query("region_id")
	if regionID != "" {
		query = query.Where("region_id = ?", regionID)
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
		"code": 200,
		"data": assets,
		"total": total,
		"page": page,
		"page_size": pageSize,
	})
}

// GetAsset 获取单个资产
func GetAsset(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var asset models.Asset
	if err := database.GetDB().Preload("Region").First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": asset})
}

// CreateAsset 创建资产
func CreateAsset(c *gin.Context) {
	var input struct {
		ComputerName string `json:"computer_name" binding:"required"`
		RegionID     uint   `json:"region_id" binding:"required"`
		IPAddress    string `json:"ip_address"`
		OSType       string `json:"os_type" binding:"required"`
		Purpose      string `json:"purpose"`
		AssetLevel   string `json:"asset_level"`
		Status       string `json:"status"`
		Remark       string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	asset := models.Asset{
		ComputerName: input.ComputerName,
		RegionID:     input.RegionID,
		IPAddress:    input.IPAddress,
		OSType:       input.OSType,
		Purpose:      input.Purpose,
		AssetLevel:   input.AssetLevel,
		Status:       input.Status,
		Remark:       input.Remark,
	}

	if err := database.GetDB().Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}

	// 重新查询以获取关联的区域信息
	database.GetDB().Preload("Region").First(&asset, asset.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "ComputerName", FieldLabel: "计算机名", NewValue: asset.ComputerName},
		{FieldName: "RegionID", FieldLabel: "区域ID", NewValue: fmt.Sprintf("%d", asset.RegionID)},
		{FieldName: "IPAddress", FieldLabel: "IP地址", NewValue: asset.IPAddress},
		{FieldName: "OSType", FieldLabel: "操作系统", NewValue: asset.OSType},
		{FieldName: "Purpose", FieldLabel: "用途", NewValue: asset.Purpose},
		{FieldName: "AssetLevel", FieldLabel: "资产等级", NewValue: asset.AssetLevel},
		{FieldName: "Status", FieldLabel: "状态", NewValue: asset.Status},
		{FieldName: "Remark", FieldLabel: "备注", NewValue: asset.Remark},
	}
	services.LogOperation(username, displayName, "创建资产", "asset", asset.ID, asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": asset})
}

// UpdateAsset 更新资产
func UpdateAsset(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var asset models.Asset
	if err := database.GetDB().First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}

	// 保存旧值快照
	oldAsset := asset

	var input struct {
		ComputerName string `json:"computer_name" binding:"required"`
		RegionID     uint   `json:"region_id" binding:"required"`
		IPAddress    string `json:"ip_address"`
		OSType       string `json:"os_type" binding:"required"`
		Purpose      string `json:"purpose"`
		AssetLevel   string `json:"asset_level"`
		Status       string `json:"status"`
		Remark       string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	asset.ComputerName = input.ComputerName
	asset.RegionID = input.RegionID
	asset.IPAddress = input.IPAddress
	asset.OSType = input.OSType
	asset.Purpose = input.Purpose
	asset.AssetLevel = input.AssetLevel
	asset.Status = input.Status
	asset.Remark = input.Remark

	if err := database.GetDB().Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 重新查询以获取关联的区域信息
	database.GetDB().Preload("Region").First(&asset, asset.ID)

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
