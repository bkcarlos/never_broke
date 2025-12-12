package budget

// MonthlyBudget 月度预算
type MonthlyBudget struct {
	Month       string  `json:"month"`       // 'YYYY-MM'
	TotalBudget float64 `json:"totalBudget"` // 当月总预算
}

