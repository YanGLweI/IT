package main

import (
	"fmt"
	"log"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/routes"
)

func main() {
	// 加载配置文件
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Println("配置文件加载成功!")

	// 初始化数据库
	database.InitDB()

	// 设置路由
	r := routes.SetupRouter()

	fmt.Printf("服务器启动在 %s\n", config.Cfg.Server.Port)
	r.Run(config.Cfg.Server.Port)
}
