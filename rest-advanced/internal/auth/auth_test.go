package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "unit-test-secret-unit-test-secret-1"

func TestPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(hash, "correct-horse") {
		t.Error("хэш содержит пароль")
	}
	if !CheckPassword(hash, "correct-horse") {
		t.Error("верный пароль не прошёл проверку")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("неверный пароль прошёл проверку")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	m := NewTokenManager(secret, time.Hour)
	raw, err := m.Issue(42)
	if err != nil {
		t.Fatal(err)
	}
	id, err := m.Parse(raw)
	if err != nil || id != 42 {
		t.Fatalf("Parse = %d, %v; ожидалось 42, nil", id, err)
	}
}

func TestTokenExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewTokenManager(secret, time.Hour)
	m.now = func() time.Time { return now }

	raw, _ := m.Issue(1)

	m.now = func() time.Time { return now.Add(59 * time.Minute) }
	if _, err := m.Parse(raw); err != nil {
		t.Errorf("токен должен быть валиден через 59 минут: %v", err)
	}

	m.now = func() time.Time { return now.Add(61 * time.Minute) }
	if _, err := m.Parse(raw); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("через 61 минуту ожидалась ErrInvalidToken, получили %v", err)
	}
}

func TestTokenRejectsOtherAlgorithms(t *testing.T) {
	m := NewTokenManager(secret, time.Hour)

	// alg=none — классическая атака: токен без подписи.
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := m.Parse(none); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("токен alg=none принят: %v", err)
	}

	// HS512 с верным ключом — другой алгоритм, тоже отклоняем.
	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(secret))
	if _, err := m.Parse(hs512); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("токен HS512 принят: %v", err)
	}
}

func TestTokenRequiresExpiration(t *testing.T) {
	m := NewTokenManager(secret, time.Hour)
	raw, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "1"}).SignedString([]byte(secret))
	if _, err := m.Parse(raw); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("токен без exp принят: %v", err)
	}
}
