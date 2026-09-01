package config

import (
	"os"
	"strconv"
	"time"
)

// 配置
type Config struct {
	Addr          string
	Env           string
	PaymentName   string
	DBDSN         string
	AuthSecret    string
	TokenTTL      time.Duration
	AdminUsername string
	AdminPassword string
	GatewayBaseURL string
}

// 获取环境变量
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// 获取环境变量 int
func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// 加载配置
func Load() Config {
	return Config{
		Addr:          getenv("PAYMENT_CENTER_ADDR", ":8080"),
		Env:           getenv("APP_ENV", "dev"),
		PaymentName:   getenv("PAYMENT_CENTER_NAME", "payment-center"),
		DBDSN:         getenv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/payment_center?charset=utf8mb4&parseTime=true&loc=Local"),
		AuthSecret:    getenv("AUTH_SECRET", "dev-payment-center-secret"),
		TokenTTL:      time.Duration(getenvInt("AUTH_TOKEN_TTL_HOURS", 12)) * time.Hour,
		AdminUsername: getenv("ADMIN_USERNAME", "admin"),
		AdminPassword: getenv("ADMIN_PASSWORD", "admin123"),
		GatewayBaseURL: getenv("GATEWAY_BASE_URL", "http://127.0.0.1:8080"),
	}
}
