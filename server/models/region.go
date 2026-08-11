package models

import (
	"time"
	"gorm.io/gorm"
)

// Region 区域模型
type Region struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name         string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name" binding:"required"`
	Description  string `gorm:"type:varchar(500)" json:"description"`
	NetworkLevel int  `gorm:"type:int;default:0" json:"network_level"`
	Subnet       string `gorm:"type:varchar(50)" json:"subnet"` // 子网（CIDR 格式，如 192.168.1.0/24），可选
	SortOrder    int  `gorm:"default:0" json:"sort_order"`     // 排序序号
}

// TableName 指定表名
func (Region) TableName() string {
	return "regions"
}
