package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config 应用配置，从环境变量加载
type Config struct {
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	AESKey      string
	ServerPort  string
	LogLevel    string
	// OAuth 配置
	OAuthGitHubClientID     string
	OAuthGitHubClientSecret string
	// S3 兼容对象存储配置
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3Region    string
}

// Load 从环境变量加载配置
func Load() *Config {
	// 尝试加载 .env 文件，忽略错误（生产环境可能不存在 .env）
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://deployhub:deployhub@localhost:5432/deployhub?sslmode=disable"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		AESKey:                  getEnv("AES_KEY", ""),
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		LogLevel:                getEnv("LOG_LEVEL", "debug"),
		OAuthGitHubClientID:     getEnv("OAUTH_GITHUB_CLIENT_ID", ""),
		OAuthGitHubClientSecret: getEnv("OAUTH_GITHUB_CLIENT_SECRET", ""),
		S3Endpoint:              getEnv("S3_ENDPOINT", ""),
		S3AccessKey:             getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:             getEnv("S3_SECRET_KEY", ""),
		S3Bucket:                getEnv("S3_BUCKET", "deployhub"),
		S3Region:                getEnv("S3_REGION", "us-east-1"),
	}
}

// Validate 校验必填配置项
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return ErrMissingJWTSecret
	}
	if c.AESKey == "" {
		return ErrMissingAESKey
	}
	if len(c.AESKey) != 64 {
		return ErrInvalidAESKeyLength
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
