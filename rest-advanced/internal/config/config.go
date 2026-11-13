// Package config читает настройки приложения из переменных окружения.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const minSecretLen = 32

// Config — все настройки сервиса в одном месте.
type Config struct {
	Port        string        // порт HTTP-сервера
	DatabaseURL string        // "postgres://..." или путь к SQLite-файлу
	JWTSecret   string        // ключ подписи токенов
	TokenTTL    time.Duration // срок жизни токена
}

// Load собирает конфигурацию и проверяет обязательные значения.
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "app.db"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TokenTTL:    24 * time.Hour,
	}

	if v := os.Getenv("TOKEN_TTL"); v != "" {
		ttl, err := time.ParseDuration(v)
		if err != nil || ttl <= 0 {
			return Config{}, fmt.Errorf("TOKEN_TTL: нужна положительная длительность вроде 30m или 24h, получено %q", v)
		}
		cfg.TokenTTL = ttl
	}

	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET не задан")
	}
	if len(cfg.JWTSecret) < minSecretLen {
		return Config{}, fmt.Errorf("JWT_SECRET слишком короткий: нужно минимум %d символов", minSecretLen)
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
