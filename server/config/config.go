package config

import (
	"os"
	"strconv"
)

// Config 服务端总配置
type Config struct {
	Server   ServerConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	Minio    MinioConfig
}

type ServerConfig struct {
	Port    string
	JWTSecret string
	JWTExpireDays int // 默认 7，记住我 30
	WebOrigin string // CORS 允许来源
}

type MySQLConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string // 签名或公网可访问前缀
}

// Load 从环境变量加载配置，未设置时使用默认值(本地开发)
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:       env("SERVER_PORT", "8080"),
			JWTSecret:  env("JWT_SECRET", "simplechat-dev-secret-change-me"),
			JWTExpireDays: envInt("JWT_EXPIRE_DAYS", 7),
			WebOrigin:  env("WEB_ORIGIN", "http://localhost:5173"),
		},
		MySQL: MySQLConfig{
			User:     env("MYSQL_USER", "simplechat"),
			Password: env("MYSQL_PASSWORD", "simplechat123456"),
			Host:     env("MYSQL_HOST", "127.0.0.1"),
			Port:     env("MYSQL_PORT", "3306"),
			DBName:   env("MYSQL_DB", "simplechat"),
		},
		Redis: RedisConfig{
			Addr:     env("REDIS_ADDR", "127.0.0.1:6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
		},
		Minio: MinioConfig{
			Endpoint:  env("MINIO_ENDPOINT", "127.0.0.1:9000"),
			AccessKey: env("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: env("MINIO_SECRET_KEY", "minioadmin123"),
			Bucket:    env("MINIO_BUCKET", "simplechat-files"),
			UseSSL:    envBool("MINIO_USE_SSL", false),
			PublicURL: env("MINIO_PUBLIC_URL", "http://127.0.0.1:9000"),
		},
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}