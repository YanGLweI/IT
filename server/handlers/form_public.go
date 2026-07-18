package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListPublicForms 获取已发布表单列表（公开接口，无需认证）
func ListPublicForms(c *gin.Context) {
	var items []models.FormVaultItem

	query := database.GetDB().Model(&models.FormVaultItem{}).Where("is_published = ?", true)

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	var total int64
	query.Count(&total)

	if err := query.Order("sort_order ASC, published_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 获取所有已使用的分类列表
	var categories []string
	database.GetDB().Model(&models.FormVaultItem{}).
		Where("is_published = ? AND category != ''", true).
		Distinct().
		Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "total": total, "page_size": pageSize, "categories": categories})
}

// PublicDownloadForm 公开下载表单文件（无需认证）
func PublicDownloadForm(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().Where("is_published = ?", true).First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在或未发布"})
		return
	}

	// 递增下载计数
	database.GetDB().Model(&models.FormVaultItem{}).Where("id = ?", id).UpdateColumn("download_count", database.GetDB().Raw("download_count + 1"))

	serveFormFile(c, &item)
}

// PublicPreviewForm 公开预览表单文件（无需认证）
func PublicPreviewForm(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.FormVaultItem
	if err := database.GetDB().Where("is_published = ?", true).First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "表单不存在或未发布"})
		return
	}

	serveFormFileInline(c, &item)
}

// getFileExt 获取文件扩展名
func getFileExt(fileName string) string {
	return strings.ToLower(filepath.Ext(fileName))
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
