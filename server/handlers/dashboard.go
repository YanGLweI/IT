package handlers

import (
	"net/http"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// RegionStat 区域统计
type RegionStat struct {
	RegionName string `json:"region_name"`
	Count      int64  `json:"count"`
}

// OSStat 操作系统统计
type OSStat struct {
	OSType string `json:"os_type"`
	Count  int64  `json:"count"`
}

// StatusStat 状态统计
type StatusStat struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// DashboardSummary 看板统计
func DashboardSummary(c *gin.Context) {
	db := database.GetDB()

	// 总资产数
	var totalAssets int64
	db.Model(&models.Asset{}).Count(&totalAssets)

	// 总区域数
	var totalRegions int64
	db.Model(&models.Region{}).Count(&totalRegions)

	// 区域资产统计
	var regionStats []RegionStat
	db.Model(&models.Asset{}).
		Select("regions.name as region_name, count(assets.id) as count").
		Joins("LEFT JOIN regions ON assets.region_id = regions.id").
		Group("regions.name").
		Scan(&regionStats)

	// 操作系统统计
	var osStats []OSStat
	db.Model(&models.Asset{}).
		Select("os_type, count(*) as count").
		Group("os_type").
		Scan(&osStats)

	// 状态统计
	var statusStats []StatusStat
	db.Model(&models.Asset{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total_assets":  totalAssets,
			"total_regions": totalRegions,
			"region_stats":  regionStats,
			"os_stats":      osStats,
			"status_stats":  statusStats,
		},
	})
}
