package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/middleware"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/models"
)

// Profile — GET /profile. Требует токен; id берётся из него, а не из URL,
// поэтому чужой профиль запросить нельзя.
func (h *Handler) Profile(c *gin.Context) {
	id := c.MustGet(middleware.UserIDKey).(uint)

	var user models.User
	err := h.db.WithContext(c.Request.Context()).First(&user, id).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Токен валиден, но пользователя уже удалили.
		errorJSON(c, http.StatusUnauthorized, "пользователь не найден")
		return
	case err != nil:
		slog.Error("загрузка профиля", "err", err)
		errorJSON(c, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	c.JSON(http.StatusOK, user)
}
