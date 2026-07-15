package models

import "time"

// ITGuide IT指南主表
type ITGuide struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	GuideType   string     `gorm:"type:varchar(20);not null" json:"guide_type"` // "step" | "video"
	Category    string     `gorm:"type:varchar(100)" json:"category"`
	SortOrder   int        `gorm:"type:int;default:0" json:"sort_order"`
	IsPublished bool       `gorm:"type:tinyint(1);default:0" json:"is_published"`
	PublishedAt *time.Time `json:"published_at"`
}

// TableName 指定表名
func (ITGuide) TableName() string {
	return "it_guides"
}

// ITGuideStep 步骤指南-步骤表（仅 guide_type=step 时使用）
type ITGuideStep struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	GuideID     uint   `gorm:"type:int unsigned;not null;index" json:"guide_id"`
	StepNumber  int    `gorm:"type:int;not null" json:"step_number"`
	Title       string `gorm:"type:varchar(200)" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	SortOrder   int    `gorm:"type:int;default:0" json:"sort_order"`
}

// TableName 指定表名
func (ITGuideStep) TableName() string {
	return "it_guide_steps"
}

// ITGuideMedia 指南媒体资源表（图片/视频）
type ITGuideMedia struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	GuideID   uint   `gorm:"type:int unsigned;not null;index" json:"guide_id"`
	StepID    uint   `gorm:"type:int unsigned;default:0;index" json:"step_id"` // 0 表示属于指南本身（视频指南）
	MediaType string `gorm:"type:varchar(20);not null" json:"media_type"`      // "image" | "video"
	FileName  string `gorm:"type:varchar(300)" json:"file_name"`
	FilePath  string `gorm:"type:varchar(500)" json:"file_path"`
	FileSize  int64  `gorm:"type:bigint" json:"file_size"`
	FileType  string `gorm:"type:varchar(255)" json:"file_type"`
	EmbedURL  string `gorm:"type:varchar(500)" json:"embed_url"` // 嵌入视频URL（如B站iframe src）
	SortOrder int    `gorm:"type:int;default:0" json:"sort_order"`
}

// TableName 指定表名
func (ITGuideMedia) TableName() string {
	return "it_guide_media"
}
