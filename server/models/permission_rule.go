package models

import "time"

// PermissionRule 岗位权限设置规则
type PermissionRule struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PositionName string    `gorm:"size:100;uniqueIndex;not null" json:"position_name"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	RulesJSON    string    `gorm:"type:text" json:"rules_json"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
