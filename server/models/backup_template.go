package models

import "time"

// BackupTemplate 备份与恢复记录表模板版本
type BackupTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Version     string    `gorm:"type:varchar(50);not null" json:"version"`       // 版本号（如 IT03-1.0）
	Description string    `gorm:"type:varchar(500)" json:"description"`           // 版本说明
	FileName    string    `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath    string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64     `gorm:"type:bigint" json:"file_size"`
	FileType    string    `gorm:"type:varchar(255)" json:"file_type"`
	IsCurrent   bool      `gorm:"type:tinyint(1);default:0" json:"is_current"`    // 是否当前版本
}

// TableName 指定表名
func (BackupTemplate) TableName() string {
	return "backup_templates"
}
