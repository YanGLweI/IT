package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// firewallNodeVO 拓扑节点视图对象
type firewallNodeVO struct {
	ID        uint   `json:"id"`
	AssetID   uint   `json:"asset_id"`
	ParentID  uint   `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
	Remark    string `json:"remark"`
	Invalid   bool   `json:"invalid"` // 资产已被删除
	Asset     struct {
		ID           uint   `json:"id"`
		ComputerName string `json:"computer_name"`
		IPAddress    string `json:"ip_address"`
	} `json:"asset"`
	Regions []firewallRegionVO `json:"regions"`
}

// firewallRegionVO 节点下挂靠的区域视图对象
type firewallRegionVO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Assets []struct {
		ID           uint   `json:"id"`
		ComputerName string `json:"computer_name"`
		IPAddress    string `json:"ip_address"`
	} `json:"assets"`
}

// GetFirewallTopologyTree 获取防火墙拓扑结构（扁平节点列表，含各节点挂靠区域及区域下资产）
func GetFirewallTopologyTree(c *gin.Context) {
	db := database.GetDB()

	var nodes []models.FirewallNode
	if err := db.Preload("Asset").Order("sort_order ASC, id ASC").Find(&nodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	var links []models.RegionFirewall
	if err := db.Preload("Region").Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 查询全部资产，按区域分组（用于展示各区域下的资产与IP）
	var assets []models.Asset
	if err := db.Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	assetsByRegion := make(map[uint][]models.Asset)
	for _, a := range assets {
		assetsByRegion[a.RegionID] = append(assetsByRegion[a.RegionID], a)
	}

	// 区域挂靠按节点ID分组
	linksByNode := make(map[uint][]models.RegionFirewall)
	for _, l := range links {
		linksByNode[l.FirewallNodeID] = append(linksByNode[l.FirewallNodeID], l)
	}

	result := make([]firewallNodeVO, 0, len(nodes))
	for _, n := range nodes {
		vo := firewallNodeVO{
			ID:        n.ID,
			AssetID:   n.AssetID,
			ParentID:  n.ParentID,
			SortOrder: n.SortOrder,
			Remark:    n.Remark,
			Regions:   []firewallRegionVO{},
		}
		if n.Asset.ID == 0 {
			// 资产已被删除，标记为无效节点
			vo.Invalid = true
		} else {
			vo.Asset.ID = n.Asset.ID
			vo.Asset.ComputerName = n.Asset.ComputerName
			vo.Asset.IPAddress = n.Asset.IPAddress
		}
		for _, l := range linksByNode[n.ID] {
			rv := firewallRegionVO{ID: l.RegionID, Name: l.Region.Name, Assets: []struct {
				ID           uint   `json:"id"`
				ComputerName string `json:"computer_name"`
				IPAddress    string `json:"ip_address"`
			}{}}
			for _, a := range assetsByRegion[l.RegionID] {
				rv.Assets = append(rv.Assets, struct {
					ID           uint   `json:"id"`
					ComputerName string `json:"computer_name"`
					IPAddress    string `json:"ip_address"`
				}{ID: a.ID, ComputerName: a.ComputerName, IPAddress: a.IPAddress})
			}
			vo.Regions = append(vo.Regions, rv)
		}
		result = append(result, vo)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// getFirewallRegionID 获取名为"Firewall"的区域ID（防火墙设备所在区域）
func getFirewallRegionID() (uint, bool) {
	var region models.Region
	if err := database.GetDB().Where("name = ?", "Firewall").First(&region).Error; err != nil {
		return 0, false
	}
	return region.ID, true
}

// CreateFirewallNode 新增防火墙拓扑节点
func CreateFirewallNode(c *gin.Context) {
	var input struct {
		AssetID   uint   `json:"asset_id" binding:"required"`
		ParentID  uint   `json:"parent_id"`
		SortOrder int    `json:"sort_order"`
		Remark    string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	db := database.GetDB()

	// 校验资产存在且属于"Firewall"区域
	var asset models.Asset
	if err := db.First(&asset, input.AssetID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "所选防火墙设备不存在"})
		return
	}
	fwRegionID, ok := getFirewallRegionID()
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未找到\"Firewall\"区域，请先在区域管理中创建"})
		return
	}
	if asset.RegionID != fwRegionID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "所选设备不在\"Firewall\"区域下"})
		return
	}

	// 校验未重复建节点
	var count int64
	db.Model(&models.FirewallNode{}).Where("asset_id = ?", input.AssetID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该防火墙设备已建立拓扑节点"})
		return
	}

	// 校验上级节点存在
	if input.ParentID != 0 {
		var parent models.FirewallNode
		if err := db.First(&parent, input.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "上级防火墙节点不存在"})
			return
		}
	}

	node := models.FirewallNode{
		AssetID:   input.AssetID,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
		Remark:    input.Remark,
	}
	if err := db.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "AssetID", FieldLabel: "防火墙设备", NewValue: asset.ComputerName},
		{FieldName: "ParentID", FieldLabel: "上级防火墙", NewValue: strconv.Itoa(int(input.ParentID))},
		{FieldName: "SortOrder", FieldLabel: "排序", NewValue: strconv.Itoa(input.SortOrder)},
		{FieldName: "Remark", FieldLabel: "备注", NewValue: input.Remark},
	}
	services.LogOperation(username, displayName, "新增防火墙拓扑节点", "firewall_topology", node.ID, asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": node})
}

// isDescendantOf 判断 candidateID 是否为 nodeID 的后代（含自身），用于环检测
func isDescendantOf(nodeID, candidateID uint, nodes map[uint]uint) bool {
	cur := candidateID
	for cur != 0 {
		if cur == nodeID {
			return true
		}
		parent, ok := nodes[cur]
		if !ok {
			return false
		}
		cur = parent
	}
	return false
}

// UpdateFirewallNode 更新防火墙拓扑节点（调整上级/排序/备注）
func UpdateFirewallNode(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db := database.GetDB()

	var node models.FirewallNode
	if err := db.Preload("Asset").First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "节点不存在"})
		return
	}
	oldNode := node

	var input struct {
		ParentID  uint   `json:"parent_id"`
		SortOrder int    `json:"sort_order"`
		Remark    string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if input.ParentID != 0 {
		if input.ParentID == node.ID {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "上级防火墙不能选择自身"})
			return
		}
		// 加载全部节点的父子关系做环检测
		var all []models.FirewallNode
		db.Select("id, parent_id").Find(&all)
		parentMap := make(map[uint]uint, len(all))
		for _, n := range all {
			parentMap[n.ID] = n.ParentID
		}
		if _, ok := parentMap[input.ParentID]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "上级防火墙节点不存在"})
			return
		}
		if isDescendantOf(node.ID, input.ParentID, parentMap) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能将下级防火墙设置为上级，会形成环路"})
			return
		}
	}

	node.ParentID = input.ParentID
	node.SortOrder = input.SortOrder
	node.Remark = input.Remark

	if err := db.Save(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := services.DiffStructs(oldNode, node, services.GetFieldLabels("firewall_topology"))
	services.LogOperation(username, displayName, "更新防火墙拓扑节点", "firewall_topology", node.ID, node.Asset.ComputerName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": node})
}

// DeleteFirewallNode 删除防火墙拓扑节点
func DeleteFirewallNode(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db := database.GetDB()

	var node models.FirewallNode
	if err := db.Preload("Asset").First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "节点不存在"})
		return
	}

	// 存在子节点时拒绝删除
	var childCount int64
	db.Model(&models.FirewallNode{}).Where("parent_id = ?", node.ID).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该节点下存在子防火墙，请先迁移子节点"})
		return
	}

	nodeName := node.Asset.ComputerName
	if nodeName == "" {
		nodeName = "无效节点#" + strconv.Itoa(int(node.ID))
	}

	// 删除该节点的所有区域挂靠记录
	if err := db.Where("firewall_node_id = ?", node.ID).Delete(&models.RegionFirewall{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	if err := db.Delete(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := services.DiffStructs(node, models.FirewallNode{}, services.GetFieldLabels("firewall_topology"))
	services.LogOperation(username, displayName, "删除防火墙拓扑节点", "firewall_topology", node.ID, nodeName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// SaveRegionFirewall 整体保存区域与防火墙的挂靠关系（全量替换）
func SaveRegionFirewall(c *gin.Context) {
	var input struct {
		Items []struct {
			RegionID       uint `json:"region_id" binding:"required"`
			FirewallNodeID uint `json:"firewall_node_id"` // 0 = 取消挂靠
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	db := database.GetDB()

	// 校验所有防火墙节点存在
	nodeIDs := make(map[uint]bool)
	for _, item := range input.Items {
		if item.FirewallNodeID != 0 {
			nodeIDs[item.FirewallNodeID] = true
		}
	}
	for nodeID := range nodeIDs {
		var n models.FirewallNode
		if err := db.First(&n, nodeID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "防火墙节点不存在(ID:" + strconv.Itoa(int(nodeID)) + ")"})
			return
		}
	}

	// 记录旧关系用于日志
	var oldLinks []models.RegionFirewall
	db.Preload("Region").Find(&oldLinks)
	oldMap := make(map[uint]uint) // region_id -> firewall_node_id
	for _, l := range oldLinks {
		oldMap[l.RegionID] = l.FirewallNodeID
	}

	// 事务内全量替换
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&models.RegionFirewall{}).Error; err != nil {
			return err
		}
		for _, item := range input.Items {
			if item.FirewallNodeID == 0 {
				continue
			}
			link := models.RegionFirewall{RegionID: item.RegionID, FirewallNodeID: item.FirewallNodeID}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败"})
		return
	}

	// 记录操作日志：输出发生变更的区域
	username, displayName, approver := services.GetUserContext(c)
	regionNames := make(map[uint]string)
	var regions []models.Region
	db.Find(&regions)
	for _, r := range regions {
		regionNames[r.ID] = r.Name
	}
	var details []services.LogDetail
	newMap := make(map[uint]uint)
	for _, item := range input.Items {
		newMap[item.RegionID] = item.FirewallNodeID
	}
	for _, item := range input.Items {
		if oldMap[item.RegionID] != item.FirewallNodeID {
			details = append(details, services.LogDetail{
				FieldName:  "RegionID",
				FieldLabel: "区域",
				OldValue:   regionNames[item.RegionID] + " -> 节点" + strconv.Itoa(int(oldMap[item.RegionID])),
				NewValue:   regionNames[item.RegionID] + " -> 节点" + strconv.Itoa(int(item.FirewallNodeID)),
			})
		}
	}
	services.LogOperation(username, displayName, "保存区域防火墙分配", "firewall_topology", 0, "区域分配", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "保存成功"})
}
