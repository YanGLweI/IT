package models

import (
	"time"

	"gorm.io/gorm"
)

// OSType 操作系统类型模型
type OSType struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name" binding:"required"`
}

// TableName 指定表名
func (OSType) TableName() string {
	return "os_types"
}
