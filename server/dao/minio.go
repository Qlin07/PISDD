package dao

import (
	"context"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"simplechat/server/config"
)

// MinioClient MinIO客户端
var MinioClient *minio.Client

// InitMinio 初始化MinIO客户端并确保bucket存在
func InitMinio(cfg *config.MinioConfig) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("初始化MinIO失败: %v", err)
	}
	ctx := context.Background()
	// 等待MinIO就绪, 避免后端启动时对象存储未就绪导致退出
	var exists bool
	if err := waitReady(30*time.Second, func() error {
		var e error
		exists, e = client.BucketExists(ctx, cfg.Bucket)
		return e
	}); err != nil {
		log.Fatalf("检查Bucket失败(等待就绪超时): %v", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("创建Bucket失败: %v", err)
		}
	}
	MinioClient = client
}