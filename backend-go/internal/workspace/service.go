package workspace

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"math"
	"neverbroke/internal/categories"
	"neverbroke/internal/curatedimport"
	"neverbroke/internal/models"
	"sort"
	"strings"
	"time"
)

type Entry struct {
	FromAccountID  string         `json:"fromAccountId,omitempty"`
	ToAccountID    string         `json:"toAccountId,omitempty"`
	Key            string         `json:"key"`
	TransactionID  string         `json:"transactionId,omitempty"`
	Kind           string         `json:"kind"`
	Date           string         `json:"date"`
	Currency       string         `json:"currency"`
	Cents          *int64         `json:"cents"`
	Description    string         `json:"description"`
	Merchant       string         `json:"merchant"`
	Category       string         `json:"category"`
	CategoryName   string         `json:"categoryName"`
	AccountID      string         `json:"accountId"`
	AccountName    string         `json:"accountName"`
	Bank           string         `json:"bank"`
	Channel        string         `json:"channel"`
	Confirmation   string         `json:"confirmation"`
	Tags           []string       `json:"tags"`
	RelatedKeys    []string       `json:"relatedKeys"`
	RelationStatus string         `json:"relationStatus"`
	Note           string         `json:"note"`
	Metadata       map[string]any `json:"metadata"`
	Source         bool           `json:"source"`
	Version        int            `json:"version"`
	Edit           Edit           `json:"edit"`
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func Text(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		a := []string{}
		for _, s := range x {
			a = append(a, Text(s))
		}
		return strings.Join(a, "、")
	case float64:
		return fmt.Sprint(x)
	}
	return ""
}
func accountName(a models.Account) string {
	bank := str(a.Institution)
	if bank == "cmb" {
		bank = "招商银行"
	}
	if bank == "ceb" {
		bank = "光大银行"
	}
	if bank == "" {
		bank = a.Type
	}
	if str(a.CardLast4) != "" {
		bank += " · " + str(a.CardLast4)
	}
	return bank
}
func Load(db *gorm.DB, user string) ([]Entry, []models.Account, error) {
	var txs []models.Transaction
	var imports []models.CuratedImportRecord
	var states []State
	var accounts []models.Account
	for _, dest := range []any{&txs, &imports, &states, &accounts} {
		if e := db.Where("user_id = ?", user).Find(dest).Error; e != nil {
			return nil, nil, e
		}
	}
	am := map[string]models.Account{}
	for _, a := range accounts {
		am[a.ID] = a
	}
	sm := map[string]State{}
	for _, s := range states {
		sm[s.Key] = s
	}
	tm := map[string]models.Transaction{}
	for _, t := range txs {
		tm[t.ID] = t
	}
	out := make([]Entry, 0, len(txs)+len(imports))
	for _, i := range imports {
		var r curatedimport.Record
		if e := json.Unmarshal([]byte(i.Payload), &r); e != nil {
			return nil, nil, e
		}
		x := Entry{Key: r.Key, TransactionID: i.TransactionID, Kind: r.Kind, Date: r.Date, Currency: r.Currency, Cents: r.AmountCents, Description: r.Description, Category: r.Category, Tags: r.Tags, RelatedKeys: r.RelatedKeys, RelationStatus: r.RelationStatus, Note: r.Note, Metadata: r.Metadata, Source: true}
		if r.Kind == "expense" || r.Kind == "refund" {
			t, exists := tm[i.TransactionID]
			if !exists || r.AmountCents == nil {
				return nil, nil, fmt.Errorf("来源记录 %s 与账本缺少对应关系", r.Key)
			}
			expected := *r.AmountCents
			if r.Kind == "refund" {
				expected = -expected
			}
			if t.Type != "expense" || int64(math.Round(t.Amount*100)) != expected || am[str(t.AccountID)].Currency != r.Currency {
				return nil, nil, fmt.Errorf("来源记录 %s 与账本金额或币种不一致，请先对账", r.Key)
			}
		}
		if t, ok := tm[i.TransactionID]; ok {
			x.AccountID = str(t.AccountID)
			delete(tm, i.TransactionID)
		}
		// Statements and schedules are not ledger rows, but still belong to an account.
		if x.AccountID == "" {
			label := Text(r.Metadata["账户"])
			last4 := Text(r.Metadata["卡号尾号"])
			matches := []models.Account{}
			for _, a := range accounts {
				card := str(a.CardLast4)
				if card != "" && a.Currency == x.Currency && (last4 == card || strings.Contains(label, card)) {
					matches = append(matches, a)
				}
			}
			if len(matches) == 1 {
				x.AccountID = matches[0].ID
			}
			x.AccountName = label
			if strings.Contains(label, "招商") {
				x.Bank = "cmb"
			}
			if strings.Contains(label, "光大") {
				x.Bank = "ceb"
			}
		}
		x.Merchant = Text(r.Metadata["商户归一"])
		if x.Merchant == "" {
			x.Merchant = r.Description
		}
		x.Channel = Text(r.Metadata["支付渠道"])
		if x.Channel == "" {
			x.Channel = Text(r.Metadata["付款渠道"])
		}
		x.Confirmation = "unknown"
		status := Text(r.Metadata["账单状态"])
		if strings.Contains(status, "正式") {
			x.Confirmation = "formal"
		} else if strings.Contains(status, "摘要") || strings.Contains(status, "待确认") || strings.Contains(strings.Join(r.Tags, " "), "未出账") {
			x.Confirmation = "pending"
		}
		if x.Kind == "schedule" {
			x.Edit.Status = "estimated"
			if strings.Contains(Text(r.Metadata["事项类型"]), "正式") {
				x.Edit.Status = "billed"
			}
		}
		out = append(out, x)
	}
	for _, t := range tm {
		cents := int64(math.Round(t.Amount * 100))
		kind := t.Type
		if kind == "expense" && cents < 0 {
			kind = "refund"
			cents = -cents
		}
		date := t.OccurredAt
		if len(date) > 10 {
			date = date[:10]
		}
		a := am[str(t.AccountID)]
		currency := a.Currency
		if currency == "" {
			currency = "CNY"
		}
		out = append(out, Entry{FromAccountID: str(t.FromAccountID), ToAccountID: str(t.ToAccountID), Key: "tx:" + t.ID, TransactionID: t.ID, Kind: kind, Date: date, Currency: currency, Cents: &cents, Description: str(t.Remark), Merchant: str(t.Remark), Category: str(t.CategoryKey), AccountID: str(t.AccountID), Confirmation: "manual", Metadata: map[string]any{"分期计划": str(t.InstallmentPlanID)}})
	}
	native, err := nativePlans(db, user, am)
	if err != nil {
		return nil, nil, err
	}
	out = append(out, native...)
	for n := range out {
		x := &out[n]
		if s, ok := sm[x.Key]; ok {
			x.Version = s.Version
			if e := json.Unmarshal([]byte(s.Payload), &x.Edit); e != nil {
				return nil, nil, e
			}
			if x.Edit.Category != "" {
				x.Category = x.Edit.Category
			}
			if x.Edit.Tags != nil {
				x.Tags = x.Edit.Tags
			}
			if x.Edit.RelationReviewed {
				x.RelatedKeys = x.Edit.RelatedKeys
				x.RelationStatus = "人工已确认"
				if len(x.RelatedKeys) == 0 {
					x.RelationStatus = "已解除关联"
				}
			}
		}
		cat := categories.Resolve(x.Category)
		x.Category = cat.ID
		x.CategoryName = cat.Name
		if a, ok := am[x.AccountID]; ok {
			x.AccountName = accountName(a)
			x.Bank = str(a.Institution)
		}
		if x.AccountName == "" {
			x.AccountName = Text(x.Metadata["账户"])
		}
		if x.Channel == "" {
			x.Channel = "未标注"
		}
		if x.Tags == nil {
			x.Tags = []string{}
		}
		if x.RelatedKeys == nil {
			x.RelatedKeys = []string{}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date == out[j].Date {
			return out[i].Key < out[j].Key
		}
		return out[i].Date > out[j].Date
	})
	return out, accounts, nil
}

type Filter struct{ From, To, Currency, Bank, Account, Category, Kind, Confirmation, Query, Queue, Channel string }

func Matches(x Entry, f Filter) bool {
	if f.From != "" && x.Date < f.From || f.To != "" && x.Date > f.To || f.Currency != "" && x.Currency != f.Currency || f.Bank != "" && x.Bank != f.Bank || f.Account != "" && x.AccountID != f.Account || f.Category != "" && x.Category != f.Category || f.Kind != "" && f.Kind != "ledger" && x.Kind != f.Kind || f.Confirmation != "" && x.Confirmation != f.Confirmation || f.Channel != "" && x.Channel != f.Channel {
		return false
	}
	if f.Kind == "ledger" && x.Kind != "expense" && x.Kind != "refund" && x.Kind != "income" && x.Kind != "transfer" {
		return false
	}
	if f.Query != "" && !strings.Contains(strings.ToLower(x.Description+" "+x.Merchant+" "+x.Note+" "+strings.Join(x.Tags, " ")), strings.ToLower(f.Query)) {
		return false
	}
	switch f.Queue {
	case "category":
		return x.Kind == "expense" && (x.Category == "uncategorized" || x.Category == "platform") && !x.Edit.Reviewed
	case "refund":
		return x.Kind == "refund" && !x.Edit.RelationReviewed
	case "receipt":
		return x.Kind == "receipt" && len(x.RelatedKeys) == 0
	case "pending":
		return x.Confirmation == "pending"
	}
	return true
}

type Metric struct {
	Gross       int64 `json:"gross"`
	Refund      int64 `json:"refund"`
	Net         int64 `json:"net"`
	Income      int64 `json:"income"`
	Count       int   `json:"count"`
	RefundCount int   `json:"refundCount"`
	Average     int64 `json:"average"`
	Pending     int64 `json:"pending"`
}

func (m *Metric) Add(x Entry) {
	if x.Cents == nil {
		return
	}
	v := *x.Cents
	switch x.Kind {
	case "expense":
		m.Gross += v
		m.Count++
		if x.Confirmation == "pending" {
			m.Pending += v
		}
	case "refund":
		m.Refund += v
		m.RefundCount++
	case "income":
		m.Income += v
	}
	m.Net = m.Gross - m.Refund
	if m.Count > 0 {
		m.Average = int64(math.Round(float64(m.Gross) / float64(m.Count)))
	}
}

type Group struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Metric
}
type Analysis struct {
	Metric
	Categories []Group  `json:"categories"`
	Merchants  []Group  `json:"merchants"`
	Channels   []Group  `json:"channels"`
	Trend      []Group  `json:"trend"`
	First      string   `json:"first"`
	Last       string   `json:"last"`
	Months     []string `json:"months"`
}

func Analyze(entries []Entry, f Filter, year string) Analysis {
	r := Analysis{Categories: []Group{}, Merchants: []Group{}, Channels: []Group{}, Trend: []Group{}, Months: []string{}}
	cm, mm, pm, dm := map[string]*Group{}, map[string]*Group{}, map[string]*Group{}, map[string]*Group{}
	months := map[string]bool{}
	add := func(m map[string]*Group, key, name string, x Entry) {
		if m[key] == nil {
			m[key] = &Group{Key: key, Name: name}
		}
		m[key].Add(x)
	}
	for _, x := range entries {
		if !Matches(x, f) || (x.Kind != "expense" && x.Kind != "refund" && x.Kind != "income") {
			continue
		}
		r.Add(x)
		if r.First == "" || x.Date < r.First {
			r.First = x.Date
		}
		if x.Date > r.Last {
			r.Last = x.Date
		}
		if len(x.Date) >= 7 {
			months[x.Date[:7]] = true
		}
		if x.Kind != "income" {
			add(cm, x.Category, x.CategoryName, x)
			add(mm, x.Merchant, x.Merchant, x)
			add(pm, x.Channel, x.Channel, x)
		}
		date := x.Date
		if year != "" && len(date) >= 7 {
			date = date[:7]
		}
		add(dm, date, date, x)
	}
	flatten := func(m map[string]*Group) []Group {
		v := []Group{}
		for _, g := range m {
			v = append(v, *g)
		}
		sort.Slice(v, func(i, j int) bool {
			if v[i].Gross == v[j].Gross {
				return v[i].Key < v[j].Key
			}
			return v[i].Gross > v[j].Gross
		})
		return v
	}
	r.Categories = flatten(cm)
	r.Merchants = flatten(mm)
	if len(r.Merchants) > 15 {
		r.Merchants = r.Merchants[:15]
	}
	r.Channels = flatten(pm)
	if year != "" {
		for m := 1; m <= 12; m++ {
			key := fmt.Sprintf("%s-%02d", year, m)
			if dm[key] == nil {
				dm[key] = &Group{Key: key, Name: key}
			}
		}
	} else if a, e := time.Parse("2006-01-02", f.From); e == nil {
		if b, e := time.Parse("2006-01-02", f.To); e == nil && b.Sub(a).Hours() <= 24*366 {
			for d := a; !d.After(b); d = d.AddDate(0, 0, 1) {
				k := d.Format("2006-01-02")
				if dm[k] == nil {
					dm[k] = &Group{Key: k, Name: k}
				}
			}
		}
	}
	r.Trend = flatten(dm)
	sort.Slice(r.Trend, func(i, j int) bool { return r.Trend[i].Key < r.Trend[j].Key })
	for m := range months {
		r.Months = append(r.Months, m)
	}
	sort.Strings(r.Months)
	return r
}

type Quality struct {
	Receipts         int   `json:"receipts"`
	MatchedReceipts  int   `json:"matchedReceipts"`
	Refunds          int   `json:"refunds"`
	ConfirmedRefunds int   `json:"confirmedRefunds"`
	Unclassified     int   `json:"unclassified"`
	Platform         int   `json:"platform"`
	FeeCents         int64 `json:"feeCents"`
	FeeRecords       int   `json:"feeRecords"`
}

func DataQuality(items []Entry, currency string) Quality {
	var q Quality
	for _, x := range items {
		if x.Currency != currency {
			continue
		}
		switch x.Kind {
		case "receipt":
			q.Receipts++
			if len(x.RelatedKeys) > 0 {
				q.MatchedReceipts++
			}
		case "refund":
			q.Refunds++
			if x.Edit.RelationReviewed {
				q.ConfirmedRefunds++
			}
		case "expense":
			if x.Category == "uncategorized" {
				q.Unclassified++
			}
			if x.Category == "platform" {
				q.Platform++
			}
		case "reference":
			typ := Text(x.Metadata["交易类型"]) + x.Description
			if strings.Contains(typ, "利息") || strings.Contains(typ, "手续费") {
				q.FeeRecords++
				if x.Cents != nil {
					q.FeeCents += *x.Cents
				}
			}
		}
	}
	return q
}
