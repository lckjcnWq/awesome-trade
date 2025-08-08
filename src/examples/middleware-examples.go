package examples

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 1. 日志中间件示例
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// 2. CORS中间件示例
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 3. 认证中间件示例
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// 这里应该验证JWT token
		// 为了示例，我们简单检查token是否为"valid-token"
		if token != "valid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", "123")
		c.Set("username", "john_doe")

		c.Next()
	}
}

// 4. 限流中间件示例
func RateLimitMiddleware(maxRequests int, duration time.Duration) gin.HandlerFunc {
	// 简单的内存限流器（生产环境建议使用Redis）
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// 清理过期的请求记录
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < duration {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// 检查是否超过限制
		if len(clients[clientIP]) >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// 记录当前请求
		clients[clientIP] = append(clients[clientIP], now)

		c.Next()
	}
}

// 5. 请求ID中间件示例
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成或获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		// 存储到上下文中
		c.Set("request_id", requestID)

		c.Next()
	}
}

// 6. 超时中间件示例
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 在goroutine中处理请求
		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
			})
			c.Abort()
		}
	}
}

// 7. 安全头中间件示例
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的HTTP头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// 8. 错误处理中间件示例
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Something went wrong",
			})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// 9. 响应时间中间件示例
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		c.Header("X-Response-Time", duration.String())

		// 记录慢请求
		if duration > 1*time.Second {
			log.Printf("Slow request: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// 10. 内容压缩中间件示例（需要额外的包）
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持gzip
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Header("Content-Encoding", "gzip")
			// 这里应该实现gzip压缩逻辑
		}

		c.Next()
	}
}

// 辅助函数
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// 中间件使用示例
func SetupMiddlewareExamples(r *gin.Engine) {
	// 全局中间件
	r.Use(LoggerMiddleware())
	r.Use(ErrorHandlerMiddleware())
	r.Use(CORSMiddleware())
	r.Use(SecurityHeadersMiddleware())
	r.Use(ResponseTimeMiddleware())

	// 带限流的API组
	api := r.Group("/api")
	api.Use(RateLimitMiddleware(100, time.Minute)) // 每分钟100个请求
	api.Use(RequestIDMiddleware())
	{
		api.GET("/public", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "This is a public endpoint"})
		})
	}

	// 需要认证的API组
	protected := r.Group("/protected")
	protected.Use(AuthMiddleware())
	protected.Use(TimeoutMiddleware(30 * time.Second))
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id")
			username := c.GetString("username")

			c.JSON(200, gin.H{
				"user_id":  userID,
				"username": username,
				"message":  "This is a protected endpoint",
			})
		})
	}
}
