package models

import "time"

// DepartmentPosition 部门岗位关联
type DepartmentPosition struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DepartmentID uint      `gorm:"index;not null" json:"department_id"`
	PositionName string    `gorm:"size:100;not null" json:"position_name"`
	CreatedAt    time.Time `json:"created_at"`
}
