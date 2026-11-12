# 01. Основы синтаксиса

[← Установка](00-setup.md) · [К оглавлению](../README.md) · [Дальше: коллекции →](02-collections.md)

## Переменные

```go
var name string = "Go" // полная форма
var year = 2009        // тип выведен автоматически: int
count := 10            // короткая форма, работает только внутри функций
var ready bool         // значение не задано — будет false
x, y := 1, "два"       // несколько переменных сразу

fmt.Println(name, year, count, ready, x, y)
```

Отличия от Python:

- Тип переменной фиксируется при объявлении и больше не меняется.
  `count = "десять"` — ошибка компиляции.
- `:=` — объявление новой переменной, `=` — присваивание существующей.
  Это разные операции, путаница между ними — частая ошибка новичка.
- Имена пишутся в `camelCase`, а не в `snake_case`: `userName`, `maxRetries`.

## Нулевые значения (zero values)

В Go не бывает неинициализированных переменных и нет универсального `None`.
У каждого типа есть значение по умолчанию:

| Тип | Zero value |
|---|---|
| `int`, `float64` и другие числа | `0` |
| `string` | `""` |
| `bool` | `false` |
| указатель, слайс, map, канал, функция, интерфейс | `nil` |
| структура | структура, где каждое поле — свой zero value |

`nil` — ближайший аналог `None`, но он есть только у перечисленных «ссылочных» типов.
У `int` или `string` значения `nil` быть не может.

## Базовые типы

| Go | Комментарий | Python |
|---|---|---|
| `bool` | `true` / `false` (с маленькой буквы) | `bool` |
| `string` | неизменяемая последовательность байт в UTF-8 | `str` |
| `int` | 64 бита на современных платформах | `int` |
| `int8` … `int64`, `uint8` … `uint64` | целые фиксированного размера | — |
| `float64`, `float32` | числа с плавающей точкой | `float` |
| `byte` | синоним `uint8` | элемент `bytes` |
| `rune` | синоним `int32`, один символ Unicode | символ `str` |

> ⚠️ `int` в Go **переполняется**, а не растёт бесконечно, как в Python.
> Для больших чисел есть пакет `math/big`.

## Преобразование типов — только явное

```go
i := 42
f := float64(i)         // int -> float64
s := strconv.Itoa(i)    // int -> "42"
n, err := strconv.Atoi("123") // "123" -> int, может вернуть ошибку

fmt.Println(f, s, n, err)

var a int = 7
var b float64 = 2
// fmt.Println(a / b)        // ошибка компиляции: разные типы
fmt.Println(float64(a) / b)  // 3.5
```

Арифметика, которая ведёт себя не как в Python:

```go
fmt.Println(7 / 2)   // 3   — деление целых даёт целое (в Python: 3.5)
fmt.Println(-7 / 2)  // -3  — округление к нулю (в Python -7 // 2 == -4)
fmt.Println(-7 % 2)  // -1  — знак как у делимого (в Python: 1)
fmt.Println(7 / 2.0) // 3.5
```

Операторов `**` и `//` нет: степень — `math.Pow`, а `/` для целых уже целочисленное.

## Константы и iota

```go
const Pi = 3.14159

// iota — счётчик внутри блока const, аналог auto() из enum
const (
	StatusNew    = iota // 0
	StatusActive        // 1
	StatusDone          // 2
)

// «Перечисление» в Go — это свой тип плюс набор констант
type Weekday int

const (
	Monday Weekday = iota + 1 // 1
	Tuesday                   // 2
	Wednesday                 // 3
)
```

Отдельного типа `enum` в языке нет — используют этот приём.

## if

```go
score := 87

if score >= 90 {
	fmt.Println("отлично")
} else if score >= 70 {
	fmt.Println("хорошо")
} else {
	fmt.Println("надо подтянуть")
}

// if с инициализацией: n видна только внутри if/else
if n := len("привет"); n > 10 {
	fmt.Println("длинная строка")
} else {
	fmt.Println("байт в строке:", n)
}
```

- Скобки вокруг условия не нужны, фигурные скобки обязательны.
- Условие — строго `bool`. Никакой «истинности» значений: `if items` или `if s`
  не скомпилируются. Пиши `if len(items) > 0`, `if s != ""`, `if p != nil`.
- Тернарного оператора (`a if cond else b`) нет.
- Логические операторы: `&&`, `||`, `!` вместо `and`, `or`, `not`.

## for — единственный цикл

```go
// классический
for i := 0; i < 3; i++ {
	fmt.Println(i)
}

// как while
n := 3
for n > 0 {
	n--
}

// как while True
for {
	break
}

// range по числу (Go 1.22+) — аналог range(3)
for i := range 3 {
	fmt.Println(i)
}

// range по коллекции — аналог enumerate()
for i, v := range []string{"a", "b"} {
	fmt.Println(i, v)
}

// индекс не нужен — ставим _
for _, v := range []string{"a", "b"} {
	fmt.Println(v)
}
```

`break` и `continue` работают как в Python. Конструкции `for ... else` нет.

## switch

```go
// по значению; break писать не нужно — «проваливания» нет
switch day := "сб"; day {
case "сб", "вс":
	fmt.Println("выходной")
default:
	fmt.Println("будни")
}

// без выражения — замена длинной цепочке if / else if
temp := 25
switch {
case temp < 0:
	fmt.Println("мороз")
case temp < 20:
	fmt.Println("прохладно")
default:
	fmt.Println("тепло")
}
```

По роли это ближе всего к `match` из Python 3.10, но без сопоставления по структуре.

## Функции

```go
package main

import (
	"errors"
	"fmt"
)

// типы пишутся после имён; одинаковые типы подряд можно сократить
func add(a, b int) int {
	return a + b
}

// несколько возвращаемых значений
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// идиома Go: последним значением возвращается ошибка
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("деление на ноль")
	}
	return a / b, nil
}

// переменное число аргументов — аналог *args
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// функция возвращает функцию — замыкание
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	fmt.Println(add(2, 3)) // 5

	q, r := divmod(17, 5)
	fmt.Println(q, r) // 3 2

	res, err := divide(1, 0)
	if err != nil {
		fmt.Println("ошибка:", err)
	} else {
		fmt.Println(res)
	}

	fmt.Println(sum(1, 2, 3))        // 6
	fmt.Println(sum([]int{4, 5}...)) // 9 — «распаковка» слайса, как *list

	next := counter()
	fmt.Println(next(), next(), next()) // 1 2 3
}
```

Чего в функциях Go **нет** по сравнению с Python:

| Python | Go |
|---|---|
| значения по умолчанию `def f(x=10)` | нет — передают структуру с настройками или делают несколько функций |
| именованные аргументы `f(x=1)` | нет |
| `**kwargs` | нет |
| декораторы `@decorator` | синтаксиса нет; оборачивают функцию вручную: `f = wrap(f)` |
| перегрузка / `functools.singledispatch` | нет; используют интерфейсы или дженерики |
| `lambda x: x + 1` | анонимная функция `func(x int) int { return x + 1 }` |

## defer — «выполни при выходе из функции»

```go
func process() {
	fmt.Println("открыли ресурс")
	defer fmt.Println("закрыли ресурс") // выполнится в самом конце

	fmt.Println("работаем")
}
```

Это замена `with` и `finally`. Подробно — в [главе про ошибки](05-errors.md).

## Вывод и форматирование

`fmt.Printf` и `fmt.Sprintf` заменяют f-строки:

```go
name, age := "Аня", 30

fmt.Println("Имя:", name, "возраст:", age)    // пробелы между аргументами ставятся сами
fmt.Printf("Имя: %s, возраст: %d\n", name, age) // перевод строки — вручную
msg := fmt.Sprintf("%s (%d)", name, age)      // вернуть строку, как f"{name} ({age})"
fmt.Println(msg)
```

| Глагол | Значение |
|---|---|
| `%v` | значение в формате по умолчанию — подходит для чего угодно |
| `%+v` | структура с именами полей |
| `%#v` | значение в синтаксисе Go (аналог `repr()`) |
| `%T` | тип значения (аналог `type(x)`) |
| `%d` | целое число |
| `%f`, `%.2f` | вещественное число, с точностью |
| `%s` | строка |
| `%q` | строка в кавычках |
| `%t` | bool |
| `%w` | оборачивание ошибки (только в `fmt.Errorf`) |

## Комментарии и документация

```go
// Однострочный комментарий.

/* Многострочный
   комментарий. */

// Add возвращает сумму a и b.
// Комментарий прямо перед объявлением — это и есть документация (аналог docstring).
func Add(a, b int) int {
	return a + b
}
```

## Практика

1. Напиши `fizzbuzz` от 1 до 30 через `for` и `switch` без выражения.
2. Функция `minMax(nums ...int) (min, max int, err error)`, которая возвращает
   ошибку при пустом вводе.
3. Замыкание-генератор чисел Фибоначчи: каждый вызов возвращает следующее число.
4. Проверь на практике, что выводят `-7 / 2` и `-7 % 2`, сравни с Python.

[Дальше: коллекции →](02-collections.md)
