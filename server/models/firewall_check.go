package models

import (
	"time"

	"gorm.io/gorm"
)

// FirewallCheck 防火墙检查记录
type FirewallCheck struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 基础字段
	Year       int    `gorm:"type:int;not null" json:"year"`
	Quarter    int    `gorm:"type:int;not null" json:"quarter"`
	ReportDate string `gorm:"type:varchar(20)" json:"report_date"`

	// 关联资产（防火墙设备）
	AssetID uint   `gorm:"not null" json:"asset_id"`
	Asset   Asset  `gorm:"foreignKey:AssetID" json:"asset,omitempty"`

	// 检查结果
	CheckResult string `gorm:"type:varchar(20);default:'compliant'" json:"check_result"` // compliant / non_compliant

	// 整改报告文件（不合规时上传）
	RectFileName string `gorm:"type:varchar(300)" json:"rect_file_name"`
	RectFilePath string `gorm:"type:varchar(500)" json:"rect_file_path"`
	RectFileSize int64  `gorm:"type:bigint" json:"rect_file_size"`
	RectFileType string `gorm:"type:varchar(255)" json:"rect_file_type"`
}

// TableName 指定表名
func (FirewallCheck) TableName() string {
	return "firewall_checks"
}
