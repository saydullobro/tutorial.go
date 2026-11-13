# rest — мини REST API

Минимальный HTTP-сервис на стандартной библиотеке Go (`net/http`), без
сторонних зависимостей. Один эндпоинт, который возвращает `Hello, World!`.

| Метод | Путь | Ответ |
|---|---|---|
| `GET` | `/hello` | `200`, `{"message":"Hello, World!"}` |

Любой другой метод на `/hello` даёт `405`, неизвестный путь — `404`.

## Запуск

```bash
cd rest
```

```bash
go run .
```

Сервер слушает порт `8080`. Другой порт задаётся переменной окружения:

```bash
PORT=9000 go run .
```

## Проверка

```bash
curl -i localhost:8080/hello
```

```text
HTTP/1.1 200 OK
Content-Type: application/json

{"message":"Hello, World!"}
```

## Тесты

```bash
go test -v ./...
```

## Файлы

| Файл | Что внутри |
|---|---|
| `go.mod` | описание модуля (аналог `pyproject.toml`) |
| `main.go` | обработчик, роутер и запуск сервера |
| `main_test.go` | тесты через `httptest`, без поднятия реального сервера |

## Сравнение с FastAPI

```python
from fastapi import FastAPI

app = FastAPI()

@app.get("/hello")
def hello():
    return {"message": "Hello, World!"}
```

Отличия в Go-версии:

- маршрут регистрируется явно — `mux.HandleFunc("GET /hello", helloHandler)` — вместо декоратора;
- ответ описан структурой `Message` с JSON-тегами, а не словарём;
- отдельный сервер вроде `uvicorn` не нужен: `http.ListenAndServe` и есть сервер.

Подробнее про `net/http` и JSON — в [главе 9](../docs/09-stdlib.md),
про тесты — в [главе 8](../docs/08-testing.md).
