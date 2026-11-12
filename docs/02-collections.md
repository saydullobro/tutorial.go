# 02. Коллекции: массивы, слайсы, map, строки

[← Основы](01-basics.md) · [К оглавлению](../README.md) · [Дальше: структуры и методы →](03-structs-methods.md)

Соответствие на верхнем уровне:

| Python | Go |
|---|---|
| `list` | слайс `[]T` |
| `tuple` фиксированной длины | массив `[N]T` или структура |
| `dict` | `map[K]V` |
| `set` | `map[T]struct{}` |
| `str` | `string` |
| `bytes` / `bytearray` | `[]byte` |

Ключевое отличие: коллекции **однородные**. `[]int` хранит только `int`.

## Массивы

Длина — часть типа: `[3]int` и `[4]int` — разные типы. В прикладном коде
массивы встречаются редко, почти всегда используют слайсы.

```go
var a [3]int          // [0 0 0]
b := [3]int{1, 2, 3}
c := b                // массив — значение: присваивание КОПИРУЕТ все элементы
c[0] = 100

fmt.Println(a, b, c)  // [0 0 0] [1 2 3] [100 2 3]
```

## Слайсы

Слайс — это «окно» в массив: указатель на данные, длина (`len`) и ёмкость (`cap`).

```go
s := []int{1, 2, 3}
s = append(s, 4, 5)        // append ВОЗВРАЩАЕТ слайс — результат надо присвоить
fmt.Println(s, len(s))     // [1 2 3 4 5] 5

fmt.Println(s[0])          // 1
fmt.Println(s[len(s)-1])   // 5 — отрицательных индексов нет
fmt.Println(s[1:3])        // [2 3]
fmt.Println(s[:2], s[3:])  // [1 2] [4 5]

empty := make([]int, 0, 100) // длина 0, память заранее выделена под 100 элементов
zeros := make([]int, 5)      // [0 0 0 0 0]
fmt.Println(empty, zeros)

var nothing []int            // nil-слайс: len == 0, append работает
nothing = append(nothing, 1)
fmt.Println(nothing)
```

Выход за границы — паника в рантайме (аналог `IndexError`). Шага в срезах
(`s[::2]`) и отрицательных индексов (`s[-1]`) нет.

### ⚠️ Срез не копирует данные

В Python `a[1:3]` создаёт новый список. В Go — новое «окно» в **ту же память**:

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]    // [2 3]
b[0] = 99
fmt.Println(a) // [1 99 3 4 5] — изменился и исходный слайс

b = append(b, 100) // у b есть запас ёмкости, поэтому запись идёт в память a
fmt.Println(a)     // [1 99 3 100 5]

c := slices.Clone(a[1:3]) // настоящая независимая копия
c[0] = 0
fmt.Println(a, c) // [1 99 3 100 5] [0 3]
```

Правило: если срез должен жить своей жизнью — делай `slices.Clone`.

### Операции над слайсами

| Python | Go |
|---|---|
| `xs.append(x)` | `xs = append(xs, x)` |
| `xs.extend(ys)` | `xs = append(xs, ys...)` |
| `len(xs)` | `len(xs)` |
| `x in xs` | `slices.Contains(xs, x)` |
| `xs.index(x)` | `slices.Index(xs, x)` (вернёт `-1`, если нет) |
| `xs.sort()` | `slices.Sort(xs)` |
| `sorted(xs)` | `slices.Sorted(slices.Values(xs))` или клон + `slices.Sort` |
| `xs.sort(key=...)` | `slices.SortFunc(xs, func(a, b T) int {...})` |
| `xs.reverse()` | `slices.Reverse(xs)` |
| `max(xs)`, `min(xs)` | `slices.Max(xs)`, `slices.Min(xs)` |
| `xs.insert(i, x)` | `xs = slices.Insert(xs, i, x)` |
| `del xs[i]` | `xs = slices.Delete(xs, i, i+1)` |
| `xs.copy()` | `slices.Clone(xs)` |
| `xs == ys` | `slices.Equal(xs, ys)` (оператор `==` для слайсов не работает) |
| `sum(xs)` | цикл |

### Вместо list comprehension — обычный цикл

```go
nums := []int{1, 2, 3, 4, 5, 6}

// Python: squares = [n * n for n in nums if n % 2 == 0]
squares := make([]int, 0, len(nums))
for _, n := range nums {
	if n%2 == 0 {
		squares = append(squares, n*n)
	}
}
fmt.Println(squares) // [4 16 36]
```

Да, это длиннее. В Go сознательно выбирают явность вместо краткости.

### Сортировка по ключу

```go
type person struct {
	name string
	age  int
}

people := []person{{"Аня", 30}, {"Боб", 25}, {"Вика", 35}}

// Python: people.sort(key=lambda p: p.age)
slices.SortFunc(people, func(a, b person) int {
	return cmp.Compare(a.age, b.age)
})
fmt.Println(people) // [{Боб 25} {Аня 30} {Вика 35}]
```

## Map

```go
ages := map[string]int{
	"аня": 30,
	"боб": 25, // запятая после последнего элемента обязательна
}

ages["вика"] = 35
fmt.Println(ages["аня"]) // 30

// Отсутствующий ключ НЕ вызывает ошибку, а возвращает zero value.
fmt.Println(ages["никто"]) // 0

// Проверка наличия — идиома «comma ok»
if age, ok := ages["боб"]; ok {
	fmt.Println("нашли:", age)
}

delete(ages, "боб")
fmt.Println(len(ages)) // 2

// Порядок обхода СЛУЧАЙНЫЙ и меняется от запуска к запуску
for name, age := range ages {
	fmt.Println(name, age)
}

// Нужен порядок — сортируем ключи
for _, name := range slices.Sorted(maps.Keys(ages)) {
	fmt.Println(name, ages[name])
}
```

| Python | Go |
|---|---|
| `d[k]` (бросает `KeyError`) | `d[k]` (вернёт zero value) |
| `d.get(k, default)` | `v, ok := d[k]` и проверка `ok` |
| `k in d` | `_, ok := d[k]` |
| `del d[k]` | `delete(d, k)` |
| `d.keys()` / `d.values()` | `maps.Keys(d)` / `maps.Values(d)` (итераторы) |
| `d.items()` | `for k, v := range d` |
| `d.clear()` | `clear(d)` |
| порядок вставки сохраняется | порядок обхода случайный |

### ⚠️ Запись в nil-map — паника

```go
var m map[string]int   // nil-map: читать можно, писать нельзя
fmt.Println(m["a"])    // 0

// m["a"] = 1          // panic: assignment to entry in nil map

m = make(map[string]int) // перед записью map надо создать
m["a"] = 1
fmt.Println(m)
```

### Замены Counter, defaultdict и set

```go
words := []string{"go", "python", "go", "rust", "go"}

// collections.Counter
counts := make(map[string]int)
for _, w := range words {
	counts[w]++ // отсутствующий ключ даёт 0, поэтому это просто работает
}
fmt.Println(counts) // map[go:3 python:1 rust:1]

// defaultdict(list)
byLen := make(map[int][]string)
for _, w := range words {
	byLen[len(w)] = append(byLen[len(w)], w)
}
fmt.Println(byLen)

// set
seen := make(map[string]struct{})
for _, w := range words {
	seen[w] = struct{}{}
}
if _, ok := seen["rust"]; ok {
	fmt.Println("rust есть, уникальных слов:", len(seen))
}
```

Ключом map может быть любой сравнимый тип: числа, строки, массивы, структуры
из сравнимых полей. Слайс или map ключом быть не могут.

## Строки

Строка в Go — неизменяемая последовательность **байт** (обычно UTF-8), а не символов.

```go
s := "привет"

fmt.Println(len(s))                    // 12 — это байты, а не символы!
fmt.Println(utf8.RuneCountInString(s)) // 6  — количество символов

fmt.Println(s[0]) // 208 — первый БАЙТ, а не буква «п»

// range по строке идёт по символам (рунам); i — позиция в байтах
for i, r := range "ёж" {
	fmt.Println(i, string(r)) // 0 ё, затем 2 ж
}

// индексация по символам — через []rune
runes := []rune(s)
fmt.Println(string(runes[0]), string(runes[len(runes)-1])) // п т
```

Для чисто ASCII-строк байты и символы совпадают, поэтому проблема
проявляется именно на кириллице, эмодзи и т. п.

### Пакет strings

| Python | Go |
|---|---|
| `s.upper()`, `s.lower()` | `strings.ToUpper(s)`, `strings.ToLower(s)` |
| `s.strip()` | `strings.TrimSpace(s)` |
| `s.split(",")` | `strings.Split(s, ",")` |
| `s.split()` | `strings.Fields(s)` |
| `",".join(xs)` | `strings.Join(xs, ",")` |
| `"x" in s` | `strings.Contains(s, "x")` |
| `s.startswith(p)`, `s.endswith(p)` | `strings.HasPrefix(s, p)`, `strings.HasSuffix(s, p)` |
| `s.replace(a, b)` | `strings.ReplaceAll(s, a, b)` |
| `s.find(x)` | `strings.Index(s, x)` |
| `s * 3` | `strings.Repeat(s, 3)` |
| `int(s)`, `float(s)` | `strconv.Atoi(s)`, `strconv.ParseFloat(s, 64)` |
| `str(n)` | `strconv.Itoa(n)` или `fmt.Sprint(n)` |
| `f"{x}"` | `fmt.Sprintf("%v", x)` |

Методов у строк нет — всё делается функциями пакета `strings`.

### Сборка строк и «сырые» строки

```go
// Склейка в цикле через + создаёт новую строку каждый раз.
// Для большого количества частей используй strings.Builder (аналог "".join / io.StringIO).
var sb strings.Builder
for i := range 3 {
	fmt.Fprintf(&sb, "строка %d\n", i)
}
fmt.Print(sb.String())

// Обратные кавычки — «сырая» многострочная строка, как r"""..."""
query := `SELECT *
FROM users
WHERE name = "аня"`
fmt.Println(query)
```

Одинарные кавычки — это **не строка**, а один символ типа `rune`: `'a'`, `'я'`.

## Практика

1. Функция `unique(xs []string) []string` — убирает дубликаты, сохраняя порядок.
2. Подсчёт частоты слов в тексте с выводом топ-3 (map + сортировка).
3. Разворот строки, корректно работающий с кириллицей.
4. Воспроизведи ловушку со срезом и `append`, затем исправь через `slices.Clone`.

[Дальше: структуры и методы →](03-structs-methods.md)
