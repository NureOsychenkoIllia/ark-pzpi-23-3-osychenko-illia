package config

import (
	"os"
)

// Config містить конфігурацію додатку
type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

// Load завантажує конфігурацію з змінних середовища
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/busoptima?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:        getEnv("PORT", "8080"),
	}
}

// getEnv повертає значення змінної середовища або значення за замовчуванням
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}