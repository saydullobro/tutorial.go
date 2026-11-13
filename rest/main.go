// Мини REST API: один эндпоинт, который возвращает "Hello, World!" в JSON.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Message — тело ответа. Теги задают имена полей в JSON.
type Message struct {
	Message string `json:"message"`
}

// helloHandler отвечает на GET /hello.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Message{Message: "Hello, World!"}); err != nil {
		log.Println("запись ответа:", err)
	}
}

// newRouter собирает маршруты. Вынесено из main, чтобы роутер можно было тестировать.
func newRouter() http.Handler {
	mux := http.NewServeMux()
	// Метод в шаблоне (Go 1.22+): на POST /hello сервер сам ответит 405.
	mux.HandleFunc("GET /hello", helloHandler)
	return mux
}

func main() {
	// Порт берём из окружения, по умолчанию 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Println("слушаем http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, newRouter()))
}
