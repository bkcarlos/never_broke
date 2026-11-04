package workspace

import "time"

// State overlays source data; version checks prevent silent lost updates.
type State struct {
	UserID    string    `gorm:"primaryKey" json:"-"`
	Key       string    `gorm:"primaryKey" json:"key"`
	Version   int       `json:"version"`
	Payload   string    `gorm:"type:text" json:"-"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Change struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"-"`
	Key       string    `gorm:"index" json:"key"`
	Before    string    `gorm:"type:text" json:"before"`
	After     string    `gorm:"type:text" json:"after"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}
type Allocation struct {
	Key   string `json:"key"`
	Cents int64  `json:"cents"`
}
type Edit struct {
	Category         string       `json:"category,omitempty"`
	Tags             []string     `json:"tags,omitempty"`
	Reviewed         bool         `json:"reviewed"`
	RelationReviewed bool         `json:"relationReviewed"`
	Allocations      []Allocation `json:"allocations"`
	RelatedKeys      []string     `json:"relatedKeys"`
	Status           string       `json:"status,omitempty"`
	ReplacedBy       string       `json:"replacedBy,omitempty"`
	Note             string       `json:"note,omitempty"`
}
type Setting struct {
	AccountID string    `json:"accountId,omitempty"`
	EndDate   string    `json:"endDate,omitempty"`
	UserID    string    `gorm:"primaryKey" json:"-"`
	Key       string    `gorm:"primaryKey" json:"key"`
	Kind      string    `gorm:"index" json:"kind"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Cents     int64     `json:"cents"`
	Category  string    `json:"category"`
	Match     string    `json:"match"`
	Date      string    `json:"date"`
	Period    string    `json:"period"`
	Enabled   bool      `json:"enabled"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Batch struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"-"`
	Status    string    `json:"status"`
	New       int       `json:"new"`
	Updated   int       `json:"updated"`
	Skipped   int       `json:"skipped"`
	Details   string    `gorm:"type:text" json:"-"`
	Error     string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

func (State) TableName() string   { return "workspace_states" }
func (Change) TableName() string  { return "workspace_changes" }
func (Setting) TableName() string { return "workspace_settings" }
func (Batch) TableName() string   { return "workspace_batches" }

// SourceAlias records a manually confirmed identity change (summary -> statement).
// The original ledger ID stays stable and existing receipt links remain valid.
type SourceAlias struct {
	UserID    string `gorm:"primaryKey" json:"-"`
	SourceKey string `gorm:"primaryKey" json:"sourceKey"`
	TargetKey string `json:"targetKey"`
	BatchID   string `json:"batchId"`
}

func (SourceAlias) TableName() string { return "workspace_source_aliases" }

type Attachment struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"-"`
	RecordKey string    `gorm:"index" json:"recordKey"`
	Name      string    `json:"name"`
	Mime      string    `json:"mime"`
	Size      int       `json:"size"`
	Data      []byte    `gorm:"type:bytea" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Attachment) TableName() string { return "workspace_attachments" }

type SyncRun struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"-"`
	Status    string    `json:"status"`
	Records   int       `json:"records"`
	Conflicts int       `json:"conflicts"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"createdAt"`
}

func (SyncRun) TableName() string { return "workspace_sync_runs" }
