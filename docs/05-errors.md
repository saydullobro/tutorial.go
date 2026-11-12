# 05. Ошибки, defer, panic

[← Интерфейсы и дженерики](04-interfaces-generics.md) · [К оглавлению](../README.md) · [Дальше: конкурентность →](06-concurrency.md)

## Ошибка — это обычное значение

В Go нет исключений и `try/except`. Функция, которая может не сработать,
возвращает ошибку **последним значением**, а вызывающий обязан её проверить.

```go
func readConfig(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// добавляем контекст и оборачиваем исходную ошибку через %w
		return nil, fmt.Errorf("чтение конфига %s: %w", path, err)
	}
	return data, nil
}
```

`error` — встроенный интерфейс с одним методом `Error() string`.
`nil` означает «ошибки нет».

Да, `if err != nil` будет в коде очень много. Это осознанное решение языка:
каждое место, где что-то может пойти не так, видно глазами, а не спрятано
в стеке вызовов.

| Python | Go |
|---|---|
| `raise ValueError("...")` | `return errors.New("...")` |
| `try: ... except E:` | `if err != nil { ... }` |
| `except NotFound:` | `if errors.Is(err, ErrNotFound)` |
| `except MyError as e:` | `var e *MyError; if errors.As(err, &e)` |
| `raise NewError(...) from e` | `fmt.Errorf("контекст: %w", err)` |
| `finally:` / `with ...:` | `defer` |
| необработанное исключение, `assert` | `panic` |

## Создание, оборачивание и проверка ошибок

```go
package main

import (
	"errors"
	"fmt"
)

// Сигнальная (sentinel) ошибка: заранее известное значение для сравнения.
var ErrNotFound = errors.New("не найдено")

// Свой тип ошибки — когда вместе с ошибкой нужны данные.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("поле %q: %s", e.Field, e.Msg)
}

var users = map[int]string{1: "Аня"}

func findUser(id int) (string, error) {
	if id <= 0 {
		return "", &ValidationError{Field: "id", Msg: "должен быть положительным"}
	}
	name, ok := users[id]
	if !ok {
		return "", ErrNotFound
	}
	return name, nil
}

func greet(id int) (string, error) {
	name, err := findUser(id)
	if err != nil {
		// %w сохраняет исходную ошибку внутри новой
		return "", fmt.Errorf("приветствие пользователя %d: %w", id, err)
	}
	return "Привет, " + name, nil
}

func main() {
	for _, id := range []int{1, 2, -5} {
		msg, err := greet(id)

		var ve *ValidationError
		switch {
		case err == nil:
			fmt.Println(msg)
		case errors.Is(err, ErrNotFound): // «это та самая ошибка?» — проверяет всю цепочку
			fmt.Println("нет такого пользователя:", err)
		case errors.As(err, &ve): // «есть ли в цепочке ошибка такого типа?»
			fmt.Println("некорректный ввод, поле:", ve.Field)
		default:
			fmt.Println("неизвестная ошибка:", err)
		}
	}
}
```

Вывод:

```text
Привет, Аня
нет такого пользователя: приветствие пользователя 2: не найдено
некорректный ввод, поле: id
```

Важно:

- Сравнивай через `errors.Is`, а не `==`: обёрнутая ошибка простым сравнением
  не найдётся.
- `%w` делает исходную ошибку доступной для `Is`/`As`. `%v` просто вставляет
  её текст — цепочка теряется.
- Несколько ошибок объединяются через `errors.Join(err1, err2)`.

## Правила хорошего тона

1. **Не игнорируй ошибки.** `result, _ := doSomething()` — только когда ошибка
   действительно невозможна или неважна.
2. **Добавляй контекст при возврате наверх**: что именно делали. Из цепочки
   получается понятное сообщение: `запуск сервера: чтение конфига app.yaml: open app.yaml: no such file or directory`.
3. **Обрабатывай ошибку один раз**: либо залогируй, либо верни выше. Не оба сразу.
4. **Текст ошибки** — с маленькой буквы, без точки в конце, без слов «ошибка при».
5. **Короткая область видимости**: `if err := f(); err != nil { ... }`.
6. Первым делом проверяй ошибку, «счастливый путь» идёт без вложенности:

```go
func process(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("чтение: %w", err)
	}

	if len(data) == 0 {
		return errors.New("файл пуст")
	}

	fmt.Println("байт прочитано:", len(data))
	return nil
}
```

## defer

`defer` откладывает вызов до выхода из текущей функции — при любом способе
выхода: `return`, ошибка, паника. Это замена `with` и `finally`.

```go
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() // пишем сразу после успешного открытия

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
```

Три свойства, которые надо помнить:

```go
func demo() {
	// 1. Порядок выполнения — обратный (LIFO)
	defer fmt.Println("первый defer — выполнится последним")
	defer fmt.Println("второй defer — выполнится первым")

	// 2. Аргументы вычисляются в момент defer, а не в момент вызова
	x := 1
	defer fmt.Println("x был:", x) // напечатает 1
	x = 100

	fmt.Println("тело функции, x =", x)
}
```

3\. `defer` срабатывает при выходе из **функции**, а не из блока. `defer` внутри
цикла накопит все вызовы до конца функции — если в цикле открываются файлы,
вынеси тело цикла в отдельную функцию.

## panic и recover

`panic` — аварийная остановка: раскручивает стек, выполняет все `defer`
и завершает программу с трассировкой. Это **не** механизм обработки ошибок.

Паникует сам рантайм: выход за границы слайса, разыменование `nil`,
запись в nil-map, деление целого на ноль.

Когда вызывать `panic` самому:

- ошибка программиста, после которой продолжать бессмысленно (нарушен инвариант);
- сбой при инициализации программы: `regexp.MustCompile`, `template.Must` —
  префикс `Must` по соглашению означает «паникует при ошибке».

Для всего, что может случиться в нормальной работе (нет файла, плохой ввод,
сеть недоступна), возвращай `error`.

`recover` перехватывает панику, но только внутри отложенной функции:

```go
package main

import "fmt"

func safeDiv(a, b int) (res int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// превращаем панику в обычную ошибку через именованный результат
			err = fmt.Errorf("перехвачена паника: %v", r)
		}
	}()
	return a / b, nil
}

func main() {
	fmt.Println(safeDiv(10, 2)) // 5 <nil>
	fmt.Println(safeDiv(1, 0))  // 0 перехвачена паника: runtime error: integer divide by zero
}
```

На практике `recover` пишут редко и на границах: например, HTTP-сервер
перехватывает панику в обработчике, чтобы один сломанный запрос не уронил
весь процесс.

## Практика

1. Функция `parseAge(s string) (int, error)`: оборачивает ошибку `strconv.Atoi`
   и возвращает `*ValidationError`, если возраст вне диапазона 0–150.
2. Объяви `ErrInsufficientFunds` и верни её из метода `Withdraw`; проверь
   через `errors.Is` после оборачивания.
3. Напиши функцию с тремя `defer` и предскажи порядок вывода до запуска.
4. Обёртка `safely(f func()) error`, превращающая панику в ошибку.

[Дальше: конкурентность →](06-concurrency.md)
