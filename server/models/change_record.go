package models

import (
	"time"

	"gorm.io/gorm"
)

// ChangeRecord 变更记录扫描件
type ChangeRecord struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Year        int            `gorm:"type:int;not null" json:"year"`
	Month       int            `gorm:"type:int;not null" json:"month"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	FileName    string         `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath    string         `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64          `gorm:"type:bigint" json:"file_size"`
	FileType    string         `gorm:"type:varchar(255)" json:"file_type"`
}

// TableName 指定表名
func (ChangeRecord) TableName() string {
	return "change_records"
}
