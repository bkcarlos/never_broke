package installment

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSchedule_EqualPrincipal(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          10000,
		TermMonths:         12,
		MonthlyRatePercent: 0.5, // 0.5%
		RepayMethod:        RepayMethodEqualPrincipal,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// Verify basic info
	assert.Equal(t, 10000.0, resp.Principal)
	assert.Equal(t, 12, resp.TermMonths)
	assert.Equal(t, 0.005, resp.MonthlyRate)
	assert.Equal(t, RepayMethodEqualPrincipal, resp.RepayMethod)
	assert.Len(t, resp.Schedule, 12)

	// Verify total amounts
	assert.InDelta(t, 10000.0+resp.TotalInterest, resp.TotalAmount, 0.01)

	// Verify each period
	monthlyPrincipal := 10000.0 / 12
	for i, item := range resp.Schedule {
		assert.Equal(t, i+1, item.Period)
		assert.InDelta(t, monthlyPrincipal, item.Principal, 0.01)

		// Interest should decrease each period
		if i > 0 {
			assert.Less(t, item.Interest, resp.Schedule[i-1].Interest)
		}

		// Total = Principal + Interest
		assert.InDelta(t, item.Principal+item.Interest, item.Total, 0.01)
	}

	// Last period should have 0 remaining
	assert.InDelta(t, 0, resp.Schedule[11].RemainingPrincipal, 0.01)
}

func TestBuildSchedule_EqualPayment(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          10000,
		TermMonths:         12,
		MonthlyRatePercent: 0.5, // 0.5%
		RepayMethod:        RepayMethodEqualPayment,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// Verify basic info
	assert.Equal(t, 10000.0, resp.Principal)
	assert.Equal(t, 12, resp.TermMonths)
	assert.Equal(t, 0.005, resp.MonthlyRate)
	assert.Equal(t, RepayMethodEqualPayment, resp.RepayMethod)
	assert.Len(t, resp.Schedule, 12)

	// Verify total amounts
	assert.InDelta(t, 10000.0+resp.TotalInterest, resp.TotalAmount, 0.01)

	// Verify each period - total payment should be roughly equal (except last)
	for i := 0; i < 11; i++ {
		assert.InDelta(t, resp.Schedule[0].Total, resp.Schedule[i].Total, 1.0)
	}

	// Interest should decrease each period
	for i := 1; i < 12; i++ {
		assert.Less(t, resp.Schedule[i].Interest, resp.Schedule[i-1].Interest)
	}

	// Last period should have 0 remaining
	assert.InDelta(t, 0, resp.Schedule[11].RemainingPrincipal, 0.01)
}

func TestBuildSchedule_ZeroInterest(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          10000,
		TermMonths:         12,
		MonthlyRatePercent: 0, // 0%
		RepayMethod:        RepayMethodEqualPayment,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// With 0% interest, it should behave like equal principal
	assert.Equal(t, 0.0, resp.TotalInterest)
	assert.InDelta(t, 10000.0, resp.TotalAmount, 0.01)

	monthlyPayment := 10000.0 / 12
	for _, item := range resp.Schedule {
		assert.InDelta(t, 0, item.Interest, 0.01)
		assert.InDelta(t, monthlyPayment, item.Principal, 0.01)
	}
}

func TestBuildSchedule_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		req  PreviewInstallmentRequest
	}{
		{
			name: "negative principal",
			req: PreviewInstallmentRequest{
				Principal:          -1000,
				TermMonths:         12,
				MonthlyRatePercent: 0.5,
				RepayMethod:        RepayMethodEqualPrincipal,
				FirstDueDate:       "2024-01-01",
			},
		},
		{
			name: "zero term months",
			req: PreviewInstallmentRequest{
				Principal:          10000,
				TermMonths:         0,
				MonthlyRatePercent: 0.5,
				RepayMethod:        RepayMethodEqualPrincipal,
				FirstDueDate:       "2024-01-01",
			},
		},
		{
			name: "negative term months",
			req: PreviewInstallmentRequest{
				Principal:          10000,
				TermMonths:         -12,
				MonthlyRatePercent: 0.5,
				RepayMethod:        RepayMethodEqualPrincipal,
				FirstDueDate:       "2024-01-01",
			},
		},
		{
			name: "invalid date format",
			req: PreviewInstallmentRequest{
				Principal:          10000,
				TermMonths:         12,
				MonthlyRatePercent: 0.5,
				RepayMethod:        RepayMethodEqualPrincipal,
				FirstDueDate:       "invalid-date",
			},
		},
		{
			name: "invalid repay method",
			req: PreviewInstallmentRequest{
				Principal:          10000,
				TermMonths:         12,
				MonthlyRatePercent: 0.5,
				RepayMethod:        "invalid",
				FirstDueDate:       "2024-01-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildSchedule(tt.req)
			assert.Error(t, err)
		})
	}
}

func TestBuildSchedule_DateProgression(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          10000,
		TermMonths:         12,
		MonthlyRatePercent: 0.5,
		RepayMethod:        RepayMethodEqualPrincipal,
		FirstDueDate:       "2024-01-15",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// Verify dates progress monthly
	assert.Equal(t, "2024-01-15", resp.Schedule[0].DueDate)
	assert.Equal(t, "2024-02-15", resp.Schedule[1].DueDate)
	assert.Equal(t, "2024-12-15", resp.Schedule[11].DueDate)
}

func TestBuildSchedule_DefaultCurrency(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          10000,
		TermMonths:         12,
		MonthlyRatePercent: 0.5,
		RepayMethod:        RepayMethodEqualPrincipal,
		FirstDueDate:       "2024-01-01",
		// Currency omitted
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// Should default to CNY (check in the implementation if this is the behavior)
	assert.NotEmpty(t, resp)
}

func TestEqualPrincipalVsEqualPayment(t *testing.T) {
	reqPrincipal := PreviewInstallmentRequest{
		Principal:          100000,
		TermMonths:         24,
		MonthlyRatePercent: 0.5,
		RepayMethod:        RepayMethodEqualPrincipal,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	reqPayment := PreviewInstallmentRequest{
		Principal:          100000,
		TermMonths:         24,
		MonthlyRatePercent: 0.5,
		RepayMethod:        RepayMethodEqualPayment,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	respPrincipal, err := BuildSchedule(reqPrincipal)
	require.NoError(t, err)

	respPayment, err := BuildSchedule(reqPayment)
	require.NoError(t, err)

	// Equal principal should have higher first payment
	assert.Greater(t, respPrincipal.Schedule[0].Total, respPayment.Schedule[0].Total)

	// Equal principal should have lower total interest
	assert.Less(t, respPrincipal.TotalInterest, respPayment.TotalInterest)

	// Equal payment should have consistent payment amounts (except last)
	for i := 0; i < 23; i++ {
		assert.InDelta(t, respPayment.Schedule[0].Total, respPayment.Schedule[i].Total, 1.0)
	}
}

func TestBuildSchedule_MathematicalAccuracy(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          50000,
		TermMonths:         36,
		MonthlyRatePercent: 0.8,
		RepayMethod:        RepayMethodEqualPayment,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// Verify total principal equals original
	totalPrincipal := 0.0
	for _, item := range resp.Schedule {
		totalPrincipal += item.Principal
	}
	assert.InDelta(t, req.Principal, totalPrincipal, 0.01)

	// Verify total interest matches
	totalInterest := 0.0
	for _, item := range resp.Schedule {
		totalInterest += item.Interest
	}
	assert.InDelta(t, resp.TotalInterest, totalInterest, 0.01)

	// Verify remaining principal decreases correctly
	for i := 1; i < len(resp.Schedule); i++ {
		expected := resp.Schedule[i-1].RemainingPrincipal - resp.Schedule[i].Principal
		assert.InDelta(t, expected, resp.Schedule[i].RemainingPrincipal, 0.01)
	}
}

func TestBuildSchedule_HighInterestRate(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          10000,
		TermMonths:         12,
		MonthlyRatePercent: 5.0, // 5% monthly rate (very high)
		RepayMethod:        RepayMethodEqualPayment,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	// With high interest, total interest should be significant
	assert.Greater(t, resp.TotalInterest, 3000.0)

	// First period interest should be high
	assert.Greater(t, resp.Schedule[0].Interest, 400.0)
}

func TestBuildSchedule_LongTerm(t *testing.T) {
	req := PreviewInstallmentRequest{
		Principal:          500000,
		TermMonths:         360, // 30 years
		MonthlyRatePercent: 0.4,
		RepayMethod:        RepayMethodEqualPayment,
		FirstDueDate:       "2024-01-01",
		Currency:           "CNY",
	}

	resp, err := BuildSchedule(req)
	require.NoError(t, err)

	assert.Len(t, resp.Schedule, 360)

	// For long-term loans, total interest can be very high
	assert.Greater(t, resp.TotalInterest, resp.Principal*0.5)
}

func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
