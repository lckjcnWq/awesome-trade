package config

import (
	"os"
)

// Config 应用配置结构
type Config struct {
	Port        string `json:"port"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// Load 加载应用配置
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnv("DEBUG", "true") == "true",
	}

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
