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
	Locale       string    `gorm:"not null;default:zh-CN" json:"locale"`    // 用户语言/区域设置
	ShowAmount   bool      `gorm:"not null;default:true" json:"showAmount"` // 是否显示金额
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserResponse 用户信息响应（仅返回非敏感信息）
type UserResponse struct {
	Email      string `json:"email"`
	Nickname   string `json:"nickname"`
	Locale     string `json:"locale"`
	ShowAmount bool   `json:"showAmount"`
}

// UserSettings 用户设置（仅设置相关字段）
type UserSettings struct {
	Locale     string `json:"locale"`
	ShowAmount bool   `json:"showAmount"`
}
