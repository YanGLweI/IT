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
		// 区域管理
		api.GET("/regions", handlers.ListRegions)
		api.POST("/regions", handlers.CreateRegion)
		api.PUT("/regions/:id", handlers.UpdateRegion)
		api.DELETE("/regions/:id", handlers.DeleteRegion)

		// 操作系统类型管理
		api.GET("/os-types", handlers.ListOSTypes)
		api.POST("/os-types", handlers.CreateOSType)
		api.PUT("/os-types/:id", handlers.UpdateOSType)
		api.DELETE("/os-types/:id", handlers.DeleteOSType)

		// 资产管理
		api.GET("/assets", handlers.ListAssets)
		api.GET("/assets/:id", handlers.GetAsset)
		api.POST("/assets", handlers.CreateAsset)
		api.PUT("/assets/:id", handlers.UpdateAsset)
		api.DELETE("/assets/:id", handlers.DeleteAsset)

		// 看板统计
		api.GET("/dashboard/summary", handlers.DashboardSummary)

		// IT政策管理
		api.GET("/policies", handlers.ListPolicies)
		api.POST("/policies", handlers.CreatePolicy)
		api.PUT("/policies/:id", handlers.UpdatePolicy)
		api.PUT("/policies/:id/file", handlers.ReplacePolicyFile)
		api.DELETE("/policies/:id", handlers.DeletePolicy)
		api.GET("/policies/:id/preview", handlers.PreviewPolicy)
		api.GET("/policies/:id/download", handlers.DownloadPolicy)

		// 网络拓扑图
		api.GET("/topologies", handlers.ListTopologies)
		api.POST("/topologies", handlers.CreateTopology)
		api.PUT("/topologies/:id", handlers.UpdateTopology)
		api.PUT("/topologies/:id/file", handlers.ReplaceTopologyFile)
		api.DELETE("/topologies/:id", handlers.DeleteTopology)
		api.GET("/topologies/:id/preview", handlers.PreviewTopology)
		api.GET("/topologies/:id/download", handlers.DownloadTopology)
	}

	return r
}
