package models

import (
	"time"

	"gorm.io/gorm"
)

// PenetrationTest 渗透测试报告
type PenetrationTest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 基础字段
	TestType    string `gorm:"type:varchar(20);default:'internal'" json:"test_type"` // internal / external
	Year        int    `gorm:"type:int;not null" json:"year"`
	ReportDate  string `gorm:"type:varchar(20)" json:"report_date"`
	VulnCount   int    `gorm:"type:int;default:0" json:"vuln_count"`     // 可渗透漏洞数
	Description string `gorm:"type:text" json:"description"`             // 结果描述

	// 报告文件
	FileName string `gorm:"type:varchar(300);not null" json:"file_name"`
	FilePath string `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize int64  `gorm:"type:bigint" json:"file_size"`
	FileType string `gorm:"type:varchar(255)" json:"file_type"`

	// 关联漏洞扫描报告（多对多）
	VulnerabilityScans []VulnerabilityScan `gorm:"many2many:penetration_test_vulnerability_scans;" json:"vulnerability_scans"`
}

// TableName 指定表名
func (PenetrationTest) TableName() string {
	return "penetration_tests"
}
