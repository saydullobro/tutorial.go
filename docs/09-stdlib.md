# 09. Стандартная библиотека

[← Тестирование](08-testing.md) · [К оглавлению](../README.md) · [Дальше: шпаргалка Python → Go →](10-python-to-go.md)

Стандартная библиотека Go покрывает большую часть задач бэкенда: HTTP-сервер
и клиент, JSON, SQL, криптография, логирование. Многие сервисы пишут вообще
без фреймворков.

| Задача | Python | Go |
|---|---|---|
| форматированный вывод | `print`, f-строки | `fmt` |
| строки | методы `str` | `strings`, `strconv`, `unicode/utf8` |
| регулярные выражения | `re` | `regexp` |
| файлы и ОС | `os`, `pathlib`, `shutil` | `os`, `path/filepath`, `io`, `bufio` |
| JSON | `json` | `encoding/json` |
| дата и время | `datetime`, `time` | `time` |
| HTTP-клиент | `requests`, `httpx` | `net/http` |
| HTTP-сервер | `flask`, `fastapi` | `net/http` |
| аргументы командной строки | `argparse` | `flag` |
| логирование | `logging` | `log/slog` |
| базы данных | DB-API, `sqlalchemy` | `database/sql` + драйвер |
| математика, случайные числа | `math`, `random` | `math`, `math/rand/v2` |
| сортировка, коллекции | `sorted`, `itertools` | `slices`, `maps`, `sort` |
| шаблоны | `jinja2` | `text/template`, `html/template` |
| вшитые в бинарник файлы | `importlib.resources` | `embed` |

Документация по любому пакету: `https://pkg.go.dev/<путь>`, например
<https://pkg.go.dev/net/http>.

## Файлы

```go
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.MkdirTemp("", "demo")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// filepath.Join — аналог os.path.join / оператора / в pathlib
	path := filepath.Join(dir, "notes.txt")

	// Записать файл целиком. 0o644 — права доступа.
	if err := os.WriteFile(path, []byte("первая\nвторая\nтретья\n"), 0o644); err != nil {
		log.Fatal(err)
	}

	// Прочитать файл целиком — аналог Path.read_text()
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("прочитано %d байт\n", len(data))

	// Читать построчно — аналог for line in f
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		fmt.Println(n, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Список файлов в папке — аналог os.listdir
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range entries {
		fmt.Println(e.Name(), e.IsDir())
	}
}
```

Ещё полезное: `os.Create` (создать/перезаписать), `os.OpenFile` (с флагами,
например дозапись), `os.Stat` (информация о файле), `filepath.WalkDir`
(рекурсивный обход, аналог `os.walk`), `errors.Is(err, os.ErrNotExist)`
(проверка «файла нет»).

## JSON

Работа идёт через структуры с тегами — по смыслу это как модели `pydantic`.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Address struct {
	City string `json:"city"`
}

type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email,omitempty"` // пропустить, если пусто
	Password string   `json:"-"`               // никогда не сериализовать
	Tags     []string `json:"tags"`
	Address  *Address `json:"address,omitempty"`
}

func main() {
	u := User{ID: 1, Name: "Аня", Password: "секрет", Tags: []string{"admin"}}

	// структура -> JSON (аналог json.dumps)
	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// JSON -> структура (аналог json.loads + валидация типов)
	raw := `{"id": 2, "name": "Боб", "tags": ["dev"], "address": {"city": "Ташкент"}, "лишнее": true}`

	var parsed User
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil { // нужен указатель!
		log.Fatal(err)
	}
	fmt.Println(parsed.Name, parsed.Tags, parsed.Address.City)

	// Структура заранее неизвестна — разбираем в map
	var anything map[string]any
	if err := json.Unmarshal([]byte(raw), &anything); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T\n", anything["id"]) // float64 — все числа JSON становятся float64
}
```

Правила, о которые спотыкаются:

- Сериализуются **только экспортируемые поля** (с большой буквы). Поле
  `name string` в JSON не попадёт и не прочитается — молча.
- Без тега имя в JSON будет как у поля: `Name`, а не `name`.
- Неизвестные поля во входном JSON игнорируются; отсутствующие получают zero value.
- Отличить «поля не было» от «поле равно 0» можно через указатель: `*int`.
- nil-слайс превращается в `null`, пустой `[]string{}` — в `[]`.
- Для потоков (тело HTTP-запроса, файл) используй `json.NewDecoder(r).Decode(&v)`
  и `json.NewEncoder(w).Encode(v)`.

## Время

```go
now := time.Now()

// Формат задаётся НЕ кодами вроде %Y-%m-%d, а образцовой датой:
// Mon Jan 2 15:04:05 MST 2006  (1 2 3 4 5 6 — месяц, день, час, минута, секунда, год)
fmt.Println(now.Format("2006-01-02 15:04:05"))
fmt.Println(now.Format(time.RFC3339))

t, err := time.Parse("2006-01-02", "2026-03-15")
if err != nil {
	log.Fatal(err)
}
fmt.Println(t.Weekday(), t.AddDate(0, 1, 0).Format("02.01.2006"))

// Длительность — отдельный тип time.Duration (аналог timedelta)
timeout := 1500 * time.Millisecond
fmt.Println(timeout, timeout.Seconds()) // 1.5s 1.5

// Замер времени выполнения
start := time.Now()
time.Sleep(10 * time.Millisecond)
fmt.Println("прошло:", time.Since(start).Round(time.Millisecond))

fmt.Println(now.After(t), now.Sub(t) > 24*time.Hour, now.UTC().Unix())
```

`time.Sleep(5)` — это 5 **наносекунд**. Всегда умножай на единицу: `5 * time.Second`.

## HTTP-сервер

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Server struct {
	mu     sync.Mutex // каждый запрос обрабатывается в своей горутине — состояние надо защищать
	tasks  map[int]Task
	nextID int
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("запись ответа:", err)
	}
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
		http.Error(w, "нужен JSON с полем title", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.nextID++
	task := Task{ID: s.nextID, Title: in.Title}
	s.tasks[task.ID] = task
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id")) // параметр пути {id}
	if err != nil {
		http.Error(w, "некорректный id", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	task, ok := s.tasks[id]
	s.mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func main() {
	s := &Server{tasks: make(map[int]Task)}

	mux := http.NewServeMux()
	// Метод и параметры пути в шаблоне — Go 1.22+
	mux.HandleFunc("GET /tasks", s.list)
	mux.HandleFunc("POST /tasks", s.create)
	mux.HandleFunc("GET /tasks/{id}", s.get)

	log.Println("слушаем http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Проверка:

```bash
curl -X POST localhost:8080/tasks -d '{"title": "выучить Go"}'
```

```bash
curl localhost:8080/tasks/1
```

Обработчик — это функция с сигнатурой `func(http.ResponseWriter, *http.Request)`.
Middleware — функция, которая принимает обработчик и возвращает обёрнутый
(то, для чего в Python были бы декораторы).

Отдельный WSGI/ASGI-сервер вроде `gunicorn` или `uvicorn` не нужен:
`net/http` — это и есть production-сервер.

## HTTP-клиент

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Repo struct {
	FullName string `json:"full_name"`
	Stars    int    `json:"stargazers_count"`
}

func fetchRepo(ctx context.Context, client *http.Client, name string) (Repo, error) {
	var repo Repo

	url := "https://api.github.com/repos/" + name
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return repo, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return repo, fmt.Errorf("запрос %s: %w", url, err)
	}
	defer resp.Body.Close() // тело ответа нужно закрывать ВСЕГДА

	// Код 4xx/5xx — это НЕ ошибка для клиента, проверяй сам (аналог raise_for_status)
	if resp.StatusCode != http.StatusOK {
		return repo, fmt.Errorf("запрос %s: статус %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return repo, fmt.Errorf("разбор ответа: %w", err)
	}
	return repo, nil
}

func main() {
	// У клиента по умолчанию НЕТ таймаута — всегда задавай его явно.
	client := &http.Client{Timeout: 10 * time.Second}

	repo, err := fetchRepo(context.Background(), client, "golang/go")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(repo.FullName, repo.Stars)
}
```

## Логирование, флаги, окружение

```go
package main

import (
	"flag"
	"log/slog"
	"os"
)

func main() {
	// Флаги командной строки — аналог argparse
	port := flag.Int("port", 8080, "порт сервера")
	debug := flag.Bool("debug", false, "подробные логи")
	flag.Parse()

	// Переменные окружения
	dsn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		dsn = "postgres://localhost/dev"
	}

	// Структурированное логирование
	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	slog.Info("запуск", "port", *port, "dsn", dsn)
	slog.Debug("видно только с -debug")

	if *port < 1024 {
		slog.Error("привилегированный порт", "port", *port)
		os.Exit(1) // код возврата; defer при os.Exit НЕ выполняются
	}
}
```

Запуск: `go run . -port 9000 -debug`.

## Что дальше, когда стандартной библиотеки мало

| Задача | Популярный выбор |
|---|---|
| HTTP-роутер с middleware | [`chi`](https://github.com/go-chi/chi), [`gin`](https://github.com/gin-gonic/gin), [`echo`](https://github.com/labstack/echo) |
| PostgreSQL | [`pgx`](https://github.com/jackc/pgx) |
| SQL → типизированный код | [`sqlc`](https://sqlc.dev) |
| ORM | [`gorm`](https://gorm.io), [`ent`](https://entgo.io) |
| миграции | [`goose`](https://github.com/pressly/goose), [`golang-migrate`](https://github.com/golang-migrate/migrate) |
| CLI | [`cobra`](https://github.com/spf13/cobra) |
| конфигурация | [`viper`](https://github.com/spf13/viper), [`envconfig`](https://github.com/kelseyhightower/envconfig) |
| валидация | [`validator`](https://github.com/go-playground/validator) |
| gRPC | [`grpc-go`](https://github.com/grpc/grpc-go) |
| тесты | [`testify`](https://github.com/stretchr/testify), [`testcontainers-go`](https://golang.testcontainers.org) |

В Go-сообществе принято сначала попробовать стандартную библиотеку и
подключать зависимость только при реальной необходимости.

## Практика

1. CLI-утилита `wc`: считает строки, слова и байты в файле, путь — через `flag`.
2. Прочитай JSON-файл со списком пользователей и выведи отсортированных по возрасту.
3. Добавь в сервер задач `DELETE /tasks/{id}` и `PATCH /tasks/{id}`.
4. Напиши middleware, который логирует метод, путь и время обработки запроса.
5. Напиши тесты на обработчики через `httptest`.

[Дальше: шпаргалка Python → Go →](10-python-to-go.md)
