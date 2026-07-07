package models

import (
	"time"

	"gorm.io/gorm"
)

// BackupRecord 备份记录
type BackupRecord struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 申请日期
	ApplicationDate string `gorm:"type:varchar(20)" json:"application_date"`

	// 备份源（关联资产）
	BackupSourceAssetID uint  `gorm:"not null" json:"backup_source_asset_id"`
	BackupSourceAsset   Asset `gorm:"foreignKey:BackupSourceAssetID" json:"backup_source_asset,omitempty"`

	// 备份对象
	BackupTargetType string `gorm:"type:varchar(50)" json:"backup_target_type"` // "系统", "磁盘分区", "目录", "配置文件", "其他"
	BackupTarget     string `gorm:"type:varchar(200)" json:"backup_target"`

	// 备份工具
	BackupTool string `gorm:"type:varchar(100)" json:"backup_tool"` // "Veeam Backup", "FortiConfBak", "HuaweiConfBak"

	// 备份介质（关联资产）
	BackupMediumAssetID uint  `gorm:"not null" json:"backup_medium_asset_id"`
	BackupMediumAsset   Asset `gorm:"foreignKey:BackupMediumAssetID" json:"backup_medium_asset,omitempty"`

	// 备份频率（单选）
	BackupFrequency string `gorm:"type:varchar(50)" json:"backup_frequency"` // "每天", "每周", "每月"

	// 保留策略（单选，选"其它"时存储用户输入的自定义内容）
	RetentionPolicy string `gorm:"type:varchar(200)" json:"retention_policy"` // "7天", "15天", "1月", "3月", "1年", 或自定义内容

	// 全量备份策略（单选）
	FullBackupStrategy string `gorm:"type:varchar(50)" json:"full_backup_strategy"` // "每天", "每周"

	// 所属部门
	DepartmentID uint       `gorm:"not null" json:"department_id"`
	Department   Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`

	// 申请表文件
	FileName string `gorm:"type:varchar(300)" json:"file_name"`
	FilePath string `gorm:"type:varchar(500)" json:"file_path"`
	FileSize int64  `gorm:"type:bigint" json:"file_size"`
	FileType string `gorm:"type:varchar(255)" json:"file_type"`

	// 恢复记录（一对多）
	Recoveries []BackupRecovery `gorm:"foreignKey:BackupRecordID" json:"recoveries,omitempty"`
}

// TableName 指定表名
func (BackupRecord) TableName() string {
	return "backup_records"
}

// BackupRecovery 恢复还原记录
type BackupRecovery struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联备份记录
	BackupRecordID uint `gorm:"not null;index" json:"backup_record_id"`

	// 恢复类型
	RecoveryType string `gorm:"type:varchar(50)" json:"recovery_type"` // "恢复测试", "故障恢复"

	// 恢复结果
	RecoveryResult string `gorm:"type:varchar(20)" json:"recovery_result"` // "成功", "失败"

	// 恢复日期
	RecoveryDate string `gorm:"type:varchar(20)" json:"recovery_date"`

	// 上传记录文件
	FileName string `gorm:"type:varchar(300)" json:"file_name"`
	FilePath string `gorm:"type:varchar(500)" json:"file_path"`
	FileSize int64  `gorm:"type:bigint" json:"file_size"`
	FileType string `gorm:"type:varchar(255)" json:"file_type"`
}

// TableName 指定表名
func (BackupRecovery) TableName() string {
	return "backup_recoveries"
}
