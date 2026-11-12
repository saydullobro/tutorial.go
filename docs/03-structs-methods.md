# 03. Структуры, указатели и методы

[← Коллекции](02-collections.md) · [К оглавлению](../README.md) · [Дальше: интерфейсы и дженерики →](04-interfaces-generics.md)

В Go нет классов и нет наследования. Вместо них: **структуры** (данные),
**методы** (функции, привязанные к типу), **встраивание** (композиция)
и **интерфейсы** (поведение, [следующая глава](04-interfaces-generics.md)).

## Структуры

```go
package main

import "fmt"

type User struct {
	ID    int
	Name  string
	Email string
	admin bool // с маленькой буквы — поле видно только внутри пакета
}

func main() {
	u := User{ID: 1, Name: "Аня", Email: "anya@example.com"}
	u.Name = "Анна"

	var empty User // все поля — zero values

	fmt.Println(u.Name, u.admin) // Анна false
	fmt.Printf("%+v\n", u)       // {ID:1 Name:Анна Email:anya@example.com admin:false}
	fmt.Printf("%+v\n", empty)   // {ID:0 Name: Email: admin:false}
}
```

Это ближайший аналог `@dataclass`. Добавить поле «на лету», как в Python
(`u.whatever = 1`), нельзя — набор полей фиксирован.

## Указатели

Указатель — это адрес значения в памяти. В Python всё и так передаётся по ссылке,
поэтому там ты о них не думал. В Go выбор явный.

```go
x := 10
p := &x         // & — взять адрес; p имеет тип *int
fmt.Println(*p) // 10 — * — получить значение по адресу
*p = 20
fmt.Println(x)  // 20

var q *int             // zero value указателя — nil
fmt.Println(q == nil)  // true
// fmt.Println(*q)     // panic: nil pointer dereference — аналог AttributeError на None
```

Арифметики указателей, как в C, нет. Память освобождает сборщик мусора.

## ⚠️ Главное отличие от Python: присваивание копирует

```go
package main

import "fmt"

type User struct {
	Name string
}

func renameCopy(u User) { // получает КОПИЮ
	u.Name = "изменено"
}

func renamePtr(u *User) { // получает адрес оригинала
	u.Name = "изменено" // u.Name — сокращение для (*u).Name
}

func main() {
	u1 := User{Name: "Аня"}
	u2 := u1 // полная копия структуры
	u2.Name = "Боб"
	fmt.Println(u1.Name, u2.Name) // Аня Боб — в Python обе переменные указывали бы на один объект

	renameCopy(u1)
	fmt.Println(u1.Name) // Аня — оригинал не изменился

	renamePtr(&u1)
	fmt.Println(u1.Name) // изменено
}
```

Что копируется, а что нет:

| Тип | При присваивании и передаче в функцию |
|---|---|
| числа, строки, bool, массивы, структуры | копируется значение целиком |
| указатель | копируется адрес — оба указывают на одно значение |
| слайс, map, канал | копируется небольшой «заголовок», данные остаются общими |

Поэтому изменение элемента map или слайса внутри функции видно снаружи,
а изменение поля структуры, переданной по значению, — нет.

## Методы

Метод — это функция с **получателем** (receiver). Получатель играет роль `self`,
но называют его коротко, по первой букве типа.

```go
package main

import (
	"errors"
	"fmt"
)

type Account struct {
	Owner   string
	balance int
}

// Конструкторов в языке нет. По соглашению пишут функцию NewXxx.
func NewAccount(owner string) (*Account, error) {
	if owner == "" {
		return nil, errors.New("владелец не указан")
	}
	return &Account{Owner: owner}, nil
}

// Получатель-значение: метод работает с копией, изменить оригинал не может.
func (a Account) Balance() int {
	return a.balance
}

// Получатель-указатель: метод может менять оригинал.
func (a *Account) Deposit(amount int) error {
	if amount <= 0 {
		return errors.New("сумма должна быть положительной")
	}
	a.balance += amount
	return nil
}

// String() — аналог __str__: его автоматически использует fmt.
func (a Account) String() string {
	return fmt.Sprintf("%s: %d ₽", a.Owner, a.balance)
}

func main() {
	acc, err := NewAccount("Аня")
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := acc.Deposit(500); err != nil {
		fmt.Println(err)
	}

	fmt.Println(acc.Balance()) // 500
	fmt.Println(acc)           // Аня: 500 ₽
}
```

### Значение или указатель в получателе?

Бери **указатель** (`func (a *Account)`), если:

- метод меняет состояние;
- структура большая и копировать её дорого;
- внутри есть то, что копировать нельзя (`sync.Mutex`).

Бери **значение** (`func (p Point)`), если тип маленький и неизменяемый
(координаты, деньги, время).

Практическое правило: если хотя бы одному методу нужен указатель — делай
указателем получатель у всех методов типа. Не смешивай.

Методы можно объявлять не только у структур, но и у любого своего типа:

```go
type Celsius float64

func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}
```

## Встраивание вместо наследования

```go
package main

import "fmt"

type Animal struct {
	Name string
}

func (a Animal) Describe() string {
	return "Меня зовут " + a.Name
}

type Dog struct {
	Animal // встраивание: поле без имени
	Breed  string
}

func (d Dog) Speak() string {
	return "Гав!"
}

func main() {
	d := Dog{Animal: Animal{Name: "Шарик"}, Breed: "дворняга"}

	fmt.Println(d.Name)       // поле «поднялось» наверх: то же, что d.Animal.Name
	fmt.Println(d.Describe()) // метод тоже
	fmt.Println(d.Speak())
}
```

Это **композиция**, а не наследование:

- `Dog` **не является** `Animal`. Передать `Dog` туда, где ждут `Animal`, нельзя
  (можно передать `d.Animal`).
- `super()` нет. Нужен «родительский» метод — вызывай явно: `d.Animal.Describe()`.
- Виртуальных методов нет: `Animal.Describe` ничего не знает о `Dog`
  и не вызовет его «переопределённые» методы.

Полиморфизм в Go делается через интерфейсы, а не через иерархию типов.

## Теги структур

Теги — метаданные полей, которые читают библиотеки (JSON, БД, валидация):

```go
type Product struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price,omitempty"` // не выводить, если 0
	Notes string  `json:"-"`               // никогда не выводить
}
```

Как это работает с JSON — в [главе про стандартную библиотеку](09-stdlib.md).

## Сравнение и анонимные структуры

```go
type Point struct{ X, Y int }

a, b := Point{1, 2}, Point{1, 2}
fmt.Println(a == b) // true — структуры сравниваются по полям, __eq__ писать не нужно

// Анонимная структура — удобна для разовых данных и табличных тестов
cfg := struct {
	Host string
	Port int
}{"localhost", 8080}
fmt.Println(cfg.Host, cfg.Port)
```

`==` не скомпилируется, если в структуре есть слайс, map или функция.

## Шпаргалка «класс Python → Go»

| Python | Go |
|---|---|
| `class User:` | `type User struct { ... }` |
| `__init__` | функция `NewUser(...)` |
| `self` | получатель: `func (u *User) ...` |
| `@dataclass` | обычная структура |
| `__str__` | метод `String() string` |
| `__eq__` | оператор `==` или свой метод `Equal` |
| `_private` (по соглашению) | имя с маленькой буквы (проверяет компилятор) |
| `@property` | обычный метод: `u.FullName()` |
| `@staticmethod`, `@classmethod` | обычная функция пакета |
| наследование | встраивание + интерфейсы |
| `isinstance(x, T)` | type assertion / type switch |
| `None`-объект | `nil`-указатель |

## Практика

1. Тип `Stack` с методами `Push`, `Pop`, `Len`. Какой получатель нужен и почему?
2. Тип `Money` (копейки в `int64`) с методами `Add` и `String()`.
3. Встрой `Timestamps{CreatedAt, UpdatedAt}` в структуры `User` и `Post`.
4. Передай структуру в функцию по значению и по указателю, убедись в разнице.

[Дальше: интерфейсы и дженерики →](04-interfaces-generics.md)
