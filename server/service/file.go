package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"

	"simplechat/server/config"
	"simplechat/server/dao"
	"simplechat/server/model"
	"simplechat/server/util"
)

// 上传大小限制
const (
	MaxImageSize = 10 << 20                 // 图片 10MB
	MaxFileSize  = 100 << 20                // 文件 100MB
)

var forbiddenExts = map[string]bool{
	".exe": true, ".bat": true, ".sh": true, ".cmd": true,
	".dll": true, ".msi": true, ".com": true, ".scr": true, ".ps1": true,
}

// FileService 文件传输
type FileService struct {
	cfg *config.MinioConfig
}

func NewFileService(cfg *config.MinioConfig) *FileService { return &FileService{cfg: cfg} }

// Upload 上传文件到MinIO, isImage决定大小限制与类型
func (s *FileService) Upload(userID int64, file *multipart.FileHeader, field string) (*model.FileRecord, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if forbiddenExts[ext] {
		return nil, errors.New("禁止上传可执行文件")
	}
	if field == "image" || strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		if file.Size > MaxImageSize {
			return nil, errors.New("图片大小不能超过10MB")
		}
	} else {
		if file.Size > MaxFileSize {
			return nil, errors.New("文件大小不能超过100MB")
		}
	}

	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 计算MD5(流式)
	md5h := md5.New()
	io.Copy(md5h, f)
	md5Str := hex.EncodeToString(md5h.Sum(nil))
	f.Seek(0, io.SeekStart)

	objectName := fmt.Sprintf("chat/%d/%d_%s",
		userID, time.Now().UnixMilli(), sanitize(file.Filename))
	ctx := context.Background()
	_, err = dao.MinioClient.PutObject(ctx, s.cfg.Bucket, objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return nil, err
	}

	fileURL := fmt.Sprintf("%s/%s/%s", s.cfg.PublicURL, s.cfg.Bucket, objectName)
	mime := file.Header.Get("Content-Type")
	rec := &model.FileRecord{
		FileID:     util.GenID(),
		FileName:   file.Filename,
		FileSize:   file.Size,
		FileURL:    fileURL,
		FileMD5:    &md5Str,
		MimeType:   &mime,
		UploaderID: userID,
		UploadTime: time.Now(),
	}
	if err := dao.DB.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

func sanitize(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	for _, c := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		name = strings.ReplaceAll(name, c, "_")
	}
	if len([]rune(name)) > 80 {
		r := []rune(name)
		ext := filepath.Ext(name)
		name = string(r[:50]) + ext
	}
	return name
}