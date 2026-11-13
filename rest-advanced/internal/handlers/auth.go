package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/auth"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/models"
)

// Теги binding задают правила валидации (библиотека go-playground/validator).
type registerRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Password string `json:"password" binding:"required,min=8,max=72"` // 72 — предел bcrypt
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func errorJSON(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

// Register — POST /register. Создаёт пользователя и сразу выдаёт токен.
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorJSON(c, http.StatusBadRequest, "некорректные данные: "+err.Error())
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		errorJSON(c, http.StatusBadRequest, "некорректные данные: имя не может быть пустым")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("хэширование пароля", "err", err)
		errorJSON(c, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	user := models.User{
		Email:        normalizeEmail(req.Email),
		Name:         name,
		PasswordHash: hash,
	}

	// Уникальность email гарантирует индекс в БД: проверка «сначала SELECT,
	// потом INSERT» уязвима для гонок между двумя одновременными запросами.
	if err := h.db.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			errorJSON(c, http.StatusConflict, "пользователь с таким email уже существует")
			return
		}
		slog.Error("создание пользователя", "err", err)
		errorJSON(c, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	h.respondWithToken(c, http.StatusCreated, user)
}

// Login — POST /login. Проверяет пароль и выдаёт токен.
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorJSON(c, http.StatusBadRequest, "некорректные данные: "+err.Error())
		return
	}

	var user models.User
	err := h.db.WithContext(c.Request.Context()).
		Where("email = ?", normalizeEmail(req.Email)).
		First(&user).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Одинаковый ответ для «нет пользователя» и «неверный пароль»,
		// чтобы нельзя было перебором выяснить, какие email зарегистрированы.
		errorJSON(c, http.StatusUnauthorized, "неверный email или пароль")
		return
	case err != nil:
		slog.Error("поиск пользователя", "err", err)
		errorJSON(c, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		errorJSON(c, http.StatusUnauthorized, "неверный email или пароль")
		return
	}

	h.respondWithToken(c, http.StatusOK, user)
}

func (h *Handler) respondWithToken(c *gin.Context, status int, user models.User) {
	token, err := h.tokens.Issue(user.ID)
	if err != nil {
		slog.Error("выпуск токена", "err", err)
		errorJSON(c, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}
	c.JSON(status, tokenResponse{Token: token, User: user})
}

func normalizeEmail(s string) string {
	return strings.ToLower(s)
}
