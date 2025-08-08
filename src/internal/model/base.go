package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型结构
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User 用户模型示例
type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}
