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

			// 资产管理
			protected.GET("/assets", handlers.ListAssets)
			protected.GET("/assets/:id", handlers.GetAsset)
			protected.POST("/assets", handlers.CreateAsset)
			protected.PUT("/assets/:id", handlers.UpdateAsset)
			protected.DELETE("/assets/:id", handlers.DeleteAsset)

			// 看板统计
			protected.GET("/dashboard/summary", handlers.DashboardSummary)

			// IT政策管理
			protected.GET("/policies", handlers.ListPolicies)
			protected.POST("/policies", handlers.CreatePolicy)
			protected.PUT("/policies/:id", handlers.UpdatePolicy)
			protected.PUT("/policies/:id/file", handlers.ReplacePolicyFile)
			protected.DELETE("/policies/:id", handlers.DeletePolicy)
			protected.GET("/policies/:id/preview", handlers.PreviewPolicy)
			protected.GET("/policies/:id/download", handlers.DownloadPolicy)

			// 网络拓扑图
			protected.GET("/topologies", handlers.ListTopologies)
			protected.POST("/topologies", handlers.CreateTopology)
			protected.PUT("/topologies/:id", handlers.UpdateTopology)
			protected.PUT("/topologies/:id/file", handlers.ReplaceTopologyFile)
			protected.DELETE("/topologies/:id", handlers.DeleteTopology)
			protected.GET("/topologies/:id/preview", handlers.PreviewTopology)
			protected.GET("/topologies/:id/download", handlers.DownloadTopology)

			// 岗位权限管理
			protected.GET("/permission-rules", handlers.ListPermissionRules)
			protected.POST("/permission-rules", handlers.CreatePermissionRule)
			protected.PUT("/permission-rules/:id", handlers.UpdatePermissionRule)
			protected.DELETE("/permission-rules/:id", handlers.DeletePermissionRule)
			protected.POST("/permission-rules/systems", handlers.AddSystemToPermissions)
			protected.DELETE("/permission-rules/systems", handlers.RemoveSystemFromPermissions)
		}
	}

	return r
}
