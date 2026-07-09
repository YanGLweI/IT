package models

import "time"

// 来源类型常量
const (
	SourceTypeUpload  = "upload"  // 直接上传的文件
	SourceTypeStatic  = "static"  // 引用其他模块的静态文件（快照复制）
	SourceTypeDynamic = "dynamic" // 引用其他模块的动态生成文件（实时生成）
)

// FormVaultItem 表单保管区条目
type FormVaultItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 基本信息
	Title       string `gorm:"type:varchar(200);not null" json:"title"`           // 表单标题（展示名）
	Description string `gorm:"type:varchar(1000)" json:"description"`              // 表单描述
	Category    string `gorm:"type:varchar(100)" json:"category"`                  // 分类（如：模板、政策、检查表）
	SortOrder   int    `gorm:"type:int;default:0" json:"sort_order"`               // 排序权重

	// 来源信息
	SourceType string `gorm:"type:varchar(20);not null;default:'upload'" json:"source_type"` // upload | static | dynamic

	// upload 类型字段
	FileName string `gorm:"type:varchar(300)" json:"file_name"` // 原始文件名
	FilePath string `gorm:"type:varchar(500)" json:"file_path"` // 服务器存储路径
	FileSize int64  `gorm:"type:bigint" json:"file_size"`       // 文件大小(bytes)
	FileType string `gorm:"type:varchar(255)" json:"file_type"` // MIME 类型

	// static/dynamic 引用字段
	RefModule    string `gorm:"type:varchar(100)" json:"ref_module"`     // 引用模块标识
	RefID        uint   `gorm:"type:int unsigned" json:"ref_id"`         // 引用记录的 ID
	RefHandler   string `gorm:"type:varchar(200)" json:"ref_handler"`    // 动态生成处理器名（仅 dynamic 类型）
	SnapshotPath string `gorm:"type:varchar(500)" json:"snapshot_path"`  // static 类型的快照文件路径

	// 发布状态
	IsPublished bool       `gorm:"type:tinyint(1);default:0" json:"is_published"` // 是否已发布
	PublishedAt *time.Time `json:"published_at"`                                  // 发布时间
}

// TableName 指定表名
func (FormVaultItem) TableName() string {
	return "form_vault_items"
}
