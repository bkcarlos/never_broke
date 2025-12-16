package budget

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMonthlyBudget(t *testing.T) {
	mu.Lock()
	originalRecords := records
	records = map[string]MonthlyBudget{
		"2024-01": {Month: "2024-01", TotalBudget: 5000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		records = originalRecords
		mu.Unlock()
	}()

	t.Run("existing budget", func(t *testing.T) {
		budget, exists := GetMonthlyBudget("2024-01")
		require.True(t, exists)
		assert.Equal(t, "2024-01", budget.Month)
		assert.Equal(t, 5000.0, budget.TotalBudget)
	})

	t.Run("non-existing budget", func(t *testing.T) {
		_, exists := GetMonthlyBudget("2024-02")
		assert.False(t, exists)
	})
}

func TestUpsertMonthlyBudget(t *testing.T) {
	mu.Lock()
	originalRecords := records
	records = map[string]MonthlyBudget{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		records = originalRecords
		mu.Unlock()
	}()

	t.Run("insert new budget", func(t *testing.T) {
		newBudget := MonthlyBudget{
			Month:       "2024-01",
			TotalBudget: 6000,
		}

		result := UpsertMonthlyBudget(newBudget)
		assert.Equal(t, "2024-01", result.Month)
		assert.Equal(t, 6000.0, result.TotalBudget)

		// Verify it was stored
		stored, exists := GetMonthlyBudget("2024-01")
		require.True(t, exists)
		assert.Equal(t, 6000.0, stored.TotalBudget)
	})

	t.Run("update existing budget", func(t *testing.T) {
		// First insert
		UpsertMonthlyBudget(MonthlyBudget{Month: "2024-02", TotalBudget: 5000})

		// Then update
		updated := MonthlyBudget{Month: "2024-02", TotalBudget: 7000}
		result := UpsertMonthlyBudget(updated)
		assert.Equal(t, 7000.0, result.TotalBudget)

		// Verify update
		stored, exists := GetMonthlyBudget("2024-02")
		require.True(t, exists)
		assert.Equal(t, 7000.0, stored.TotalBudget)
	})
}

func TestConcurrentBudgetAccess(t *testing.T) {
	mu.Lock()
	originalRecords := records
	records = map[string]MonthlyBudget{
		"2024-01": {Month: "2024-01", TotalBudget: 5000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		records = originalRecords
		mu.Unlock()
	}()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = GetMonthlyBudget("2024-01")
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			month := "2024-03"
			UpsertMonthlyBudget(MonthlyBudget{
				Month:       month,
				TotalBudget: float64(idx * 100),
			})
		}(i)
	}

	wg.Wait()

	// Verify the budget exists (last write wins)
	_, exists := GetMonthlyBudget("2024-03")
	assert.True(t, exists)
}
