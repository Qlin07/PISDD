package dao

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"simplechat/server/config"
)

var (
	DB  *gorm.DB
	Rdb *redis.Client
)

// InitDB 初始化MySQL连接并自动迁移(首次建表)
func InitDB(cfg *config.MySQLConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接MySQL失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取底层DB失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db

	// 等待MySQL就绪(首次起拉、重启后数据库可能比后端晚就绪),
	// 避免后端启动时数据库未连上导致后续请求全部报错。
	if err := waitReady(30*time.Second, sqlDB.Ping); err != nil {
		log.Fatalf("连接MySQL失败(等待就绪超时): %v", err)
	}
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	// 等待Redis就绪
	if err := waitReady(30*time.Second, func() error {
		return Rdb.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("连接Redis失败(等待就绪超时): %v", err)
	}
}

// waitReady 重试执行 fn 直至成功或超时; 用于连接初始化时等待依赖服务就绪
func waitReady(timeout time.Duration, fn func() error) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			return lastErr
		}
		time.Sleep(time.Second)
	}
}

// Redis缓存Key
const (
	KeyUserOnline = "online:%d"     // 在线状态
	KeyOnlineMode = "user:%d"       // 用户信息缓存(Hash)
	KeyToken      = "token:%d"      // JWT
	KeyConvList   = "conv:list:%d"  // 会话列表缓存
	KeyUnread     = "unread:%d"     // 未读计数(Hash)
)

// ctx 通用context(带超时)
func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}