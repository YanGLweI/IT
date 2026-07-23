package handlers

import (
	"net/http"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListMenuFavorites 获取当前用户的菜单收藏列表
func ListMenuFavorites(c *gin.Context) {
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	var favorites []models.MenuFavorite
	if err := database.GetDB().
		Where("username = ?", usernameStr).
		Order("sort_order ASC, created_at ASC").
		Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": favorites})
}

// ToggleMenuFavorite 切换菜单收藏状态（per-user）
func ToggleMenuFavorite(c *gin.Context) {
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	var req struct {
		MenuIndex   string `json:"menu_index" binding:"required"`
		Icon        string `json:"icon" binding:"required"`
		Title       string `json:"title" binding:"required"`
		IsFavorited bool   `json:"is_favorited"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数无效"})
		return
	}

	db := database.GetDB()
	if req.IsFavorited {
		// 收藏：插入记录（忽略重复）
		// 先查当前最大 sort_order
		var maxOrder int
		db.Model(&models.MenuFavorite{}).
			Where("username = ?", usernameStr).
			Select("COALESCE(MAX(sort_order), 0)").
			Scan(&maxOrder)
		result := db.Where(models.MenuFavorite{Username: usernameStr, MenuIndex: req.MenuIndex}).
			Attrs(models.MenuFavorite{Icon: req.Icon, Title: req.Title, SortOrder: maxOrder + 1}).
			FirstOrCreate(&models.MenuFavorite{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "收藏失败"})
			return
		}
	} else {
		// 取消收藏：删除记录
		if err := db.Where("username = ? AND menu_index = ?", usernameStr, req.MenuIndex).Delete(&models.MenuFavorite{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消收藏失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "操作成功"})
}

// ReorderMenuFavorites 批量更新收藏排序
func ReorderMenuFavorites(c *gin.Context) {
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	var req struct {
		MenuIndices []string `json:"menu_indices" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数无效"})
		return
	}

	db := database.GetDB()
	tx := db.Begin()
	for i, menuIndex := range req.MenuIndices {
		if err := tx.Model(&models.MenuFavorite{}).
			Where("username = ? AND menu_index = ?", usernameStr, menuIndex).
			Update("sort_order", i+1).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "排序失败"})
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "排序失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "排序成功"})
}
