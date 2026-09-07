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
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("连接Redis失败: %v", err)
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