package models

import (
	"time"

	"gorm.io/gorm"
)

// ChangeRecord 变更记录扫描件
type ChangeRecord struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Year         int            `gorm:"type:int;not null" json:"year"`
	Month        int            `gorm:"type:int;not null" json:"month"`
	Description  string         `gorm:"type:varchar(500)" json:"description"`
	ApplyDate    *time.Time     `gorm:"type:date" json:"apply_date"`
	ImplementDate *time.Time    `gorm:"type:date" json:"implement_date"`
	FileName     string         `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath     string         `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize     int64          `gorm:"type:bigint" json:"file_size"`
	FileType     string         `gorm:"type:varchar(255)" json:"file_type"`
	ChangeTypes  []ChangeType   `gorm:"many2many:change_record_change_types;" json:"change_types"`
}

// TableName 指定表名
func (ChangeRecord) TableName() string {
	return "change_records"
}
