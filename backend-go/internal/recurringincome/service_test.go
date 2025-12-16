package recurringincome

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindDuePlans(t *testing.T) {
	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{
		{ID: "plan1", NextDueDate: "2024-01-10", Active: true},
		{ID: "plan2", NextDueDate: "2024-01-20", Active: true},
		{ID: "plan3", NextDueDate: "2024-01-05", Active: true},
		{ID: "plan4", NextDueDate: "2024-01-15", Active: false}, // Inactive
		{ID: "plan5", NextDueDate: "", Active: true},            // No due date
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("find plans due on exact date", func(t *testing.T) {
		result := FindDuePlans("2024-01-10")
		assert.Len(t, result, 2) // plan1 and plan3
		ids := []string{result[0].ID, result[1].ID}
		assert.Contains(t, ids, "plan1")
		assert.Contains(t, ids, "plan3")
	})

	t.Run("find plans due before date", func(t *testing.T) {
		result := FindDuePlans("2024-01-25")
		assert.Len(t, result, 3) // plan1, plan2, plan3
	})

	t.Run("no plans due", func(t *testing.T) {
		result := FindDuePlans("2024-01-01")
		assert.Len(t, result, 0)
	})

	t.Run("invalid date format", func(t *testing.T) {
		result := FindDuePlans("invalid-date")
		assert.Nil(t, result)
	})
}

func TestCalcNextDueDate(t *testing.T) {
	tests := []struct {
		name          string
		plan          RecurringIncomePlan
		expectedDelta time.Duration
	}{
		{
			name: "monthly",
			plan: RecurringIncomePlan{
				Frequency:   FrequencyMonthly,
				NextDueDate: "2024-01-15",
			},
			expectedDelta: 30 * 24 * time.Hour, // Approximately 1 month
		},
		{
			name: "bimonthly",
			plan: RecurringIncomePlan{
				Frequency:   FrequencyBimonthly,
				NextDueDate: "2024-01-15",
			},
			expectedDelta: 60 * 24 * time.Hour, // Approximately 2 months
		},
		{
			name: "quarterly",
			plan: RecurringIncomePlan{
				Frequency:   FrequencyQuarterly,
				NextDueDate: "2024-01-15",
			},
			expectedDelta: 90 * 24 * time.Hour, // Approximately 3 months
		},
		{
			name: "semiannual",
			plan: RecurringIncomePlan{
				Frequency:   FrequencySemiannual,
				NextDueDate: "2024-01-15",
			},
			expectedDelta: 180 * 24 * time.Hour, // Approximately 6 months
		},
		{
			name: "weekly",
			plan: RecurringIncomePlan{
				Frequency:   FrequencyWeekly,
				NextDueDate: "2024-01-15",
			},
			expectedDelta: 7 * 24 * time.Hour,
		},
		{
			name: "yearly",
			plan: RecurringIncomePlan{
				Frequency:   FrequencyYearly,
				NextDueDate: "2024-01-15",
			},
			expectedDelta: 365 * 24 * time.Hour, // Approximately 1 year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextDate := CalcNextDueDate(tt.plan)
			require.NotEmpty(t, nextDate)

			current, err := time.Parse("2006-01-02", tt.plan.NextDueDate)
			require.NoError(t, err)

			next, err := time.Parse("2006-01-02", nextDate)
			require.NoError(t, err)

			delta := next.Sub(current)
			// Allow some tolerance for month length variations
			assert.InDelta(t, tt.expectedDelta.Hours(), delta.Hours(), 72) // ±3 days tolerance
		})
	}

	t.Run("empty next due date", func(t *testing.T) {
		plan := RecurringIncomePlan{
			Frequency:   FrequencyMonthly,
			NextDueDate: "",
		}
		result := CalcNextDueDate(plan)
		assert.Empty(t, result)
	})

	t.Run("invalid date format", func(t *testing.T) {
		plan := RecurringIncomePlan{
			Frequency:   FrequencyMonthly,
			NextDueDate: "invalid-date",
		}
		result := CalcNextDueDate(plan)
		assert.Empty(t, result)
	})

	t.Run("unknown frequency", func(t *testing.T) {
		plan := RecurringIncomePlan{
			Frequency:   "unknown",
			NextDueDate: "2024-01-15",
		}
		result := CalcNextDueDate(plan)
		// Should return the same date for unknown frequency
		assert.Equal(t, "2024-01-15", result)
	})
}

func TestCalcNextDueDate_MonthlyProgression(t *testing.T) {
	plan := RecurringIncomePlan{
		Frequency:   FrequencyMonthly,
		NextDueDate: "2024-01-31",
	}

	// Test progression through months with different days
	nextDate := CalcNextDueDate(plan)
	// February doesn't have 31 days, so it should be adjusted
	assert.NotEmpty(t, nextDate)

	// Parse and verify it's in February or March (depending on how AddDate handles it)
	next, err := time.Parse("2006-01-02", nextDate)
	require.NoError(t, err)
	// Go's AddDate will overflow to March for Jan 31 + 1 month
	assert.True(t, next.Month() == time.February || next.Month() == time.March)
}

func TestCalcNextDueDate_YearBoundary(t *testing.T) {
	plan := RecurringIncomePlan{
		Frequency:   FrequencyMonthly,
		NextDueDate: "2024-12-15",
	}

	nextDate := CalcNextDueDate(plan)
	next, err := time.Parse("2006-01-02", nextDate)
	require.NoError(t, err)

	assert.Equal(t, 2025, next.Year())
	assert.Equal(t, time.January, next.Month())
	assert.Equal(t, 15, next.Day())
}
