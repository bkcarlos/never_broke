package memory

import (
	"neverbroke/internal/storage/models"
	"sync"
)

// Store 内存存储（临时实现，后续可替换为数据库）
type Store struct {
	mu           sync.RWMutex
	userSettings map[string]*models.UserSettings
}

// NewStore 创建新的内存存储
func NewStore() *Store {
	return &Store{
		userSettings: make(map[string]*models.UserSettings),
	}
}

// GetUserSettings 获取用户设置
func (s *Store) GetUserSettings(userID string) (*models.UserSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if settings, ok := s.userSettings[userID]; ok {
		return settings, nil
	}

	// 返回默认设置
	return &models.UserSettings{
		Locale:     "zh-CN",
		ShowAmount: true,
	}, nil
}

// UpdateUserSettings 更新用户设置
func (s *Store) UpdateUserSettings(userID string, settings *models.UserSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userSettings[userID] = settings
	return nil
}

