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
		}

		// 受保护接口（需要JWT认证）
		protected := api.Group("")
		protected.Use(middleware.JWTAuth())
		{
			// 双控验证（需要JWT但不需要双控token）
			protected.POST("/dual-control/verify", handlers.VerifyDualControl)

			// 区域管理
			protected.GET("/regions", handlers.ListRegions)
			protected.POST("/regions", handlers.CreateRegion)
			protected.PUT("/regions/:id", handlers.UpdateRegion)
			protected.DELETE("/regions/:id", handlers.DeleteRegion)

			// 操作系统类型管理
			protected.GET("/os-types", handlers.ListOSTypes)
			protected.POST("/os-types", handlers.CreateOSType)
			protected.PUT("/os-types/:id", handlers.UpdateOSType)
			protected.DELETE("/os-types/:id", handlers.DeleteOSType)

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
			protected.POST("/permission-rules/reorder", handlers.ReorderPermissionRule)
			protected.PUT("/permission-rules/systems/reorder", handlers.ReorderSystemInPermissions)

			// 部门管理 - 查询（不需要双控）
			protected.GET("/departments", handlers.ListDepartments)
			protected.GET("/departments/:id/positions", handlers.ListDepartmentPositions)

			// 用户权限管理 - 查询（不需要双控）
			protected.GET("/user-permissions", handlers.ListUserPermissions)
			protected.GET("/user-permissions/:id", handlers.GetUserPermission)

			// ============ 双控保护接口（需要JWT + 双控验证）============
			dual := protected.Group("")
			dual.Use(middleware.DualControl())
			{
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
			}
		}
	}

	return r
}
