package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	AppEnv   string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		d.Host, d.Port, d.User, d.Password, d.Name,
	)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:   getEnv("PORT", "8080"),
		AppEnv: getEnv("APP_ENV", "development"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "crm_platform"),
		},
	}

	return cfg, nil
}

func getEnv[T any](key string, defaultValue T) T {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	switch any(defaultValue).(type) {
	case string:
		return any(value).(T)
	case int:
		if v, err := strconv.Atoi(value); err == nil {
			return any(v).(T)
		}
	case int64:
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return any(v).(T)
		}
	case int32:
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			return any(int32(v)).(T)
		}
	case float64:
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return any(v).(T)
		}
	case bool:
		if v, err := strconv.ParseBool(value); err == nil {
			return any(v).(T)
		}
	}

	return defaultValue
}
