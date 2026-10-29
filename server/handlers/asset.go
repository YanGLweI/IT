package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
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
// assets.id 上有 6 个 RESTRICT 外键（asset_software / firewall_nodes / firewall_checks /
// backup_records 源与介质 / assets 自引用），必须按子→父顺序在同一事务内清理，否则 MySQL 1451 会让删除失败。
func DeleteAsset(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db := database.GetDB()

	var asset models.Asset
	if err := db.First(&asset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}

	// 预检：存活的业务历史记录与下挂虚拟机不能随资产静默销毁，明确拒绝并告知去哪处理
	var vmCount, liveChecks, liveSrc, liveMed int64
	db.Model(&models.Asset{}).Where("host_asset_id = ?", id).Count(&vmCount)
	db.Unscoped().Model(&models.FirewallCheck{}).Where("asset_id = ? AND deleted_at IS NULL", id).Count(&liveChecks)
	db.Unscoped().Model(&models.BackupRecord{}).Where("backup_source_asset_id = ? AND deleted_at IS NULL", id).Count(&liveSrc)
	db.Unscoped().Model(&models.BackupRecord{}).Where("backup_medium_asset_id = ? AND deleted_at IS NULL", id).Count(&liveMed)
	if vmCount+liveChecks+liveSrc+liveMed > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf(
			"该资产仍有未处理的关联记录，无法删除：虚拟机 %d 台、防火墙检查记录 %d 条、备份记录（作为备份源 %d 条 / 作为备份介质 %d 条）。请先在资产管理、防火墙检查、备份记录模块迁移或删除。",
			vmCount, liveChecks, liveSrc, liveMed)})
		return
	}

	// 防火墙拓扑节点：节点本身随资产清理，但有子防火墙时不能删（否则子节点失去父级）
	var nodeIDs []uint
	db.Model(&models.FirewallNode{}).Where("asset_id = ?", id).Pluck("id", &nodeIDs)
	if len(nodeIDs) > 0 {
		var childNodeCount int64
		db.Model(&models.FirewallNode{}).Where("parent_id IN ? AND asset_id <> ?", nodeIDs, id).Count(&childNodeCount)
		if childNodeCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf(
				"该资产是防火墙拓扑节点，其下还有 %d 个子防火墙，请先在防火墙拓扑模块迁移子节点。", childNodeCount)})
			return
		}
	}

	// 软删残行必须 Unscoped 统计：普通查询看不到它们，但它们仍占用外键，会让资产永久删不掉
	var deadCheckCount, deadBackupCount int64
	db.Unscoped().Model(&models.FirewallCheck{}).Where("asset_id = ? AND deleted_at IS NOT NULL", id).Count(&deadCheckCount)
	db.Unscoped().Model(&models.BackupRecord{}).
		Where("(backup_source_asset_id = ? OR backup_medium_asset_id = ?) AND deleted_at IS NOT NULL", id, id).
		Count(&deadBackupCount)

	var softwareLinkCount int64
	db.Model(&models.AssetSoftware{}).Where("asset_id = ?", id).Count(&softwareLinkCount)

	deadBackupIDs := make([]uint, 0)
	if deadBackupCount > 0 {
		db.Unscoped().Model(&models.BackupRecord{}).
			Where("(backup_source_asset_id = ? OR backup_medium_asset_id = ?) AND deleted_at IS NOT NULL", id, id).
			Pluck("id", &deadBackupIDs)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1) 拓扑节点与区域挂靠：region_firewalls.firewall_node_id 没有数据库外键，
		//    不显式清理不会报错，只会静默留下指向已删节点的悬挂引用
		if len(nodeIDs) > 0 {
			if err := tx.Where("firewall_node_id IN ?", nodeIDs).Delete(&models.RegionFirewall{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("id IN ?", nodeIDs).Delete(&models.FirewallNode{}).Error; err != nil {
				return err
			}
		}

		// 2) 资产对应表的第三方软件关联
		if err := tx.Where("asset_id = ?", id).Delete(&models.AssetSoftware{}).Error; err != nil {
			return err
		}

		// 3) 已软删的防火墙检查记录
		if deadCheckCount > 0 {
			if err := tx.Unscoped().Where("asset_id = ? AND deleted_at IS NOT NULL", id).Delete(&models.FirewallCheck{}).Error; err != nil {
				return err
			}
		}

		// 4) 已软删的备份记录：先清恢复记录，backup_recoveries 对 backup_records.id 有 RESTRICT 外键
		if len(deadBackupIDs) > 0 {
			if err := tx.Unscoped().Where("backup_record_id IN ?", deadBackupIDs).Delete(&models.BackupRecovery{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("id IN ?", deadBackupIDs).Delete(&models.BackupRecord{}).Error; err != nil {
				return err
			}
		}

		// 5) 资产本体
		return tx.Unscoped().Delete(&asset).Error
	})
	if err != nil {
		log.Printf("删除资产失败 (id=%d, computer_name=%s): %v", id, asset.ComputerName, err)
		if isForeignKeyBlock(err) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf(
				"该资产仍存在外键保护的关联记录，未能自动清理。请检查：资产对应表、防火墙拓扑节点、备份记录（源/介质）、防火墙检查记录，或联系管理员。原始错误：%s", err.Error())})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("asset")
	details := services.DiffStructs(asset, models.Asset{}, fieldLabels)
	if softwareLinkCount+int64(len(nodeIDs))+deadCheckCount+deadBackupCount > 0 {
		details = append(details, services.LogDetail{
			FieldName:  "CascadeCleanup",
			FieldLabel: "级联清理",
			NewValue: fmt.Sprintf("第三方软件关联 %d 条、防火墙拓扑节点 %d 个、已软删防火墙检查 %d 条、已软删备份记录 %d 条",
				softwareLinkCount, len(nodeIDs), deadCheckCount, deadBackupCount),
		})
	}
	services.LogOperation(username, displayName, "删除资产", "asset", asset.ID, asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// isForeignKeyBlock 判定 MySQL 1451/1452：子表仍存在引用，RESTRICT 拒绝了父行删除
func isForeignKeyBlock(err error) bool {
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		return myErr.Number == 1451 || myErr.Number == 1452
	}
	msg := err.Error()
	return strings.Contains(msg, "Error 1451") || strings.Contains(msg, "Error 1452")
}
