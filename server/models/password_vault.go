package models

import "time"

// PasswordCategory 密码分类
type PasswordCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name      string `gorm:"type:varchar(50);not null;uniqueIndex" json:"name"`
	Icon      string `gorm:"type:varchar(50);not null" json:"icon"`
	IsPreset  bool   `gorm:"type:tinyint(1);default:0" json:"is_preset"`
	SortOrder int    `gorm:"type:int;default:0" json:"sort_order"`

	// 非数据库字段，用于返回条目数量
	EntryCount int64 `gorm:"-" json:"entry_count"`
}

func (PasswordCategory) TableName() string {
	return "password_categories"
}

// PasswordEntry 密码条目
type PasswordEntry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	CategoryID        uint   `gorm:"type:int unsigned;not null" json:"category_id"`
	Icon              string `gorm:"type:varchar(50);not null" json:"icon"`
	Name              string `gorm:"type:varchar(200);not null" json:"name"`
	Username          string `gorm:"type:varchar(200);not null" json:"username"`
	EncryptedPassword string `gorm:"type:text;not null" json:"-"` // AES加密密文，不返回前端
	URL               string `gorm:"type:varchar(500)" json:"url"`
	Port              int    `gorm:"type:int" json:"port"`
	Notes             string `gorm:"type:text" json:"notes"`
	IsStarred         bool   `gorm:"type:tinyint(1);default:0" json:"is_starred"`
	CreatedBy         string `gorm:"type:varchar(100);not null" json:"created_by"`
	UpdatedBy         string `gorm:"type:varchar(100);not null" json:"updated_by"`

	// 关联字段
	CategoryName string   `gorm:"-" json:"category_name"`
	Viewers      []string `gorm:"-" json:"viewers"`
}

func (PasswordEntry) TableName() string {
	return "password_entries"
}

// PasswordEntryViewer 密码查看授权
type PasswordEntryViewer struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	EntryID  uint   `gorm:"type:int unsigned;not null;uniqueIndex:idx_entry_user" json:"entry_id"`
	Username string `gorm:"type:varchar(100);not null;uniqueIndex:idx_entry_user" json:"username"`
}

func (PasswordEntryViewer) TableName() string {
	return "password_entry_viewers"
}

// PasswordViewLog 密码查看日志
type PasswordViewLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EntryID   uint      `gorm:"type:int unsigned;not null" json:"entry_id"`
	EntryName string    `gorm:"type:varchar(200);not null" json:"entry_name"`
	Viewer    string    `gorm:"type:varchar(100);not null" json:"viewer"`
	ViewedAt  time.Time `gorm:"not null" json:"viewed_at"`
}

func (PasswordViewLog) TableName() string {
	return "password_view_logs"
}
