package models

import "time"

// ChangeType 变更类型
type ChangeType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	SortOrder int       `gorm:"type:int;default:0" json:"sort_order"`
}

// TableName 指定表名
func (ChangeType) TableName() string {
	return "change_types"
}
