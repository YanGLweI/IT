package models

import (
	"time"

	"gorm.io/gorm"
)

// Calendar 日程主表
type Calendar struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Title          string         `gorm:"type:varchar(200);not null" json:"title"`
	Description    string         `gorm:"type:text" json:"description"`
	StartTime      time.Time      `gorm:"type:datetime;not null" json:"start_time"`
	EndTime        time.Time      `gorm:"type:datetime;not null" json:"end_time"`
	IsAllDay       bool           `gorm:"type:boolean;default:false" json:"is_all_day"`
	RepeatRuleJSON string         `gorm:"type:json" json:"repeat_rule_json"`
	CreatedBy      string         `gorm:"type:varchar(100);not null;index:idx_created_by" json:"created_by"`

	Participants []CalendarParticipant `gorm:"foreignKey:CalendarID" json:"participants,omitempty"`
}

// TableName 指定表名
func (Calendar) TableName() string {
	return "calendars"
}

// CalendarParticipant 日程参与者关联表
type CalendarParticipant struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CalendarID  uint      `gorm:"type:int;not null;index:idx_calendar_id" json:"calendar_id"`
	UserDN      string    `gorm:"type:varchar(500);not null;index:idx_user_dn" json:"user_dn"`
	DisplayName string    `gorm:"type:varchar(100)" json:"display_name"`
	NotifiedAt  *time.Time `gorm:"type:datetime" json:"notified_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (CalendarParticipant) TableName() string {
	return "calendar_participants"
}

// CalendarNotification 日程通知记录表（消息中心持久化）
type CalendarNotification struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	CalendarID  uint       `gorm:"type:int;not null;index:idx_calendar_id" json:"calendar_id"`
	UserDN      string     `gorm:"type:varchar(500);not null;index:idx_user_date" json:"user_dn"`
	NotifyType  string     `gorm:"type:varchar(20);default:'login'" json:"notify_type"`
	NotifyTime  time.Time  `gorm:"type:datetime;not null;index:idx_notify_time" json:"notify_time"`
	SentAt      time.Time  `gorm:"type:datetime;not null" json:"sent_at"`
	ReadAt      *time.Time `gorm:"type:datetime" json:"read_at,omitempty"`
	PopupShown  bool       `gorm:"type:boolean;default:false" json:"popup_shown"`
	CreatedAt   time.Time  `json:"created_at"`

	// 关联字段（不映射到数据库）
	CalendarTitle string `gorm:"-" json:"calendar_title,omitempty"`
	StartTime     *time.Time `gorm:"-" json:"start_time,omitempty"`
	EndTime       *time.Time `gorm:"-" json:"end_time,omitempty"`
	IsAllDay      bool   `gorm:"-" json:"is_all_day,omitempty"`
}

// TableName 指定表名
func (CalendarNotification) TableName() string {
	return "calendar_notifications"
}
