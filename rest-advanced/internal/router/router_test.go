package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/auth"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/database"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/router"
)

const testSecret = "test-secret-test-secret-test-secret-123"

// newServer поднимает приложение на отдельной SQLite в памяти для каждого теста.
func newServer(t *testing.T) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.Open("file:" + t.Name() + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	return router.New(db, auth.NewTokenManager(testSecret, time.Hour))
}

// call выполняет запрос и возвращает код и разобранное JSON-тело.
func call(t *testing.T, h http.Handler, method, path, token string, body any) (int, map[string]any) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("тело ответа не JSON: %q", rec.Body.String())
	}
	return rec.Code, out
}

func register(t *testing.T, h http.Handler, email string) string {
	t.Helper()
	code, out := call(t, h, "POST", "/register", "", map[string]string{
		"email": email, "name": "Аня", "password": "correct-horse",
	})
	if code != http.StatusCreated {
		t.Fatalf("register: код %d, тело %v", code, out)
	}
	return out["token"].(string)
}

func TestRegister(t *testing.T) {
	h := newServer(t)

	code, out := call(t, h, "POST", "/register", "", map[string]string{
		"email": "Anya@Example.COM", "name": "Аня", "password": "correct-horse",
	})
	if code != http.StatusCreated {
		t.Fatalf("код %d, тело %v", code, out)
	}
	if out["token"] == "" {
		t.Error("токен не выдан")
	}

	user := out["user"].(map[string]any)
	if user["email"] != "anya@example.com" {
		t.Errorf("email = %v, ожидался нормализованный anya@example.com", user["email"])
	}
	for _, leaked := range []string{"password", "password_hash", "PasswordHash"} {
		if _, ok := user[leaked]; ok {
			t.Errorf("в ответе утекло поле %q", leaked)
		}
	}
}

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{"пустое тело", map[string]string{}},
		{"плохой email", map[string]string{"email": "not-an-email", "name": "А", "password": "correct-horse"}},
		{"короткий пароль", map[string]string{"email": "a@b.co", "name": "А", "password": "short"}},
		{"пароль длиннее 72", map[string]string{"email": "a@b.co", "name": "А", "password": strings.Repeat("x", 73)}},
		{"нет имени", map[string]string{"email": "a@b.co", "password": "correct-horse"}},
		{"имя из пробелов", map[string]string{"email": "a@b.co", "name": "   ", "password": "correct-horse"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newServer(t)
			if code, out := call(t, h, "POST", "/register", "", tt.body); code != http.StatusBadRequest {
				t.Errorf("код %d, ожидался 400; тело %v", code, out)
			}
		})
	}
}

func TestRegisterDuplicate(t *testing.T) {
	h := newServer(t)
	register(t, h, "anya@example.com")

	// Тот же email в другом регистре — это тот же пользователь.
	code, _ := call(t, h, "POST", "/register", "", map[string]string{
		"email": "ANYA@example.com", "name": "Двойник", "password": "correct-horse",
	})
	if code != http.StatusConflict {
		t.Errorf("код %d, ожидался 409", code)
	}
}

func TestLogin(t *testing.T) {
	h := newServer(t)
	register(t, h, "anya@example.com")

	tests := []struct {
		name     string
		email    string
		password string
		want     int
	}{
		{"верные данные", "anya@example.com", "correct-horse", http.StatusOK},
		{"email в другом регистре", "Anya@Example.com", "correct-horse", http.StatusOK},
		{"неверный пароль", "anya@example.com", "wrong-password", http.StatusUnauthorized},
		{"нет такого пользователя", "ghost@example.com", "correct-horse", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, out := call(t, h, "POST", "/login", "", map[string]string{
				"email": tt.email, "password": tt.password,
			})
			if code != tt.want {
				t.Fatalf("код %d, ожидался %d; тело %v", code, tt.want, out)
			}
			if tt.want == http.StatusOK && out["token"] == "" {
				t.Error("токен не выдан")
			}
		})
	}

	// Сообщения для «нет пользователя» и «неверный пароль» не должны различаться.
	_, a := call(t, h, "POST", "/login", "", map[string]string{"email": "anya@example.com", "password": "wrong-password"})
	_, b := call(t, h, "POST", "/login", "", map[string]string{"email": "ghost@example.com", "password": "wrong-password"})
	if a["error"] != b["error"] {
		t.Errorf("ответы различаются и раскрывают наличие email: %q и %q", a["error"], b["error"])
	}
}

func TestProfile(t *testing.T) {
	h := newServer(t)
	token := register(t, h, "anya@example.com")

	t.Run("с токеном", func(t *testing.T) {
		code, out := call(t, h, "GET", "/profile", token, nil)
		if code != http.StatusOK {
			t.Fatalf("код %d, тело %v", code, out)
		}
		if out["email"] != "anya@example.com" || out["name"] != "Аня" {
			t.Errorf("неожиданный профиль: %v", out)
		}
		if _, ok := out["password_hash"]; ok {
			t.Error("в профиле утёк хэш пароля")
		}
	})

	t.Run("после логина", func(t *testing.T) {
		_, out := call(t, h, "POST", "/login", "", map[string]string{"email": "anya@example.com", "password": "correct-horse"})
		if code, _ := call(t, h, "GET", "/profile", out["token"].(string), nil); code != http.StatusOK {
			t.Errorf("код %d, ожидался 200", code)
		}
	})

	t.Run("без токена", func(t *testing.T) {
		if code, _ := call(t, h, "GET", "/profile", "", nil); code != http.StatusUnauthorized {
			t.Errorf("код %d, ожидался 401", code)
		}
	})

	t.Run("мусорный токен", func(t *testing.T) {
		if code, _ := call(t, h, "GET", "/profile", "not.a.jwt", nil); code != http.StatusUnauthorized {
			t.Errorf("код %d, ожидался 401", code)
		}
	})

	t.Run("токен с чужой подписью", func(t *testing.T) {
		forged, _ := auth.NewTokenManager("another-secret-another-secret-12345", time.Hour).Issue(1)
		if code, _ := call(t, h, "GET", "/profile", forged, nil); code != http.StatusUnauthorized {
			t.Errorf("код %d, ожидался 401", code)
		}
	})

	t.Run("просроченный токен", func(t *testing.T) {
		expired, _ := auth.NewTokenManager(testSecret, -time.Minute).Issue(1)
		if code, _ := call(t, h, "GET", "/profile", expired, nil); code != http.StatusUnauthorized {
			t.Errorf("код %d, ожидался 401", code)
		}
	})

	t.Run("токен несуществующего пользователя", func(t *testing.T) {
		orphan, _ := auth.NewTokenManager(testSecret, time.Hour).Issue(9999)
		if code, _ := call(t, h, "GET", "/profile", orphan, nil); code != http.StatusUnauthorized {
			t.Errorf("код %d, ожидался 401", code)
		}
	})
}

func TestHealth(t *testing.T) {
	if code, out := call(t, newServer(t), "GET", "/health", "", nil); code != http.StatusOK || out["status"] != "ok" {
		t.Errorf("код %d, тело %v", code, out)
	}
}
