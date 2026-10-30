// Package curatedimport imports reviewed historical records atomically. It does
// not replay old transactions against today's account balances.
package curatedimport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"neverbroke/internal/models"
	"regexp"
	"strings"
	"time"
)

type Account struct {
	Key         string `json:"key"`
	Institution string `json:"institution"`
	CardLast4   string `json:"cardLast4"`
	Currency    string `json:"currency"`
}
type Record struct {
	Key            string         `json:"key"`
	Kind           string         `json:"kind"`
	Date           string         `json:"date"`
	Currency       string         `json:"currency"`
	AmountCents    *int64         `json:"amountCents,omitempty"`
	AccountKey     string         `json:"accountKey,omitempty"`
	Description    string         `json:"description"`
	Category       string         `json:"category,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	RelatedKeys    []string       `json:"relatedKeys,omitempty"`
	RelationStatus string         `json:"relationStatus,omitempty"`
	Note           string         `json:"note,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}
type Bundle struct {
	Version  int       `json:"version"`
	Accounts []Account `json:"accounts"`
	Records  []Record  `json:"records"`
}
type Summary struct {
	Records         int              `json:"records"`
	New             int              `json:"new"`
	Skipped         int              `json:"skipped"`
	ByKind          map[string]int   `json:"byKind"`
	NetExpenseCents map[string]int64 `json:"netExpenseCents"`
	BalancePolicy   string           `json:"balancePolicy"`
}

var token = regexp.MustCompile(`^[A-Za-z0-9:_-]{1,180}$`)

func ID(user, key string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("neverbroke:curated:"+user+":"+key)).String()
}
func Validate(b Bundle) error {
	if b.Version != 1 || len(b.Records) == 0 || len(b.Records) > 20000 {
		return errors.New("expected version 1 and 1–20000 records")
	}
	ac := map[string]Account{}
	rs := map[string]Record{}
	for _, a := range b.Accounts {
		if !token.MatchString(a.Key) || (a.Currency != "CNY" && a.Currency != "USD") || (a.Institution != "cmb" && a.Institution != "ceb") || !regexp.MustCompile(`^\d{4}$`).MatchString(a.CardLast4) {
			return fmt.Errorf("invalid account %s", a.Key)
		}
		if _, ok := ac[a.Key]; ok {
			return fmt.Errorf("duplicate account %s", a.Key)
		}
		ac[a.Key] = a
	}
	for _, r := range b.Records {
		if !token.MatchString(r.Key) || len(r.Description) > 2000 || len(r.Category) > 50 || len(r.Note) > 20000 {
			return fmt.Errorf("invalid record %s", r.Key)
		}
		if _, ok := rs[r.Key]; ok {
			return fmt.Errorf("duplicate key %s", r.Key)
		}
		if _, err := time.Parse("2006-01-02", r.Date); err != nil {
			return fmt.Errorf("invalid date for %s", r.Key)
		}
		if r.Currency != "CNY" && r.Currency != "USD" {
			return fmt.Errorf("invalid currency for %s", r.Key)
		}
		switch r.Kind {
		case "expense", "refund":
			a, ok := ac[r.AccountKey]
			if !ok || a.Currency != r.Currency || r.AmountCents == nil || *r.AmountCents <= 0 || *r.AmountCents > 9999999999999 {
				return fmt.Errorf("invalid transaction %s", r.Key)
			}
		case "receipt", "statement", "reference", "daily_summary":
		case "schedule":
			if r.AmountCents != nil && *r.AmountCents < 0 {
				return fmt.Errorf("negative schedule %s", r.Key)
			}
		default:
			return fmt.Errorf("invalid kind for %s", r.Key)
		}
		rs[r.Key] = r
	}
	for _, r := range b.Records {
		for _, k := range r.RelatedKeys {
			target, ok := rs[k]
			if !ok {
				return fmt.Errorf("missing related record %s", k)
			}
			if r.Kind == "refund" && (target.Kind != "expense" || target.Currency != r.Currency || target.AccountKey != r.AccountKey) {
				return fmt.Errorf("invalid refund relation %s", r.Key)
			}
		}
	}
	return nil
}

// Run checks every record even on repeat imports. Changed content under the same
// key is a conflict, never a silent overwrite. A user row lock serializes imports.
func Run(ctx context.Context, db *gorm.DB, user string, b Bundle, commit bool) (Summary, error) {
	out := Summary{Records: len(b.Records), ByKind: map[string]int{}, NetExpenseCents: map[string]int64{}, BalancePolicy: "历史导入不改动现有账户余额；新账户余额需另行核实设置"}
	if err := Validate(b); err != nil {
		return out, err
	}
	err := db.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		var uid int64
		if err := tx.Raw("SELECT id FROM users WHERE id::text = ? FOR UPDATE", user).Scan(&uid).Error; err != nil {
			return err
		}
		if uid == 0 {
			return errors.New("user not found")
		}
		var old []models.CuratedImportRecord
		if err := tx.Where("user_id = ?", user).Find(&old).Error; err != nil {
			return err
		}
		oldMap := map[string]models.CuratedImportRecord{}
		for _, r := range old {
			oldMap[r.SourceKey] = r
		}
		newRows := []models.CuratedImportRecord{}
		newTx := []models.Transaction{}
		accountIDs := map[string]string{}
		for _, a := range b.Accounts {
			var found []models.Account
			if err := tx.Where("user_id = ? AND institution = ? AND card_last4 = ? AND currency = ?", user, a.Institution, a.CardLast4, a.Currency).Find(&found).Error; err != nil {
				return err
			}
			if len(found) > 1 {
				return fmt.Errorf("ambiguous account %s", a.Key)
			}
			if len(found) == 1 {
				accountIDs[a.Key] = found[0].ID
				continue
			}
			id := ID(user, "account:"+a.Key)
			accountIDs[a.Key] = id
			if commit {
				a := a
				if err := tx.Create(&models.Account{ID: id, UserID: user, Type: "credit", Institution: &a.Institution, CardLast4: &a.CardLast4, Currency: a.Currency, Balance: 0}).Error; err != nil {
					return err
				}
			}
		}
		for _, r := range b.Records {
			out.ByKind[r.Kind]++
			if r.Kind == "expense" {
				out.NetExpenseCents[r.Currency] += *r.AmountCents
			}
			if r.Kind == "refund" {
				out.NetExpenseCents[r.Currency] -= *r.AmountCents
			}
			payload, _ := json.Marshal(r)
			hash := sha256.Sum256(payload)
			digest := hex.EncodeToString(hash[:])
			if prior, ok := oldMap[r.Key]; ok {
				if prior.Digest != digest {
					return fmt.Errorf("record %s changed; reconcile before reimport", r.Key)
				}
				out.Skipped++
				continue
			}
			out.New++
			row := models.CuratedImportRecord{ID: ID(user, r.Key), UserID: user, SourceKey: r.Key, Kind: r.Kind, Currency: r.Currency, Date: r.Date, Payload: string(payload), Digest: digest}
			if r.Kind == "expense" || r.Kind == "refund" {
				// A refund is a negative expense, not income. Only this reviewed importer
				// writes signed expense amounts; the ordinary entry endpoint stays positive.
				amt := float64(*r.AmountCents) / 100
				if r.Kind == "refund" {
					amt = -amt
				}
				note := r.Description
				if r.Kind == "refund" {
					note = "退款｜" + note
				}
				if len(r.Tags) > 0 {
					note += "\n标签：" + strings.Join(r.Tags, "、")
				}
				if r.Note != "" {
					note += "\n" + r.Note
				}
				if r.RelationStatus != "" {
					note += "\n关联状态：" + r.RelationStatus
				}
				aid := accountIDs[r.AccountKey]
				category := r.Category
				row.TransactionID = row.ID
				newTx = append(newTx, models.Transaction{ID: row.ID, UserID: user, Type: "expense", Amount: amt, AccountID: &aid, CategoryKey: &category, Remark: &note, OccurredAt: r.Date + "T00:00:00+08:00"})
			}
			newRows = append(newRows, row)
		}
		if !commit {
			return nil
		}
		if len(newTx) > 0 {
			if err := tx.CreateInBatches(newTx, 200).Error; err != nil {
				return err
			}
		}
		if len(newRows) > 0 {
			return tx.CreateInBatches(newRows, 200).Error
		}
		return nil
	})
	return out, err
}
