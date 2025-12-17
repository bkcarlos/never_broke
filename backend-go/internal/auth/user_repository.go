package auth

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

// Repository 用户仓储接口
type Repository interface {
	// User operations
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
}

// GormRepository GORM 实现的用户仓储
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository 创建 GORM 仓储
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// CreateUser 创建用户
// 安全说明：
// 1. 邮箱统一转为小写存储（防止大小写绕过）
// 2. 检查邮箱唯一性
func (r *GormRepository) CreateUser(user *User) error {
	// 邮箱转小写（如果不为空）
	if user.Email != "" {
		user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	}

	// 检查邮箱是否已存在（如果不为空）
	if user.Email != "" {
		var count int64
		if err := r.db.Model(&User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrEmailExists
		}
	}

	return r.db.Create(user).Error
}

// GetUserByEmail 根据邮箱获取用户
func (r *GormRepository) GetUserByEmail(email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func (r *GormRepository) GetUserByID(id int) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
