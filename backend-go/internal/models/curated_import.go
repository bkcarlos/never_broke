package models

import "time"

// CuratedImportRecord preserves receipt links, uncertain refund matches and
// repayment forecasts separately from actual ledger entries. No email is needed.
type CuratedImportRecord struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID        string    `gorm:"not null;uniqueIndex:idx_import_user_key" json:"-"`
	SourceKey     string    `gorm:"not null;uniqueIndex:idx_import_user_key" json:"key"`
	Kind          string    `gorm:"not null;index" json:"kind"`
	Currency      string    `gorm:"not null" json:"currency"`
	Date          string    `gorm:"not null;index" json:"date"`
	Payload       string    `gorm:"type:text;not null" json:"-"`
	Digest        string    `gorm:"not null" json:"-"`
	TransactionID string    `gorm:"index" json:"transactionId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}
