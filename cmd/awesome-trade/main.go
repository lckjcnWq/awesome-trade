package main

import (
	"awesome-trade/internal/config"
	"awesome-trade/internal/handler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化处理器
	h := handler.New(cfg)

	// 设置路由
	http.HandleFunc("/", h.Home)
	http.HandleFunc("/health", h.Health)

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("🚀 Awesome Trade 服务启动在端口 %s\n", cfg.Port)
	fmt.Printf("📊 访问 http://localhost%s 查看应用\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
