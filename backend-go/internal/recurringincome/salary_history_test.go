package recurringincome

import (
	"github.com/stretchr/testify/require"
	"neverbroke/internal/payroll"
	"testing"
)

func TestSalaryHistoryRaisePolicyAndAnnualTax(t *testing.T) {
	v := UpsertSalaryPlanRequest{StartDate: "2026-01-01", PreTaxMonthly: 40000, PayDay: 31, CityCode: "beijing", AccountID: "a", Currency: "CNY", BaseMode: "policy", Contributions: &SalaryContributions{SocialBase: 35811, HousingBase: 35811, PensionRate: .08, MedicalRate: .02, UnemploymentRate: .005, HousingRate: .12}}
	versions, err := MergeSalaryVersion(nil, v)
	require.NoError(t, err)
	raise := v
	raise.StartDate = "2026-07-01"
	raise.PreTaxMonthly = 50000
	raise.AccountID = "b"
	versions, err = MergeSalaryVersion(versions, raise)
	require.NoError(t, err)
	items, err := PreviewSalaryHistory(versions, 2026, payroll.NewService(nil).GetRegionRuleAt)
	require.NoError(t, err)
	require.Len(t, items, 12)
	require.Equal(t, "2026-02-28", items[1].Date)
	require.Equal(t, 40000.0, items[5].PreTax)
	require.Equal(t, 50000.0, items[6].PreTax)
	require.Equal(t, 35811.0, items[5].Configuration.SocialBase)
	require.Equal(t, 36348.0, items[6].Configuration.SocialBase)
	require.Equal(t, "a", items[5].Config.AccountID)
	require.Equal(t, "b", items[6].Config.AccountID)
	require.Equal(t, 290000.0, items[6].CumulativePreTax)
	require.Equal(t, roundSalary(items[5].CumulativeTax+items[6].TaxAmount), items[6].CumulativeTax)
	require.Greater(t, items[6].TaxAmount, 5000.0)
	for _, i := range items {
		require.InDelta(t, i.PreTax, i.SocialSecurity+i.TaxAmount+i.PostTax, .001)
		require.Empty(t, i.Warning)
	}
	next, err := PreviewSalaryHistory(versions, 2027, payroll.NewService(nil).GetRegionRuleAt)
	require.NoError(t, err)
	require.Equal(t, 50000.0, next[0].CumulativePreTax)
	require.NotEmpty(t, next[6].Warning)
	// A correction updates one month boundary, preserving the other version.
	raise.PreTaxMonthly = 51000
	versions, err = MergeSalaryVersion(versions, raise)
	require.NoError(t, err)
	require.Len(t, versions, 2)
	require.Equal(t, 40000.0, versions[0].PreTaxMonthly)
}

func TestSalaryHistoryFixedBasesNextMonthAndDeductions(t *testing.T) {
	v := UpsertSalaryPlanRequest{StartDate: "2025-12-01", PreTaxMonthly: 20000, PayDay: 10, PayMonthOffset: 1, Currency: "CNY", BaseMode: "fixed", Contributions: &SalaryContributions{SocialBase: 10000, HousingBase: 10000, PensionRate: .08, HousingRate: .12}}
	later := v
	later.StartDate = "2026-07-01"
	later.PreTaxMonthly = 30000
	items, err := PreviewSalaryHistory([]UpsertSalaryPlanRequest{v, later}, 2026, payroll.NewService(nil).GetRegionRuleAt)
	require.NoError(t, err)
	require.Equal(t, "2025-12", items[0].SalaryMonth)
	require.Equal(t, "2026-01-10", items[0].Date)
	require.Equal(t, 20000.0, items[6].PreTax)
	require.Equal(t, 30000.0, items[7].PreTax)
	require.Equal(t, 2000.0, items[7].SocialSecurity)
	// Negative monthly taxable income must reduce YTD taxable income without refunding tax.
	low := v
	low.StartDate = "2026-01-01"
	low.PayMonthOffset = 0
	low.PreTaxMonthly = 3000
	low.Contributions = &SalaryContributions{}
	high := low
	high.StartDate = "2026-02-01"
	high.PreTaxMonthly = 10000
	items, err = PreviewSalaryHistory([]UpsertSalaryPlanRequest{low, high}, 2026, payroll.NewService(nil).GetRegionRuleAt)
	require.NoError(t, err)
	require.Equal(t, 90.0, items[1].TaxAmount)
	high.PayMonthOffset = 1
	_, err = MergeSalaryVersion([]UpsertSalaryPlanRequest{low}, high)
	require.Error(t, err)
}
