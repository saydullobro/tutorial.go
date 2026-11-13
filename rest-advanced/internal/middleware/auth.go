// Package middleware содержит промежуточные обработчики Gin.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/auth"
)

// UserIDKey — ключ, под которым id пользователя лежит в gin.Context.
const UserIDKey = "userID"

// Auth требует заголовок "Authorization: Bearer <token>". При успехе кладёт
// id пользователя в контекст, иначе отвечает 401 и прерывает цепочку.
func Auth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, raw, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || raw == "" {
			abort(c, "нужен заголовок Authorization: Bearer <token>")
			return
		}

		id, err := tokens.Parse(raw)
		if err != nil {
			abort(c, "недействительный или просроченный токен")
			return
		}

		c.Set(UserIDKey, id)
		c.Next()
	}
}

func abort(c *gin.Context, msg string) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
