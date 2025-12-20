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
