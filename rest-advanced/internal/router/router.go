// Package router собирает маршруты приложения.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/auth"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/handlers"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/middleware"
)

// New создаёт Gin-движок с маршрутами. Recovery превращает панику в обработчике
// в ответ 500, а Logger пишет строку на каждый запрос.
func New(db *gorm.DB, tokens *auth.TokenManager) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	h := handlers.New(db, tokens)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Группа с middleware: всё внутри доступно только с токеном.
	protected := r.Group("/", middleware.Auth(tokens))
	protected.GET("/profile", h.Profile)

	return r
}
