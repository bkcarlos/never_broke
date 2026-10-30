package curatedimport

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"neverbroke/internal/http/response"
	"neverbroke/internal/models"
	"strconv"
)

func ImportHandler(db *gorm.DB, commit bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<20)
		var b Bundle
		if err := c.ShouldBindJSON(&b); err != nil {
			response.BadRequest(c, "invalid or oversized import bundle")
			return
		}
		summary, err := Run(c.Request.Context(), db, c.GetString("user_id"), b, commit)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		response.OK(c, summary)
	}
}
func ListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := db.WithContext(c.Request.Context()).Where("user_id = ?", c.GetString("user_id")).Model(&models.CuratedImportRecord{})
		if k := c.Query("key"); k != "" {
			q = q.Where("source_key = ?", k)
		}
		if k := c.Query("kind"); k != "" {
			q = q.Where("kind = ?", k)
		}
		if k := c.Query("currency"); k != "" {
			q = q.Where("currency = ?", k)
		}
		if k := c.Query("date"); k != "" {
			q = q.Where("date = ?", k)
		}
		if k := c.Query("transactionId"); k != "" {
			q = q.Where("transaction_id = ?", k)
		}
		offset, _ := strconv.Atoi(c.Query("offset"))
		if offset < 0 {
			offset = 0
		}
		var total int64
		if err := q.Count(&total).Error; err != nil {
			response.InternalError(c, "failed to count records")
			return
		}
		var rows []models.CuratedImportRecord
		if err := q.Order("date DESC, source_key").Limit(100).Offset(offset).Find(&rows).Error; err != nil {
			response.InternalError(c, "failed to read records")
			return
		}
		items := make([]Record, 0, len(rows))
		for _, r := range rows {
			var item Record
			if err := json.Unmarshal([]byte(r.Payload), &item); err != nil {
				response.InternalError(c, "invalid stored record")
				return
			}
			items = append(items, item)
		}
		response.OK(c, gin.H{"items": items, "total": total, "offset": offset, "hasMore": int64(offset+len(items)) < total})
	}
}
