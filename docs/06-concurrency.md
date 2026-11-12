# 06. Конкурентность: горутины и каналы

[← Ошибки](05-errors.md) · [К оглавлению](../README.md) · [Дальше: пакеты и модули →](07-packages-modules.md)

Это главная причина, по которой Python-разработчики приходят в Go.

## Чем это отличается от Python

| | Python | Go |
|---|---|---|
| Единица работы | поток, корутина, процесс | горутина |
| Параллельность на ядрах | мешает GIL; нужен `multiprocessing` | да, по умолчанию на всех ядрах |
| Стоимость | поток ≈ мегабайты стека | горутина ≈ 2 КБ, можно запускать сотни тысяч |
| Синтаксис | `async def` / `await`, «раскрашенные» функции | обычные функции + слово `go` |
| Блокирующий вызов | в `asyncio` блокирует весь event loop | блокирует только свою горутину |
| Обмен данными | `queue.Queue`, `asyncio.Queue` | каналы |
| Ожидание нескольких событий | `asyncio.wait`, `asyncio.gather` | `select`, `sync.WaitGroup` |
| Отмена и таймауты | `asyncio.timeout`, `task.cancel()` | `context.Context` |

Ключевое: в Go **нет разделения на sync и async функции**. Любую функцию можно
запустить в горутине, и любой блокирующий вызов (сеть, диск, `time.Sleep`)
планировщик обслуживает сам. Библиотек-двойников вроде `requests` / `aiohttp` нет.

## Горутины

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done() // сообщаем «я закончил» при любом выходе из функции

	time.Sleep(100 * time.Millisecond) // имитация работы
	fmt.Println("воркер", id, "готов")
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)         // +1 к счётчику ДО запуска горутины
		go worker(i, &wg) // go — запустить функцию в новой горутине
	}

	wg.Wait() // ждём, пока счётчик не станет 0
	fmt.Println("все закончили")
}
```

Три воркера отработают за ~100 мс, а не за 300. Порядок вывода не определён.

> ⚠️ Когда `main` завершается, программа завершается сразу — **не дожидаясь**
> остальных горутин. Без `wg.Wait()` эта программа ничего бы не вывела.

`sync.WaitGroup` — аналог `asyncio.gather` для случая «просто дождаться всех».
Начиная с Go 1.25 пару `wg.Add(1)` + `go` + `defer wg.Done()` можно заменить
одним вызовом `wg.Go(func() { ... })`.

## Каналы

Канал — типизированная труба между горутинами. Девиз Go:

> Не общайтесь через общую память — делите память через общение.

```go
package main

import "fmt"

func main() {
	ch := make(chan int) // небуферизованный канал

	go func() {
		ch <- 42 // отправка: блокируется, пока кто-то не прочитает
	}()

	v := <-ch // чтение: блокируется, пока кто-то не отправит
	fmt.Println(v)

	// Буферизованный канал: отправка не блокируется, пока в буфере есть место
	buf := make(chan string, 2)
	buf <- "a"
	buf <- "b"
	fmt.Println(<-buf, <-buf)
}
```

Небуферизованный канал — это ещё и точка синхронизации: отправитель и
получатель «встречаются». Буферизованный ведёт себя как `queue.Queue(maxsize=N)`.

### Закрытие канала и range

```go
package main

import "fmt"

// chan<- int — канал только для отправки, <-chan int — только для чтения.
// Направление в сигнатуре защищает от ошибок и документирует намерение.
func generate(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // закрывает канал ТОЛЬКО отправитель
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in { // range читает, пока канал не закроют
			out <- v * v
		}
	}()
	return out
}

func main() {
	// конвейер (pipeline): generate -> square -> вывод
	for v := range square(generate(5)) {
		fmt.Println(v) // 1 4 9 16 25
	}

	ch := make(chan int, 1)
	ch <- 7
	close(ch)

	v, ok := <-ch
	fmt.Println(v, ok) // 7 true  — в буфере ещё было значение
	v, ok = <-ch
	fmt.Println(v, ok) // 0 false — канал закрыт и пуст
}
```

### select

`select` ждёт сразу несколько операций с каналами и выполняет ту, что готова первой.

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	result := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		result <- "готово"
	}()

	select {
	case res := <-result:
		fmt.Println("получили:", res)
	case <-time.After(time.Second): // таймаут
		fmt.Println("слишком долго")
	}

	// default делает select неблокирующим
	select {
	case res := <-result:
		fmt.Println(res)
	default:
		fmt.Println("данных пока нет")
	}
}
```

## Общая память: Mutex и гонки данных

Если несколько горутин меняют одну переменную без синхронизации — это
**гонка данных** (data race). В отличие от Python, GIL тебя не прикроет:
результат будет неверным или программа упадёт.

```go
package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n[key]++
}

func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n[key]
}

func main() {
	c := Counter{n: make(map[string]int)}
	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc("клики")
		}()
	}

	wg.Wait()
	fmt.Println(c.Get("клики")) // всегда 1000
}
```

Убери `Lock`/`Unlock` — и программа упадёт с `fatal error: concurrent map writes`.

Встроенный детектор гонок — запускай с ним тесты всегда:

```bash
go test -race ./...
```

```bash
go run -race .
```

Что выбрать:

| Задача | Инструмент |
|---|---|
| передать данные или владение ими между горутинами | канал |
| защитить общее состояние (кэш, счётчик, map) | `sync.Mutex` / `sync.RWMutex` |
| простой числовой счётчик | `sync/atomic` |
| выполнить что-то ровно один раз | `sync.Once` |
| дождаться группы горутин | `sync.WaitGroup` |
| дождаться группы и получить первую ошибку | `errgroup.Group` |

## context: отмена и таймауты

`context.Context` передаёт сигнал отмены и дедлайн вниз по цепочке вызовов.
По соглашению это **первый параметр** функции с именем `ctx`.

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func slowOperation(ctx context.Context) error {
	select {
	case <-time.After(2 * time.Second): // работа, которая длится 2 секунды
		return nil
	case <-ctx.Done(): // контекст отменён или истёк таймаут
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel() // всегда вызывай cancel, чтобы освободить ресурсы

	err := slowOperation(ctx)
	fmt.Println(err) // context deadline exceeded
}
```

Почти все функции стандартной библиотеки, работающие с сетью и БД, принимают
`ctx`: HTTP-запросы, `database/sql` и т. д. В HTTP-обработчике контекст
запроса — `r.Context()`: он отменяется, когда клиент разорвал соединение.

## Паттерн: пул воркеров

Ограничиваем число одновременно работающих горутин — аналог
`ThreadPoolExecutor(max_workers=3)`.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func process(job int) int {
	time.Sleep(50 * time.Millisecond)
	return job * 2
}

func main() {
	const numWorkers = 3

	jobs := make(chan int)
	results := make(chan int)

	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs { // каждый воркер берёт задачи из общего канала
				results <- process(job)
			}
		}()
	}

	// отправляем задачи и закрываем канал — сигнал «задач больше не будет»
	go func() {
		for i := 1; i <= 9; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// когда все воркеры закончат — закрываем results, чтобы завершился range ниже
	go func() {
		wg.Wait()
		close(results)
	}()

	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("сумма:", sum) // 90
}
```

## Паттерн: errgroup

Пакет `golang.org/x/sync/errgroup` — группа горутин, которая возвращает первую
ошибку и отменяет остальных. Ближе всего к `asyncio.TaskGroup`.

```bash
go get golang.org/x/sync/errgroup
```

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/sync/errgroup"
)

func check(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("запрос %s: %w", url, err)
	}
	defer resp.Body.Close()

	fmt.Println(url, resp.Status)
	return nil
}

func main() {
	urls := []string{"https://go.dev", "https://pkg.go.dev"}

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(5) // не больше 5 запросов одновременно

	for _, url := range urls {
		g.Go(func() error {
			return check(ctx, url)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("ошибка:", err)
	}
}
```

## Типичные ошибки

| Ошибка | Что происходит | Как избежать |
|---|---|---|
| `main` не ждёт горутины | программа завершается раньше них | `WaitGroup`, канал, `errgroup` |
| Отправка в канал, из которого никто не читает | горутина зависает навсегда (утечка) | буфер, `select` с `ctx.Done()` |
| Все горутины заблокированы | `fatal error: all goroutines are asleep - deadlock!` | проверь, кто закрывает канал и кто читает |
| Отправка в закрытый канал | `panic` | закрывает только отправитель, один раз |
| Чтение/запись nil-канала | блокировка навсегда | создавай канал через `make` |
| Общая переменная без защиты | гонка данных | мьютекс, канал, `-race` в тестах |
| Копирование структуры с `sync.Mutex` | каждая копия получает свой замок | передавай по указателю; `go vet` это ловит |
| Горутина без условия выхода | утечка памяти | всегда продумывай, как она завершится |

> До Go 1.22 переменная цикла была общей на все итерации, и замыкание в `go func()`
> внутри `for` видело её последнее значение. В старом коде и статьях из-за этого
> встречается строка `i := i`. В Go 1.22+ у каждой итерации своя переменная.

## Практика

1. Скачай 10 URL параллельно и выведи код ответа и время каждого запроса.
2. Перепиши то же через пул из 3 воркеров.
3. Добавь общий таймаут 2 секунды через `context.WithTimeout`.
4. Напиши потокобезопасный кэш `Cache[K comparable, V any]` с `sync.RWMutex`
   и проверь его тестом с флагом `-race`.
5. Намеренно получи deadlock и прочитай сообщение рантайма.

[Дальше: пакеты и модули →](07-packages-modules.md)
