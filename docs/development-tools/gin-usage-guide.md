# Gin框架详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [路由和HTTP方法](#路由和HTTP方法)
3. [中间件](#中间件)
4. [请求处理](#请求处理)
5. [响应处理](#响应处理)
6. [参数绑定和验证](#参数绑定和验证)
7. [文件处理](#文件处理)
8. [错误处理](#错误处理)
9. [高级特性](#高级特性)

## 基础概念

### 1.1 创建Gin应用

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    // 创建默认的Gin引擎（包含Logger和Recovery中间件）
    r := gin.Default()
    
    // 或者创建不带中间件的引擎
    // r := gin.New()
    
    // 启动服务器
    r.Run(":8080") // 默认在0.0.0.0:8080启动服务
}
```

### 1.2 Gin模式

```go
// 设置Gin模式
gin.SetMode(gin.DebugMode)   // 开发模式（默认）
gin.SetMode(gin.ReleaseMode) // 生产模式
gin.SetMode(gin.TestMode)    // 测试模式
```

## 路由和HTTP方法

### 2.1 基本路由

```go
func main() {
    r := gin.Default()
    
    // GET请求
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    
    // POST请求
    r.POST("/users", createUser)
    
    // PUT请求
    r.PUT("/users/:id", updateUser)
    
    // DELETE请求
    r.DELETE("/users/:id", deleteUser)
    
    // PATCH请求
    r.PATCH("/users/:id", patchUser)
    
    // HEAD请求
    r.HEAD("/users", headUsers)
    
    // OPTIONS请求
    r.OPTIONS("/users", optionsUsers)
    
    r.Run()
}
```

### 2.2 路由参数

```go
// 路径参数
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.String(200, "User ID: %s", id)
})

// 通配符参数
r.GET("/files/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.String(200, "File path: %s", filepath)
})

// 查询参数
r.GET("/search", func(c *gin.Context) {
    query := c.Query("q")           // 获取查询参数
    page := c.DefaultQuery("page", "1") // 带默认值的查询参数
    c.JSON(200, gin.H{
        "query": query,
        "page":  page,
    })
})
```

### 2.3 路由组

```go
func main() {
    r := gin.Default()
    
    // 创建路由组
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", getUsers)
        v1.POST("/users", createUser)
        
        // 嵌套路由组
        userGroup := v1.Group("/users")
        {
            userGroup.GET("/:id", getUser)
            userGroup.PUT("/:id", updateUser)
            userGroup.DELETE("/:id", deleteUser)
        }
    }
    
    // 带中间件的路由组
    authorized := r.Group("/admin")
    authorized.Use(AuthRequired())
    {
        authorized.GET("/dashboard", dashboard)
        authorized.GET("/users", adminGetUsers)
    }
    
    r.Run()
}
```

## 中间件

### 3.1 使用内置中间件

```go
func main() {
    r := gin.New()
    
    // 使用Logger中间件
    r.Use(gin.Logger())
    
    // 使用Recovery中间件
    r.Use(gin.Recovery())
    
    // 使用CORS中间件
    r.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    })
    
    r.Run()
}
```

### 3.2 自定义中间件

```go
// 认证中间件
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // 验证token逻辑
        if !validateToken(token) {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC1123),
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

// 限流中间件
func RateLimitMiddleware(limit int) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(limit), limit)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Too many requests"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 请求处理

### 4.1 获取请求数据

```go
// 获取JSON数据
func createUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 处理用户创建逻辑
    c.JSON(201, user)
}

// 获取表单数据
func handleForm(c *gin.Context) {
    name := c.PostForm("name")
    email := c.DefaultPostForm("email", "default@example.com")
    
    c.JSON(200, gin.H{
        "name":  name,
        "email": email,
    })
}

// 获取文件上传
func uploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 保存文件
    if err := c.SaveUploadedFile(file, "./uploads/"+file.Filename); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "File uploaded successfully"})
}
```

### 4.2 请求头处理

```go
func handleHeaders(c *gin.Context) {
    // 获取请求头
    userAgent := c.GetHeader("User-Agent")
    contentType := c.GetHeader("Content-Type")
    
    // 设置响应头
    c.Header("X-Custom-Header", "custom-value")
    
    c.JSON(200, gin.H{
        "user_agent":   userAgent,
        "content_type": contentType,
    })
}
```

## 响应处理

### 5.1 不同类型的响应

```go
func responses(c *gin.Context) {
    // JSON响应
    c.JSON(200, gin.H{"message": "success"})
    
    // XML响应
    c.XML(200, gin.H{"message": "success"})
    
    // YAML响应
    c.YAML(200, gin.H{"message": "success"})
    
    // 字符串响应
    c.String(200, "Hello %s", "World")
    
    // HTML响应
    c.HTML(200, "index.html", gin.H{
        "title": "Main website",
    })
    
    // 重定向
    c.Redirect(302, "/new-path")
    
    // 文件响应
    c.File("./static/image.png")
    
    // 数据响应
    c.Data(200, "image/png", []byte("..."))
}
```

### 5.2 状态码处理

```go
func statusCodes(c *gin.Context) {
    // 设置状态码
    c.Status(201)
    
    // 带状态码的JSON响应
    c.JSON(http.StatusCreated, gin.H{"message": "Created"})
    
    // 中止请求并返回状态码
    c.AbortWithStatus(404)
    
    // 中止请求并返回JSON
    c.AbortWithStatusJSON(400, gin.H{"error": "Bad Request"})
}
```

## 参数绑定和验证

### 6.1 结构体绑定

```go
type User struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"gte=0,lte=130"`
    Password string `json:"password" binding:"required,min=6"`
}

func createUser(c *gin.Context) {
    var user User

    // 绑定JSON数据
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 绑定查询参数
    if err := c.ShouldBindQuery(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 绑定表单数据
    if err := c.ShouldBind(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, user)
}
```

### 6.2 自定义验证器

```go
import "github.com/go-playground/validator/v10"

// 自定义验证函数
func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    return len(username) >= 3 && len(username) <= 20
}

func init() {
    // 注册自定义验证器
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("username", validateUsername)
    }
}

type RegisterRequest struct {
    Username string `json:"username" binding:"required,username"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
```

### 6.3 多种绑定方式

```go
func bindingExamples(c *gin.Context) {
    // URI绑定
    type UriBinding struct {
        ID int `uri:"id" binding:"required"`
    }
    var uri UriBinding
    if err := c.ShouldBindUri(&uri); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Header绑定
    type HeaderBinding struct {
        Rate   int    `header:"Rate"`
        Domain string `header:"Domain"`
    }
    var header HeaderBinding
    if err := c.ShouldBindHeader(&header); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "uri":    uri,
        "header": header,
    })
}
```

## 文件处理

### 7.1 单文件上传

```go
func uploadSingle(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "No file uploaded"})
        return
    }

    // 验证文件类型
    if !isValidFileType(file.Header.Get("Content-Type")) {
        c.JSON(400, gin.H{"error": "Invalid file type"})
        return
    }

    // 验证文件大小（例如：最大10MB）
    if file.Size > 10*1024*1024 {
        c.JSON(400, gin.H{"error": "File too large"})
        return
    }

    // 生成唯一文件名
    filename := generateUniqueFilename(file.Filename)
    filepath := "./uploads/" + filename

    if err := c.SaveUploadedFile(file, filepath); err != nil {
        c.JSON(500, gin.H{"error": "Failed to save file"})
        return
    }

    c.JSON(200, gin.H{
        "message":  "File uploaded successfully",
        "filename": filename,
        "size":     file.Size,
    })
}
```

### 7.2 多文件上传

```go
func uploadMultiple(c *gin.Context) {
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    files := form.File["files"]
    var uploadedFiles []string

    for _, file := range files {
        filename := generateUniqueFilename(file.Filename)
        filepath := "./uploads/" + filename

        if err := c.SaveUploadedFile(file, filepath); err != nil {
            c.JSON(500, gin.H{"error": "Failed to save file: " + file.Filename})
            return
        }

        uploadedFiles = append(uploadedFiles, filename)
    }

    c.JSON(200, gin.H{
        "message": "Files uploaded successfully",
        "files":   uploadedFiles,
        "count":   len(uploadedFiles),
    })
}
```

### 7.3 文件下载

```go
func downloadFile(c *gin.Context) {
    filename := c.Param("filename")
    filepath := "./uploads/" + filename

    // 检查文件是否存在
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(404, gin.H{"error": "File not found"})
        return
    }

    // 设置下载头
    c.Header("Content-Description", "File Transfer")
    c.Header("Content-Transfer-Encoding", "binary")
    c.Header("Content-Disposition", "attachment; filename="+filename)
    c.Header("Content-Type", "application/octet-stream")

    c.File(filepath)
}
```

## 错误处理

### 8.1 全局错误处理

```go
func ErrorHandler() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        if err, ok := recovered.(string); ok {
            c.JSON(500, gin.H{
                "error":   "Internal Server Error",
                "message": err,
            })
        }
        c.AbortWithStatus(500)
    })
}

func main() {
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(ErrorHandler())

    r.Run()
}
```

### 8.2 自定义错误类型

```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return e.Message
}

func NewAPIError(code int, message string, details ...string) *APIError {
    err := &APIError{
        Code:    code,
        Message: message,
    }
    if len(details) > 0 {
        err.Details = details[0]
    }
    return err
}

func handleAPIError(c *gin.Context, err error) {
    if apiErr, ok := err.(*APIError); ok {
        c.JSON(apiErr.Code, apiErr)
        return
    }

    c.JSON(500, gin.H{
        "error":   "Internal Server Error",
        "message": err.Error(),
    })
}
```

## 高级特性

### 9.1 模板渲染

```go
func main() {
    r := gin.Default()

    // 加载HTML模板
    r.LoadHTMLGlob("templates/*")

    // 设置静态文件路径
    r.Static("/static", "./static")

    r.GET("/", func(c *gin.Context) {
        c.HTML(200, "index.html", gin.H{
            "title": "Gin Framework",
            "users": []string{"Alice", "Bob", "Charlie"},
        })
    })

    r.Run()
}
```

### 9.2 WebSocket支持

```go
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func websocketHandler(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Print("upgrade failed: ", err)
        return
    }
    defer conn.Close()

    for {
        mt, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("read failed:", err)
            break
        }

        // 回显消息
        err = conn.WriteMessage(mt, message)
        if err != nil {
            log.Println("write failed:", err)
            break
        }
    }
}
```

### 9.3 优雅关闭

```go
import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    r := gin.Default()

    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    // 在goroutine中启动服务器
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // 等待中断信号来优雅地关闭服务器
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // 5秒的超时时间来关闭服务器
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exiting")
}
```

### 9.4 测试

```go
import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestPingRoute(t *testing.T) {
    // 设置测试模式
    gin.SetMode(gin.TestMode)

    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // 创建测试请求
    req, _ := http.NewRequest("GET", "/ping", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    // 断言
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "pong")
}

func TestCreateUser(t *testing.T) {
    gin.SetMode(gin.TestMode)

    r := gin.Default()
    r.POST("/users", createUser)

    user := User{
        Name:  "John Doe",
        Email: "john@example.com",
        Age:   30,
    }

    jsonData, _ := json.Marshal(user)
    req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, 201, w.Code)
}
```

## 实际项目示例

### 10.1 完整的用户管理API

```go
// models/user.go
type User struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"unique;not null" binding:"required,min=3,max=20"`
    Email    string `json:"email" gorm:"unique;not null" binding:"required,email"`
    Password string `json:"-" gorm:"not null" binding:"required,min=6"`
    IsActive bool   `json:"is_active" gorm:"default:true"`
}

// handlers/user.go
type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    users, total, err := h.userService.GetUsers(page, limit)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "users": users,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    user, err := h.userService.CreateUser(&req)
    if err != nil {
        if errors.Is(err, service.ErrUserExists) {
            c.JSON(409, gin.H{"error": "User already exists"})
            return
        }
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid user ID"})
        return
    }

    user, err := h.userService.GetUserByID(uint(id))
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            c.JSON(404, gin.H{"error": "User not found"})
            return
        }
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, user)
}
```

### 10.2 中间件组合使用

```go
func SetupRoutes(r *gin.Engine) {
    // 全局中间件
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(CORSMiddleware())
    r.Use(RateLimitMiddleware(100)) // 每秒100个请求

    // 健康检查（无需认证）
    r.GET("/health", healthHandler.CheckHealth)

    // API v1
    v1 := r.Group("/api/v1")
    v1.Use(RequestIDMiddleware())
    {
        // 公开端点
        auth := v1.Group("/auth")
        {
            auth.POST("/login", authHandler.Login)
            auth.POST("/register", authHandler.Register)
        }

        // 需要认证的端点
        protected := v1.Group("/")
        protected.Use(AuthMiddleware())
        {
            users := protected.Group("/users")
            users.Use(LoggingMiddleware())
            {
                users.GET("/", userHandler.GetUsers)
                users.POST("/", userHandler.CreateUser)
                users.GET("/:id", userHandler.GetUser)
                users.PUT("/:id", userHandler.UpdateUser)
                users.DELETE("/:id", userHandler.DeleteUser)
            }

            // 管理员端点
            admin := protected.Group("/admin")
            admin.Use(AdminMiddleware())
            {
                admin.GET("/stats", adminHandler.GetStats)
                admin.GET("/logs", adminHandler.GetLogs)
            }
        }
    }
}
```

这个详细的Gin使用指南涵盖了从基础到高级的所有重要概念。您可以根据项目需要参考相应的部分来实现功能。
```
```
