package incomeplan

import "sync"

var (
	mu    sync.RWMutex
	plans = []IncomePlan{}
)

// ListPlans 列出所有收入计划
func ListPlans() []IncomePlan {
	mu.RLock()
	defer mu.RUnlock()
	copied := make([]IncomePlan, len(plans))
	copy(copied, plans)
	return copied
}

// AddPlan 添加收入计划
func AddPlan(p IncomePlan) IncomePlan {
	mu.Lock()
	defer mu.Unlock()
	plans = append(plans, p)
	return p
}

// UpdatePlan 更新收入计划
func UpdatePlan(p IncomePlan) (IncomePlan, bool) {
	mu.Lock()
	defer mu.Unlock()
	for i := range plans {
		if plans[i].ID == p.ID {
			plans[i] = p
			return plans[i], true
		}
	}
	return IncomePlan{}, false
}

// DeletePlan 删除收入计划
func DeletePlan(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	for i := range plans {
		if plans[i].ID == id {
			plans = append(plans[:i], plans[i+1:]...)
			return true
		}
	}
	return false
}






