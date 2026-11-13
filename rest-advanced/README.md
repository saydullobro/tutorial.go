# rest-advanced — REST API на Gin + GORM

Учебный, но близкий к реальному сервис: регистрация, вход и профиль
пользователя с JWT-авторизацией.

- **[Gin](https://github.com/gin-gonic/gin)** — самый популярный веб-фреймворк Go: роутер, middleware, валидация.
- **[GORM](https://gorm.io)** — самый популярный ORM Go: модели, миграции, запросы.
- **SQLite** по умолчанию (файл, ничего ставить не нужно) или **PostgreSQL** — выбор по `DATABASE_URL`.
- **JWT** (HS256) для токенов, **bcrypt** для паролей.

> Нужен Go 1.26: актуальные версии зависимостей требуют его. Если у тебя
> старше, команда `go` сама скачает нужный toolchain при первом запуске.

## API

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| `POST` | `/register` | открыт | создать пользователя, сразу получить токен |
| `POST` | `/login` | открыт | войти по email и паролю, получить токен |
| `GET` | `/profile` | **с токеном** | данные текущего пользователя |
| `GET` | `/health` | открыт | проверка, что сервис жив |

Коды ответов: `201` создан, `200` успех, `400` неверные данные, `401` нет или
неверный токен / пароль, `409` email уже занят, `500` внутренняя ошибка.
Ошибки всегда имеют вид `{"error": "..."}`.

## Запуск

Секрет для подписи токенов обязателен, не короче 32 символов:

```bash
cd rest-advanced
```

```bash
export JWT_SECRET=$(openssl rand -hex 32)
```

```bash
go run ./cmd/server
```

Сервер слушает `:8080`, база создаётся в файле `app.db` (он в `.gitignore`).

### Настройки (переменные окружения)

| Переменная | По умолчанию | Описание |
|---|---|---|
| `JWT_SECRET` | — (**обязательна**) | ключ подписи токенов, минимум 32 символа |
| `PORT` | `8080` | порт сервера |
| `DATABASE_URL` | `app.db` | путь к SQLite-файлу или `postgres://user:pass@host:5432/db` |
| `TOKEN_TTL` | `24h` | срок жизни токена, например `30m` |
| `GIN_MODE` | `debug` | в production ставь `release` |

## Пример сессии

```bash
curl -s -X POST localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"anya@example.com","name":"Аня","password":"correct-horse"}'
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {"id": 1, "email": "anya@example.com", "name": "Аня", "created_at": "...", "updated_at": "..."}
}
```

```bash
curl -s -X POST localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"anya@example.com","password":"correct-horse"}'
```

Токен передаётся в заголовке `Authorization`:

```bash
curl -s localhost:8080/profile -H "Authorization: Bearer $TOKEN"
```

```json
{"id": 1, "email": "anya@example.com", "name": "Аня", "created_at": "...", "updated_at": "..."}
```

Правила регистрации: `email` валидный (до 255 символов), `name` не пустое
(до 100), `password` от 8 до 72 символов (72 — предел bcrypt).

## Структура проекта

```text
rest-advanced/
├── cmd/server/main.go          # точка входа: сборка зависимостей, graceful shutdown
└── internal/
    ├── config/                 # чтение и проверка переменных окружения
    ├── database/               # подключение GORM (SQLite / PostgreSQL), AutoMigrate
    ├── models/                 # модель User — структура, отображаемая на таблицу
    ├── auth/                   # bcrypt-пароли и выпуск/проверка JWT
    ├── middleware/             # Auth: проверка токена перед защищёнными маршрутами
    ├── handlers/               # Register, Login, Profile
    └── router/                 # маршруты + интеграционные тесты
```

Зависимости передаются явно (`handlers.New(db, tokens)`), без глобальных
переменных — поэтому тесты поднимают приложение на отдельной базе в памяти.

## Тесты

```bash
go test -race ./...
```

Покрыто: регистрация и валидация, дубликат email (в том числе в другом
регистре), вход, все варианты неверного токена (нет, мусор, чужая подпись,
просрочен, `alg=none`, другой алгоритм, пользователь удалён), конфигурация.

## Решения, о которых стоит знать

- **Пароли** хранятся только как bcrypt-хэш; поле `PasswordHash` помечено
  `json:"-"` и никогда не попадает в ответ.
- **Вход не выдаёт, какой email зарегистрирован**: «нет пользователя» и
  «неверный пароль» дают одинаковый ответ.
- **Email** приводится к нижнему регистру, уникальность гарантирует индекс БД
  (проверка «SELECT, потом INSERT» ломается при двух одновременных запросах).
- **JWT**: принимается только HS256, срок действия (`exp`) обязателен,
  `alg=none` отклоняется.
- **Профиль** определяется по токену, а не по id в URL — чужой запрос невозможен.
- **Логи GORM** не содержат значений параметров: иначе email и хэши паролей
  попадали бы в лог.
- **`AutoMigrate`** удобен для обучения; в production миграции обычно ведут
  отдельным инструментом ([goose](https://github.com/pressly/goose),
  [golang-migrate](https://github.com/golang-migrate/migrate)).

Чего здесь нет (хороший следующий шаг): refresh-токены, ограничение частоты
запросов на `/login`, подтверждение email, смена пароля, OpenAPI-описание.

> PostgreSQL-ветка подключается через `DATABASE_URL=postgres://...`, но в
> этом проекте проверялась только на SQLite: сервера PostgreSQL при разработке
> не было.

## Для тех, кто пришёл из Python

| FastAPI / SQLAlchemy | Здесь |
|---|---|
| `@app.post("/register")` | `r.POST("/register", h.Register)` |
| Pydantic-модель запроса | структура с тегами `binding:"required,email"` |
| `Depends(get_current_user)` | middleware `Auth` + `c.Set` / `c.MustGet` |
| `HTTPException(409)` | `c.JSON(http.StatusConflict, ...)` |
| модель SQLAlchemy | структура с тегами `gorm:"..."` |
| `Base.metadata.create_all` / Alembic | `db.AutoMigrate` / goose |
| `session.query(User).filter_by(...)` | `db.Where("email = ?", e).First(&user)` |
| `passlib`, `python-jose` | `golang.org/x/crypto/bcrypt`, `golang-jwt/jwt` |
| `uvicorn main:app` | `go run ./cmd/server` (сервер встроен) |

Основы из соответствующих глав документации: [структуры](../docs/03-structs-methods.md),
[ошибки](../docs/05-errors.md), [пакеты](../docs/07-packages-modules.md),
[тесты](../docs/08-testing.md), [стандартная библиотека](../docs/09-stdlib.md).
