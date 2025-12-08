package installment

import (
	"errors"
	"math"
	"time"
)

var ErrInvalidInput = errors.New("invalid input")

// BuildSchedule 根据还款方式构建分期计划
func BuildSchedule(req PreviewInstallmentRequest) (PreviewInstallmentResponse, error) {
	if req.Currency == "" {
		req.Currency = "CNY"
	}
	switch req.RepayMethod {
	case RepayMethodEqualPrincipal:
		return buildEqualPrincipalSchedule(req)
	case RepayMethodEqualPayment:
		return buildEqualPaymentSchedule(req)
	default:
		return PreviewInstallmentResponse{}, ErrInvalidInput
	}
}

// buildEqualPrincipalSchedule 等额本金计算
// 每期本金 = 本金 / N
// 每期利息 = 剩余本金 × 月利率
func buildEqualPrincipalSchedule(req PreviewInstallmentRequest) (PreviewInstallmentResponse, error) {
	if req.TermMonths <= 0 || req.Principal <= 0 {
		return PreviewInstallmentResponse{}, ErrInvalidInput
	}

	monthlyRate := req.MonthlyRatePercent / 100.0
	dueDate, err := time.Parse("2006-01-02", req.FirstDueDate)
	if err != nil {
		return PreviewInstallmentResponse{}, err
	}

	monthlyPrincipal := req.Principal / float64(req.TermMonths)
	remaining := req.Principal
	var schedule []InstallmentPreviewItem
	totalInterest := 0.0

	for i := 0; i < req.TermMonths; i++ {
		interest := remaining * monthlyRate
		total := monthlyPrincipal + interest
		remainingAfter := remaining - monthlyPrincipal

		item := InstallmentPreviewItem{
			Period:             i + 1,
			DueDate:            dueDate.AddDate(0, i, 0).Format("2006-01-02"),
			Principal:          monthlyPrincipal,
			Interest:           interest,
			Total:              total,
			RemainingPrincipal: remainingAfter,
		}
		schedule = append(schedule, item)
		remaining = remainingAfter
		totalInterest += interest
	}

	resp := PreviewInstallmentResponse{
		Principal:     req.Principal,
		TotalInterest: totalInterest,
		TotalAmount:   req.Principal + totalInterest,
		TermMonths:    req.TermMonths,
		MonthlyRate:   monthlyRate,
		RepayMethod:   RepayMethodEqualPrincipal,
		Schedule:      schedule,
	}
	return resp, nil
}

// buildEqualPaymentSchedule 等额本息计算
// 每期还款额固定：A = P * r * (1 + r)^n / ((1 + r)^n - 1)
// 当期利息 = 剩余本金 × r
// 当期本金 = 当期还款额 - 利息
func buildEqualPaymentSchedule(req PreviewInstallmentRequest) (PreviewInstallmentResponse, error) {
	if req.TermMonths <= 0 || req.Principal <= 0 {
		return PreviewInstallmentResponse{}, ErrInvalidInput
	}

	monthlyRate := req.MonthlyRatePercent / 100.0
	dueDate, err := time.Parse("2006-01-02", req.FirstDueDate)
	if err != nil {
		return PreviewInstallmentResponse{}, err
	}

	n := float64(req.TermMonths)
	r := monthlyRate

	// 如果利率为0，退化为等额本金但总额均分
	if r == 0 {
		return buildEqualPrincipalSchedule(req)
	}

	r1n := math.Pow(1+r, n)
	// 每期还款额（固定）
	monthlyPayment := req.Principal * r * r1n / (r1n - 1)

	remaining := req.Principal
	var schedule []InstallmentPreviewItem
	totalInterest := 0.0

	for i := 0; i < req.TermMonths; i++ {
		interest := remaining * r
		principal := monthlyPayment - interest
		// 最后一期间对 principal/remaining 做少量误差修正
		if i == req.TermMonths-1 {
			principal = remaining
		}
		total := principal + interest
		remainingAfter := remaining - principal

		item := InstallmentPreviewItem{
			Period:             i + 1,
			DueDate:            dueDate.AddDate(0, i, 0).Format("2006-01-02"),
			Principal:          principal,
			Interest:           interest,
			Total:              total,
			RemainingPrincipal: remainingAfter,
		}
		schedule = append(schedule, item)
		remaining = remainingAfter
		totalInterest += interest
	}

	resp := PreviewInstallmentResponse{
		Principal:     req.Principal,
		TotalInterest: totalInterest,
		TotalAmount:   req.Principal + totalInterest,
		TermMonths:    req.TermMonths,
		MonthlyRate:   monthlyRate,
		RepayMethod:   RepayMethodEqualPayment,
		Schedule:      schedule,
	}
	return resp, nil
}
