package recurringincome

import (
	"neverbroke/internal/citysocialrate"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// 初始化社保费率数据
	citysocialrate.InitDefaultRates()

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

func TestCalculateTax(t *testing.T) {
	tests := []struct {
		name                    string
		cumulativeTaxableIncome float64
		expectedTaxApprox       float64
	}{
		{
			name:                    "below first bracket",
			cumulativeTaxableIncome: 10000,
			expectedTaxApprox:       300, // 10000 * 0.03
		},
		{
			name:                    "in second bracket",
			cumulativeTaxableIncome: 100000,
			expectedTaxApprox:       7480, // 100000 * 0.10 - 2520
		},
		{
			name:                    "in third bracket",
			cumulativeTaxableIncome: 200000,
			expectedTaxApprox:       23080, // 200000 * 0.20 - 16920
		},
		{
			name:                    "zero income",
			cumulativeTaxableIncome: 0,
			expectedTaxApprox:       0,
		},
		{
			name:                    "exactly at bracket threshold",
			cumulativeTaxableIncome: 36000,
			expectedTaxApprox:       1080, // 36000 * 0.03
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tax := CalculateTax(tt.cumulativeTaxableIncome)
			assert.InDelta(t, tt.expectedTaxApprox, tax, 1.0)
		})
	}
}

func TestGetSocialSecurityRate(t *testing.T) {
	tests := []struct {
		name         string
		cityCode     string
		expectedRate float64
	}{
		{"beijing", "beijing", 0.202},
		{"shanghai", "shanghai", 0.175},
		{"guangzhou", "guangzhou", 0.152},
		{"shenzhen", "shenzhen", 0.155},
		{"unknown city", "unknown", 0.182}, // default
		{"empty city", "", 0.182},          // default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := GetSocialSecurityRate(tt.cityCode)
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

func TestMonthlyDeduction(t *testing.T) {
	// Just verify the constant is set correctly
	assert.Equal(t, 5000.0, MonthlyDeduction)
}

func TestCalculateTax_ProgressiveBrackets(t *testing.T) {
	// Test that tax increases progressively
	income1 := 30000.0
	income2 := 100000.0
	income3 := 500000.0

	tax1 := CalculateTax(income1)
	tax2 := CalculateTax(income2)
	tax3 := CalculateTax(income3)

	// Higher income should have higher tax
	assert.Less(t, tax1, tax2)
	assert.Less(t, tax2, tax3)

	// Effective tax rate should increase
	effectiveRate1 := tax1 / income1
	effectiveRate2 := tax2 / income2
	effectiveRate3 := tax3 / income3

	assert.Less(t, effectiveRate1, effectiveRate2)
	assert.Less(t, effectiveRate2, effectiveRate3)
}

func TestCalculateTax_NegativeIncome(t *testing.T) {
	// Test edge case of negative income
	tax := CalculateTax(-1000)
	// Should return negative tax (which doesn't make sense in real life, but tests the function)
	assert.Less(t, tax, 0.0)
}

func TestCalculateTax_VeryHighIncome(t *testing.T) {
	// Test highest bracket
	income := 2000000.0 // 2 million
	tax := CalculateTax(income)

	// Should use highest rate (45%)
	expectedTax := income*0.45 - 181920
	assert.InDelta(t, expectedTax, tax, 1.0)
}

func TestSalaryPreviewMonthEndAndAnnualReset(t *testing.T) {
	rate := 0.2
	req := SalaryPreviewRequest{PreTaxAmount: 20000, StartDate: "2024-01-31", PayDay: 31, Frequency: FrequencyMonthly, SocialRate: &rate}
	preview, err := PreviewSalaryYear(req, 2024)
	assert.NoError(t, err)
	assert.Len(t, preview.Items, 12)
	assert.Equal(t, "2024-02-29", preview.Items[1].Date)
	assert.Equal(t, "2024-04-30", preview.Items[3].Date)
	assert.Equal(t, 330.0, preview.Items[0].TaxAmount)
	assert.Greater(t, preview.Items[11].TaxAmount, preview.Items[0].TaxAmount)
	next, err := PreviewSalaryYear(req, 2025)
	assert.NoError(t, err)
	assert.Equal(t, "2025-02-28", next.Items[1].Date)
	assert.Equal(t, preview.Items[0].TaxAmount, next.Items[0].TaxAmount)
	assert.Equal(t, 20000.0, next.Items[0].CumulativePreTax)

	req.StartDate = "2024-12-20"
	req.PayDay = 10
	preview, err = PreviewSalaryToYearEnd(req)
	assert.NoError(t, err)
	assert.Equal(t, 2025, preview.Year)
	assert.Equal(t, "2025-01-10", preview.Items[0].Date)
}

func TestSalaryPreviewItemizedContributionsAndReconciliation(t *testing.T) {
	req := SalaryPreviewRequest{PreTaxAmount: 30000.01, StartDate: "2026-06-15", PayDay: 15, Frequency: FrequencyMonthly, ExtraDeductions: 2000,
		Contributions: &SalaryContributions{SocialBase: 25000.05, HousingBase: 28000.09, PensionRate: 0.08, MedicalRate: 0.02, UnemploymentRate: 0.002, HousingRate: 0.12}}
	preview, err := PreviewSalaryYear(req, 2026)
	assert.NoError(t, err)
	assert.Len(t, preview.Items, 7)
	first := preview.Items[0]
	assert.Equal(t, "2026-06-15", first.Date)
	assert.Equal(t, 2000.0, first.Contributions.Pension)
	assert.Equal(t, 500.0, first.Contributions.Medical)
	assert.Equal(t, 50.0, first.Contributions.Unemployment)
	assert.Equal(t, 3360.01, first.Contributions.Housing)
	assert.Equal(t, 5910.01, first.SocialSecurity)
	assert.Zero(t, first.Contributions.Injury)
	assert.Zero(t, first.Contributions.Maternity)
	assert.Equal(t, 17090.0, first.TaxableIncome)
	assert.Equal(t, 512.7, first.TaxAmount)
	for _, item := range preview.Items {
		assert.InDelta(t, item.PreTax, item.PostTax+item.SocialSecurity+item.TaxAmount, 0.00001)
	}
	req.Contributions = nil
	rate := 0.1
	req.SocialRate = &rate
	legacy, err := PreviewSalaryYear(req, 2026)
	assert.NoError(t, err)
	assert.Nil(t, legacy.Items[0].Contributions)
	assert.Equal(t, 3000.0, legacy.Items[0].SocialSecurity)
	req.ExtraDeductions = 50000
	lowTax, err := PreviewSalaryYear(req, 2026)
	assert.NoError(t, err)
	assert.Zero(t, lowTax.Items[0].TaxAmount)
	assert.Equal(t, 27000.01, lowTax.Items[0].PostTax)
}

func TestSalaryPreviewRejectsInvalidConfigurations(t *testing.T) {
	base := SalaryPreviewRequest{PreTaxAmount: 10000, StartDate: "2026-01-01", PayDay: 10, Frequency: FrequencyMonthly}
	for _, change := range []func(*SalaryPreviewRequest){
		func(r *SalaryPreviewRequest) { r.PayDay = 0 },
		func(r *SalaryPreviewRequest) { r.PayDay = 32 },
		func(r *SalaryPreviewRequest) { r.PreTaxAmount = -1 },
		func(r *SalaryPreviewRequest) { r.ExtraDeductions = -1 },
		func(r *SalaryPreviewRequest) { r.StartDate = "invalid" },
		func(r *SalaryPreviewRequest) { r.Contributions = &SalaryContributions{SocialBase: -1} },
		func(r *SalaryPreviewRequest) { r.Contributions = &SalaryContributions{PensionRate: 1.1} },
		func(r *SalaryPreviewRequest) {
			r.Contributions = &SalaryContributions{SocialBase: 20000, PensionRate: 1}
		},
	} {
		req := base
		change(&req)
		_, err := PreviewSalaryYear(req, 2026)
		assert.Error(t, err)
	}
}

func TestSalaryPreviewCadence(t *testing.T) {
	rate := 0.0
	req := SalaryPreviewRequest{PreTaxAmount: 10000, StartDate: "2026-11-30", PayDay: 31, Frequency: FrequencyQuarterly, SocialRate: &rate}
	preview, err := PreviewSalaryYear(req, 2027)
	assert.NoError(t, err)
	assert.Len(t, preview.Items, 4)
	assert.Equal(t, "2027-02-28", preview.Items[0].Date)
	assert.Equal(t, "2027-05-31", preview.Items[1].Date)
	req.Frequency = FrequencyWeekly
	req.StartDate = "2026-12-25"
	preview, err = PreviewSalaryYear(req, 2027)
	assert.NoError(t, err)
	assert.Equal(t, "2027-01-01", preview.Items[0].Date)
	assert.Equal(t, "2027-01-08", preview.Items[1].Date)
}
