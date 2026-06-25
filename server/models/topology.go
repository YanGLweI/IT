package models

import (
	"time"
	"gorm.io/gorm"
)

// Topology 网络拓扑图模型
type Topology struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string `gorm:"type:varchar(200);not null" json:"name" binding:"required"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	FileName    string `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath    string `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64  `gorm:"type:bigint" json:"file_size"`
}

// TableName 指定表名
func (Topology) TableName() string {
	return "topologies"
}
