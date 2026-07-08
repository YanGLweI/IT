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

// VulnTrendItem 漏洞趋势数据点（按季度）
type VulnTrendItem struct {
	Year          int `json:"year"`
	Quarter       int `json:"quarter"`
	CriticalCount int `json:"critical_count"`
	HighCount     int `json:"high_count"`
	MediumCount   int `json:"medium_count"`
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

	// 备份资产数（去重的备份源资产）
	var backupAssets int64
	db.Table("backup_records").
		Where("deleted_at IS NULL").
		Select("COUNT(DISTINCT backup_source_asset_id)").
		Scan(&backupAssets)

	// 未修复漏洞总数
	var totalUnfixedVulns int64
	db.Table("vulnerability_scans").
		Where("status = ? AND deleted_at IS NULL", "unfixed").
		Select("COALESCE(SUM(critical_count),0) + COALESCE(SUM(high_count),0) + COALESCE(SUM(medium_count),0)").
		Scan(&totalUnfixedVulns)

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

	// 漏洞趋势（按季度聚合，最近8个季度）
	var vulnTrend []VulnTrendItem
	cutoffQ := now.Year()*4 + int(now.Month()-1)/3 - 7 // 最近8个季度的起始季度编号
	db.Model(&models.VulnerabilityScan{}).
		Select("year, quarter, SUM(critical_count) as critical_count, SUM(high_count) as high_count, SUM(medium_count) as medium_count").
		Where("deleted_at IS NULL AND (year * 4 + quarter) >= ?", cutoffQ).
		Group("year, quarter").
		Order("year ASC, quarter ASC").
		Scan(&vulnTrend)

	// 近30天操作日志趋势
	var trendStats []TrendStat
	thirtyDaysAgo := now.AddDate(0, 0, -29)
	db.Model(&models.OperationLog{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, count(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("DATE_FORMAT(created_at, '%Y-%m-%d')").
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
			"total_unfixed_vulns":   totalUnfixedVulns,
			"backup_assets":         backupAssets,
			"region_stats":          regionStats,
			"os_stats":              osStats,
			"status_stats":          statusStats,
			"vuln_trend":            vulnTrend,
			"trend_stats":           trendStats,
			"software_update_stats": softwareUpdateStats,
		},
	})
}
