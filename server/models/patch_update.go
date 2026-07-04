package models

import (
	"time"

	"gorm.io/gorm"
)

// PatchUpdate 补丁更新（系统合规性报表）
type PatchUpdate struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 基础字段
	Year  int `gorm:"type:int;not null" json:"year"`
	Month int `gorm:"type:int;not null" json:"month"`

	TotalAssets        int    `gorm:"type:int;not null" json:"total_assets"`
	Compliance         string `gorm:"type:varchar(20);default:'compliant'" json:"compliance"` // compliant / non_compliant
	NonCompliantAssets int    `gorm:"type:int;default:0" json:"non_compliant_assets"`

	// 合规性报表文件
	FileName string `gorm:"type:varchar(300)" json:"file_name"`
	FilePath string `gorm:"type:varchar(500)" json:"file_path"`
	FileSize int64  `gorm:"type:bigint" json:"file_size"`
	FileType string `gorm:"type:varchar(255)" json:"file_type"`

	// 修复报表文件（不合规时上传）
	FixFileName string `gorm:"type:varchar(300)" json:"fix_file_name"`
	FixFilePath string `gorm:"type:varchar(500)" json:"fix_file_path"`
	FixFileSize int64  `gorm:"type:bigint" json:"fix_file_size"`
	FixFileType string `gorm:"type:varchar(255)" json:"fix_file_type"`
}

// TableName 指定表名
func (PatchUpdate) TableName() string {
	return "patch_updates"
}
