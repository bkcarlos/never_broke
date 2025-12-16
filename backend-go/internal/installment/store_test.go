package installment

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPlansByAccount(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []InstallmentPlan{
		{ID: "plan1", AccountID: "acc1", Principal: 10000},
		{ID: "plan2", AccountID: "acc2", Principal: 20000},
		{ID: "plan3", AccountID: "acc1", Principal: 15000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("account with multiple plans", func(t *testing.T) {
		result := ListPlansByAccount("acc1")
		assert.Len(t, result, 2)
		assert.Equal(t, "plan1", result[0].ID)
		assert.Equal(t, "plan3", result[1].ID)
	})

	t.Run("account with one plan", func(t *testing.T) {
		result := ListPlansByAccount("acc2")
		assert.Len(t, result, 1)
		assert.Equal(t, "plan2", result[0].ID)
	})

	t.Run("account with no plans", func(t *testing.T) {
		result := ListPlansByAccount("acc3")
		assert.Len(t, result, 0)
	})
}

func TestAddPlan(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []InstallmentPlan{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	newPlan := InstallmentPlan{
		ID:         "new-plan",
		AccountID:  "acc1",
		Principal:  10000,
		TermMonths: 12,
		Status:     InstallmentStatusActive,
	}

	result := AddPlan(newPlan)
	assert.Equal(t, "new-plan", result.ID)

	// Verify it was added
	list := ListPlansByAccount("acc1")
	assert.Len(t, list, 1)
	assert.Equal(t, "new-plan", list[0].ID)
}

func TestFindPlanByID(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []InstallmentPlan{
		{ID: "plan1", AccountID: "acc1", Principal: 10000},
		{ID: "plan2", AccountID: "acc2", Principal: 20000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("found", func(t *testing.T) {
		plan, found := FindPlanByID("plan1")
		require.True(t, found)
		assert.Equal(t, "plan1", plan.ID)
		assert.Equal(t, "acc1", plan.AccountID)
		assert.Equal(t, 10000.0, plan.Principal)
	})

	t.Run("not found", func(t *testing.T) {
		_, found := FindPlanByID("non-existent")
		assert.False(t, found)
	})
}

func TestConcurrentInstallmentAccess(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []InstallmentPlan{
		{ID: "concurrent-test", AccountID: "acc1", Principal: 10000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = FindPlanByID("concurrent-test")
			_ = ListPlansByAccount("acc1")
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			AddPlan(InstallmentPlan{
				ID:        "plan-" + string(rune(idx)),
				AccountID: "acc2",
				Principal: float64(idx * 1000),
			})
		}(i)
	}

	wg.Wait()

	// Verify plans were added
	list := ListPlansByAccount("acc2")
	assert.NotEmpty(t, list)
}
