package models

import "time"

// MenuFavorite 菜单收藏（per-user）
type MenuFavorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_user_menu" json:"username"`
	MenuIndex string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_user_menu" json:"menu_index"` // 路由路径，如 /assets
	Icon      string    `gorm:"type:varchar(50);not null" json:"icon"`                                  // svg 图标名
	Title     string    `gorm:"type:varchar(50);not null" json:"title"`                                 // 菜单显示名
	SortOrder int       `gorm:"default:0" json:"sort_order"`                                            // 排序序号
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (MenuFavorite) TableName() string {
	return "menu_favorites"
}
