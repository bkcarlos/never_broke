package payroll

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	"time"
)

// SourceDocument archives a retrieval of the official notice; it does not approve
// new numeric rules. Policy changes still require review of the effective dates.
type SourceDocument struct {
	ID        string `gorm:"primaryKey;size:64"`
	URL       string `gorm:"type:text"`
	Name      string
	SHA256    string `gorm:"size:64"`
	FetchedAt time.Time
	Content   []byte `gorm:"type:bytea"`
}

func ArchiveOfficialSources(ctx context.Context, db *gorm.DB) (int, error) {
	if err := db.AutoMigrate(&SourceDocument{}); err != nil {
		return 0, err
	}
	client := &http.Client{Timeout: 25 * time.Second}
	seen := map[string]bool{}
	count := 0
	for _, rule := range officialRules {
		for _, source := range rule.Sources {
			if seen[source.URL] {
				continue
			}
			seen[source.URL] = true
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
			if err != nil {
				return count, err
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; NeverBroke policy archive)")
			res, err := client.Do(req)
			if err != nil {
				return count, fmt.Errorf("%s: %w", source.Name, err)
			}
			raw, readErr := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024+1))
			res.Body.Close()
			if readErr != nil {
				return count, readErr
			}
			if res.StatusCode != 200 || len(raw) == 0 || len(raw) > 4*1024*1024 {
				return count, fmt.Errorf("%s: invalid document response %d", source.Name, res.StatusCode)
			}
			hash := sha256.Sum256(raw)
			id := sha256.Sum256([]byte(source.URL))
			row := SourceDocument{ID: hex.EncodeToString(id[:]), URL: source.URL, Name: source.Name, SHA256: hex.EncodeToString(hash[:]), FetchedAt: time.Now().UTC(), Content: raw}
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}
