package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http/httptest"
	"neverbroke/internal/curatedimport"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"
	pgstore "neverbroke/internal/repository/postgres"
	"os"
	"strings"
	"testing"
)

func TestWorkspacePostgres(t *testing.T) {
	if os.Getenv("WORKSPACE_TEST_DATABASE") != "1" {
		t.Skip("requires disposable neverbroke_workspace_test_* database")
	}
	db, e := gorm.Open(postgres.Open(dbconfig.LoadFromEnv().DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, e)
	var name string
	require.NoError(t, db.Raw("SELECT current_database()").Scan(&name).Error)
	require.True(t, strings.HasPrefix(name, "neverbroke_workspace_test_"), "refuse production")
	require.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS users (id bigint PRIMARY KEY)").Error)
	require.NoError(t, db.Exec("INSERT INTO users(id) VALUES(1),(2) ON CONFLICT DO NOTHING").Error)
	require.NoError(t, db.AutoMigrate(&models.Account{}, &models.Transaction{}, &models.CuratedImportRecord{}, &State{}, &Change{}, &Setting{}, &Batch{}, &SourceAlias{}, &Attachment{}, &SyncRun{}))
	engine := gin.New()
	g := engine.Group("/api", func(c *gin.Context) {
		u := c.GetHeader("X-Test-User")
		if u == "" {
			u = "1"
		}
		c.Set("user_id", u)
	})
	Register(g, db)
	g.POST("/transactions", ledger.NewHandler(nil, db).CreateTransaction)
	call := func(method, path, user string, body any) (int, map[string]any) {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api"+path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User", user)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		var r struct {
			Code    int
			Data    map[string]any
			Message string
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &r), w.Body.String())
		if r.Code != 0 {
			t.Log(r.Message)
		}
		return r.Code, r.Data
	}
	b := curatedimport.Bundle{Version: 1, Accounts: []curatedimport.Account{{Key: "a", Institution: "cmb", CardLast4: "1234", Currency: "CNY"}}, Records: []curatedimport.Record{{Key: "buy", Kind: "expense", Date: "2026-10-30", Currency: "CNY", AmountCents: cents(2600), AccountKey: "a", Description: "lunch", Category: "餐饮外卖"}, {Key: "refund", Kind: "refund", Date: "2026-10-30", Currency: "CNY", AmountCents: cents(1100), AccountKey: "a", Description: "lunch", Category: "餐饮外卖"}, {Key: "receipt", Kind: "receipt", Date: "2026-10-30", Currency: "CNY", AmountCents: cents(2600), Description: "receipt"}}}
	code, p := call("POST", "/workspace/import/preview", "1", ImportRequest{Bundle: b})
	require.Zero(t, code)
	code, batch := call("POST", "/workspace/import/commit", "1", ImportRequest{Bundle: b, Token: p["token"].(string)})
	require.Zero(t, code)
	require.Equal(t, float64(3), batch["new"])
	code, _ = call("POST", "/workspace/import/commit", "1", ImportRequest{Bundle: b, Token: p["token"].(string)})
	require.NotZero(t, code)
	code, p = call("POST", "/workspace/import/preview", "1", ImportRequest{Bundle: b})
	require.Zero(t, code)
	require.Equal(t, float64(3), p["skipped"])
	code, _ = call("POST", "/workspace/import/commit", "1", ImportRequest{Bundle: b, Token: p["token"].(string)})
	require.Zero(t, code)
	entries, accounts, e := Load(db, "1")
	require.NoError(t, e)
	require.Len(t, entries, 3)
	require.Equal(t, int64(1500), Analyze(entries, Filter{Currency: "CNY"}, "").Net)
	require.Zero(t, accounts[0].Balance)
	other, _, e := Load(db, "2")
	require.NoError(t, e)
	require.Empty(t, other)
	req := EditRequest{Reason: "confirm receipt", Items: []EditItem{{Key: "receipt", Version: 0, Edit: Edit{RelationReviewed: true, RelatedKeys: []string{"buy"}}}}}
	code, _ = call("POST", "/workspace/edit", "2", req)
	require.NotZero(t, code)
	code, _ = call("POST", "/workspace/edit", "1", req)
	require.Zero(t, code)
	code, _ = call("POST", "/workspace/edit", "1", req)
	require.NotZero(t, code)
	entries, _, e = Load(db, "1")
	require.NoError(t, e)
	require.Equal(t, int64(1500), Analyze(entries, Filter{Currency: "CNY"}, "").Net)
	code, _ = call("POST", "/workspace/batches/"+batch["id"].(string)+"/undo", "1", nil)
	require.NotZero(t, code)
	b.Records[0].Category = "交通出行"
	code, p = call("POST", "/workspace/import/preview", "1", ImportRequest{Bundle: b})
	require.Zero(t, code)
	code, revision := call("POST", "/workspace/import/commit", "1", ImportRequest{Bundle: b, Token: p["token"].(string), Resolutions: map[string]Resolution{"buy": {Action: "update"}}})
	require.Zero(t, code)
	require.Equal(t, float64(1), revision["updated"])
	code, _ = call("POST", "/workspace/batches/"+revision["id"].(string)+"/undo", "1", nil)
	require.Zero(t, code)
	entries, _, e = Load(db, "1")
	require.NoError(t, e)
	require.Equal(t, int64(1500), Analyze(entries, Filter{Currency: "CNY"}, "").Net)
	code, _ = call("POST", "/transactions", "1", map[string]any{"type": "refund", "amount": 11, "categoryKey": "food", "accountId": accounts[0].ID, "occurredAt": "2026-10-30T12:00:00+08:00"})
	require.Zero(t, code)
	entries, accounts, e = Load(db, "1")
	require.NoError(t, e)
	require.Equal(t, int64(400), Analyze(entries, Filter{Currency: "CNY"}, "").Net)
	require.Equal(t, -11.0, accounts[0].Balance)
	code, _ = call("POST", "/workspace/edit", "1", EditRequest{Reason: "invalid allocation", Items: []EditItem{{Key: "buy", Version: 0, Edit: Edit{Category: "shopping"}}, {Key: "refund", Version: 0, Edit: Edit{RelationReviewed: true, RelatedKeys: []string{"buy"}, Allocations: []Allocation{{"buy", 9900}}}}}})
	require.NotZero(t, code)
	var n int64
	require.NoError(t, db.Model(&State{}).Where("user_id = '1' AND key = 'buy'").Count(&n).Error)
	require.Zero(t, n)
	// Formal statements replace a provisional row without doubling the ledger.
	provisional := curatedimport.Record{Key: "summary", Kind: "expense", Date: "2026-10-29", Currency: "CNY", AmountCents: cents(5000), AccountKey: "a", Description: "gift", Metadata: map[string]any{"账单状态": []any{"每日摘要"}}}
	small := curatedimport.Bundle{Version: 1, Records: []curatedimport.Record{provisional}}
	code, p = call("POST", "/workspace/import/preview", "1", ImportRequest{Bundle: small})
	require.Zero(t, code)
	code, initial := call("POST", "/workspace/import/commit", "1", ImportRequest{Bundle: small, Token: p["token"].(string)})
	require.Zero(t, code)
	formal := provisional
	formal.Key = "formal"
	formal.Metadata = map[string]any{"账单状态": []any{"正式账单"}}
	small.Records = []curatedimport.Record{provisional, formal}
	code, p = call("POST", "/workspace/import/preview", "1", ImportRequest{Bundle: small})
	require.Zero(t, code)
	require.Equal(t, float64(1), p["conflicts"])
	code, replaceBatch := call("POST", "/workspace/import/commit", "1", ImportRequest{Bundle: small, Token: p["token"].(string), Resolutions: map[string]Resolution{"formal": {Action: "replace", Target: "summary"}}})
	require.Zero(t, code)
	require.Equal(t, float64(1), replaceBatch["updated"])
	code, p = call("POST", "/workspace/import/preview", "1", ImportRequest{Bundle: small})
	require.Zero(t, code)
	require.Zero(t, p["conflicts"])
	require.Equal(t, float64(1), p["skipped"])
	entries, _, e = Load(db, "1")
	require.NoError(t, e)
	require.Equal(t, int64(5400), Analyze(entries, Filter{Currency: "CNY"}, "").Net)
	code, _ = call("POST", "/workspace/batches/"+replaceBatch["id"].(string)+"/undo", "1", nil)
	require.Zero(t, code)
	code, _ = call("POST", "/workspace/batches/"+initial["id"].(string)+"/undo", "1", nil)
	require.Zero(t, code)
	// Restore original amounts and source identities after both reversals.
	entries, _, e = Load(db, "1")
	require.NoError(t, e)
	require.Equal(t, int64(400), Analyze(entries, Filter{Currency: "CNY"}, "").Net)

	code, _ = call("POST", "/workspace/edit", "1", EditRequest{Reason: "category review", Items: []EditItem{{Key: "buy", Version: 0, Edit: Edit{Category: "shopping", Reviewed: true}}}})
	require.Zero(t, code)
	legacy, e := pgstore.NewPgTransactionRepository(db).ListByUserBetween(context.Background(), "1", "2026-10-01", "2026-11-01")
	require.NoError(t, e)
	var source models.CuratedImportRecord
	require.NoError(t, db.Where("user_id = '1' AND source_key = 'buy'").First(&source).Error)
	matched := false
	for _, tr := range legacy {
		if tr.ID == source.TransactionID {
			require.Equal(t, "shopping", tr.CategoryKey)
			matched = true
		}
	}
	require.True(t, matched)
	require.NoError(t, db.Model(&models.Transaction{}).Where("id = ?", source.TransactionID).Update("amount", 1).Error)
	_, _, e = Load(db, "1")
	require.ErrorContains(t, e, "金额或币种不一致")

}
