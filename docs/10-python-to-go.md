# 10. Шпаргалка Python → Go

[← Стандартная библиотека](09-stdlib.md) · [К оглавлению](../README.md)

Справочник для быстрого поиска «как это пишется в Go». Подробности — в
соответствующих главах.

## Синтаксис

| Python | Go |
|---|---|
| `x = 10` | `x := 10` (объявление), `x = 10` (присваивание) |
| `x: int = 10` | `var x int = 10` |
| `PI = 3.14` | `const Pi = 3.14` |
| `None` | `nil` (только для указателей, слайсов, map, каналов, функций, интерфейсов) |
| `True`, `False` | `true`, `false` |
| `and`, `or`, `not` | `&&`, `\|\|`, `!` |
| `x ** 2` | `x * x` или `math.Pow(x, 2)` |
| `a // b` | `a / b` для целых (но округление к нулю!) |
| `a if cond else b` | нет; обычный `if` / `else` |
| `x += 1` | `x++` или `x += 1` |
| `# комментарий` | `// комментарий` |
| `"""docstring"""` | комментарий `// Имя ...` перед объявлением |
| `pass` | пустой блок `{}` |
| `print(x)` | `fmt.Println(x)` |
| `f"{name}: {n:.2f}"` | `fmt.Sprintf("%s: %.2f", name, n)` |
| `type(x)` | `fmt.Sprintf("%T", x)` |
| `isinstance(x, T)` | `v, ok := x.(T)` или type switch |
| `if __name__ == "__main__":` | `func main()` в `package main` |

## Управление

| Python | Go |
|---|---|
| `if x:` | `if x != 0`, `if s != ""`, `if len(xs) > 0`, `if p != nil` |
| `elif` | `else if` |
| `for i in range(n):` | `for i := range n` |
| `for i in range(a, b):` | `for i := a; i < b; i++` |
| `for x in xs:` | `for _, x := range xs` |
| `for i, x in enumerate(xs):` | `for i, x := range xs` |
| `for k, v in d.items():` | `for k, v := range d` |
| `while cond:` | `for cond { }` |
| `while True:` | `for { }` |
| `match x:` | `switch x { }` |
| `with open(p) as f:` | `f, err := os.Open(p)` + `defer f.Close()` |
| `try / except` | `if err != nil` |
| `raise E("...")` | `return errors.New("...")` |
| `assert cond` | `if !cond { panic("...") }` (в тестах — `t.Fatalf`) |

## Функции и типы

| Python | Go |
|---|---|
| `def f(a: int, b: int) -> int:` | `func f(a, b int) int` |
| `return a, b` | `return a, b` с сигнатурой `(int, int)` |
| `def f(*args):` | `func f(args ...int)` |
| `f(*xs)` | `f(xs...)` |
| `def f(x=10):` | нет; структура с параметрами |
| `lambda x: x + 1` | `func(x int) int { return x + 1 }` |
| `@decorator` | `f = wrap(f)` |
| генератор с `yield` | итератор `iter.Seq[T]` или канал |
| `class A:` | `type A struct { }` |
| `def method(self):` | `func (a *A) Method()` |
| `__init__` | `func NewA(...) *A` |
| `__str__` | `String() string` |
| `class B(A):` | встраивание: `type B struct { A }` |
| `Protocol` / `ABC` | `interface` |
| `Generic[T]` | `type Box[T any] struct { }` |
| `Optional[int]` | `*int` или пара `(int, bool)` |
| `Enum` | свой тип + `const ( ... iota )` |
| `_private` | имя с маленькой буквы |

## Коллекции

| Python | Go |
|---|---|
| `xs = [1, 2, 3]` | `xs := []int{1, 2, 3}` |
| `xs = []` | `var xs []int` |
| `xs.append(x)` | `xs = append(xs, x)` |
| `xs[-1]` | `xs[len(xs)-1]` |
| `xs[1:3]` (копия) | `xs[1:3]` (**та же память**), копия — `slices.Clone` |
| `x in xs` | `slices.Contains(xs, x)` |
| `sorted(xs)`, `xs.sort()` | `slices.Sort(xs)` |
| `[f(x) for x in xs if c(x)]` | цикл `for` + `append` |
| `d = {"a": 1}` | `d := map[string]int{"a": 1}` |
| `d.get(k)` | `v, ok := d[k]` |
| `k in d` | `_, ok := d[k]` |
| `del d[k]` | `delete(d, k)` |
| `set()` | `map[T]struct{}` |
| `Counter` | `map[T]int` и `m[k]++` |
| `defaultdict(list)` | `map[K][]V` и `m[k] = append(m[k], v)` |
| `(1, "a")` | структура или несколько возвращаемых значений |
| `len(x)` | `len(x)` |

## Конкурентность

| Python | Go |
|---|---|
| `threading.Thread(target=f).start()` | `go f()` |
| `async def` / `await` | обычная функция; ожидание — через канал или `WaitGroup` |
| `asyncio.gather(...)` | `sync.WaitGroup` / `errgroup.Group` |
| `asyncio.TaskGroup` | `errgroup.Group` |
| `queue.Queue` | `chan T` |
| `threading.Lock` | `sync.Mutex` |
| `asyncio.wait_for(x, timeout)` | `context.WithTimeout` |
| `ThreadPoolExecutor(n)` | пул воркеров на канале или `errgroup` с `SetLimit(n)` |
| `asyncio.sleep(1)` | `time.Sleep(time.Second)` |

## Инструменты

| Python | Go |
|---|---|
| `python app.py` | `go run .` |
| `pip install x` / `uv add x` | `go get x` |
| `pyproject.toml` | `go.mod` |
| `venv` | не нужен |
| `pytest` | `go test ./...` |
| `black` / `ruff format` | `gofmt` / `go fmt` |
| `ruff check` / `flake8` | `go vet`, `golangci-lint` |
| `mypy` | компилятор |
| `pipx install` | `go install пакет@latest` |
| PyPI | <https://pkg.go.dev> |
| `pdb` | [`dlv`](https://github.com/go-delve/delve) (Delve) |
| `cProfile` | `pprof` |

## Топ-15 ловушек для Python-разработчика

1. **`:=` против `=`.** `:=` во вложенном блоке создаёт *новую* переменную,
   которая затеняет внешнюю:

   ```go
   count := 0
   if true {
   	count := 10 // новая переменная, внешняя count не изменилась!
   	_ = count
   }
   fmt.Println(count) // 0
   ```

2. **Срез слайса не копирует данные** — изменения видны в исходном
   ([глава 2](02-collections.md)).
3. **`append` возвращает результат** — `append(xs, x)` без присваивания бесполезен.
4. **Структуры копируются** при присваивании и передаче в функцию; чтобы
   менять оригинал, нужен указатель ([глава 3](03-structs-methods.md)).
5. **Копия в `range`.** Переменная цикла — копия элемента:

   ```go
   type item struct{ n int }
   items := []item{{1}, {2}}

   for _, it := range items {
   	it.n *= 10 // меняем копию
   }
   fmt.Println(items) // [{1} {2}]

   for i := range items {
   	items[i].n *= 10 // меняем сам элемент
   }
   fmt.Println(items) // [{10} {20}]
   ```

6. **`len("привет") == 12`** — длина в байтах, а не в символах.
7. **Порядок обхода map случайный.**
8. **Запись в nil-map — паника**; nil-слайс при этом безопасен.
9. **Отсутствующий ключ map возвращает zero value**, а не ошибку.
10. **Целочисленное деление**: `7 / 2 == 3`, `-7 / 2 == -3`, `int` переполняется.
11. **Нет «истинности»**: `if xs` не скомпилируется.
12. **Интерфейс с nil-указателем внутри не равен nil**
    ([глава 4](04-interfaces-generics.md)).
13. **`main` не ждёт горутины** ([глава 6](06-concurrency.md)).
14. **GIL нет** — общие данные между горутинами нужно защищать.
15. **`defer` срабатывает при выходе из функции**, а не из блока или итерации цикла.

## Как перестроить мышление

| Привычка из Python | Подход в Go |
|---|---|
| «Сначала напишу, типы потом» | Сначала типы и сигнатуры — компилятор помогает проектировать |
| «Проще попросить прощения» (EAFP, `try/except`) | Проверяй явно: `if err != nil`, `v, ok := m[k]` |
| Иерархии классов | Маленькие интерфейсы и композиция |
| Короче — лучше (comprehensions, однострочники) | Понятнее — лучше; обычный цикл — это нормально |
| Магия: декораторы, метаклассы, `__getattr__` | Явный код, минимум скрытого поведения |
| Фреймворк на каждую задачу | Сначала стандартная библиотека |
| Много способов сделать одно и то же | Обычно один общепринятый способ |
| Спор о стиле кода | `gofmt` решает за всех |

Короткие цитаты из [Go Proverbs](https://go-proverbs.github.io/), которые
стоит держать в голове:

- *Clear is better than clever.* — Понятное лучше хитрого.
- *Errors are values.* — Ошибки — это значения.
- *The bigger the interface, the weaker the abstraction.* — Чем больше интерфейс,
  тем слабее абстракция.
- *A little copying is better than a little dependency.* — Немного копирования
  лучше, чем лишняя зависимость.
- *Make the zero value useful.* — Делай нулевое значение полезным.

[← К оглавлению](../README.md)
