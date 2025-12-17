package auth

import (
	"time"
)

// User 用户模型
type User struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"` // 永不序列化到 JSON
	Nickname     string    `gorm:"not null" json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserResponse 用户信息响应（仅返回非敏感信息）
type UserResponse struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}
