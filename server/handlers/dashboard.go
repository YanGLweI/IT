package handlers

import (
	"net/http"
	"time"

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

// LevelStat 资产等级统计
type LevelStat struct {
	Level string `json:"level"`
	Count int64  `json:"count"`
}

// TrendStat 操作日志趋势
type TrendStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// SoftwareUpdateStat 软件更新状态统计
type SoftwareUpdateStat struct {
	NeedUpdate bool  `json:"need_update"`
	Count      int64 `json:"count"`
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

	// SFTP账号数
	var totalSftpAccounts int64
	db.Model(&models.SftpAccount{}).Count(&totalSftpAccounts)

	// 用户/岗位数
	var totalUserPermissions int64
	db.Model(&models.UserPermission{}).Count(&totalUserPermissions)

	// 需更新软件数
	var needUpdateSoftware int64
	db.Model(&models.ApprovedSoftware{}).Where("need_update = ?", true).Count(&needUpdateSoftware)

	// 本月操作次数
	var monthlyOpCount int64
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	db.Model(&models.OperationLog{}).Where("created_at >= ?", monthStart).Count(&monthlyOpCount)

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
		Select("os_types.name as os_type, count(assets.id) as count").
		Joins("LEFT JOIN os_types ON assets.os_type_id = os_types.id").
		Group("os_types.name").
		Scan(&osStats)

	// 状态统计
	var statusStats []StatusStat
	db.Model(&models.Asset{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats)

	// 资产等级统计
	var levelStats []LevelStat
	db.Model(&models.Asset{}).
		Select("COALESCE(asset_level, '未分级') as level, count(*) as count").
		Group("asset_level").
		Scan(&levelStats)

	// 近30天操作日志趋势
	var trendStats []TrendStat
	thirtyDaysAgo := now.AddDate(0, 0, -29)
	db.Model(&models.OperationLog{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&trendStats)

	// 软件更新状态
	var softwareUpdateStats []SoftwareUpdateStat
	db.Model(&models.ApprovedSoftware{}).
		Select("need_update, count(*) as count").
		Group("need_update").
		Scan(&softwareUpdateStats)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total_assets":          totalAssets,
			"total_regions":         totalRegions,
			"total_sftp_accounts":   totalSftpAccounts,
			"total_user_permissions": totalUserPermissions,
			"need_update_software":  needUpdateSoftware,
			"monthly_op_count":      monthlyOpCount,
			"region_stats":          regionStats,
			"os_stats":              osStats,
			"status_stats":          statusStats,
			"level_stats":           levelStats,
			"trend_stats":           trendStats,
			"software_update_stats": softwareUpdateStats,
		},
	})
}
