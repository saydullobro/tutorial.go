package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken — токен с неверной подписью, просрочен или повреждён.
var ErrInvalidToken = errors.New("недействительный токен")

// TokenManager выпускает и проверяет JWT, подписанные HS256.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time // подменяется в тестах
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// Issue выпускает токен для пользователя; id лежит в стандартном поле sub.
func (m *TokenManager) Issue(userID uint) (string, error) {
	now := m.now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse проверяет подпись и срок действия и возвращает id пользователя.
func (m *TokenManager) Parse(raw string) (uint, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(raw, &claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		// Явно разрешаем только HS256 — защита от подмены алгоритма.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(m.now),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: некорректный sub", ErrInvalidToken)
	}
	return uint(id), nil
}
