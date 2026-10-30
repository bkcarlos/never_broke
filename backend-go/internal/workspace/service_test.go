package workspace

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"neverbroke/internal/categories"
	"neverbroke/internal/curatedimport"
	"neverbroke/internal/models"
	"testing"
)

func cents(v int64) *int64 { return &v }
func TestMetricsCurrencyScopeAndRefundPeriod(t *testing.T) {
	rows := []Entry{{Key: "buy", Kind: "expense", Date: "2026-01-01", Currency: "CNY", Cents: cents(2600), Category: "food", CategoryName: "餐饮外卖"}, {Key: "refund", Kind: "refund", Date: "2026-02-01", Currency: "CNY", Cents: cents(1100), Category: "food"}, {Key: "usd", Kind: "expense", Date: "2026-01-01", Currency: "USD", Cents: cents(999999)}, {Key: "receipt", Kind: "receipt", Date: "2026-01-01", Currency: "CNY", Cents: cents(2600)}, {Key: "plan", Kind: "schedule", Date: "2026-01-01", Currency: "CNY", Cents: cents(5500)}}
	r := Analyze(rows, Filter{Currency: "CNY"}, "2026")
	require.Equal(t, int64(1500), r.Net)
	require.Equal(t, int64(2600), r.Gross)
	require.Equal(t, int64(1100), r.Refund)
	require.Zero(t, r.Income)
	require.Equal(t, 1, r.Count)
	require.Equal(t, int64(2600), r.Average)
	require.Len(t, r.Trend, 12)
	require.Equal(t, int64(-1100), r.Trend[1].Net)
	feb := Analyze(rows, Filter{Currency: "CNY", From: "2026-02-01", To: "2026-02-28"}, "")
	require.Equal(t, int64(-1100), feb.Net)
	require.Len(t, feb.Trend, 28)
	empty := Analyze(rows, Filter{Currency: "USD", From: "2030-01-01", To: "2030-12-31"}, "2030")
	require.Zero(t, empty.Net)
	require.Len(t, empty.Trend, 12)
	require.Equal(t, categories.Resolve("餐饮外卖").ID, categories.Resolve("food").ID)
	require.Equal(t, "healthcare", categories.Resolve("healthcare").Aliases[2])
}
func TestRefundAllocations(t *testing.T) {
	buy := Entry{Key: "b", Kind: "expense", Currency: "CNY", AccountID: "a", Cents: cents(2600)}
	r := Entry{Key: "r", Kind: "refund", Currency: "CNY", AccountID: "a", Cents: cents(1100)}
	all := map[string]Entry{"b": buy, "r": r}
	edit := Edit{RelationReviewed: true, RelatedKeys: []string{"b"}, Allocations: []Allocation{{"b", 1100}}}
	require.NoError(t, validateEdit(r, edit, all))
	other := r
	other.Key = "r2"
	other.Edit = Edit{RelationReviewed: true, Allocations: []Allocation{{"b", 1600}}}
	all[other.Key] = other
	require.ErrorContains(t, validateEdit(r, edit, all), "超过")
	delete(all, "r2")
	buy.Currency = "USD"
	all["b"] = buy
	require.Error(t, validateEdit(r, edit, all))
	buy.Currency = "CNY"
	buy.AccountID = "someone"
	all["b"] = buy
	require.Error(t, validateEdit(r, edit, all))
	require.NoError(t, validateEdit(r, Edit{RelationReviewed: true}, all))
}
func TestImportPreviewConflictsAndVersions(t *testing.T) {
	b := curatedimport.Bundle{Version: 1, Accounts: []curatedimport.Account{{Key: "a", Institution: "cmb", CardLast4: "1234", Currency: "CNY"}}, Records: []curatedimport.Record{{Key: "a", Kind: "expense", Currency: "CNY", Date: "2026-10-30", Description: "Lunch", AccountKey: "a", AmountCents: cents(2600)}}}
	p, e := preview(b, nil)
	require.NoError(t, e)
	require.Equal(t, 1, p.New)
	require.Zero(t, p.Conflicts)
	raw, _ := json.Marshal(b.Records[0])
	old := []models.CuratedImportRecord{{SourceKey: "a", Payload: string(raw), Digest: digest(b.Records[0])}}
	same, e := preview(b, old)
	require.NoError(t, e)
	require.Equal(t, 1, same.Skipped)
	require.NotEqual(t, p.Token, same.Token)
	b.Records[0].Category = "food"
	change, e := preview(b, old)
	require.NoError(t, e)
	require.Equal(t, 1, change.Updated)
	require.Equal(t, 1, change.Conflicts)
	b.Records[0].Key = "another"
	dupe, e := preview(b, old)
	require.NoError(t, e)
	require.Equal(t, "duplicate", dupe.Differences[0].Type)
	require.Equal(t, []string{"a"}, dupe.Differences[0].Candidates)
}

func TestProjectionUnknownBalancesAndPlanStates(t *testing.T) {
	rows := []Entry{
		{Kind: "schedule", Date: "2026-10-30", Currency: "CNY", Cents: cents(2600), Edit: Edit{Status: "billed"}},
		{Kind: "schedule", Date: "2026-10-30", Currency: "CNY", Cents: cents(1100), Edit: Edit{Status: "estimated"}},
		{Kind: "schedule", Date: "2026-10-30", Currency: "CNY", Cents: cents(9900), Edit: Edit{Status: "replaced"}},
		{Kind: "schedule", Date: "2026-10-30", Currency: "CNY", Cents: cents(9900), Edit: Edit{Status: "paid"}},
		{Kind: "schedule", Date: "2026-10-30", Currency: "USD", Cents: cents(9900), Edit: Edit{Status: "billed"}},
	}
	acc := []models.Account{{ID: "cash", Type: "debit", Currency: "CNY", Balance: 100}}
	p := Project(rows, acc, nil, "CNY", "2026-10-30", 2)
	require.Nil(t, p.Starting)
	require.Nil(t, p.Points[0].Remaining)
	require.Equal(t, int64(2600), p.Points[0].Billed)
	require.Equal(t, int64(1100), p.Points[0].Estimated)
	settings := []Setting{{Kind: "balance", Match: "cash"}}
	p = Project(rows, acc, settings, "CNY", "2026-10-30", 2)
	require.Equal(t, int64(6300), *p.Points[0].Remaining)
	rows = append(rows, Entry{Kind: "schedule", Date: "2026-10-30", Currency: "CNY", Edit: Edit{Status: "billed"}})
	p = Project(rows, acc, settings, "CNY", "2026-10-30", 2)
	require.Equal(t, 1, p.UnknownAmounts)
	require.Nil(t, p.Points[0].Remaining)
}
