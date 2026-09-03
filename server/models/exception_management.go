package models

import (
	"time"

	"gorm.io/gorm"
)

// ExceptionManagement 例外管理 - 用于留档电脑补丁升级例外的授权文件
type ExceptionManagement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 基本信息
	ApplyDate string `gorm:"type:varchar(20)" json:"apply_date"`           // 申请日期
	Applicant string `gorm:"type:varchar(50);not null" json:"applicant"`     // 申请人
	Reason    string `gorm:"type:text" json:"reason"`                        // 例外情况说明
	EndDate   string `gorm:"type:varchar(20)" json:"end_date"`               // 持续到

	// 文件信息
	FileName  string `gorm:"type:varchar(300);not null" json:"file_name"`    // 文件名
	FilePath  string `gorm:"type:varchar(1000);not null" json:"file_path"`    // 文件路径
	FileSize  int64  `gorm:"type:bigint" json:"file_size"`                   // 文件大小
	FileType  string `gorm:"type:varchar(255)" json:"file_type"`             // 文件类型
}
