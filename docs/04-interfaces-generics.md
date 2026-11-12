# 04. Интерфейсы и дженерики

[← Структуры и методы](03-structs-methods.md) · [К оглавлению](../README.md) · [Дальше: ошибки →](05-errors.md)

## Интерфейсы

Интерфейс — это набор методов. Тип реализует интерфейс **автоматически**,
если у него есть все эти методы. Слова `implements` и наследования от базового
класса нет.

Это утиная типизация из Python, только проверяемая компилятором.
Ближайший аналог — `typing.Protocol`.

```go
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
}

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

// Функция принимает что угодно, у чего есть метод Area().
func totalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func main() {
	shapes := []Shape{Circle{R: 1}, Rect{W: 2, H: 3}}
	fmt.Printf("%.2f\n", totalArea(shapes)) // 9.14
}
```

`Circle` и `Rect` ничего не знают про `Shape`. Интерфейс можно объявить
позже и в другом пакете — и существующие типы начнут ему соответствовать.

### Принципы

- **Интерфейсы маленькие.** Один-два метода — норма. Имя часто заканчивается
  на `-er`: `Reader`, `Writer`, `Stringer`.
- **Интерфейс объявляет потребитель**, а не автор типа. Нужен в функции только
  метод `Save` — объяви рядом интерфейс с одним `Save`.
- **«Принимай интерфейсы, возвращай конкретные типы».**
- Не создавай интерфейс «на будущее». Он нужен, когда есть хотя бы две
  реализации или нужно подменять зависимость в тестах.

### Интерфейсы стандартной библиотеки, которые встречаются везде

| Интерфейс | Метод | Аналог в Python |
|---|---|---|
| `error` | `Error() string` | `Exception` |
| `fmt.Stringer` | `String() string` | `__str__` |
| `io.Reader` | `Read(p []byte) (n int, err error)` | file-like объект с `.read()` |
| `io.Writer` | `Write(p []byte) (n int, err error)` | file-like объект с `.write()` |
| `io.Closer` | `Close() error` | `.close()` / `__exit__` |

Файл, сетевое соединение, HTTP-тело, буфер в памяти, gzip-поток — всё это
`io.Reader`/`io.Writer`. Поэтому функция, принимающая `io.Reader`, работает с
любым источником данных.

### any, type assertion и type switch

`any` (он же `interface{}`) — пустой интерфейс, которому соответствует любой тип.
Аналог `object` / `Any`. Чтобы достать конкретное значение, тип надо проверить:

```go
package main

import "fmt"

func describe(v any) string {
	// type switch — аналог цепочки isinstance()
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("целое %d", x)
	case string:
		return fmt.Sprintf("строка длиной %d байт", len(x))
	case []int:
		return fmt.Sprintf("слайс из %d элементов", len(x))
	case fmt.Stringer:
		return "умеет String(): " + x.String()
	default:
		return fmt.Sprintf("что-то другое: %T", x)
	}
}

func main() {
	fmt.Println(describe(42))
	fmt.Println(describe("go"))
	fmt.Println(describe(3.14))

	var v any = "привет"

	// type assertion с проверкой
	if s, ok := v.(string); ok {
		fmt.Println("это строка:", s)
	}

	// без проверки — паника, если тип не тот
	// n := v.(int) // panic: interface conversion
}
```

Используй `any` только там, где тип действительно неизвестен (например,
произвольный JSON). В остальных случаях лучше конкретный интерфейс или дженерик.

### Проверка реализации на этапе компиляции

```go
// Если Circle перестанет реализовывать Shape, эта строка не скомпилируется.
var _ Shape = Circle{}
```

### ⚠️ Методы с получателем-указателем

Если методы объявлены на `*T`, интерфейс реализует только **указатель**:

```go
type Saver interface{ Save() }

type File struct{ name string }

func (f *File) Save() {}

var _ Saver = &File{} // работает
// var _ Saver = File{} // не скомпилируется: метод Save объявлен у *File
```

### ⚠️ nil внутри интерфейса

Интерфейс хранит пару «тип + значение». Он равен `nil`, только когда пусты оба.

```go
package main

import "fmt"

type MyErr struct{}

func (*MyErr) Error() string { return "сломалось" }

func bad() error {
	var e *MyErr // nil-указатель
	return e     // интерфейс уже НЕ nil: внутри (тип *MyErr, значение nil)
}

func good() error {
	return nil
}

func main() {
	fmt.Println(bad() != nil)  // true — «ошибка есть», хотя её нет
	fmt.Println(good() != nil) // false
}
```

Правило: из функции, возвращающей `error`, в случае успеха возвращай
буквально `nil`, а не переменную с типизированным nil-указателем.

## Дженерики

Дженерики позволяют писать функции и типы, работающие с разными типами без
потери типобезопасности. Аналог `TypeVar` / `Generic[T]` из `typing`.

```go
package main

import (
	"cmp"
	"fmt"
)

// T и U — параметры типа; any — ограничение «любой тип»
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

// cmp.Ordered — все типы, которые можно сравнивать через < и >
func Largest[T cmp.Ordered](xs []T) (T, bool) {
	var zero T
	if len(xs) == 0 {
		return zero, false
	}
	best := xs[0]
	for _, x := range xs[1:] {
		if x > best {
			best = x
		}
	}
	return best, true
}

// Собственное ограничение. ~int означает «int и любые типы на его основе».
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

// Обобщённый тип
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return last, true
}

func main() {
	nums := []int{1, 2, 3, 4}

	// параметры типа обычно выводятся сами
	strs := Map(nums, func(n int) string { return fmt.Sprint(n * n) })
	fmt.Println(strs) // [1 4 9 16]

	fmt.Println(Filter(nums, func(n int) bool { return n%2 == 0 })) // [2 4]
	fmt.Println(Largest([]string{"go", "python", "rust"}))          // rust true
	fmt.Println(Sum([]float64{1.5, 2.5}))                           // 4

	var st Stack[string]
	st.Push("a")
	st.Push("b")
	v, ok := st.Pop()
	fmt.Println(v, ok) // b true
}
```

Встроенные ограничения: `any`, `comparable` (типы с `==`, годятся в ключи map),
`cmp.Ordered` (типы с `<`, `>`).

Когда использовать:

- ✅ контейнеры и утилиты над коллекциями (`Map`, `Filter`, `Stack[T]`, `Cache[K, V]`);
- ❌ если достаточно обычного интерфейса — бери интерфейс;
- ❌ не пиши дженерик заранее: сначала конкретный код, обобщай при повторении.

Многое уже есть в стандартной библиотеке: пакеты `slices`, `maps`, функции
`min`, `max`.

## Итераторы (Go 1.23+)

Аналог генераторов Python. Итератор — функция, которая «отдаёт» значения через
`yield`, и по ней можно пройти обычным `for ... range`.

```go
package main

import (
	"fmt"
	"iter"
)

// Countdown — аналог генератора Python:
//
//	def countdown(n):
//	    while n > 0:
//	        yield n
//	        n -= 1
func Countdown(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := n; i > 0; i-- {
			if !yield(i) { // false — вызывающий сделал break
				return
			}
		}
	}
}

func main() {
	for v := range Countdown(3) {
		fmt.Println(v) // 3, 2, 1
	}
}
```

На итераторах построены `maps.Keys`, `slices.Values`, `slices.Sorted`,
`slices.Collect` (собрать итератор в слайс — как `list(gen)`).

## Практика

1. Интерфейс `Notifier` с методом `Notify(msg string) error` и две реализации:
   `EmailNotifier` и `ConsoleNotifier`. Функция принимает `[]Notifier`.
2. Реализуй `fmt.Stringer` для своего типа и проверь, что `fmt.Println` его использует.
3. Обобщённая функция `Keys[K comparable, V any](m map[K]V) []K`.
4. Обобщённый тип `Set[T comparable]` с методами `Add`, `Has`, `Len`.
5. Итератор `Fib() iter.Seq[int]` — бесконечные числа Фибоначчи; возьми первые 10 через `break`.

[Дальше: ошибки →](05-errors.md)
