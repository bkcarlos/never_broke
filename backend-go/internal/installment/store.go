package installment

import "sync"

var (
	mu    sync.RWMutex
	plans = []InstallmentPlan{}
)

// ListPlansByAccount 根据账户ID列出分期计划
func ListPlansByAccount(accountID string) []InstallmentPlan {
	mu.RLock()
	defer mu.RUnlock()
	var result []InstallmentPlan
	for _, p := range plans {
		if p.AccountID == accountID {
			result = append(result, p)
		}
	}
	return result
}

// AddPlan 添加分期计划
func AddPlan(p InstallmentPlan) InstallmentPlan {
	mu.Lock()
	defer mu.Unlock()
	plans = append(plans, p)
	return p
}

// FindPlanByID 根据ID查找分期计划
func FindPlanByID(id string) (*InstallmentPlan, bool) {
	mu.RLock()
	defer mu.RUnlock()
	for i := range plans {
		if plans[i].ID == id {
			plan := plans[i]
			return &plan, true
		}
	}
	return nil, false
}
