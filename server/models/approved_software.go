package models

import (
	"time"

	"gorm.io/gorm"
)

// ApprovedSoftware 核准软件目录
type ApprovedSoftware struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name           string `gorm:"type:varchar(200);not null" json:"name" binding:"required"`
	Version        string `gorm:"type:varchar(100)" json:"version"`
	LatestVersion  string `gorm:"type:varchar(100)" json:"latest_version"`
	NeedUpdate     bool   `gorm:"default:false" json:"need_update"`
	UpdateReason   string `gorm:"type:varchar(500)" json:"update_reason"`
	Vendor         string `gorm:"type:varchar(200)" json:"vendor"`
	VendorWebsite  string `gorm:"type:varchar(500)" json:"vendor_website"`
	LicenseType    string `gorm:"type:varchar(20);default:商用" json:"license_type"` // 商用 / 开源
	Purpose        string `gorm:"type:varchar(500)" json:"purpose"`
}

// TableName 指定表名
func (ApprovedSoftware) TableName() string {
	return "approved_software"
}

// AssetSoftware 资产与核准软件关联表
type AssetSoftware struct {
	ID                   uint              `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	AssetID              uint              `gorm:"not null;uniqueIndex:idx_asset_software" json:"asset_id"`
	ApprovedSoftwareID   uint              `gorm:"not null;uniqueIndex:idx_asset_software" json:"approved_software_id"`
	Asset                Asset             `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	ApprovedSoftware     ApprovedSoftware  `gorm:"foreignKey:ApprovedSoftwareID" json:"approved_software,omitempty"`
}

// TableName 指定表名
func (AssetSoftware) TableName() string {
	return "asset_software"
}
