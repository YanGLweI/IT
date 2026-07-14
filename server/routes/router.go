package routes

import (
	"it-platform-server/handlers"
	"it-platform-server/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.Cors())

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// API路由组
	api := r.Group("/api")
	{
		// 公开接口（无需认证）
		public := api.Group("")
		{
			public.POST("/login", handlers.Login)
			public.GET("/public-key", handlers.GetPublicKey)
			public.POST("/refresh-token", handlers.RefreshToken)

			// 免登录表单下载（公开接口）
			public.GET("/public/forms", handlers.ListPublicForms)
			public.GET("/public/forms/:id/download", handlers.PublicDownloadForm)
			public.GET("/public/forms/:id/preview", handlers.PublicPreviewForm)

			// IT指南公开接口
			public.GET("/public/it-guides", handlers.ListPublicITGuides)
			public.GET("/public/it-guides/:id", handlers.GetPublicITGuide)
		}

		// 受保护接口（需要JWT认证）
		protected := api.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 双控验证（需要JWT但不需要双控token）
			protected.POST("/dual-control/verify", handlers.VerifyDualControl)

			// 区域管理 - 查询（不需要双控）
			protected.GET("/regions", handlers.ListRegions)

			// 操作系统类型管理 - 查询（不需要双控）
			protected.GET("/os-types", handlers.ListOSTypes)

			// 资产管理 - 查询（不需要双控）
			protected.GET("/assets", handlers.ListAssets)
			protected.GET("/assets/:id", handlers.GetAsset)

			// 看板统计
			protected.GET("/dashboard/summary", handlers.DashboardSummary)

			// IT政策管理 - 查询（不需要双控）
			protected.GET("/policies", handlers.ListPolicies)
			protected.POST("/policies", handlers.CreatePolicy)
			protected.GET("/policies/:id/preview", handlers.PreviewPolicy)
			protected.GET("/policies/:id/download", handlers.DownloadPolicy)

			// 网络拓扑图 - 查询（不需要双控）
			protected.GET("/topologies", handlers.ListTopologies)
			protected.POST("/topologies", handlers.CreateTopology)
			protected.GET("/topologies/:id/preview", handlers.PreviewTopology)
			protected.GET("/topologies/:id/download", handlers.DownloadTopology)

			// 岗位权限管理 - 查询和排序（不需要双控）
			protected.GET("/permission-rules", handlers.ListPermissionRules)
			protected.GET("/permission-rules/position", handlers.GetPositionPermissions)
			protected.POST("/permission-rules/reorder", handlers.ReorderPermissionRule)
			protected.PUT("/permission-rules/systems/reorder", handlers.ReorderSystemInPermissions)
			protected.GET("/permission-rules/export-change-record", handlers.ExportChangeRecord)

			// 部门管理 - 查询（不需要双控）
			protected.GET("/departments", handlers.ListDepartments)
			protected.GET("/departments/:id/positions", handlers.ListDepartmentPositions)

			// 用户权限管理 - 查询（不需要双控）
			protected.GET("/user-permissions", handlers.ListUserPermissions)
			protected.GET("/user-permissions/:id", handlers.GetUserPermission)
			protected.GET("/user-permissions/export-confirmation", handlers.ExportDepartmentConfirmation)

			// SFTP账号管理 - 查询（不需要双控）
			protected.GET("/sftp-servers", handlers.ListSftpServers)
			protected.GET("/sftp-accounts", handlers.ListSftpAccounts)
			protected.GET("/sftp-accounts/export-confirmation", handlers.ExportSftpConfirmation)

			// 第三方应用管理 - 查询（不需要双控）
			protected.GET("/approved-software/need-update", handlers.ListApprovedSoftwareNeedUpdate)
			protected.GET("/approved-software", handlers.ListApprovedSoftware)
			protected.GET("/asset-software", handlers.ListAssetSoftware)
			protected.GET("/asset-software/export-patch-update", handlers.ExportPatchUpdateRecord)
			protected.GET("/asset-software/:id/links", handlers.GetAssetSoftwareLinks)

			// 月度检查历史 - 查询（不需要双控）
			protected.GET("/monthly-checks", handlers.ListMonthlyChecks)
			protected.GET("/monthly-checks/:id/download", handlers.DownloadMonthlyCheck)
			protected.GET("/monthly-checks/:id/preview", handlers.PreviewMonthlyCheck)

			// 季度检查历史 - 查询（不需要双控）
			protected.GET("/quarterly-checks", handlers.ListQuarterlyChecks)
			protected.GET("/quarterly-checks/:id/download", handlers.DownloadQuarterlyCheck)
			protected.GET("/quarterly-checks/:id/preview", handlers.PreviewQuarterlyCheck)

			// 用户变更记录历史 - 查询（不需要双控）
			protected.GET("/user-change-histories", handlers.ListUserChangeHistories)
			protected.GET("/user-change-histories/:id/download", handlers.DownloadUserChangeHistory)
			protected.GET("/user-change-histories/:id/preview", handlers.PreviewUserChangeHistory)

			// 变更管理 - 查询（不需要双控）
			protected.GET("/change-types", handlers.ListChangeTypes)
			protected.POST("/change-types/reorder", handlers.ReorderChangeType)
			protected.GET("/change-record-templates", handlers.ListChangeRecordTemplates)
			protected.GET("/change-record-templates/current", handlers.GetCurrentChangeRecordTemplate)
			protected.GET("/change-record-templates/:id/download", handlers.DownloadChangeRecordTemplate)
			protected.GET("/change-record-templates/:id/preview", handlers.PreviewChangeRecordTemplate)
			protected.GET("/change-records", handlers.ListChangeRecords)
			protected.GET("/change-records/:id/preview", handlers.PreviewChangeRecord)
			protected.GET("/change-records/:id/download", handlers.DownloadChangeRecord)

			// 漏洞扫描 - 查询（不需要双控）
			protected.GET("/vulnerability-scans", handlers.ListVulnerabilityScans)
			protected.GET("/vulnerability-scans/stats", handlers.GetVulnerabilityScanStats)
			protected.GET("/vulnerability-scans/:id/download", handlers.DownloadVulnerabilityScan)
			protected.GET("/vulnerability-scans/:id/preview", handlers.PreviewVulnerabilityScan)
			protected.GET("/vulnerability-scans/:id/fix-preview", handlers.PreviewFixReport)
			protected.GET("/vulnerability-scans/:id/fix-download", handlers.DownloadFixReport)
			protected.GET("/vulnerability-scans/:id/rect-preview", handlers.PreviewRectReport)
			protected.GET("/vulnerability-scans/:id/rect-download", handlers.DownloadRectReport)

			// 系统加固 - 查询（不需要双控）
			protected.GET("/system-hardening/export-checklist", handlers.ExportSystemHardeningChecklist)
			protected.GET("/system-hardening", handlers.ListSystemHardeningHistories)
			protected.GET("/system-hardening/:id/download", handlers.DownloadSystemHardeningHistory)
			protected.GET("/system-hardening/:id/preview", handlers.PreviewSystemHardeningHistory)

			// 渗透测试 - 查询（不需要双控）
			protected.GET("/penetration-tests", handlers.ListPenetrationTests)
			protected.GET("/penetration-tests/:id/preview", handlers.PreviewPenetrationTest)
			protected.GET("/penetration-tests/:id/download", handlers.DownloadPenetrationTest)

			// 防火墙检查 - 查询（不需要双控）
			protected.GET("/firewall-checks", handlers.ListFirewallChecks)
			protected.GET("/firewall-checks/:id/preview", handlers.PreviewFirewallCheckReport)
			protected.GET("/firewall-checks/:id/download", handlers.DownloadFirewallCheckReport)
			protected.GET("/firewall-checks/:id/rect-preview", handlers.PreviewFirewallRectReport)
			protected.GET("/firewall-checks/:id/rect-download", handlers.DownloadFirewallRectReport)

			// 补丁更新 - 查询（不需要双控）
			protected.GET("/patch-updates", handlers.ListPatchUpdates)
			protected.GET("/patch-updates/:id/preview", handlers.PreviewPatchUpdate)
			protected.GET("/patch-updates/:id/download", handlers.DownloadPatchUpdate)
			protected.GET("/patch-updates/:id/fix-preview", handlers.PreviewPatchFixReport)
			protected.GET("/patch-updates/:id/fix-download", handlers.DownloadPatchFixReport)

			// 备份管理 - 查询（不需要双控）
			protected.GET("/backups", handlers.ListBackups)
			protected.GET("/backups/:id/preview", handlers.PreviewBackup)
			protected.GET("/backups/:id/download", handlers.DownloadBackup)
			protected.GET("/backup-recoveries/:id/preview", handlers.PreviewBackupRecovery)
			protected.GET("/backup-recoveries/:id/download", handlers.DownloadBackupRecovery)
			protected.GET("/backup-templates", handlers.ListBackupTemplates)
			protected.GET("/backup-templates/:id/download", handlers.DownloadBackupTemplate)
			protected.GET("/backup-templates/:id/preview", handlers.PreviewBackupTemplate)

			// 日志管理 - 查询（不需要双控）
			protected.GET("/login-logs", handlers.ListLoginLogs)
			protected.GET("/operation-logs", handlers.ListOperationLogs)
			protected.GET("/operation-logs/:id/details", handlers.GetOperationLogDetails)
			protected.POST("/logout", handlers.Logout)

			// 表单发布 - 读操作（不需要双控）
			protected.GET("/form-vault", handlers.ListFormVaultItems)
			protected.GET("/form-vault/cross-module-sources", handlers.ListCrossModuleSources)
			protected.GET("/form-vault/cross-module-sources/:module/files", handlers.ListCrossModuleFiles)
			protected.GET("/form-vault/cross-module/generators/:name/params", handlers.GetGeneratorParams)
			protected.GET("/form-vault/:id/preview", handlers.PreviewFormVaultItem)
			protected.GET("/form-vault/:id/download", handlers.DownloadFormVaultItem)

			// IT指南 - 读操作（不需要双控）
			protected.GET("/it-guides", handlers.ListITGuides)
			protected.GET("/it-guides/:id", handlers.GetITGuide)
			protected.GET("/it-guides/:id/steps", handlers.ListITGuideSteps)

			// 日历日程 - 读操作（不需要双控）
			protected.GET("/calendars", handlers.ListCalendars)
			protected.GET("/calendars/:id", handlers.GetCalendar)
			protected.GET("/calendars/today-notifications", handlers.GetTodayNotifications)
			protected.GET("/calendars/unread-count", handlers.GetUnreadCount)
			protected.GET("/calendars/pending-notifications", handlers.GetPendingNotifications)
			protected.PUT("/calendars/notifications/:id/read", handlers.MarkNotificationRead)
			protected.PUT("/calendars/notifications/:id/popup-shown", handlers.MarkNotificationPopupShown)
			protected.POST("/calendars/check-conflict", handlers.CheckConflict)

			// 密码本 - 读操作（不需要双控）
			protected.GET("/password-categories", handlers.ListPasswordCategories)
			protected.GET("/password-entries", handlers.ListPasswordEntries)
			protected.GET("/password-view-logs", handlers.ListPasswordViewLogs)
			protected.POST("/password-entries/:id/unlock", handlers.UnlockPasswordEntry)
			protected.PUT("/password-entries/:id/star", handlers.TogglePasswordEntryStar)
			protected.PUT("/password-categories/:id/sort", handlers.SortPasswordCategory)

			// LDAP用户获取（不需要双控）
			protected.GET("/ldap/users", handlers.GetLDAPUsers)

			// 日历日程 - 写操作（不需要双控，仅需JWT）
			protected.POST("/calendars", handlers.CreateCalendar)
			protected.PUT("/calendars/:id", handlers.UpdateCalendar)
			protected.DELETE("/calendars/:id", handlers.DeleteCalendar)

			// ============ 双控保护接口（需要JWT + 双控验证）============
			dual := protected.Group("")
			dual.Use(middleware.DualControl())
			{
				// 区域管理 - 写操作（需要双控）
				dual.POST("/regions", handlers.CreateRegion)
				dual.PUT("/regions/:id", handlers.UpdateRegion)
				dual.DELETE("/regions/:id", handlers.DeleteRegion)

				// 操作系统类型管理 - 写操作（需要双控）
				dual.POST("/os-types", handlers.CreateOSType)
				dual.PUT("/os-types/:id", handlers.UpdateOSType)
				dual.DELETE("/os-types/:id", handlers.DeleteOSType)

				// 资产管理 - 修改删除
				dual.POST("/assets", handlers.CreateAsset)
				dual.PUT("/assets/:id", handlers.UpdateAsset)
				dual.DELETE("/assets/:id", handlers.DeleteAsset)
				
				// IT政策 - 修改删除
				dual.PUT("/policies/:id", handlers.UpdatePolicy)
				dual.PUT("/policies/:id/file", handlers.ReplacePolicyFile)
				dual.DELETE("/policies/:id", handlers.DeletePolicy)

				// 网络拓扑图 - 修改删除
				dual.PUT("/topologies/:id", handlers.UpdateTopology)
				dual.PUT("/topologies/:id/file", handlers.ReplaceTopologyFile)
				dual.DELETE("/topologies/:id", handlers.DeleteTopology)

				// 岗位权限管理配置（所有写操作）
				dual.POST("/permission-rules", handlers.CreatePermissionRule)
				dual.PUT("/permission-rules/:id", handlers.UpdatePermissionRule)
				dual.DELETE("/permission-rules/:id", handlers.DeletePermissionRule)
				dual.POST("/permission-rules/systems", handlers.AddSystemToPermissions)
				dual.DELETE("/permission-rules/systems", handlers.RemoveSystemFromPermissions)
				dual.PUT("/permission-rules/systems/rename", handlers.RenameSystemInPermissions)
				dual.POST("/permission-rules/systems/roles", handlers.ManageRolesInSystem)

				// 部门管理 - 写操作（需要双控）
				dual.POST("/departments", handlers.CreateDepartment)
				dual.PUT("/departments/:id", handlers.UpdateDepartment)
				dual.DELETE("/departments/:id", handlers.DeleteDepartment)
				dual.POST("/departments/:id/positions", handlers.AddDepartmentPosition)
				dual.DELETE("/departments/:id/positions/:pid", handlers.RemoveDepartmentPosition)

				// 用户权限管理 - 写操作（需要双控）
				dual.POST("/user-permissions", handlers.CreateUserPermission)
				dual.PUT("/user-permissions/:id", handlers.UpdateUserPermission)
				dual.DELETE("/user-permissions/:id", handlers.DeleteUserPermission)

				// SFTP服务器管理 - 写操作（需要双控）
				dual.POST("/sftp-servers", handlers.CreateSftpServer)
				dual.PUT("/sftp-servers/:id", handlers.UpdateSftpServer)
				dual.DELETE("/sftp-servers/:id", handlers.DeleteSftpServer)

				// SFTP账号管理 - 写操作（需要双控）
				dual.POST("/sftp-accounts", handlers.CreateSftpAccount)
				dual.PUT("/sftp-accounts/:id", handlers.UpdateSftpAccount)
				dual.DELETE("/sftp-accounts/:id", handlers.DeleteSftpAccount)

				// 第三方应用管理 - 写操作（需要双控）
				dual.POST("/approved-software", handlers.CreateApprovedSoftware)
				dual.PUT("/approved-software/:id", handlers.UpdateApprovedSoftware)
				dual.DELETE("/approved-software/:id", handlers.DeleteApprovedSoftware)
				dual.PUT("/asset-software/:id/links", handlers.UpdateAssetSoftwareLinks)

				// 月度检查历史 - 写操作（需要双控）
				dual.POST("/monthly-checks", handlers.CreateMonthlyCheck)
				dual.PUT("/monthly-checks/:id", handlers.UpdateMonthlyCheck)
				dual.DELETE("/monthly-checks/:id", handlers.DeleteMonthlyCheck)

				// 季度检查历史 - 写操作（需要双控）
				dual.POST("/quarterly-checks", handlers.CreateQuarterlyCheck)
				dual.PUT("/quarterly-checks/:id", handlers.UpdateQuarterlyCheck)
				dual.DELETE("/quarterly-checks/:id", handlers.DeleteQuarterlyCheck)

				// 用户变更记录历史 - 写操作（需要双控）
				dual.POST("/user-change-histories", handlers.CreateUserChangeHistory)
				dual.PUT("/user-change-histories/:id", handlers.UpdateUserChangeHistory)
				dual.DELETE("/user-change-histories/:id", handlers.DeleteUserChangeHistory)

				// 变更管理 - 写操作（需要双控）
				dual.POST("/change-types", handlers.CreateChangeType)
				dual.PUT("/change-types/:id", handlers.UpdateChangeType)
				dual.DELETE("/change-types/:id", handlers.DeleteChangeType)
				dual.POST("/change-record-templates", handlers.UploadChangeRecordTemplate)
				dual.DELETE("/change-record-templates/:id", handlers.DeleteChangeRecordTemplate)
				dual.POST("/change-records", handlers.CreateChangeRecord)
				dual.PUT("/change-records/:id", handlers.UpdateChangeRecord)
				dual.DELETE("/change-records/:id", handlers.DeleteChangeRecord)

				// 漏洞扫描 - 写操作（需要双控）
				dual.POST("/vulnerability-scans", handlers.CreateVulnerabilityScan)
				dual.PUT("/vulnerability-scans/:id", handlers.UpdateVulnerabilityScan)
				dual.DELETE("/vulnerability-scans/:id", handlers.DeleteVulnerabilityScan)
				dual.PUT("/vulnerability-scans/:id/fix", handlers.FixVulnerabilityScan)
				dual.DELETE("/vulnerability-scans/:id/fix", handlers.DeleteFixReport)
				dual.PUT("/vulnerability-scans/:id/rect", handlers.UploadRectReport)
				dual.DELETE("/vulnerability-scans/:id/rect", handlers.DeleteRectReport)

				// 系统加固 - 写操作（需要双控）
				dual.POST("/system-hardening", handlers.CreateSystemHardeningHistory)
				dual.PUT("/system-hardening/:id", handlers.UpdateSystemHardeningHistory)
				dual.DELETE("/system-hardening/:id", handlers.DeleteSystemHardeningHistory)

				// 渗透测试 - 写操作（需要双控）
				dual.POST("/penetration-tests", handlers.CreatePenetrationTest)
				dual.PUT("/penetration-tests/:id", handlers.UpdatePenetrationTest)
				dual.DELETE("/penetration-tests/:id", handlers.DeletePenetrationTest)

				// 防火墙检查 - 写操作（需要双控）
				dual.POST("/firewall-checks", handlers.CreateFirewallCheck)
				dual.PUT("/firewall-checks/:id", handlers.UpdateFirewallCheck)
				dual.DELETE("/firewall-checks/:id", handlers.DeleteFirewallCheck)
				dual.PUT("/firewall-checks/:id/rect", handlers.UploadFirewallRectReport)
				dual.DELETE("/firewall-checks/:id/rect", handlers.DeleteFirewallRectReport)

				// 补丁更新 - 写操作（需要双控）
				dual.POST("/patch-updates", handlers.CreatePatchUpdate)
				dual.PUT("/patch-updates/:id", handlers.UpdatePatchUpdate)
				dual.DELETE("/patch-updates/:id", handlers.DeletePatchUpdate)
				dual.PUT("/patch-updates/:id/fix", handlers.UploadPatchFixReport)
				dual.DELETE("/patch-updates/:id/fix", handlers.DeletePatchFixReport)

				// 备份管理 - 写操作（需要双控）
				dual.POST("/backups", handlers.CreateBackup)
				dual.PUT("/backups/:id", handlers.UpdateBackup)
				dual.DELETE("/backups/:id", handlers.DeleteBackup)
				dual.POST("/backups/:id/recoveries", handlers.CreateBackupRecovery)
				dual.PUT("/backup-recoveries/:id", handlers.UpdateBackupRecovery)
				dual.DELETE("/backup-recoveries/:id", handlers.DeleteBackupRecovery)
				dual.POST("/backup-templates", handlers.UploadBackupTemplate)
				dual.DELETE("/backup-templates/:id", handlers.DeleteBackupTemplate)

				// 表单发布 - 写操作（需要双控）
				dual.POST("/form-vault", handlers.UploadFormVaultItem)
				dual.PUT("/form-vault/:id", handlers.UpdateFormVaultItem)
				dual.DELETE("/form-vault/:id", handlers.DeleteFormVaultItem)
				dual.PUT("/form-vault/:id/publish", handlers.PublishFormVaultItem)
				dual.PUT("/form-vault/:id/unpublish", handlers.UnpublishFormVaultItem)
				dual.POST("/form-vault/cross-module", handlers.CreateCrossModuleRef)

				// IT指南 - 写操作（需要双控）
				dual.POST("/it-guides", handlers.CreateITGuide)
				dual.PUT("/it-guides/:id", handlers.UpdateITGuide)
				dual.DELETE("/it-guides/:id", handlers.DeleteITGuide)
				dual.PUT("/it-guides/:id/publish", handlers.PublishITGuide)
				dual.PUT("/it-guides/:id/unpublish", handlers.UnpublishITGuide)
				dual.POST("/it-guides/:id/steps", handlers.CreateITGuideStep)
				dual.PUT("/it-guides/:id/steps/:stepId", handlers.UpdateITGuideStep)
				dual.DELETE("/it-guides/:id/steps/:stepId", handlers.DeleteITGuideStep)
				dual.POST("/it-guides/:id/steps/reorder", handlers.ReorderITGuideSteps)
				dual.POST("/it-guides/:id/media", handlers.UploadITGuideMedia)
				dual.DELETE("/it-guides/:id/media/:mediaId", handlers.DeleteITGuideMedia)

				// 密码本 - 写操作（需要双控）
				dual.POST("/password-categories", handlers.CreatePasswordCategory)
				dual.PUT("/password-categories/:id", handlers.UpdatePasswordCategory)
				dual.DELETE("/password-categories/:id", handlers.DeletePasswordCategory)
				dual.POST("/password-entries", handlers.CreatePasswordEntry)
				dual.PUT("/password-entries/:id", handlers.UpdatePasswordEntry)
				dual.DELETE("/password-entries/:id", handlers.DeletePasswordEntry)
			}
		}
	}

	return r
}
