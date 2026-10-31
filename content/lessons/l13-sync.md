---
id: l13
week: 3
day: 3
minutes: 120
title: Пакет sync и atomic
goals: Mutex, RWMutex, Once, atomic, sync.Pool, sync.Map, Cond; понимать deadlock и гонки
lc: insert-delete-getrandom-o1 | Insert Delete GetRandom O(1) | Medium | Слайс значений + map значение→индекс; удаление подменой последним
lc: design-hashset | Design HashSet | Easy | Бакеты слайсов или простой map — тренируй структуру с методами
lc: lfu-cache | LFU Cache | Hard | Группы по частоте + минимальная частота; повышенная сложность (по желанию)
---

## sync.Mutex

Взаимное исключение: в критической секции одновременно только одна горутина.

```go run
package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	mu sync.Mutex // нулевое значение — разблокированный; ставь рядом с защищаемыми полями
	m  map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock() // defer — чтобы не забыть разблокировать при return/panic
	c.m[key]++
}

func (c *SafeCounter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[key]
}

func main() {
	c := SafeCounter{m: map[string]int{}}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc("hits") }()
	}
	wg.Wait()
	fmt.Println(c.Get("hits"))
}
```

Правила:

1. **Не копируй** структуры с мьютексом (после первого использования): методы — на указателе; `go vet` (copylocks) предупредит.
2. Держи блокировку **как можно короче**; не вызывай чужой код, колбэки, не делай I/O под мьютексом.
3. Мьютекс **не реентерабелен**: повторный `Lock()` из той же горутины = deadlock.
4. Порядок захвата нескольких мьютексов должен быть везде одинаковым (иначе A→B и B→A дадут взаимную блокировку).

## RWMutex

Много читателей **или** один писатель. Выгоден, когда чтений сильно больше и критическая секция заметная; при коротких секциях обычный `Mutex` часто быстрее.

```go
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *Cache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}
```

> [!WARNING] Нельзя «повысить» RLock до Lock — отпусти читательскую блокировку, возьми писательскую и **перепроверь условие** (между ними состояние могло измениться).

## sync.Once — ровно один раз

```go run
package main

import (
	"fmt"
	"sync"
)

var (
	once sync.Once
	cfg  map[string]string
)

func config() map[string]string {
	once.Do(func() { // даже при гонке из 100 горутин выполнится один раз; остальные ждут завершения
		fmt.Println("загружаю конфиг")
		cfg = map[string]string{"env": "prod"}
	})
	return cfg
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = config() }()
	}
	wg.Wait()
	fmt.Println(config()["env"])

	// Go 1.21+: OnceValue — ленивое вычисление значения
	get := sync.OnceValue(func() int { fmt.Println("считаю..."); return 42 })
	fmt.Println(get(), get())
}
```

Применения: ленивая инициализация синглтона, закрытие ресурса один раз (`closeOnce.Do(func() { close(ch) })`).

## sync/atomic

Атомарные операции над числами и указателями без блокировок — для счётчиков, флагов и «горячих» путей.

```go run
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var hits atomic.Int64
	var ready atomic.Bool
	var cfg atomic.Pointer[string] // атомарная подмена целого неизменяемого значения

	s := "v1"
	cfg.Store(&s)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hits.Add(2)
		}()
	}
	wg.Wait()
	ready.Store(true)

	// CompareAndSwap: «если сейчас X — заменить на Y» — основа lock-free алгоритмов
	var n atomic.Int32
	fmt.Println(n.CompareAndSwap(0, 5), n.CompareAndSwap(0, 9), n.Load())

	fmt.Println(hits.Load(), ready.Load(), *cfg.Load())
}
```

Атомики дают корректность отдельных операций, но **не** защищают инварианты из нескольких полей — для них нужен мьютекс.

## sync.Pool

Пул временных объектов для снижения нагрузки на GC (буферы, структуры при обработке запросов). Объекты могут быть удалены GC в любой момент — это **кэш**, а не хранилище.

```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func render() string {
	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset() // всегда очищай перед использованием
	b.WriteString("hello")
	return b.String()
}
```

## sync.Map

Потокобезопасная map; оптимальна в двух случаях: ключ записывается один раз, а читается много; разные горутины работают с непересекающимися ключами. В остальных случаях `map` + `Mutex`/`RWMutex` проще, типобезопаснее и часто быстрее.

## sync.Cond (редко)

Условная переменная — «спать, пока не изменится состояние под мьютексом» (`Wait`, `Signal`, `Broadcast`). Нужна для реализации блокирующих очередей; на практике почти всегда заменяется каналами. Ждать нужно в цикле: `for !condition { c.Wait() }`.

## Deadlock: типичные сценарии

```go
mu.Lock()
mu.Lock() // повторный захват — зависание навсегда

// A: Lock(m1); Lock(m2)      B: Lock(m2); Lock(m1)   → взаимная блокировка

wg.Add(1); wg.Wait()          // никто не вызвал Done
ch := make(chan int); ch <- 1 // никто не читает
```

Диагностика: `kill -QUIT <pid>` печатает стеки всех горутин; профиль pprof `goroutine`; в тестах — `-timeout`.

## Как выбрать

| Задача | Инструмент |
|---|---|
| Защитить структуру/map | `Mutex` (`RWMutex` при тяжёлых чтениях) |
| Счётчик, флаг | `atomic` |
| Разовая инициализация | `Once`/`OnceValue` |
| Передача данных, конвейер, сигналы | каналы |
| Много временных буферов | `sync.Pool` |
| Ждать группу горутин | `WaitGroup` / `errgroup` |

## Проверь себя

<details><summary>Почему нельзя копировать <code>sync.Mutex</code>?</summary>Копия содержит состояние блокировки на момент копирования; две копии не синхронизируются друг с другом. <code>go vet</code> ловит это (copylocks).</details>

<details><summary>Чем <code>RWMutex</code> отличается от <code>Mutex</code>?</summary>Позволяет нескольким читателям работать одновременно; писатель — эксклюзивно.</details>

<details><summary>Что произойдёт при concurrent write в обычный map?</summary>Runtime обнаружит это и аварийно завершит программу: <code>fatal error: concurrent map writes</code> (не panic — её нельзя перехватить).</details>

<details><summary>Когда использовать atomic вместо mutex?</summary>Для независимых простых значений (счётчик, флаг, указатель на неизменяемую конфигурацию); для многополевых инвариантов — mutex.</details>
