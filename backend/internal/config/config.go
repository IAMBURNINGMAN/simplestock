package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	SessionKey  string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", getEnv("DB_USER", "postgres"), getEnv("DB_PASSWORD", "postgres"), getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432"), getEnv("DB_NAME", "simplestock"))),
		SessionKey:  getEnv("SESSION_KEY", "simplestock-secret-key-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
