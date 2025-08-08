package examples

import (
	"awesome-trade/src/internal/model"
	"awesome-trade/src/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserRequest 用户请求结构
type UserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ExampleHandler 示例处理器
type ExampleHandler struct{}

// NewExampleHandler 创建示例处理器
func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{}
}

// 1. 基础路由示例
func (h *ExampleHandler) BasicRoutes(c *gin.Context) {
	// 获取路径参数
	id := c.Param("id")

	// 获取查询参数
	name := c.Query("name")
	page := c.DefaultQuery("page", "1")

	// 返回JSON响应
	utils.Success(c, gin.H{
		"id":   id,
		"name": name,
		"page": page,
	})
}

// 2. 请求绑定示例
func (h *ExampleHandler) CreateUser(c *gin.Context) {
	var req UserRequest

	// 绑定JSON数据并验证
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	// 模拟创建用户
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		IsActive: true,
	}

	// 这里应该调用服务层保存用户
	// userService.CreateUser(&user)

	utils.Success(c, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// 3. 文件上传示例
func (h *ExampleHandler) UploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "No file uploaded")
		return
	}

	// 验证文件大小（最大5MB）
	if file.Size > 5*1024*1024 {
		utils.BadRequest(c, "File size too large (max 5MB)")
		return
	}

	// 生成文件名
	filename := time.Now().Format("20060102150405") + "_" + file.Filename
	filepath := "./uploads/" + filename

	// 保存文件
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		utils.InternalServerError(c, "Failed to save file")
		return
	}

	utils.Success(c, gin.H{
		"message":  "File uploaded successfully",
		"filename": filename,
		"size":     file.Size,
	})
}

// 4. 多文件上传示例
func (h *ExampleHandler) UploadMultipleFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequest(c, "Failed to parse form")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.BadRequest(c, "No files uploaded")
		return
	}

	var uploadedFiles []string
	for _, file := range files {
		filename := time.Now().Format("20060102150405") + "_" + file.Filename
		filepath := "./uploads/" + filename

		if err := c.SaveUploadedFile(file, filepath); err != nil {
			utils.InternalServerError(c, "Failed to save file: "+file.Filename)
			return
		}

		uploadedFiles = append(uploadedFiles, filename)
	}

	utils.Success(c, gin.H{
		"message": "Files uploaded successfully",
		"files":   uploadedFiles,
		"count":   len(uploadedFiles),
	})
}

// 5. 表单数据处理示例
func (h *ExampleHandler) HandleForm(c *gin.Context) {
	// 获取表单数据
	name := c.PostForm("name")
	email := c.DefaultPostForm("email", "")
	age, _ := strconv.Atoi(c.PostForm("age"))

	// 获取多选框数据
	hobbies := c.PostFormArray("hobbies")

	utils.Success(c, gin.H{
		"name":    name,
		"email":   email,
		"age":     age,
		"hobbies": hobbies,
	})
}

// 6. Cookie处理示例
func (h *ExampleHandler) SetCookie(c *gin.Context) {
	// 设置Cookie
	c.SetCookie(
		"session_id", // name
		"abc123",     // value
		3600,         // maxAge (秒)
		"/",          // path
		"localhost",  // domain
		false,        // secure
		true,         // httpOnly
	)

	utils.Success(c, gin.H{
		"message": "Cookie set successfully",
	})
}

func (h *ExampleHandler) GetCookie(c *gin.Context) {
	// 获取Cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		utils.BadRequest(c, "Session cookie not found")
		return
	}

	utils.Success(c, gin.H{
		"session_id": sessionID,
	})
}

// 7. 请求头处理示例
func (h *ExampleHandler) HandleHeaders(c *gin.Context) {
	// 获取请求头
	userAgent := c.GetHeader("User-Agent")
	authorization := c.GetHeader("Authorization")
	contentType := c.GetHeader("Content-Type")

	// 设置响应头
	c.Header("X-API-Version", "v1.0")
	c.Header("X-Response-Time", time.Now().Format(time.RFC3339))

	utils.Success(c, gin.H{
		"user_agent":    userAgent,
		"authorization": authorization,
		"content_type":  contentType,
	})
}

// 8. 重定向示例
func (h *ExampleHandler) RedirectExample(c *gin.Context) {
	redirectType := c.Query("type")

	switch redirectType {
	case "permanent":
		c.Redirect(http.StatusMovedPermanently, "/new-location")
	case "temporary":
		c.Redirect(http.StatusFound, "/new-location")
	default:
		c.Redirect(http.StatusFound, "/")
	}
}

// 9. 流式响应示例
func (h *ExampleHandler) StreamResponse(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.Header("Transfer-Encoding", "chunked")

	for i := 0; i < 10; i++ {
		c.SSEvent("message", gin.H{
			"time":    time.Now().Format(time.RFC3339),
			"counter": i,
		})
		c.Writer.Flush()
		time.Sleep(time.Second)
	}
}

// 10. 错误处理示例
func (h *ExampleHandler) ErrorExample(c *gin.Context) {
	errorType := c.Query("type")

	switch errorType {
	case "400":
		utils.BadRequest(c, "This is a bad request error")
	case "401":
		utils.Unauthorized(c, "This is an unauthorized error")
	case "500":
		utils.InternalServerError(c, "This is an internal server error")
	default:
		utils.Success(c, gin.H{
			"message": "No error occurred",
		})
	}
}

// SetupExampleRoutes 设置示例路由
func SetupExampleRoutes(r *gin.Engine) {
	handler := NewExampleHandler()

	examples := r.Group("/examples")
	{
		// 基础路由
		examples.GET("/basic/:id", handler.BasicRoutes)

		// 请求处理
		examples.POST("/users", handler.CreateUser)
		examples.POST("/form", handler.HandleForm)

		// 文件上传
		examples.POST("/upload", handler.UploadFile)
		examples.POST("/upload-multiple", handler.UploadMultipleFiles)

		// Cookie处理
		examples.POST("/set-cookie", handler.SetCookie)
		examples.GET("/get-cookie", handler.GetCookie)

		// 请求头处理
		examples.GET("/headers", handler.HandleHeaders)

		// 重定向
		examples.GET("/redirect", handler.RedirectExample)

		// 流式响应
		examples.GET("/stream", handler.StreamResponse)

		// 错误处理
		examples.GET("/error", handler.ErrorExample)
	}
}
