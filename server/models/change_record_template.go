package models

import "time"

// ChangeRecordTemplate 变更记录表模板版本
type ChangeRecordTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Version     string    `gorm:"type:varchar(50);not null" json:"version"`       // 版本号（如 IT02-3.0）
	Description string    `gorm:"type:varchar(500)" json:"description"`           // 版本说明
	FileName    string    `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath    string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize    int64     `gorm:"type:bigint" json:"file_size"`
	FileType    string    `gorm:"type:varchar(255)" json:"file_type"`
	IsCurrent   bool      `gorm:"type:tinyint(1);default:0" json:"is_current"`    // 是否当前版本
}

// TableName 指定表名
func (ChangeRecordTemplate) TableName() string {
	return "change_record_templates"
}
