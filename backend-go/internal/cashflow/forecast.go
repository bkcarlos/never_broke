package cashflow

import (
	"fmt"
	"math"
	"strings"
	"time"

	"neverbroke/internal/models"
	"neverbroke/internal/workspace"
)

func validateRequest(req CashflowForecastRequest) (time.Time, error) {
	start, err := time.Parse(time.DateOnly, req.StartDate)
	if err != nil || start.Year() < 1900 || start.Year() > 2200 {
		return start, fmt.Errorf("预测日期无效")
	}
	if req.Days < 1 || req.Days > 366 {
		return start, fmt.Errorf("预测天数须为 1–366")
	}
	if req.Base != ForecastBaseTotalAssets && req.Base != ForecastBaseAccount {
		return start, fmt.Errorf("预测范围无效")
	}
	if req.Base == ForecastBaseAccount && req.AccountID == "" {
		return start, fmt.Errorf("请选择账户")
	}
	if req.SafeAmount < 0 || math.IsNaN(req.SafeAmount) || math.IsInf(req.SafeAmount, 0) {
		return start, fmt.Errorf("安全垫金额无效")
	}
	return start, nil
}

func units(cents int64) float64          { return float64(cents) / 100 }
func cents(value float64) int64          { return int64(math.Round(value * 100)) }
func amountPointer(cents int64) *float64 { v := units(cents); return &v }
func activeExpense(s workspace.Setting, currency string) bool {
	return (s.Kind == "fixed_expense" || s.Kind == "subscription") && s.Enabled && s.Currency == currency
}
func matchesExpense(x workspace.Entry, s workspace.Setting) bool {
	if s.AccountID != "" && x.AccountID != s.AccountID {
		return false
	}
	keyword := strings.ToLower(strings.TrimSpace(s.Match))
	if keyword == "" {
		keyword = strings.ToLower(strings.TrimSpace(s.Name))
	}
	return keyword != "" && strings.Contains(strings.ToLower(x.Description+" "+x.Merchant), keyword)
}

// Estimate ordinary spending from up to three previous months plus the elapsed
// current month. Fixed expenses and installment repayments are not learned twice.
func spendingRate(items []workspace.Entry, settings []workspace.Setting, currency, account string, start time.Time) (int64, string, int) {
	month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	oldest := month.AddDate(0, -3, 0).Format(time.DateOnly)
	monthKey, today := month.Format(time.DateOnly), start.Format(time.DateOnly)
	var historical, current int64
	first := ""
	currentObserved := false
	for _, x := range items {
		if x.Currency != currency || (account != "" && x.AccountID != account) || x.Date < oldest || x.Date >= today || x.Cents == nil || (x.Kind != "expense" && x.Kind != "refund") {
			continue
		}
		if x.Date < monthKey {
			if first == "" || x.Date < first {
				first = x.Date
			}
		} else {
			currentObserved = true
		}
		fixed := false
		for _, s := range settings {
			if activeExpense(s, currency) && matchesExpense(x, s) {
				fixed = true
				break
			}
		}
		if fixed || workspace.Text(x.Metadata["分期计划"]) != "" {
			continue
		}
		value := *x.Cents
		if x.Kind == "refund" {
			value = -value
		}
		if x.Date < monthKey {
			historical += value
		} else {
			current += value
		}
	}
	elapsed := start.Day() - 1
	historyDays := 0
	var historicalRate, currentRate float64
	if first != "" {
		firstDate, _ := time.Parse(time.DateOnly, first)
		historyDays = int(month.Sub(firstDate).Hours() / 24)
		historicalRate = float64(max(int64(0), historical)) / float64(max(1, historyDays))
	}
	if elapsed > 0 {
		currentRate = float64(max(int64(0), current)) / float64(elapsed)
	}
	if historyDays > 0 {
		rate := historicalRate
		if currentObserved {
			rate = historicalRate*0.7 + currentRate*0.3
		}
		return int64(math.Round(rate)), "history", historyDays
	}
	if currentObserved {
		return int64(math.Round(currentRate)), "current", elapsed
	}
	return 0, "unknown", 0
}

func buildForecast(req CashflowForecastRequest, currency string, items []workspace.Entry, accounts []models.Account, settings []workspace.Setting, incomes []incomeEvent, budgets map[string]int64) (CashflowForecastResponse, error) {
	start, err := validateRequest(req)
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	end := start.AddDate(0, 0, req.Days)
	out := CashflowForecastResponse{Currency: currency, StartDate: req.StartDate, EndDate: end.Format(time.DateOnly), Base: req.Base, AccountID: req.AccountID, SafeAmount: req.SafeAmount, Points: []DailyForecastPoint{}, Months: []MonthlyForecast{}, Warnings: []string{}, DangerDates: []string{}, WarningDates: []string{}, Status: ForecastStatusSafe, SpendingKnown: true}
	warn := func(message string) {
		for _, old := range out.Warnings {
			if old == message {
				return
			}
		}
		out.Warnings = append(out.Warnings, message)
	}
	selected := ""
	if req.Base == ForecastBaseAccount {
		selected = req.AccountID
	}
	accountMap := map[string]models.Account{}
	selectedAccounts := []models.Account{}
	for _, a := range accounts {
		accountMap[a.ID] = a
		if a.Currency == currency && !a.Archived && a.Type != "credit" && (selected == "" || a.ID == selected) {
			selectedAccounts = append(selectedAccounts, a)
		}
	}
	if selected != "" && len(selectedAccounts) != 1 {
		return out, fmt.Errorf("请选择同币种的现金类账户")
	}
	projection := workspace.Project(nil, selectedAccounts, settings, currency, req.StartDate, req.Days)
	out.BalanceKnown = projection.Starting != nil
	balance := int64(0)
	if out.BalanceKnown {
		balance = *projection.Starting
	} else {
		warn("账户余额尚未全部核验；可查看收支估算，暂不判断余额及资金缺口。")
	}
	flows := map[string][]ForecastEvent{}
	add := func(date, kind, name string, value int64) {
		flows[date] = append(flows[date], ForecastEvent{Kind: kind, Name: name, Amount: units(value)})
	}

	// Scheduled credit repayments refer to the debt account, not the bank used to
	// pay it. Do not silently charge every debt to a selected bank account.
	for _, x := range items {
		if x.Kind != "schedule" || x.Currency != currency || x.Date < req.StartDate || x.Date >= out.EndDate || x.Edit.Status == "paid" || x.Edit.Status == "cancelled" || x.Edit.Status == "replaced" {
			continue
		}
		if selected != "" && x.AccountID != selected {
			out.BalanceKnown = false
			warn("存在尚未指定扣款账户的还款事项，单账户预测未计入这些还款；请切换全部现金账户查看资金缺口。")
			continue
		}
		if x.Cents == nil {
			out.UnknownAmounts++
			continue
		}
		add(x.Date, "repayment", x.Description, *x.Cents)
	}
	if out.UnknownAmounts > 0 {
		out.BalanceKnown = false
		warn("有还款金额待确认，收支合计尚不完整，暂不判断资金缺口。")
	}

	usedActual := map[string]bool{}
	for _, income := range incomes {
		if income.Currency != currency || (selected != "" && income.AccountID != selected) {
			continue
		}
		// Realized entries are already in the starting balance.
		realized := false
		for _, x := range items {
			if x.Kind == "income" && x.Date <= req.StartDate && x.Date == income.Date && x.AccountID == income.AccountID && x.Cents != nil && !usedActual[x.Key] && *x.Cents == income.Cents && x.Description == income.Name {
				realized = true
				usedActual[x.Key] = true
				break
			}
		}
		if !realized {
			add(income.Date, "income", income.Name, income.Cents)
		}
		if income.Warning != "" {
			warn(income.Warning)
		}
	}
	for _, s := range settings {
		if !activeExpense(s, currency) {
			continue
		}
		if err := workspace.ValidateRecurringExpense(s); err != nil {
			out.BalanceKnown = false
			warn("固定支出「" + s.Name + "」配置不完整，请编辑后保存。")
			continue
		}
		if selected != "" && s.AccountID != selected {
			if s.AccountID == "" {
				out.BalanceKnown = false
				warn("有固定支出未指定扣款账户，未计入单账户预测；请补充账户或切换全部现金账户。")
			}
			continue
		}
		first, _ := time.Parse(time.DateOnly, s.Date)
		for _, date := range workspace.RecurrenceDates(first, s.Period, s.EndDate, start, end) {
			paid := false
			// Exact date + keyword + amount, one actual entry per occurrence. Avoid
			// broad month matching that would swallow two weekly payments.
			for _, x := range items {
				if x.Kind == "expense" && x.Date <= req.StartDate && x.Date == date && x.Currency == currency && x.Cents != nil && *x.Cents == s.Cents && !usedActual[x.Key] && matchesExpense(x, s) {
					paid = true
					usedActual[x.Key] = true
					break
				}
			}
			if !paid {
				add(date, "fixed", s.Name, s.Cents)
			}
		}
	}
	rate, source, historyDays := spendingRate(items, settings, currency, selected, start)
	out.HistoryDays = historyDays
	if source == "history" {
		warn("日常消费按近三个月可用记录估算；有本月记录时，历史日均占 70%、本月日均占 30%。匹配固定支出关键词的消费不重复计入日均。")
	}
	if source == "current" {
		warn("暂无前三个月消费记录，日常消费暂按本月已过天数的日均估算。")
	}
	warn("按现金收付口径汇总：信用卡历史消费不再计作现金流出，还款在到期日计入；未来日常消费按当日预留资金估算。同日收支按日终净额计算。")
	// Cash movements already included in the current account balance are shown
	// only in the month's actual totals, never applied to the balance again.
	actualIn, actualOut := map[string]int64{}, map[string]int64{}
	ordinaryToday := int64(0)
	for _, x := range items {
		if x.Date > req.StartDate || len(x.Date) < 7 || x.Cents == nil {
			continue
		}
		a, exists := accountMap[x.AccountID]
		scope := x.Currency == currency && (selected == "" || x.AccountID == selected)
		cash := scope && exists && a.Type != "credit"
		key := x.Date[:7]
		switch x.Kind {
		case "income":
			if cash {
				actualIn[key] += *x.Cents
			}
		case "expense", "refund":
			sign := int64(1)
			if x.Kind == "refund" {
				sign = -1
			}
			if cash {
				actualOut[key] += sign * *x.Cents
			}
			if scope && x.Date == req.StartDate {
				fixed := false
				for _, s := range settings {
					if activeExpense(s, currency) && matchesExpense(x, s) {
						fixed = true
						break
					}
				}
				if !fixed {
					ordinaryToday += sign * *x.Cents
				}
			}
		case "transfer":
			from, fromOK := accountMap[x.FromAccountID]
			to, toOK := accountMap[x.ToAccountID]
			fromCash := fromOK && from.Currency == currency && from.Type != "credit" && (selected == "" || from.ID == selected)
			toCash := toOK && to.Currency == currency && to.Type != "credit" && (selected == "" || to.ID == selected)
			if fromCash && !toCash {
				actualOut[key] += *x.Cents
			}
			if toCash && !fromCash {
				actualIn[key] += *x.Cents
			}
		}
	}
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		date, month := day.Format(time.DateOnly), day.Format("2006-01")
		if len(out.Months) == 0 || out.Months[len(out.Months)-1].Month != month {
			last := time.Date(day.Year(), day.Month()+1, 0, 0, 0, 0, 0, time.UTC)
			to := last
			if end.AddDate(0, 0, -1).Before(to) {
				to = end.AddDate(0, 0, -1)
			}
			m := MonthlyForecast{Month: month, From: day.Format(time.DateOnly), To: to.Format(time.DateOnly), Partial: to.Before(last), ActualIncome: units(actualIn[month]), ActualOutflow: units(actualOut[month]), Status: ForecastStatusSafe, SpendingSource: source}
			if source == "unknown" {
				if budgets[month] > 0 && selected == "" {
					m.SpendingSource = "budget"
				} else {
					out.SpendingKnown = false
					warn("部分月份缺少消费历史及预算，日常消费尚未计入；当前结果仅覆盖已知计划，不代表可放心消费。")
				}
			}
			out.Months = append(out.Months, m)
		}
		m := &out.Months[len(out.Months)-1]
		daily := rate
		if m.SpendingSource == "budget" {
			// Budgets cover all spending: reserve fixed obligations first. This is a
			// fallback only; exhausting a budget never zeros an observed spending rate.
			fixed := int64(0)
			for date, events := range flows {
				if strings.HasPrefix(date, month) {
					for _, e := range events {
						if e.Kind == "fixed" {
							fixed += cents(e.Amount)
						}
					}
				}
			}
			spent := int64(0)
			for _, x := range items {
				if x.Currency == currency && (selected == "" || x.AccountID == selected) && strings.HasPrefix(x.Date, month) && x.Date < req.StartDate && x.Cents != nil {
					if x.Kind == "expense" {
						spent += *x.Cents
					}
					if x.Kind == "refund" {
						spent -= *x.Cents
					}
				}
			}
			days := time.Date(day.Year(), day.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			if month == start.Format("2006-01") {
				days -= start.Day() - 1
			}
			daily = int64(math.Round(float64(max(int64(0), budgets[month]-spent-fixed)) / float64(days)))
		}
		if date == req.StartDate {
			out.DailyBudget = units(daily)
			daily = max(int64(0), daily-max(int64(0), ordinaryToday))
		}
		if daily > 0 {
			add(date, "daily", "日常消费预留", daily)
		}
		events := flows[date]
		if events == nil {
			events = []ForecastEvent{}
		}
		in, outflow := int64(0), int64(0)
		for _, e := range events {
			value := cents(e.Amount)
			switch e.Kind {
			case "income":
				in += value
				m.ForecastIncome = units(cents(m.ForecastIncome) + value)
			case "fixed":
				outflow += value
				m.FixedExpense = units(cents(m.FixedExpense) + value)
			case "daily":
				outflow += value
				m.DailyExpense = units(cents(m.DailyExpense) + value)
			case "repayment":
				outflow += value
				m.Repayment = units(cents(m.Repayment) + value)
			}
		}
		point := DailyForecastPoint{Date: date, Inflow: units(in), Outflow: units(outflow), Events: events}
		if out.BalanceKnown {
			point.OpeningBalance = amountPointer(balance)
			balance += in - outflow
			point.ClosingBalance = amountPointer(balance)
			m.ClosingBalance = amountPointer(balance)
			if m.MinBalance == nil || balance < cents(*m.MinBalance) {
				m.MinBalance = amountPointer(balance)
				m.MinBalanceDate = date
			}
			if balance < 0 {
				out.DangerDates = append(out.DangerDates, date)
				m.Status = ForecastStatusDanger
				if m.FirstGapDate == "" {
					m.FirstGapDate = date
				}
			} else if balance < cents(req.SafeAmount) {
				out.WarningDates = append(out.WarningDates, date)
				if m.Status != ForecastStatusDanger {
					m.Status = ForecastStatusWarning
				}
			}
		}
		out.Points = append(out.Points, point)
	}
	for i := range out.Months {
		m := &out.Months[i]
		m.Income = units(cents(m.ActualIncome) + cents(m.ForecastIncome))
		m.Outflow = units(cents(m.ActualOutflow) + cents(m.FixedExpense) + cents(m.DailyExpense) + cents(m.Repayment))
		m.Net = units(cents(m.Income) - cents(m.Outflow))
		if m.MinBalance != nil {
			m.GapAmount = amountPointer(max(int64(0), -cents(*m.MinBalance)))
		}
		if !out.BalanceKnown || (m.SpendingSource == "unknown" && m.Status == ForecastStatusSafe) {
			m.Status = ForecastStatus("unknown")
		}
	}
	if !out.BalanceKnown {
		out.Status = ForecastStatus("unknown")
	} else if len(out.DangerDates) > 0 {
		out.Status = ForecastStatusDanger
	} else if len(out.WarningDates) > 0 {
		out.Status = ForecastStatusWarning
	} else if !out.SpendingKnown {
		out.Status = ForecastStatus("unknown")
	}
	return out, nil
}
