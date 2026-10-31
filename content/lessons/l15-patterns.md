---
id: l15
week: 3
day: 5
minutes: 120
title: Паттерны конкурентности
goals: Worker pool, fan-out/fan-in, pipeline, ограничение параллелизма, errgroup, rate limit
lc: task-scheduler | Task Scheduler | Medium | Жадность по самой частой задаче или куча + очередь «остывания»
lc: top-k-frequent-words | Top K Frequent Words | Medium | Частоты + сортировка с составным компаратором (как в задании по частоте слов)
lc: merge-intervals | Merge Intervals | Medium | Сортировка по началу и слияние за один проход
---

## Worker Pool

Фиксированное число воркеров читает задачи из общего канала. Ограничивает использование ресурсов и сглаживает пики.

```go run
package main

import (
	"fmt"
	"sort"
	"sync"
)

type Job struct{ ID, N int }
type Result struct{ ID, Square int }

func worker(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs { // завершится, когда jobs закроют
		results <- Result{j.ID, j.N * j.N}
	}
}

func main() {
	jobs := make(chan Job)
	results := make(chan Result)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ { // 3 воркера
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	go func() { // отправитель: закрывает jobs, когда задачи кончились
		for i := 1; i <= 8; i++ {
			jobs <- Job{ID: i, N: i}
		}
		close(jobs)
	}()

	go func() { wg.Wait(); close(results) }() // закрываем results, когда все воркеры вышли

	var all []Result
	for r := range results {
		all = append(all, r)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID }) // порядок завершения не определён
	fmt.Println(all)
}
```

Схема владения: тот, кто **пишет** в канал, его и закрывает; результаты закрывает «координатор» после `wg.Wait()`.

## Pipeline

Цепочка стадий, соединённых каналами; каждая стадия — горутина(ы) с одним входом и одним выходом. Отмена через `context` предотвращает утечки, если потребитель ушёл раньше.

```go run
package main

import (
	"context"
	"fmt"
)

func gen(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func sq(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // если выйдем раньше времени — все стадии остановятся

	for v := range sq(ctx, sq(ctx, gen(ctx, 1, 2, 3))) { // n⁴
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

## Fan-out / Fan-in

**Fan-out**: несколько горутин читают из одного канала (распределяют работу). **Fan-in**: объединение нескольких каналов в один (`merge` из урока про каналы).

## Ограничение параллелизма

Семафор на буферизованном канале — самый простой вариант:

```go run
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	const limit = 3
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	var cur, peak atomic.Int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // занять слот (блок, если слотов нет)
			defer func() { <-sem }() // освободить

			n := cur.Add(1)
			for { // обновляем максимум атомарно
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			cur.Add(-1)
		}()
	}
	wg.Wait()
	fmt.Println("одновременно не больше", limit, "-> пик:", peak.Load() <= limit)
}
```

В библиотеке `golang.org/x/sync` есть `semaphore.Weighted` (взвешенный семафор) и `errgroup` с `SetLimit`.

## errgroup

Группа горутин, где **первая ошибка отменяет контекст** остальных, а `Wait()` возвращает её. Это внешний пакет `golang.org/x/sync/errgroup`; ниже минимальная собственная реализация (то же самое ты напишешь в задании), поэтому в проекте без зависимостей она тоже пригодится:

```go run
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Group struct {
	wg     sync.WaitGroup
	once   sync.Once
	err    error
	cancel context.CancelFunc
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.once.Do(func() { // запоминаем только первую ошибку
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

func main() {
	g, ctx := WithContext(context.Background())
	for i := 1; i <= 3; i++ {
		g.Go(func() error {
			select {
			case <-time.After(time.Duration(i) * 20 * time.Millisecond):
				if i == 2 {
					return errors.New("задача 2 упала")
				}
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}
	fmt.Println("итог:", g.Wait())
}
```

## Rate limiting

```go
// простейший: не чаще 1 раза в 100 мс
tick := time.NewTicker(100 * time.Millisecond)
defer tick.Stop()
for req := range requests {
	<-tick.C
	go handle(req)
}
```

Для «бёрстов» и реального использования применяют **token bucket** (`golang.org/x/time/rate`): `limiter.Wait(ctx)` / `Allow()`. Алгоритм ты реализуешь сам в неделе 8.

## Другие паттерны

- **Pub/Sub**: список каналов подписчиков под мьютексом; публикация — рассылка (неблокирующая отправка через `select` + `default` или буферы).
- **Singleflight** (`x/sync/singleflight`): одинаковые одновременные запросы к медленному ресурсу выполняются один раз, результат делится на всех. Решает «cache stampede».
- **Graceful shutdown**: сигнал → отмена контекста → дождаться воркеров (`WaitGroup`) с таймаутом.
- **Or-done / tee / bridge** — комбинаторы каналов; нужны редко.
- **Heartbeat / watchdog** — периодические сигналы жизни, `time.Ticker`.

## Как не отстрелить себе ногу

1. Для каждой горутины определи: **кто её остановит**, **что если потребитель ушёл**, **что если паника**.
2. Буфер канала — не лекарство от deadlock, а оптимизация; размер выбирай осознанно.
3. Не плоди горутины без ограничений (по горутине на элемент миллиона записей) — ограничивай пулом/семафором.
4. Тестируй с `-race` и `-count=100`; используй `goleak` для проверки утечек.
5. Предпочитай простоту: часто достаточно `WaitGroup` + мьютекс, без каналов.

## Проверь себя

<details><summary>Как завершить worker pool корректно?</summary>Закрыть канал задач; воркеры выходят из <code>range</code>; после <code>wg.Wait()</code> закрыть канал результатов.</details>

<details><summary>Как ограничить число одновременных горутин?</summary>Пул воркеров или семафор на буферизованном канале (<code>sem &lt;- struct{}{}</code> / <code>&lt;-sem</code>), либо <code>errgroup.SetLimit</code>.</details>

<details><summary>Чем errgroup лучше WaitGroup?</summary>Собирает первую ошибку, отменяет общий контекст при сбое, умеет ограничивать параллелизм.</details>

<details><summary>Как предотвратить утечку горутин в pipeline?</summary>Все отправки в канал делать через <code>select</code> с <code>ctx.Done()</code> и отменять контекст при выходе потребителя.</details>
