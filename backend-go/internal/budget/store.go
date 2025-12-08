package budget

import "sync"

var (
	mu      sync.RWMutex
	records = map[string]MonthlyBudget{} // key: month string
)

// GetMonthlyBudget 获取月度预算
func GetMonthlyBudget(month string) (MonthlyBudget, bool) {
	mu.RLock()
	defer mu.RUnlock()
	budget, exists := records[month]
	return budget, exists
}

// UpsertMonthlyBudget 更新或插入月度预算
func UpsertMonthlyBudget(b MonthlyBudget) MonthlyBudget {
	mu.Lock()
	defer mu.Unlock()
	records[b.Month] = b
	return b
}
