package repository

import (
	"gorm.io/gorm"
)

// BaseRepository 基础仓储结构
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓储实例
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{
		db: db,
	}
}

// GetDB 获取数据库连接
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}
