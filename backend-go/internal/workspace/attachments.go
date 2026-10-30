package workspace

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"mime"
	"net/http"
	"neverbroke/internal/http/response"
	"path/filepath"
	"strings"
)

func attachmentRoutes(g *gin.RouterGroup, db *gorm.DB) {
	g.POST("/workspace/attachments", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 11<<20)
		file, e := c.FormFile("file")
		if e != nil || file.Size > 10<<20 {
			response.BadRequest(c, "请选择不超过 10 MB 的 PDF、PNG 或 JPEG")
			return
		}
		key := c.PostForm("key")
		entries, _, e := Load(db, c.GetString("user_id"))
		if e != nil {
			response.InternalError(c, "读取记录失败")
			return
		}
		found := false
		for _, x := range entries {
			if x.Key == key {
				found = true
			}
		}
		if !found {
			response.BadRequest(c, "记录不存在")
			return
		}
		reader, e := file.Open()
		if e != nil {
			response.BadRequest(c, "无法读取附件")
			return
		}
		defer reader.Close()
		data, e := io.ReadAll(io.LimitReader(reader, 10<<20+1))
		if e != nil || len(data) > 10<<20 {
			response.BadRequest(c, "附件过大")
			return
		}
		typ := http.DetectContentType(data)
		if typ != "application/pdf" && typ != "image/png" && typ != "image/jpeg" {
			response.BadRequest(c, "只支持 PDF、PNG 和 JPEG")
			return
		}
		a := Attachment{ID: uuid.NewString(), UserID: c.GetString("user_id"), RecordKey: key, Name: filepath.Base(file.Filename), Mime: typ, Size: len(data), Data: data}
		if len(a.Name) > 250 {
			response.BadRequest(c, "文件名过长")
			return
		}
		if e := db.Create(&a).Error; e != nil {
			response.InternalError(c, "保存附件失败")
			return
		}
		response.OK(c, a)
	})
	g.GET("/workspace/attachments/:id", func(c *gin.Context) {
		var a Attachment
		if e := db.Where("user_id = ? AND id = ?", c.GetString("user_id"), c.Param("id")).First(&a).Error; e != nil {
			c.Status(404)
			return
		}
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "private, no-store")
		c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": a.Name}))
		c.Data(200, a.Mime, a.Data)
	})
	g.GET("/workspace/sync", func(c *gin.Context) {
		runs := []SyncRun{}
		if e := db.Where("user_id = ?", c.GetString("user_id")).Order("created_at DESC").Limit(30).Find(&runs).Error; e != nil {
			response.InternalError(c, "读取同步状态失败")
			return
		}
		response.OK(c, runs)
	})
	g.POST("/workspace/sync", func(c *gin.Context) {
		var r SyncRun
		if c.ShouldBindJSON(&r) != nil || r.Records < 0 || r.Conflicts < 0 || len(r.Error) > 1000 {
			response.BadRequest(c, "同步状态无效")
			return
		}
		if !strings.Contains("|success|failed|conflict|preview|", "|"+r.Status+"|") {
			response.BadRequest(c, "同步状态无效")
			return
		}
		r.ID = uuid.NewString()
		r.UserID = c.GetString("user_id")
		if e := db.Create(&r).Error; e != nil {
			response.InternalError(c, "保存同步状态失败")
			return
		}
		response.OK(c, r)
	})
}
