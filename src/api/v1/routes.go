package v1

import (
	"awesome-trade/src/internal/handler"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置API路由
func SetupRoutes(r *gin.Engine) {
	// 创建处理器实例
	healthHandler := handler.NewHealthHandler()

	// 健康检查路由
	r.GET("/health", healthHandler.CheckHealth)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 基础路由
		v1.GET("/ping", healthHandler.Ping)

		// 用户相关路由
		userGroup := v1.Group("/users")
		{
			userGroup.GET("/", func(c *gin.Context) {
				// TODO: 实现获取用户列表
			})
			userGroup.POST("/", func(c *gin.Context) {
				// TODO: 实现创建用户
			})
			userGroup.GET("/:id", func(c *gin.Context) {
				// TODO: 实现获取用户详情
			})
			userGroup.PUT("/:id", func(c *gin.Context) {
				// TODO: 实现更新用户
			})
			userGroup.DELETE("/:id", func(c *gin.Context) {
				// TODO: 实现删除用户
			})
		}

		// 认证相关路由
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", func(c *gin.Context) {
				// TODO: 实现用户登录
			})
			authGroup.POST("/register", func(c *gin.Context) {
				// TODO: 实现用户注册
			})
			authGroup.POST("/logout", func(c *gin.Context) {
				// TODO: 实现用户登出
			})
		}
	}
}
