package dao

import (
	"context"
	"fmt"
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
	// 公开读策略: 允许匿名 GetObject, 使 PublicURL 生成的图片/文件 URL 可直接在浏览器 "<img>" 中访问
	// (否则匿名访问 403, 前端图片/头像会加载失败显示黑底)。
	// 幂等: 每次启动重新设置相同策略, 无副作用。
	if err := setPublicReadPolicy(ctx, client, cfg.Bucket); err != nil {
		log.Fatalf("设置Bucket公开读策略失败: %v", err)
	}
	MinioClient = client
}

// setPublicReadPolicy 设置只允许匿名读取(bucket/*)的桶策略
func setPublicReadPolicy(ctx context.Context, client *minio.Client, bucket string) error {
	return client.SetBucketPolicy(ctx, bucket, fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`,
		bucket))
}