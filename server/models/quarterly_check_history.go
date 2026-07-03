package models

import (
	"time"

	"gorm.io/gorm"
)

// QuarterlyCheckHistory 季度检查历史
type QuarterlyCheckHistory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Year        int            `gorm:"type:int;not null" json:"year"`
	Quarter     int            `gorm:"type:int;not null" json:"quarter"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	FileName    string         `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath    string         `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64          `gorm:"type:bigint" json:"file_size"`
	FileType    string         `gorm:"type:varchar(255)" json:"file_type"`

	SoftwareList []ApprovedSoftware `gorm:"-" json:"software_list,omitempty"`
}

// TableName 指定表名
func (QuarterlyCheckHistory) TableName() string {
	return "quarterly_check_histories"
}

// QuarterlyCheckSoftware 季度检查与核准软件关联
type QuarterlyCheckSoftware struct {
	ID                      uint                  `gorm:"primaryKey" json:"id"`
	CreatedAt               time.Time             `json:"created_at"`
	QuarterlyCheckHistoryID uint                  `gorm:"not null;index" json:"quarterly_check_history_id"`
	ApprovedSoftwareID      uint                  `gorm:"not null" json:"approved_software_id"`
	OriginalVersion         string                `gorm:"type:varchar(100)" json:"original_version"`
	QuarterlyCheckHistory   QuarterlyCheckHistory `gorm:"foreignKey:QuarterlyCheckHistoryID" json:"-"`
	ApprovedSoftware        ApprovedSoftware      `gorm:"foreignKey:ApprovedSoftwareID" json:"approved_software,omitempty"`
}

// TableName 指定表名
func (QuarterlyCheckSoftware) TableName() string {
	return "quarterly_check_software"
}
