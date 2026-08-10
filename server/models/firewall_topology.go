package models

import (
	"time"
)

// FirewallNode 防火墙拓扑节点（防火墙设备复用资产模块中"Firewall"区域下的资产）
// 采用硬删除（与 Asset 一致）：软删除会与 asset_id 唯一索引冲突，导致删除后无法重建节点
type FirewallNode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AssetID   uint      `gorm:"uniqueIndex;not null" json:"asset_id"` // 一个防火墙资产只建一次节点
	Asset     Asset     `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	ParentID  uint      `gorm:"default:0" json:"parent_id"` // 0 = 顶级防火墙
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Remark    string    `gorm:"type:varchar(500)" json:"remark"`
}

// TableName 指定表名
func (FirewallNode) TableName() string {
	return "firewall_nodes"
}

// RegionFirewall 区域→防火墙一对一挂靠关系
type RegionFirewall struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	RegionID       uint      `gorm:"uniqueIndex;not null" json:"region_id"` // 一个区域只挂靠一个防火墙
	Region         Region    `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	FirewallNodeID uint      `gorm:"not null;index" json:"firewall_node_id"`
}

// TableName 指定表名
func (RegionFirewall) TableName() string {
	return "region_firewalls"
}
