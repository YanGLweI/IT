package models

import (
	"time"

	"gorm.io/gorm"
)

// SftpServer SFTP服务器配置
type SftpServer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name" binding:"required"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SftpServer) TableName() string {
	return "sftp_servers"
}

// SftpAccount SFTP账号
type SftpAccount struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ServerID        uint           `gorm:"index;not null" json:"server_id" binding:"required"`
	AccountName     string         `gorm:"type:varchar(100);not null" json:"account_name" binding:"required"`
	CreatedTime     string         `gorm:"type:varchar(50)" json:"created_time"`
	Validity        string         `gorm:"type:varchar(100)" json:"validity"`
	PermissionsJSON string         `gorm:"type:text" json:"permissions_json"`
	ContactPerson   string         `gorm:"type:varchar(100)" json:"contact_person"`
	Department      string         `gorm:"type:varchar(100)" json:"department"`
	WhitelistJSON   string         `gorm:"type:text" json:"whitelist_json"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (SftpAccount) TableName() string {
	return "sftp_accounts"
}
