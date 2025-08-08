package main

import (
	"awesome-trade/src/api/v1"
	"awesome-trade/src/internal/config"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin路由器
	r := gin.Default()

	// 设置路由
	v1.SetupRoutes(r)

	// 启动服务器
	port := ":" + cfg.Server.Port
	log.Printf("Starting Awesome Trade server on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
