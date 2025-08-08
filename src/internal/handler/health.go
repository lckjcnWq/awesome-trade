package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器实例
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// CheckHealth 健康检查端点
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "Awesome Trade API is running",
		"timestamp": gin.H{},
	})
}

// Ping 简单的ping端点
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
