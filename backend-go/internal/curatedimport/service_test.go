package curatedimport

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/models"
	"neverbroke/internal/money"
	repo "neverbroke/internal/repository/postgres"
	"os"
	"strings"
	"sync"
	"testing"
)

func ptr(n int64) *int64 { return &n }
func sample() Bundle {
	return Bundle{Version: 1, Accounts: []Account{{Key: "cmb", Institution: "cmb", CardLast4: "2689", Currency: "CNY"}, {Key: "usd", Institution: "ceb", CardLast4: "5188", Currency: "USD"}}, Records: []Record{
		{Key: "buy", Kind: "expense", Date: "2026-10-31", Currency: "CNY", AmountCents: ptr(10000), AccountKey: "cmb", Description: "meal", Category: "餐饮"},
		{Key: "return", Kind: "refund", Date: "2026-10-31", Currency: "CNY", AmountCents: ptr(3000), AccountKey: "cmb", Description: "partial refund", RelatedKeys: []string{"buy"}, RelationStatus: "规则关联待确认"},
		{Key: "usd", Kind: "expense", Date: "2026-10-31", Currency: "USD", AmountCents: ptr(3588), AccountKey: "usd", Description: "subscription"},
		{Key: "receipt", Kind: "receipt", Date: "2026-10-31", Currency: "CNY", AmountCents: ptr(10000), Description: "receipt", RelatedKeys: []string{"buy"}},
		{Key: "due", Kind: "schedule", Date: "2026-11-13", Currency: "CNY", AmountCents: ptr(20000), Description: "estimated installment"},
	}}
}
func clone(b Bundle) Bundle {
	p, _ := json.Marshal(b)
	var out Bundle
	json.Unmarshal(p, &out)
	return out
}
func TestValidate(t *testing.T) {
	b := sample()
	require.NoError(t, Validate(b))
	cases := []func(*Bundle){func(b *Bundle) { b.Records = append(b.Records, b.Records[0]) }, func(b *Bundle) { b.Records[0].AmountCents = ptr(-1) }, func(b *Bundle) { b.Records[0].Date = "2026-02-30" }, func(b *Bundle) { b.Records[1].RelatedKeys = []string{"missing"} }, func(b *Bundle) { b.Records[1].RelatedKeys = []string{"usd"} }, func(b *Bundle) { b.Records[0].Currency = "USD" }, func(b *Bundle) { b.Records[0].AccountKey = "missing" }}
	for _, mutate := range cases {
		x := clone(b)
		mutate(&x)
		require.Error(t, Validate(x))
	}
	require.NotEqual(t, ID("1", "buy"), ID("2", "buy"))
}

// This integration test refuses production databases. Run on a fresh dedicated
// Postgres database named neverbroke_import_test_* with IMPORT_TEST_DATABASE=1.
func TestPostgresAtomicImport(t *testing.T) {
	if os.Getenv("IMPORT_TEST_DATABASE") != "1" {
		t.Skip("dedicated Postgres integration test")
	}
	db, err := gorm.Open(postgres.Open(dbconfig.LoadFromEnv().DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	var name string
	require.NoError(t, db.Raw("SELECT current_database()").Scan(&name).Error)
	require.True(t, strings.HasPrefix(name, "neverbroke_import_test_"))
	require.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS users (id bigint PRIMARY KEY, email text)").Error)
	require.NoError(t, db.Exec("INSERT INTO users(id,email) VALUES (1,'owner@test.invalid'),(2,'other@test.invalid') ON CONFLICT DO NOTHING").Error)
	require.NoError(t, db.AutoMigrate(&models.Account{}, &models.Transaction{}, &models.CuratedImportRecord{}))
	ctx := context.Background()
	b := sample()
	out, err := Run(ctx, db, "1", b, false)
	require.NoError(t, err)
	require.Equal(t, 5, out.New)
	var n int64
	db.Model(&models.Account{}).Count(&n)
	require.Zero(t, n)
	out, err = Run(ctx, db, "1", b, true)
	require.NoError(t, err)
	require.Equal(t, int64(7000), out.NetExpenseCents["CNY"])
	require.NoError(t, db.Model(&models.Account{}).Where("user_id = ? AND currency = ?", "1", "CNY").Update("balance", 123.45).Error)
	out, err = Run(ctx, db, "1", b, true)
	require.NoError(t, err)
	require.Equal(t, 5, out.Skipped)
	require.Zero(t, out.New)
	var a models.Account
	db.Where("user_id = ? AND currency = ?", "1", "CNY").First(&a)
	require.Equal(t, 123.45, a.Balance)
	// A conflicting late record rolls back new account creation and all records.
	bad := clone(b)
	bad.Accounts = append(bad.Accounts, Account{Key: "new", Institution: "cmb", CardLast4: "9999", Currency: "CNY"})
	bad.Records[4].Description = "changed"
	_, err = Run(ctx, db, "1", bad, true)
	require.Error(t, err)
	db.Model(&models.Account{}).Where("card_last4 = ?", "9999").Count(&n)
	require.Zero(t, n)
	// Account currency partitions reports; refunds are negative expense, never income.
	store := repo.NewPgStore(db)
	txs, err := store.Transactions().ListByUserBetween(ctx, "1", "2026-10-31", "2026-11-01")
	require.NoError(t, err)
	require.Len(t, txs, 2)
	var total float64
	for _, tx := range txs {
		require.Equal(t, "expense", string(tx.Type))
		total += tx.Amount
	}
	require.Equal(t, 70.0, total)
	txs, err = store.Transactions().ListByUserBetween(money.WithCurrency(ctx, "USD"), "1", "2026-10-01", "2026-11-01")
	require.NoError(t, err)
	require.Len(t, txs, 1)
	require.Equal(t, 35.88, txs[0].Amount)
	// Concurrent imports for another owner remain isolated and idempotent.
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := Run(ctx, db, "2", b, true); errs <- e }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	db.Model(&models.Transaction{}).Where("user_id = ?", "2").Count(&n)
	require.Equal(t, int64(3), n)
	db.Model(&models.Transaction{}).Where("user_id = ?", "1").Count(&n)
	require.Equal(t, int64(3), n)
}
