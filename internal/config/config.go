package config

import "os"

type Config struct {
	Addr         string
	Env          string
	PaymentName  string
	StripeAPIKey  string
	StripeWebhook string
	DBDSN        string
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() Config {
	return Config{
		Addr:         getenv("PAYMENT_CENTER_ADDR", ":8080"),
		Env:          getenv("APP_ENV", "dev"),
		PaymentName:  getenv("PAYMENT_CENTER_NAME", "payment-center"),
		StripeAPIKey:  os.Getenv("STRIPE_API_KEY"),
		StripeWebhook: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		DBDSN:        getenv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/payment_center?charset=utf8mb4&parseTime=true&loc=Local"),
	}
}
