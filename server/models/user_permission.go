package models

import "time"

// UserPermission 用户权限
type UserPermission struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	DepartmentID    uint      `gorm:"index;not null" json:"department_id"`
	PositionName    string    `gorm:"size:100;not null" json:"position_name"`
	SystemRolesJSON string    `gorm:"type:text" json:"system_roles_json"`
	// 格式: [{"system":"防火墙","roles":["admin","viewer"]},{"system":"Windows域控","roles":["Domain Admin"]}]
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
