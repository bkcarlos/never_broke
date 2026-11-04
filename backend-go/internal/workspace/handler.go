package workspace

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"neverbroke/internal/categories"
	"neverbroke/internal/http/response"
	"neverbroke/internal/models"
	"strconv"
	"strings"
	"time"
)

func Register(g *gin.RouterGroup, db *gorm.DB) {
	attachmentRoutes(g, db)
	g.GET("/workspace/calendar", func(c *gin.Context) { calendar(c, db) })
	g.GET("/workspace/projection", func(c *gin.Context) {
		items, accounts, e := Load(db.WithContext(c.Request.Context()), c.GetString("user_id"))
		if e != nil {
			response.InternalError(c, "读取计划失败")
			return
		}
		var settings []Setting
		if e := db.Where("user_id = ?", c.GetString("user_id")).Find(&settings).Error; e != nil {
			response.InternalError(c, "读取余额状态失败")
			return
		}
		from := c.DefaultQuery("from", time.Now().Format("2006-01-02"))
		if _, e := time.Parse("2006-01-02", from); e != nil {
			response.BadRequest(c, "日期无效")
			return
		}
		response.OK(c, Project(items, accounts, settings, c.DefaultQuery("currency", "CNY"), from, 90))
	})
	g.GET("/workspace/categories", func(c *gin.Context) { response.OK(c, categories.All) })
	g.GET("/workspace", func(c *gin.Context) { read(c, db) })
	g.GET("/workspace/record", func(c *gin.Context) { detail(c, db) })
	g.POST("/workspace/edit", func(c *gin.Context) { edit(c, db) })
	g.GET("/workspace/settings", func(c *gin.Context) {
		var s []Setting
		if e := db.Where("user_id = ?", c.GetString("user_id")).Order("kind, name").Find(&s).Error; e != nil {
			response.InternalError(c, "读取设置失败")
			return
		}
		if s == nil {
			s = []Setting{}
		}
		response.OK(c, s)
	})
	g.POST("/workspace/settings", func(c *gin.Context) { setting(c, db) })
	g.POST("/workspace/import/preview", func(c *gin.Context) { importBundle(c, db, false) })
	g.POST("/workspace/import/commit", func(c *gin.Context) { importBundle(c, db, true) })
	g.GET("/workspace/batches", func(c *gin.Context) {
		var b []Batch
		if e := db.Omit("details").Where("user_id = ?", c.GetString("user_id")).Order("created_at DESC").Limit(100).Find(&b).Error; e != nil {
			response.InternalError(c, "读取批次失败")
			return
		}
		if b == nil {
			b = []Batch{}
		}
		response.OK(c, b)
	})
	g.POST("/workspace/batches/:id/undo", func(c *gin.Context) { undoBatch(c, db) })
}
func filters(c *gin.Context) (Filter, error) {
	f := Filter{From: c.Query("from"), To: c.Query("to"), Currency: c.DefaultQuery("currency", "CNY"), Bank: c.Query("bank"), Account: c.Query("account"), Category: c.Query("category"), Kind: c.Query("kind"), Confirmation: c.Query("confirmation"), Query: c.Query("q"), Queue: c.Query("queue"), Channel: c.Query("channel")}
	for _, d := range []string{f.From, f.To} {
		if d != "" {
			if _, e := time.Parse("2006-01-02", d); e != nil {
				return f, errors.New("日期格式错误")
			}
		}
	}
	if f.To != "" && f.From > f.To {
		return f, errors.New("开始日期晚于结束日期")
	}
	return f, nil
}
func read(c *gin.Context, db *gorm.DB) {
	f, e := filters(c)
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	year := c.Query("year")
	if year != "" {
		y, e := strconv.Atoi(year)
		if e != nil || y < 1900 || y > 2200 {
			response.BadRequest(c, "年份无效")
			return
		}
		f.From = year + "-01-01"
		f.To = year + "-12-31"
	}
	items, accounts, e := Load(db.WithContext(c.Request.Context()), c.GetString("user_id"))
	if e != nil {
		response.InternalError(c, "读取账本失败")
		return
	}
	rows := []Entry{}
	queue := map[string]int{"category": 0, "refund": 0, "receipt": 0, "pending": 0}
	years := map[string]bool{}
	for _, x := range items {
		if Matches(x, f) {
			rows = append(rows, x)
		}
		if x.Currency == f.Currency {
			for q := range queue {
				if Matches(x, Filter{Queue: q}) {
					queue[q]++
				}
			}
			if (x.Kind == "expense" || x.Kind == "refund") && len(x.Date) >= 4 {
				years[x.Date[:4]] = true
			}
		}
	}
	total := len(rows)
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + 50
	if end > total {
		end = total
	}
	response.OK(c, gin.H{"items": rows[offset:end], "total": total, "offset": offset, "hasMore": end < total, "quality": DataQuality(items, f.Currency), "analysis": Analyze(items, f, year), "accounts": accounts, "categories": categories.All, "queues": queue, "years": years, "scope": f})
}
func detail(c *gin.Context, db *gorm.DB) {
	items, _, e := Load(db.WithContext(c.Request.Context()), c.GetString("user_id"))
	if e != nil {
		response.InternalError(c, "读取失败")
		return
	}
	var found *Entry
	byKey := map[string]Entry{}
	for i := range items {
		byKey[items[i].Key] = items[i]
		if items[i].Key == c.Query("key") {
			found = &items[i]
		}
	}
	if found == nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	related := []Entry{}
	seen := map[string]bool{}
	for _, k := range found.RelatedKeys {
		if x, ok := byKey[k]; ok {
			related = append(related, x)
			seen[k] = true
		}
	}
	candidates := []Entry{}
	for _, x := range items {
		if found.Kind == "schedule" && x.Key != found.Key && x.Currency == found.Currency && (x.Kind == "statement" || x.Kind == "schedule" && x.Edit.Status == "billed") && (found.AccountID != "" && x.AccountID == found.AccountID || found.AccountID == "" && found.AccountName != "" && x.AccountName == found.AccountName) {
			a, _ := time.Parse("2006-01-02", found.Date)
			b, _ := time.Parse("2006-01-02", x.Date)
			if days := a.Sub(b).Hours() / 24; days >= -45 && days <= 45 {
				candidates = append(candidates, x)
			}
		}
		for _, k := range x.RelatedKeys {
			if k == found.Key && !seen[x.Key] {
				related = append(related, x)
				seen[x.Key] = true
			}
		}
		if (found.Kind == "refund" || found.Kind == "receipt") && x.Kind == "expense" && x.Currency == found.Currency && x.Key != found.Key && (found.Kind != "refund" || x.AccountID == found.AccountID) {
			a, _ := time.Parse("2006-01-02", found.Date)
			b, _ := time.Parse("2006-01-02", x.Date)
			days := a.Sub(b).Hours() / 24
			if days >= -3 && days <= 120 && (x.Merchant == found.Merchant || (found.Cents != nil && x.Cents != nil && *found.Cents == *x.Cents)) {
				candidates = append(candidates, x)
			}
		}
	}
	if len(candidates) > 100 {
		candidates = candidates[:100]
	}
	history := []Change{}
	if e := db.Where("user_id = ? AND key = ?", c.GetString("user_id"), found.Key).Order("created_at DESC").Limit(30).Find(&history).Error; e != nil {
		response.InternalError(c, "读取审计失败")
		return
	}
	attachments := []Attachment{}
	if e := db.Omit("data").Where("user_id = ? AND record_key = ?", c.GetString("user_id"), found.Key).Find(&attachments).Error; e != nil {
		response.InternalError(c, "读取附件失败")
		return
	}
	response.OK(c, gin.H{"attachments": attachments, "record": found, "related": related, "candidates": candidates, "history": history})
}
func lock(db *gorm.DB, user string) error {
	var id int64
	if e := db.Raw("SELECT id FROM users WHERE id::text = ? FOR UPDATE", user).Scan(&id).Error; e != nil {
		return e
	}
	if id == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

type EditItem struct {
	Key     string `json:"key"`
	Version int    `json:"version"`
	Edit    Edit   `json:"edit"`
}
type EditRequest struct {
	Items  []EditItem `json:"items"`
	Reason string     `json:"reason"`
}

func validateEdit(x Entry, edit Edit, all map[string]Entry) error {
	if len(edit.Note) > 10000 || len(edit.Tags) > 50 {
		return errors.New("说明或标签过长")
	}
	if edit.Category != "" && len(edit.Category) > 50 {
		return errors.New("分类过长")
	}
	if edit.Status != "" {
		if x.Kind != "schedule" {
			return errors.New("只有还款计划可设置状态")
		}
		switch edit.Status {
		case "estimated", "billed", "paid", "cancelled", "replaced":
		default:
			return errors.New("无效状态")
		}
	}
	if edit.Status == "replaced" {
		r, ok := all[edit.ReplacedBy]
		if !ok || r.Key == x.Key || (r.Kind != "statement" && r.Kind != "schedule") || r.Currency != x.Currency {
			return errors.New("请选择同币种的正式账单或还款事项")
		}
		if r.Kind == "schedule" && r.Edit.Status != "billed" {
			return errors.New("替代事项必须已出账")
		}
		if x.AccountID != "" && r.AccountID != x.AccountID || x.AccountID == "" && x.AccountName != "" && r.AccountName != x.AccountName {
			return errors.New("替代事项必须属于同一账户")
		}
	}
	if !edit.RelationReviewed {
		return nil
	}
	if x.Kind != "refund" && x.Kind != "receipt" {
		return errors.New("仅退款和凭证可以确认关联")
	}
	keys := map[string]bool{}
	for _, k := range edit.RelatedKeys {
		v, ok := all[k]
		if !ok || v.Kind != "expense" || v.Currency != x.Currency || x.Kind == "refund" && v.AccountID != x.AccountID {
			return errors.New("关联消费必须属于当前用户、同币种；退款还须同账户")
		}
		if keys[k] {
			return errors.New("重复关联")
		}
		keys[k] = true
	}
	if x.Kind == "refund" {
		var total int64
		allocated := map[string]bool{}
		for _, a := range edit.Allocations {
			v, ok := all[a.Key]
			if !ok || !keys[a.Key] || allocated[a.Key] || a.Cents <= 0 || v.Cents == nil {
				return errors.New("退款分配无效")
			}
			allocated[a.Key] = true
			total += a.Cents
			used := a.Cents
			for _, other := range all {
				if other.Key == x.Key || other.Kind != "refund" || !other.Edit.RelationReviewed {
					continue
				}
				for _, p := range other.Edit.Allocations {
					if p.Key == a.Key {
						used += p.Cents
					}
				}
			}
			if used > *v.Cents {
				return errors.New("累计退款不能超过原消费金额")
			}
		}
		if len(keys) > 0 && (x.Cents == nil || total != *x.Cents || len(allocated) != len(keys)) {
			return errors.New("分配合计必须等于本笔退款金额")
		}
		if len(keys) == 0 && len(edit.Allocations) > 0 {
			return errors.New("解除关联时请清空分配")
		}
	}
	return nil
}
func edit(c *gin.Context, db *gorm.DB) {
	var req EditRequest
	if c.ShouldBindJSON(&req) != nil || len(req.Items) == 0 || len(req.Items) > 100 || strings.TrimSpace(req.Reason) == "" {
		response.BadRequest(c, "需要 1–100 项及修改原因")
		return
	}
	user := c.GetString("user_id")
	e := db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if e := lock(tx, user); e != nil {
			return e
		}
		entries, _, e := Load(tx, user)
		if e != nil {
			return e
		}
		all := map[string]Entry{}
		for _, x := range entries {
			all[x.Key] = x
		}
		seen := map[string]bool{}
		for _, v := range req.Items {
			x, ok := all[v.Key]
			if !ok {
				return errors.New("记录不存在")
			}
			if seen[v.Key] || v.Version != x.Version {
				return errors.New("记录已更改，请刷新后重试")
			}
			seen[v.Key] = true
			x.Edit = v.Edit
			all[v.Key] = x
		}
		for _, v := range req.Items {
			if e := validateEdit(all[v.Key], v.Edit, all); e != nil {
				return e
			}
		}
		for _, v := range req.Items {
			var old State
			e := tx.Where("user_id = ? AND key = ?", user, v.Key).First(&old).Error
			if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
				return e
			}
			p, _ := json.Marshal(v.Edit)
			state := State{UserID: user, Key: v.Key, Version: v.Version + 1, Payload: string(p)}
			if e := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "key"}}, DoUpdates: clause.AssignmentColumns([]string{"version", "payload", "updated_at"})}).Create(&state).Error; e != nil {
				return e
			}
			audit := Change{ID: uuid.NewString(), UserID: user, Key: v.Key, Before: old.Payload, After: string(p), Reason: req.Reason}
			if e := tx.Create(&audit).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.OK(c, gin.H{"updated": len(req.Items)})
}
func setting(c *gin.Context, db *gorm.DB) {
	var s Setting
	if c.ShouldBindJSON(&s) != nil {
		response.BadRequest(c, "设置格式错误")
		return
	}
	s.UserID = c.GetString("user_id")
	if s.Key == "" {
		s.Key = uuid.NewString()
	}
	if len(s.Name) > 200 || len(s.Match) > 200 || len(s.Key) > 180 {
		response.BadRequest(c, "内容过长")
		return
	}
	if s.Currency != "CNY" && s.Currency != "USD" {
		response.BadRequest(c, "币种无效")
		return
	}
	if s.Cents > 9999999999999 || s.Cents < -9999999999999 {
		response.BadRequest(c, "金额超出范围")
		return
	}
	switch s.Kind {
	case "rule":
		if s.Match == "" || s.Category == "" {
			response.BadRequest(c, "规则需要关键词与分类")
			return
		}
	case "budget":
		if s.Cents < 0 || s.Category == "" {
			response.BadRequest(c, "分类预算无效")
			return
		}
	case "subscription", "fixed_expense":
		if e := ValidateRecurringExpense(s); e != nil {
			response.BadRequest(c, e.Error())
			return
		}
	case "balance":
		if _, e := time.Parse("2006-01-02", s.Date); e != nil {
			response.BadRequest(c, "核验日期无效")
			return
		}
	default:
		response.BadRequest(c, "设置类型无效")
		return
	}
	e := db.Transaction(func(tx *gorm.DB) error {
		if e := lock(tx, s.UserID); e != nil {
			return e
		}
		var old Setting
		e := tx.Where("user_id = ? AND key = ?", s.UserID, s.Key).First(&old).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if old.Version != s.Version {
			return errors.New("设置已变更，请刷新")
		}
		if old.Kind != "" && old.Kind != s.Kind {
			return errors.New("不能更改设置类型")
		}
		if (s.Kind == "subscription" || s.Kind == "fixed_expense") && s.AccountID != "" {
			var a models.Account
			if e := tx.Where("user_id = ? AND id = ? AND currency = ? AND archived = ?", s.UserID, s.AccountID, s.Currency, false).First(&a).Error; e != nil || a.Type == "credit" {
				return errors.New("请选择同币种的现金类扣款账户")
			}
		}
		if s.Kind == "balance" {
			var a models.Account
			if e := tx.Where("user_id = ? AND id = ? AND currency = ?", s.UserID, s.Match, s.Currency).First(&a).Error; e != nil {
				return errors.New("账户不存在或币种不符")
			}
			if s.Key != "balance:"+s.Match {
				return errors.New("余额核验标识无效")
			}
			if e := tx.Model(&models.Account{}).Where("user_id = ? AND id = ?", s.UserID, s.Match).Update("balance", float64(s.Cents)/100).Error; e != nil {
				return e
			}
		}
		before, _ := json.Marshal(old)
		s.Version++
		after, _ := json.Marshal(s)
		if e := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&s).Error; e != nil {
			return e
		}
		return tx.Create(&Change{ID: uuid.NewString(), UserID: s.UserID, Key: "setting:" + s.Key, Before: string(before), After: string(after), Reason: "保存" + s.Kind}).Error
	})
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.OK(c, s)
}
