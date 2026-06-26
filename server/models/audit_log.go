package models

import "time"

// LoginLog 登录日志
type LoginLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"type:varchar(100);index" json:"username"`
	DisplayName string    `gorm:"type:varchar(100)" json:"display_name"`
	Action      string    `gorm:"type:varchar(30)" json:"action"` // login_success / login_failure / logout
	IPAddress   string    `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent   string    `gorm:"type:varchar(500)" json:"user_agent"`
	Detail      string    `gorm:"type:varchar(500)" json:"detail"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}

// OperationLog 操作日志主表
type OperationLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(100);index" json:"username"`
	DisplayName  string    `gorm:"type:varchar(100)" json:"display_name"`
	Action       string    `gorm:"type:varchar(50);index" json:"action"`
	ResourceType string    `gorm:"type:varchar(50);index" json:"resource_type"`
	ResourceID   uint      `gorm:"index" json:"resource_id"`
	ResourceName string    `gorm:"type:varchar(200)" json:"resource_name"`
	Approver     string    `gorm:"type:varchar(100)" json:"approver"`
	IPAddress    string    `gorm:"type:varchar(50)" json:"ip_address"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}

// OperationLogDetail 操作日志明细
type OperationLogDetail struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OperationLogID uint      `gorm:"index" json:"operation_log_id"`
	FieldName      string    `gorm:"type:varchar(100)" json:"field_name"`
	FieldLabel     string    `gorm:"type:varchar(100)" json:"field_label"`
	OldValue       string    `gorm:"type:text" json:"old_value"`
	NewValue       string    `gorm:"type:text" json:"new_value"`
	CreatedAt      time.Time `json:"created_at"`
}

func (OperationLogDetail) TableName() string {
	return "operation_log_details"
}
