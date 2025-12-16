package recurringincome

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPlans(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{
		{ID: "plan1", Amount: 10000},
		{ID: "plan2", Amount: 20000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	result := ListPlans()
	assert.Len(t, result, 2)
	assert.Equal(t, "plan1", result[0].ID)
	assert.Equal(t, "plan2", result[1].ID)
}

func TestSavePlan(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("create new plan", func(t *testing.T) {
		newPlan := RecurringIncomePlan{
			ID:     "new-plan",
			Amount: 15000,
			Active: true,
		}

		result := SavePlan(newPlan)
		assert.Equal(t, "new-plan", result.ID)

		// Verify it was saved
		list := ListPlans()
		assert.Len(t, list, 1)
		assert.Equal(t, "new-plan", list[0].ID)
	})

	t.Run("update existing plan", func(t *testing.T) {
		// First create
		SavePlan(RecurringIncomePlan{ID: "update-plan", Amount: 10000})

		// Then update
		updated := RecurringIncomePlan{ID: "update-plan", Amount: 20000}
		result := SavePlan(updated)
		assert.Equal(t, 20000.0, result.Amount)

		// Verify update
		list := ListPlans()
		assert.Len(t, list, 2) // new-plan + update-plan
		found := false
		for _, p := range list {
			if p.ID == "update-plan" {
				assert.Equal(t, 20000.0, p.Amount)
				found = true
			}
		}
		assert.True(t, found)
	})
}

func TestGetPlanByID(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{
		{ID: "plan1", Amount: 10000},
		{ID: "plan2", Amount: 20000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("found", func(t *testing.T) {
		plan := GetPlanByID("plan1")
		require.NotNil(t, plan)
		assert.Equal(t, "plan1", plan.ID)
		assert.Equal(t, 10000.0, plan.Amount)
	})

	t.Run("not found", func(t *testing.T) {
		plan := GetPlanByID("non-existent")
		assert.Nil(t, plan)
	})
}

func TestConcurrentRecurringIncomeAccess(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{
		{ID: "concurrent-test", Amount: 10000},
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
			_ = GetPlanByID("concurrent-test")
			_ = ListPlans()
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			SavePlan(RecurringIncomePlan{
				ID:     "plan-concurrent",
				Amount: float64(idx * 1000),
			})
		}(i)
	}

	wg.Wait()

	// Verify plan exists
	plan := GetPlanByID("plan-concurrent")
	assert.NotNil(t, plan)
}
