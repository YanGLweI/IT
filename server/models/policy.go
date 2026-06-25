package models

import (
	"time"
	"gorm.io/gorm"
)

// Policy IT政策模型
type Policy struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title       string `gorm:"type:varchar(200);not null" json:"title" binding:"required"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	FileName    string `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath    string `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64  `gorm:"type:bigint" json:"file_size"`
	FileType    string `gorm:"type:varchar(255)" json:"file_type"`
}

// TableName 指定表名
func (Policy) TableName() string {
	return "policies"
}
