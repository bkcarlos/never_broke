package payroll

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// PolicyHistory keeps reviewed effective periods independent of publication dates.
type PolicyHistory struct {
	ID               string `gorm:"primaryKey;size:80"`
	RegionCode       string `gorm:"index:idx_policy_period"`
	EffectiveFrom    string `gorm:"size:10;index:idx_policy_period"`
	EffectiveThrough string `gorm:"size:10"`
	Payload          string `gorm:"type:text"`
	VerifiedAt       string `gorm:"size:10"`
}

func SeedPolicyHistory(db *gorm.DB) error {
	if err := db.AutoMigrate(&PolicyHistory{}); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, r := range officialRules {
			response, err := findOfficialRule(r.Code, r.From)
			if err != nil {
				return err
			}
			raw, err := json.Marshal(response)
			if err != nil {
				return err
			}
			row := PolicyHistory{ID: r.Code + ":" + r.From, RegionCode: r.Code, EffectiveFrom: r.From, EffectiveThrough: r.Through, Payload: string(raw), VerifiedAt: response.VerifiedAt}
			// Existing reviewed records are never silently overwritten on restart.
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormRepository) RuleAt(code, date string) (*RegionRuleResponse, error) {
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return nil, fmt.Errorf("invalid rule date")
	}
	var rows []PolicyHistory
	if err := r.db.Where("region_code = ? AND effective_from <= ? AND effective_through >= ?", code, date, date).Limit(2).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrRuleNotFound
	}
	if len(rows) != 1 {
		return nil, errors.New("overlapping policy periods")
	}
	var rule RegionRuleResponse
	if err := json.Unmarshal([]byte(rows[0].Payload), &rule); err != nil {
		return nil, err
	}
	rule.AsOf = date
	return &rule, nil
}
