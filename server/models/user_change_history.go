package models

import (
	"time"

	"gorm.io/gorm"
)

// UserChangeHistory 用户变更记录历史
type UserChangeHistory struct {
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
func (UserChangeHistory) TableName() string {
	return "user_change_histories"
}
