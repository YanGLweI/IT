package models

import (
	"time"
	"gorm.io/gorm"
)

// Asset IT资产模型
type Asset struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ComputerName   string `gorm:"type:varchar(100);not null" json:"computer_name" binding:"required"`
	RegionID       uint   `gorm:"not null" json:"region_id" binding:"required"`
	Region         Region `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	IPAddress      string `gorm:"type:varchar(50)" json:"ip_address"`
	OSType         string `gorm:"type:varchar(50);not null" json:"os_type" binding:"required"`
	Purpose        string `gorm:"type:varchar(200)" json:"purpose"`
	AssetLevel     string `gorm:"type:varchar(20)" json:"asset_level"`
	Status         string `gorm:"type:varchar(20);default:在用" json:"status"`
	Remark         string `gorm:"type:varchar(500)" json:"remark"`
}

// TableName 指定表名
func (Asset) TableName() string {
	return "assets"
}
