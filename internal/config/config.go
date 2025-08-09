package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
	ServerPort string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost: getEnv("DB_HOST", "postgres"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName: getEnv("DB_NAME", "subscriptions"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST must be set")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER must be set")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
