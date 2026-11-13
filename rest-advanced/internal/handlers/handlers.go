// Package handlers содержит HTTP-обработчики API.
package handlers

import (
	"gorm.io/gorm"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/auth"
)

// Handler хранит зависимости обработчиков. Их передают явно, а не через
// глобальные переменные — так код проще тестировать.
type Handler struct {
	db     *gorm.DB
	tokens *auth.TokenManager
}

func New(db *gorm.DB, tokens *auth.TokenManager) *Handler {
	return &Handler{db: db, tokens: tokens}
}
