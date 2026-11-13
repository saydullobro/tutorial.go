package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHello(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	newRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("код = %d, ожидался %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, ожидался application/json", ct)
	}

	var got Message
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("разбор ответа: %v", err)
	}
	if got.Message != "Hello, World!" {
		t.Errorf("message = %q, ожидалось %q", got.Message, "Hello, World!")
	}
}

func TestRouting(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{name: "верный запрос", method: http.MethodGet, path: "/hello", wantCode: http.StatusOK},
		{name: "неверный метод", method: http.MethodPost, path: "/hello", wantCode: http.StatusMethodNotAllowed},
		{name: "неизвестный путь", method: http.MethodGet, path: "/nope", wantCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			newRouter().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("%s %s: код = %d, ожидался %d", tt.method, tt.path, rec.Code, tt.wantCode)
			}
		})
	}
}
