# 08. Тестирование

[← Пакеты и модули](07-packages-modules.md) · [К оглавлению](../README.md) · [Дальше: стандартная библиотека →](09-stdlib.md)

Тестирование встроено в язык: пакет `testing` и команда `go test`. Ставить
ничего не нужно.

| pytest | Go |
|---|---|
| `test_*.py` | `*_test.go` рядом с кодом |
| `def test_add():` | `func TestAdd(t *testing.T)` |
| `assert x == y` | `if x != y { t.Errorf(...) }` |
| `@pytest.mark.parametrize` | табличные тесты + `t.Run` |
| фикстуры | обычные функции-помощники, `t.Cleanup` |
| `tmp_path` | `t.TempDir()` |
| `pytest.raises(E)` | проверка `err` через `errors.Is` |
| `pytest -k name` | `go test -run Name` |
| `pytest --cov` | `go test -cover` |
| `pytest-benchmark` | встроенные `BenchmarkXxx` |
| `hypothesis` | встроенный фаззинг `FuzzXxx` |
| `unittest.mock` | интерфейсы + поддельные реализации |

## Код, который будем тестировать

Файл `calc/calc.go`:

```go
package calc

import "errors"

var ErrDivByZero = errors.New("деление на ноль")

// Add возвращает сумму a и b.
func Add(a, b int) int {
	return a + b
}

// Div делит a на b нацело.
func Div(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivByZero
	}
	return a / b, nil
}
```

## Тесты

Файл `calc/calc_test.go`:

```go
package calc

import (
	"errors"
	"fmt"
	"testing"
)

// Простой тест: имя начинается с Test, аргумент — *testing.T.
func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d, ожидалось %d", got, want)
	}
}

// Табличный тест — основная идиома Go, аналог parametrize.
func TestDiv(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr error
	}{
		{name: "нацело", a: 10, b: 2, want: 5},
		{name: "с остатком", a: 7, b: 2, want: 3},
		{name: "отрицательное", a: -7, b: 2, want: -3},
		{name: "на ноль", a: 1, b: 0, wantErr: ErrDivByZero},
	}

	for _, tt := range tests {
		// t.Run создаёт подтест с именем — он виден в отчёте и запускается отдельно
		t.Run(tt.name, func(t *testing.T) {
			got, err := Div(tt.a, tt.b)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ошибка = %v, ожидалась %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Div(%d, %d) = %d, ожидалось %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// Бенчмарк: имя начинается с Benchmark.
func BenchmarkAdd(b *testing.B) {
	for b.Loop() { // Go 1.24+; в старом коде: for i := 0; i < b.N; i++
		Add(2, 3)
	}
}

// Пример: одновременно документация и тест.
// go test сравнивает вывод с комментарием "Output:".
func ExampleAdd() {
	fmt.Println(Add(2, 3))
	// Output: 5
}

// Фаззинг: Go сам подбирает входные данные в поисках падений.
func FuzzDiv(f *testing.F) {
	f.Add(10, 2) // начальные примеры
	f.Add(-1, 0)

	f.Fuzz(func(t *testing.T, a, b int) {
		got, err := Div(a, b)
		if b == 0 {
			if !errors.Is(err, ErrDivByZero) {
				t.Fatalf("при b == 0 ожидалась ErrDivByZero, получили %v", err)
			}
			return
		}
		if got != a/b {
			t.Errorf("Div(%d, %d) = %d", a, b, got)
		}
	})
}
```

Методы `*testing.T`:

| Метод | Действие |
|---|---|
| `t.Errorf(...)` | отметить тест упавшим и **продолжить** |
| `t.Fatalf(...)` | отметить упавшим и **остановить** этот тест |
| `t.Logf(...)` | вывод, который виден при `-v` или падении |
| `t.Run(name, fn)` | подтест |
| `t.Helper()` | пометить функцию как вспомогательную — в отчёте будет строка вызывающего |
| `t.Cleanup(fn)` | выполнить после теста (аналог teardown / `yield` в фикстуре) |
| `t.TempDir()` | временная папка, удаляется автоматически |
| `t.Setenv(k, v)` | переменная окружения на время теста |
| `t.Skip(...)` | пропустить тест |
| `t.Parallel()` | разрешить запуск параллельно с другими тестами |

Встроенного `assert` нет намеренно: сообщение об ошибке ты пишешь сам, и оно
должно объяснять, что пошло не так. Принятый формат:
`Функция(аргументы) = получили, ожидалось то-то`.

Если хочется ассертов — популярна библиотека
[`testify`](https://github.com/stretchr/testify) (`assert.Equal(t, want, got)`),
для сравнения сложных структур — [`go-cmp`](https://github.com/google/go-cmp).
Но сначала освой стандартный подход: он используется в большинстве проектов.

## Запуск

```bash
go test ./...
```

```bash
go test -v ./calc
```

```bash
go test -run 'TestDiv/на_ноль' ./calc
```

```bash
go test -race -cover ./...
```

```bash
go test -bench=. -benchmem ./calc
```

```bash
go test -fuzz=FuzzDiv -fuzztime=10s ./calc
```

| Флаг | Значение |
|---|---|
| `-v` | подробный вывод, имена всех тестов |
| `-run Регулярка` | запустить только подходящие тесты (пробелы в именах подтестов заменяются на `_`) |
| `-race` | детектор гонок данных |
| `-cover` | процент покрытия |
| `-coverprofile=c.out` | сохранить профиль; смотреть: `go tool cover -html=c.out` |
| `-count=1` | не использовать кэш результатов |
| `-bench=.` | запустить бенчмарки |
| `-short` | режим быстрых тестов (проверяется через `testing.Short()`) |
| `-failfast` | остановиться на первом упавшем тесте |

`go test` кэширует успешные результаты: если код не менялся, увидишь `(cached)`.

## Вспомогательные функции вместо фикстур

```go
package store

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTempFile — помощник: создаёт файл во временной папке.
func writeTempFile(t *testing.T, content string) string {
	t.Helper() // при падении отчёт укажет на строку вызова, а не сюда

	path := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("запись временного файла: %v", err)
	}
	return path
}

func TestReadBack(t *testing.T) {
	path := writeTempFile(t, "привет")

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "привет" {
		t.Errorf("прочитали %q", got)
	}
}
```

## Подмена зависимостей: интерфейсы вместо mock.patch

В Go нельзя «пропатчить» функцию в рантайме, как `unittest.mock.patch`.
Вместо этого зависимость принимается через интерфейс, а в тесте передаётся
поддельная реализация.

```go
package notify

import (
	"errors"
	"testing"
)

// Sender — всё, что нужно сервису от внешнего мира.
type Sender interface {
	Send(to, text string) error
}

type Service struct {
	sender Sender
}

func (s *Service) Welcome(email string) error {
	if email == "" {
		return errors.New("пустой email")
	}
	return s.sender.Send(email, "Добро пожаловать!")
}

// --- обычно это лежит в файле *_test.go ---

// fakeSender запоминает вызовы вместо реальной отправки.
type fakeSender struct {
	sent []string
	err  error
}

func (f *fakeSender) Send(to, text string) error {
	f.sent = append(f.sent, to)
	return f.err
}

func TestWelcome(t *testing.T) {
	fake := &fakeSender{}
	svc := &Service{sender: fake}

	if err := svc.Welcome("anya@example.com"); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(fake.sent) != 1 || fake.sent[0] != "anya@example.com" {
		t.Errorf("отправлено %v, ожидалось одно письмо на anya@example.com", fake.sent)
	}
}
```

Это и есть главная практическая причина, зачем нужны интерфейсы.

## Тестирование HTTP

Пакет `net/http/httptest` позволяет тестировать обработчики без запуска сервера:

```go
package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "привет, %s", r.URL.Query().Get("name"))
}

func TestHelloHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello?name=go", nil)
	rec := httptest.NewRecorder()

	helloHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("код = %d, ожидался 200", rec.Code)
	}
	if got := rec.Body.String(); got != "привет, go" {
		t.Errorf("тело = %q", got)
	}
}
```

Для тестов HTTP-клиента есть `httptest.NewServer` — настоящий локальный сервер
на случайном порту.

## Практика

1. Напиши табличные тесты для функций из практики [главы 2](02-collections.md).
2. Добейся покрытия 100 % для пакета `calc` и посмотри HTML-отчёт.
3. Напиши бенчмарк, сравнивающий склейку строк через `+` и через `strings.Builder`.
4. Протестируй сервис с зависимостью через поддельную реализацию интерфейса.

[Дальше: стандартная библиотека →](09-stdlib.md)
