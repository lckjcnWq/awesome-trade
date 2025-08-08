package examples

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewExampleHandler()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.POST("/users", handler.CreateUser)
	r.GET("/basic/:id", handler.BasicRoutes)
	r.GET("/error", handler.ErrorExample)

	return r
}

// 测试基础GET请求
func TestPingRoute(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

// 测试POST请求和JSON绑定
func TestCreateUser(t *testing.T) {
	router := setupTestRouter()

	user := UserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "User created successfully")
}

// 测试无效JSON数据
func TestCreateUserInvalidData(t *testing.T) {
	router := setupTestRouter()

	// 发送无效的JSON数据
	invalidJSON := `{"username": "ab", "email": "invalid-email"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request data")
}

// 测试路径参数
func TestBasicRoutesWithParams(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/basic/123?name=test&page=2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "123", data["id"])
	assert.Equal(t, "test", data["name"])
	assert.Equal(t, "2", data["page"])
}

// 测试错误处理
func TestErrorHandling(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		errorType      string
		expectedStatus int
		expectedCode   float64
	}{
		{"400", 400, 400},
		{"401", 401, 401},
		{"500", 500, 500},
		{"", 200, 0},
	}

	for _, tc := range testCases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error?type="+tc.errorType, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedStatus, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		if tc.expectedCode > 0 {
			assert.Equal(t, tc.expectedCode, response["code"])
		}
	}
}

// 测试中间件
func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 添加测试中间件
	r.Use(func(c *gin.Context) {
		c.Header("X-Test-Header", "test-value")
		c.Next()
	})

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "test-value", w.Header().Get("X-Test-Header"))
}

// 测试认证中间件
func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	protected := r.Group("/protected")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id")
			c.JSON(200, gin.H{"user_id": userID})
		})
	}

	// 测试无认证头的请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected/profile", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is required")

	// 测试有效token的请求
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected/profile", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "123")

	// 测试无效token的请求
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

// 基准测试
func BenchmarkPingRoute(b *testing.B) {
	router := setupTestRouter()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ping", nil)
		router.ServeHTTP(w, req)
	}
}

// 并发测试
func TestConcurrentRequests(t *testing.T) {
	router := setupTestRouter()

	const numRequests = 100
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping", nil)
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// 收集结果
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		assert.Equal(t, 200, statusCode)
	}
}
