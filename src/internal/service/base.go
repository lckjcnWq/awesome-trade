package service

// BaseService 基础服务结构
type BaseService struct {
	// 这里可以添加通用的服务依赖，如数据库连接、缓存等
}

// NewBaseService 创建基础服务实例
func NewBaseService() *BaseService {
	return &BaseService{}
}
