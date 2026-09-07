package dao

import (
	"context"
	"log"

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
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		log.Fatalf("检查Bucket失败: %v", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("创建Bucket失败: %v", err)
		}
	}
	MinioClient = client
}