package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"awesome-trade/internal/config"
)

// Handler HTTP处理器结构
type Handler struct {
	config *config.Config
}

// New 创建新的处理器实例
func New(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

// Home 首页处理器
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message":     "欢迎使用 Awesome Trade!",
		"version":     "1.0.0",
		"environment": h.config.Environment,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// Health 健康检查处理器
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "awesome-trade",
	}

	json.NewEncoder(w).Encode(response)
}
