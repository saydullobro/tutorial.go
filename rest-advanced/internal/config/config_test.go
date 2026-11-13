package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("TOKEN_TTL", "")

	t.Run("без секрета", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		if _, err := Load(); err == nil {
			t.Error("ожидалась ошибка")
		}
	})

	t.Run("короткий секрет", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "short")
		if _, err := Load(); err == nil {
			t.Error("ожидалась ошибка")
		}
	})

	t.Run("значения по умолчанию", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Port != "8080" || cfg.DatabaseURL != "app.db" || cfg.TokenTTL != 24*time.Hour {
			t.Errorf("неожиданные значения: %+v", cfg)
		}
	})

	t.Run("свой TTL", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
		t.Setenv("TOKEN_TTL", "30m")
		cfg, err := Load()
		if err != nil || cfg.TokenTTL != 30*time.Minute {
			t.Errorf("TokenTTL = %v, err = %v", cfg.TokenTTL, err)
		}
	})

	t.Run("плохой TTL", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
		t.Setenv("TOKEN_TTL", "сутки")
		if _, err := Load(); err == nil {
			t.Error("ожидалась ошибка")
		}
	})
}
