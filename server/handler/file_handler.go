package handler

import (
	"github.com/gin-gonic/gin"

	"simplechat/server/config"
	"simplechat/server/service"
)

// FileHandler 文件上传
type FileHandler struct {
	file *service.FileService
}

func NewFileHandler(cfg *config.Config) *FileHandler {
	return &FileHandler{file: service.NewFileService(&cfg.Minio)}
}

type uploadResp struct {
	FileID   int64  `json:"file_id"`
	FileURL  string `json:"file_url"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

func (h *FileHandler) Upload(c *gin.Context) {
	fieldID := "type"
	if c.Query("type") == "image" {
		fieldID = "image"
	} else {
		fieldID = "file"
	}
	fileHeader, err := c.FormFile(fieldID)
	if err != nil {
		// 也允许字段名为 file/image 通用
		fileHeader, err = c.FormFile("file")
	}
	if err != nil {
		BadRequest(c, "请选择文件(type=file 或 image 字段)")
		return
	}
	rec, err := h.file.Upload(uid(c), fileHeader, fieldID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, uploadResp{
		FileID:  rec.FileID,
		FileURL: rec.FileURL,
		FileName: rec.FileName,
		FileSize: rec.FileSize,
		MimeType: derefStr(rec.MimeType),
	})
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}