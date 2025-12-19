package auth

import (
	"sync"
	"time"
)

// RateLimiter 限流器（基于滑动窗口）
// 安全说明：
// 1. 按 IP + Email 限流（防止暴力破解）
// 2. 使用滑动窗口算法（更精确）
// 3. 内存实现（简单高效，重启后重置）
// 4. 生产环境建议使用 Redis 实现（分布式支持）
type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*attemptRecord

	// 配置
	maxAttempts int           // 最大尝试次数
	window      time.Duration // 时间窗口
}

type attemptRecord struct {
	timestamps []time.Time
}

// NewRateLimiter 创建限流器
// maxAttempts: 时间窗口内最大尝试次数
// window: 时间窗口（例如 5 分钟）
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts:    make(map[string]*attemptRecord),
		maxAttempts: maxAttempts,
		window:      window,
	}

	// 启动清理协程（每 10 分钟清理一次过期记录）
	go rl.cleanupLoop()

	return rl
}

// Check 检查是否允许请求（不记录）
// key: 通常是 "ip:email" 格式
func (rl *RateLimiter) Check(key string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	record, exists := rl.attempts[key]
	if !exists {
		return true
	}

	// 计算有效的时间戳数量
	validCount := 0
	for _, ts := range record.timestamps {
		if ts.After(cutoff) {
			validCount++
		}
	}

	return validCount < rl.maxAttempts
}

// RecordFailure 记录失败的尝试
func (rl *RateLimiter) RecordFailure(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 获取或创建记录
	record, exists := rl.attempts[key]
	if !exists {
		record = &attemptRecord{
			timestamps: []time.Time{},
		}
		rl.attempts[key] = record
	}

	// 移除过期的时间戳（滑动窗口）
	validTimestamps := []time.Time{}
	for _, ts := range record.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	record.timestamps = validTimestamps

	// 记录本次失败尝试
	record.timestamps = append(record.timestamps, now)
}

// Allow 检查并记录（已废弃，保留用于兼容）
// key: 通常是 "ip:email" 格式
func (rl *RateLimiter) Allow(key string) bool {
	return rl.Check(key)
}

// Reset 重置某个 key 的限流记录（登录成功后调用）
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// cleanupLoop 定期清理过期记录
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup 清理过期记录
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for key, record := range rl.attempts {
		// 移除过期的时间戳
		validTimestamps := []time.Time{}
		for _, ts := range record.timestamps {
			if ts.After(cutoff) {
				validTimestamps = append(validTimestamps, ts)
			}
		}

		// 如果没有有效时间戳，删除记录
		if len(validTimestamps) == 0 {
			delete(rl.attempts, key)
		} else {
			record.timestamps = validTimestamps
		}
	}
}
