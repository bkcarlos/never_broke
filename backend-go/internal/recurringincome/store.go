package recurringincome

import "sync"

var (
	mu    sync.RWMutex
	plans = []RecurringIncomePlan{}
)

// ListPlans 列出所有周期收入计划
func ListPlans() []RecurringIncomePlan {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]RecurringIncomePlan, len(plans))
	copy(result, plans)
	return result
}

// SavePlan 保存周期收入计划（新建或更新）
func SavePlan(p RecurringIncomePlan) RecurringIncomePlan {
	mu.Lock()
	defer mu.Unlock()

	for i, existing := range plans {
		if existing.ID == p.ID {
			plans[i] = p
			return p
		}
	}
	plans = append(plans, p)
	return p
}

// GetPlanByID 根据 ID 获取计划
func GetPlanByID(id string) *RecurringIncomePlan {
	mu.RLock()
	defer mu.RUnlock()

	for i := range plans {
		if plans[i].ID == id {
			result := plans[i]
			return &result
		}
	}
	return nil
}
