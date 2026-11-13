// Package database открывает подключение GORM и применяет миграции.
package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/models"
)

// Open выбирает драйвер по строке подключения и мигрирует схему.
//
//	postgres://user:pass@host:5432/db  -> PostgreSQL
//	app.db, file:x?mode=memory         -> SQLite (на чистом Go, без CGO)
func Open(dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		dialector = postgres.Open(dsn)
	} else {
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		// Переводит ошибки драйвера в общие, например gorm.ErrDuplicatedKey.
		TranslateError: true,
		Logger:         newLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}

	// AutoMigrate создаёт и дополняет таблицы по структурам. Для production
	// обычно берут отдельный инструмент миграций (goose, golang-migrate).
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("миграция: %w", err)
	}
	return db, nil
}

// newLogger настраивает лог GORM. ParameterizedQueries прячет значения
// параметров: иначе в лог попали бы email и хэши паролей из INSERT.
func newLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true, // «не найдено» — штатная ситуация, не ошибка
		ParameterizedQueries:      true,
	})
}
