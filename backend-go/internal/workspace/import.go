package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"neverbroke/internal/curatedimport"
	"neverbroke/internal/http/response"
	"neverbroke/internal/models"
	"sort"
	"strings"
)

type Resolution struct {
	Action string `json:"action"`
	Target string `json:"target"`
}
type ImportRequest struct {
	Bundle      curatedimport.Bundle  `json:"bundle"`
	Token       string                `json:"token"`
	Resolutions map[string]Resolution `json:"resolutions"`
}
type Difference struct {
	Key        string                `json:"key"`
	Type       string                `json:"type"`
	Before     *curatedimport.Record `json:"before,omitempty"`
	After      curatedimport.Record  `json:"after"`
	Candidates []string              `json:"candidates"`
}
type ImportPreview struct {
	Token       string            `json:"token"`
	New         int               `json:"new"`
	Updated     int               `json:"updated"`
	Skipped     int               `json:"skipped"`
	Conflicts   int               `json:"conflicts"`
	Differences []Difference      `json:"differences"`
	Totals      map[string]Metric `json:"totals"`
}
type Snapshot struct {
	BeforePayload string                      `json:"beforePayload"`
	BeforeDigest  string                      `json:"beforeDigest"`
	AfterDigest   string                      `json:"afterDigest"`
	Key           string                      `json:"key"`
	Before        *models.CuratedImportRecord `json:"before"`
	After         models.CuratedImportRecord  `json:"after"`
	TxBefore      *models.Transaction         `json:"txBefore"`
	TxAfter       *models.Transaction         `json:"txAfter"`
}

func digest(v any) string {
	p, _ := json.Marshal(v)
	h := sha256.Sum256(p)
	return hex.EncodeToString(h[:])
}
func sameBusiness(a, b curatedimport.Record) bool {
	return (a.Kind == "expense" || a.Kind == "refund") && a.Kind == b.Kind && a.AccountKey == b.AccountKey && a.Currency == b.Currency && a.Date == b.Date && a.AmountCents != nil && b.AmountCents != nil && *a.AmountCents == *b.AmountCents && (a.Description == b.Description || (Text(a.Metadata["商户归一"]) != "" && Text(a.Metadata["商户归一"]) == Text(b.Metadata["商户归一"])))
}
func preview(b curatedimport.Bundle, old []models.CuratedImportRecord) (ImportPreview, error) {
	if e := curatedimport.Validate(b); e != nil {
		return ImportPreview{}, e
	}
	return previewValidated(b, old)
}
func previewValidated(b curatedimport.Bundle, old []models.CuratedImportRecord) (ImportPreview, error) {
	p := ImportPreview{Differences: []Difference{}, Totals: map[string]Metric{}}
	om := map[string]models.CuratedImportRecord{}
	rs := map[string]curatedimport.Record{}
	versions := map[string]string{}
	for _, r := range old {
		om[r.SourceKey] = r
		versions[r.SourceKey] = r.Digest
		var v curatedimport.Record
		if e := json.Unmarshal([]byte(r.Payload), &v); e != nil {
			return p, e
		}
		rs[r.SourceKey] = v
	}
	p.Token = digest(struct {
		B curatedimport.Bundle
		V map[string]string
	}{b, versions})
	for _, r := range b.Records {
		m := p.Totals[r.Currency]
		m.Add(Entry{Kind: r.Kind, Cents: r.AmountCents})
		p.Totals[r.Currency] = m
		if prior, ok := om[r.Key]; ok {
			if prior.Digest == digest(r) {
				p.Skipped++
				continue
			}
			before := rs[r.Key]
			p.Updated++
			p.Conflicts++
			p.Differences = append(p.Differences, Difference{Key: r.Key, Type: "changed", Before: &before, After: r, Candidates: []string{}})
			continue
		}
		candidates := []string{}
		for key, v := range rs {
			if sameBusiness(r, v) {
				candidates = append(candidates, key)
			}
		}
		sort.Strings(candidates)
		typ := "new"
		if len(candidates) > 0 {
			typ = "duplicate"
			p.Conflicts++
		}
		p.New++
		p.Differences = append(p.Differences, Difference{Key: r.Key, Type: typ, After: r, Candidates: candidates})
		rs[r.Key] = r
	}
	return p, nil
}
func importBundle(c *gin.Context, db *gorm.DB, commit bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 24<<20)
	var req ImportRequest
	if c.ShouldBindJSON(&req) != nil {
		response.BadRequest(c, "导入请求格式无效或文件过大")
		return
	}
	user := c.GetString("user_id")
	var result ImportPreview
	batch := Batch{ID: uuid.NewString(), UserID: user, Status: "applied"}
	e := db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if e := lock(tx, user); e != nil {
			return e
		}
		var old []models.CuratedImportRecord
		if e := tx.Where("user_id = ?", user).Find(&old).Error; e != nil {
			return e
		}
		var aliases []SourceAlias
		if e := tx.Where("user_id = ?", user).Find(&aliases).Error; e != nil {
			return e
		}
		aliasMap := map[string]string{}
		for _, a := range aliases {
			aliasMap[a.SourceKey] = a.TargetKey
		}
		req.Bundle = remapBundle(req.Bundle, aliasMap)
		// Validate incremental references against the existing source records.
		full, e := completeBundle(tx, user, req.Bundle, old)
		if e != nil {
			return e
		}
		if e = curatedimport.Validate(full); e != nil {
			return e
		}
		result, e = previewValidated(req.Bundle, old)
		if e != nil {
			return e
		}
		if !commit {
			return nil
		}
		if req.Token == "" || req.Token != result.Token {
			return errors.New("导入预览已过期，请重新预览")
		}
		om := map[string]models.CuratedImportRecord{}
		for _, r := range old {
			om[r.SourceKey] = r
		}
		replacementMap := map[string]string{}
		for _, d := range result.Differences {
			choice := req.Resolutions[d.Key]
			if choice.Action != "replace" {
				continue
			}
			if d.Type != "duplicate" {
				return errors.New("只有疑似重复记录可替换已有摘要")
			}
			matched := false
			for _, k := range d.Candidates {
				if k == choice.Target {
					matched = true
				}
			}
			if !matched {
				return errors.New("替换目标不在候选中")
			}
			prior, ok := om[choice.Target]
			if !ok {
				return errors.New("替换目标须为已有记录")
			}
			var previous curatedimport.Record
			json.Unmarshal([]byte(prior.Payload), &previous)
			if !strings.Contains(Text(previous.Metadata["账单状态"]), "摘要") && !strings.Contains(strings.Join(previous.Tags, " "), "未出账") {
				return errors.New("仅允许正式账单替换待确认摘要")
			}
			if !strings.Contains(Text(d.After.Metadata["账单状态"]), "正式") {
				return errors.New("新记录必须来自正式账单")
			}
			replacementMap[d.Key] = choice.Target
		}
		if len(replacementMap) > 0 {
			req.Bundle = remapBundle(req.Bundle, replacementMap)
			for oldKey, newKey := range replacementMap {
				req.Resolutions[newKey] = Resolution{Action: "update"}
				if e := tx.Create(&SourceAlias{UserID: user, SourceKey: oldKey, TargetKey: newKey, BatchID: batch.ID}).Error; e != nil {
					return e
				}
			}
			result, e = previewValidated(req.Bundle, old)
			if e != nil {
				return e
			}
		}
		accepted := map[string]bool{}
		for _, d := range result.Differences {
			choice := req.Resolutions[d.Key]
			switch d.Type {
			case "changed":
				if choice.Action != "update" && choice.Action != "keep" {
					return fmt.Errorf("%s 需要选择保留或更新", d.Key)
				}
				if choice.Action == "keep" {
					continue
				}
			case "duplicate":
				if choice.Action != "distinct" && choice.Action != "skip" {
					return fmt.Errorf("%s 疑似重复，需要人工确认", d.Key)
				}
				if choice.Action == "skip" {
					continue
				}
			}
			accepted[d.Key] = true
		}
		// A skipped new key cannot leave a dangling relation. Resolve it in the source bundle.
		for _, r := range req.Bundle.Records {
			if !accepted[r.Key] {
				continue
			}
			for _, k := range r.RelatedKeys {
				if _, ok := om[k]; !ok && !accepted[k] {
					return errors.New("跳过记录仍被其他记录引用，请先修订关联后预览")
				}
			}
		}
		snapshots := []Snapshot{}
		filtered := req.Bundle
		filtered.Records = []curatedimport.Record{}
		for _, r := range req.Bundle.Records {
			before, exists := om[r.Key]
			if !accepted[r.Key] {
				batch.Skipped++
				continue
			}
			if exists {
				var prior curatedimport.Record
				json.Unmarshal([]byte(before.Payload), &prior)
				if prior.Kind != r.Kind || prior.Currency != r.Currency || prior.AccountKey != r.AccountKey {
					return errors.New("修订不能更改记录业务类型、币种或账户，请在来源核对")
				}
				var edits int64
				if e := tx.Model(&State{}).Where("user_id = ? AND key = ?", user, r.Key).Count(&edits).Error; e != nil {
					return e
				}
				if edits > 0 {
					return errors.New("该记录存在人工修改；请保留原值并先核对人工版本")
				}
				snap := Snapshot{Key: r.Key, Before: &before}
				after := before
				payload, _ := json.Marshal(r)
				after.Payload = string(payload)
				after.Digest = digest(r)
				if before.TransactionID != "" {
					var t models.Transaction
					if e := tx.Where("user_id = ? AND id = ?", user, before.TransactionID).First(&t).Error; e != nil {
						return e
					}
					saved := t
					snap.TxBefore = &saved
					t.Amount = float64(*r.AmountCents) / 100
					if r.Kind == "refund" {
						t.Amount = -t.Amount
					}
					category := r.Category
					remark := r.Description + "\n" + r.Note
					t.CategoryKey = &category
					t.Remark = &remark
					t.OccurredAt = r.Date + "T00:00:00+08:00"
					if e := tx.Save(&t).Error; e != nil {
						return e
					}
					snap.TxAfter = &t
				}
				after.Date = r.Date
				if e := tx.Save(&after).Error; e != nil {
					return e
				}
				snap.After = after
				snapshots = append(snapshots, snap)
				batch.Updated++
			} else {
				filtered.Records = append(filtered.Records, r)
			}
		}
		// Validate using complete accepted source records, so references to old rows remain resolvable.
		if len(filtered.Records) > 0 {
			complete := filtered
			complete.Accounts = full.Accounts
			newKeys := map[string]bool{}
			for _, r := range filtered.Records {
				newKeys[r.Key] = true
			}
			var current []models.CuratedImportRecord
			if e := tx.Where("user_id = ?", user).Find(&current).Error; e != nil {
				return e
			}
			for _, r := range current {
				var v curatedimport.Record
				json.Unmarshal([]byte(r.Payload), &v)
				complete.Records = append(complete.Records, v)
			}
			// Require account definitions for all referenced historical accounts; full export supplies these.
			if _, e := curatedimport.Run(c.Request.Context(), tx, user, complete, true); e != nil {
				return e
			}
			for key := range newKeys {
				var after models.CuratedImportRecord
				if e := tx.Where("user_id = ? AND source_key = ?", user, key).First(&after).Error; e != nil {
					return e
				}
				snap := Snapshot{Key: key, After: after}
				if after.TransactionID != "" {
					var t models.Transaction
					if e := tx.Where("user_id = ? AND id = ?", user, after.TransactionID).First(&t).Error; e != nil {
						return e
					}
					snap.TxAfter = &t
				}
				snapshots = append(snapshots, snap)
				batch.New++
			}
		}
		entries, _, e := Load(tx, user)
		if e != nil {
			return e
		}
		byKey := map[string]Entry{}
		for _, x := range entries {
			byKey[x.Key] = x
		}
		for _, x := range entries {
			if x.Edit.RelationReviewed {
				if e := validateEdit(x, x.Edit, byKey); e != nil {
					return e
				}
			}
		}
		for i := range snapshots {
			snapshots[i].AfterDigest = snapshots[i].After.Digest
			if snapshots[i].Before != nil {
				snapshots[i].BeforePayload = snapshots[i].Before.Payload
				snapshots[i].BeforeDigest = snapshots[i].Before.Digest
			}
		}
		payload, _ := json.Marshal(snapshots)
		batch.Details = string(payload)
		return tx.Create(&batch).Error
	})
	if e != nil {
		if commit {
			batch.Status = "failed"
			batch.Error = e.Error()
			batch.New = 0
			batch.Updated = 0
			batch.Details = "[]"
			db.Create(&batch)
		}
		response.BadRequest(c, e.Error())
		return
	}
	if commit {
		response.OK(c, batch)
	} else {
		response.OK(c, result)
	}
}
func undoBatch(c *gin.Context, db *gorm.DB) {
	user := c.GetString("user_id")
	e := db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if e := lock(tx, user); e != nil {
			return e
		}
		var b Batch
		if e := tx.Where("user_id = ? AND id = ?", user, c.Param("id")).First(&b).Error; e != nil {
			return e
		}
		if b.Status != "applied" {
			return errors.New("只能撤销成功且尚未撤销的批次")
		}
		var snaps []Snapshot
		if e := json.Unmarshal([]byte(b.Details), &snaps); e != nil {
			return e
		}
		for i := range snaps {
			snaps[i].After.Digest = snaps[i].AfterDigest
			if snaps[i].Before != nil {
				snaps[i].Before.UserID = user
				snaps[i].Before.Payload = snaps[i].BeforePayload
				snaps[i].Before.Digest = snaps[i].BeforeDigest
			}
		}
		keys := map[string]bool{}
		for _, s := range snaps {
			keys[s.Key] = true
			var current models.CuratedImportRecord
			if e := tx.Where("user_id = ? AND source_key = ?", user, s.Key).First(&current).Error; e != nil {
				return e
			}
			if current.Digest != s.After.Digest {
				return errors.New("记录已有后续修改，不能撤销")
			}
			if s.TxAfter != nil {
				var currentTx models.Transaction
				if e := tx.Where("user_id = ? AND id = ?", user, s.TxAfter.ID).First(&currentTx).Error; e != nil {
					return e
				}
				if digest(currentTx) != digest(s.TxAfter) {
					return errors.New("交易已有后续修改，不能撤销")
				}
			}
			var n int64
			if e := tx.Model(&State{}).Where("user_id = ? AND key = ?", user, s.Key).Count(&n).Error; e != nil {
				return e
			}
			if n > 0 {
				return errors.New("记录已人工处理，不能撤销")
			}
			if e := tx.Model(&Attachment{}).Where("user_id = ? AND record_key = ?", user, s.Key).Count(&n).Error; e != nil {
				return e
			}
			if n > 0 {
				return errors.New("记录已添加附件，不能撤销")
			}
		}
		entries, _, e := Load(tx, user)
		if e != nil {
			return e
		}
		for _, x := range entries {
			if keys[x.Key] {
				continue
			}
			for _, k := range x.RelatedKeys {
				for _, s := range snaps {
					if s.Before == nil && s.Key == k {
						return errors.New("已有核对记录引用了本批次，不能撤销")
					}
				}
			}
		}
		var all []models.CuratedImportRecord
		if e := tx.Where("user_id = ?", user).Find(&all).Error; e != nil {
			return e
		}
		for _, row := range all {
			if keys[row.SourceKey] {
				continue
			}
			var r curatedimport.Record
			json.Unmarshal([]byte(row.Payload), &r)
			for _, key := range r.RelatedKeys {
				for _, s := range snaps {
					if s.Key == key && s.Before == nil {
						return errors.New("后续记录引用了本批次，不能撤销")
					}
				}
			}
		}
		for _, s := range snaps {
			if s.Before == nil {
				if s.TxAfter != nil {
					if e := tx.Where("user_id = ? AND id = ?", user, s.TxAfter.ID).Delete(&models.Transaction{}).Error; e != nil {
						return e
					}
				}
				if e := tx.Where("user_id = ? AND source_key = ?", user, s.Key).Delete(&models.CuratedImportRecord{}).Error; e != nil {
					return e
				}
			} else {
				if e := tx.Save(s.Before).Error; e != nil {
					return e
				}
				if s.TxBefore != nil {
					if e := tx.Save(s.TxBefore).Error; e != nil {
						return e
					}
				}
			}
		}
		if e := tx.Where("user_id = ? AND batch_id = ?", user, b.ID).Delete(&SourceAlias{}).Error; e != nil {
			return e
		}
		b.Status = "undone"
		return tx.Save(&b).Error
	})
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.OK(c, gin.H{"status": "undone"})
}

func remapBundle(b curatedimport.Bundle, aliases map[string]string) curatedimport.Bundle {
	if len(aliases) == 0 {
		return b
	}
	result := b
	result.Records = []curatedimport.Record{}
	seen := map[string]int{}
	for _, r := range b.Records {
		original := r.Key
		if k, ok := aliases[r.Key]; ok {
			r.Key = k
		}
		links := []string{}
		for _, k := range r.RelatedKeys {
			if a, ok := aliases[k]; ok {
				k = a
			}
			links = append(links, k)
		}
		r.RelatedKeys = links
		if i, ok := seen[r.Key]; ok {
			if original != r.Key {
				result.Records[i] = r
			}
			continue
		}
		seen[r.Key] = len(result.Records)
		result.Records = append(result.Records, r)
	}
	return result
}
func completeBundle(db *gorm.DB, user string, b curatedimport.Bundle, old []models.CuratedImportRecord) (curatedimport.Bundle, error) {
	if b.Version != 1 || len(b.Records) == 0 {
		return b, errors.New("需要 version 1 和非空记录")
	}
	out := b
	out.Records = append([]curatedimport.Record{}, b.Records...)
	out.Accounts = append([]curatedimport.Account{}, b.Accounts...)
	keys := map[string]bool{}
	accounts := map[string]bool{}
	for _, r := range b.Records {
		keys[r.Key] = true
	}
	for _, a := range b.Accounts {
		accounts[a.Key] = true
	}
	var txs []models.Transaction
	var accs []models.Account
	if e := db.Where("user_id = ?", user).Find(&txs).Error; e != nil {
		return b, e
	}
	if e := db.Where("user_id = ?", user).Find(&accs).Error; e != nil {
		return b, e
	}
	am := map[string]models.Account{}
	for _, a := range accs {
		am[a.ID] = a
	}
	tm := map[string]models.Transaction{}
	for _, t := range txs {
		tm[t.ID] = t
	}
	for _, row := range old {
		var r curatedimport.Record
		if e := json.Unmarshal([]byte(row.Payload), &r); e != nil {
			return b, e
		}
		if !keys[r.Key] {
			out.Records = append(out.Records, r)
		}
		if r.AccountKey != "" && !accounts[r.AccountKey] {
			a, ok := am[str(tm[row.TransactionID].AccountID)]
			if !ok {
				return b, errors.New("历史记录缺少账户映射")
			}
			out.Accounts = append(out.Accounts, curatedimport.Account{Key: r.AccountKey, Institution: str(a.Institution), CardLast4: str(a.CardLast4), Currency: a.Currency})
			accounts[r.AccountKey] = true
		}
	}
	return out, nil
}
