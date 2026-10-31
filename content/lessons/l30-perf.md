---
id: l30
week: 6
day: 5
minutes: 120
title: Профилирование и оптимизация
goals: Мерить, а не гадать: бенчмарки, pprof (CPU, heap, goroutine, mutex), trace, типичные оптимизации
lc: range-sum-query-mutable | Range Sum Query - Mutable | Medium | Дерево Фенвика или сегментов: обновление и запрос за O(log n)
lc: range-sum-query-2d-immutable | Range Sum Query 2D - Immutable | Medium | Двумерные префиксные суммы
---

> «Преждевременная оптимизация — корень всех зол» (Кнут). Правильная последовательность: **сделай правильно → сделай читаемо → измерь → оптимизируй узкое место → измерь снова**.

## Бенчмарки

Бенчмарки — основа любой оптимизации: без числа нет улучшения. Их можно запускать даже из обычной программы через `testing.Benchmark`:

```go run
package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

var sink string // результат присваиваем глобальной переменной, чтобы компилятор не выбросил код

func concat(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += strconv.Itoa(i) + ","
	}
	return s
}

func builder(n int) string {
	var sb strings.Builder
	sb.Grow(n * 4)
	for i := 0; i < n; i++ {
		sb.WriteString(strconv.Itoa(i))
		sb.WriteByte(',')
	}
	return sb.String()
}

func report(name string, f func(b *testing.B)) {
	r := testing.Benchmark(f)
	fmt.Printf("%-10s %8d ns/op %6d B/op %4d allocs/op\n", name, r.NsPerOp(), r.AllocedBytesPerOp(), r.AllocsPerOp())
}

func main() {
	report("concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink = concat(500)
		}
	})
	report("builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink = builder(500)
		}
	})
}
```

В реальных проектах: `func BenchmarkX(b *testing.B)` в `_test.go`, запуск `go test -bench=. -benchmem -count=10 -run=^$`, сравнение «до/после» — `benchstat old.txt new.txt` (учитывает разброс). Единицы: `ns/op`, `B/op` (байт на операцию), `allocs/op` (число аллокаций).

Ловушки бенчмарков: компилятор вырезает «мёртвый» код (используй sink / `b.Loop()` с Go 1.24), прогрев, шум других процессов, подготовка данных внутри цикла (`b.ResetTimer()`), разные размеры входа (`b.Run` с подбенчмарками).

## pprof: где тратится время и память

Профили собираются сэмплированием. Подключение к сервису:

```go
import _ "net/http/pprof"   // регистрирует /debug/pprof/* в DefaultServeMux
go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }() // ТОЛЬКО на внутреннем интерфейсе!
```

```bash
go tool pprof -http=:9090 http://localhost:6060/debug/pprof/profile?seconds=30   # CPU
go tool pprof -http=:9090 http://localhost:6060/debug/pprof/heap                 # память (inuse/alloc)
go tool pprof http://localhost:6060/debug/pprof/goroutine                        # где «застряли» горутины
go tool pprof http://localhost:6060/debug/pprof/mutex                            # конкуренция за мьютексы
go tool pprof http://localhost:6060/debug/pprof/block                            # блокировки на каналах/sync
```

(`mutex` и `block` нужно включить: `runtime.SetMutexProfileFraction(5)`, `runtime.SetBlockProfileRate(1)`.)

Из тестов: `go test -bench=. -cpuprofile cpu.out -memprofile mem.out`, затем `go tool pprof -http=:9090 cpu.out`.

Как читать: **flame graph** (ширина блока = доля времени; ищи широкие «плато»), `top` (flat = внутри функции, cum = с вызываемыми), `list FuncName` (построчно). Профиль heap: `inuse_space` (что сейчас жив) против `alloc_space` (что аллоцировалось вообще — для давления на GC).

`go tool trace` — временная шкала: планировщик, GC, блокировки, сетевой ввод-вывод; помогает найти простаивающие ядра и «лестницы» блокировок.

## Типичные источники медленной работы и приёмы

| Симптом | Приём |
|---|---|
| Много аллокаций | предвыделение, `sync.Pool`, переиспользование буферов, избегать `interface{}`/рефлексии в горячем пути |
| Конкатенация строк | `strings.Builder`, `strconv.Append*` вместо `fmt.Sprintf` |
| Конкуренция за один мьютекс | шардирование (N независимых map+mutex по хэшу ключа), `RWMutex`, atomic, уменьшение критической секции |
| Медленный JSON | типизированные структуры вместо `map[string]any`, кодогенерация/альтернативные библиотеки |
| Большая куча с указателями | плоские массивы значений, индексы вместо указателей |
| N+1 запросов, нет индексов | батчинг, индексы, `EXPLAIN` — чаще всего узкое место **не в Go** |
| Горутина на каждую мелочь | пул воркеров, батчи |
| GC съедает CPU | уменьшить аллокации; `GOGC`/`GOMEMLIMIT` |
| Лишнее копирование | передавать срезы/указатели для больших структур, `io.Copy` потоком |
| Холодный старт | ленивая инициализация, `sync.Once` |

### PGO (Profile-Guided Optimization)

Go 1.21+: положи профиль CPU рабочего сервиса в `default.pgo` рядом с `main` — компилятор будет агрессивнее инлайнить горячие пути (прирост обычно 2–14%): `go build -pgo=auto`.

### Шардирование — ключевой приём против contention

```go
type Sharded struct {
	shards [64]struct {
		mu sync.Mutex
		m  map[string]int
		_  [40]byte // padding — защита от false sharing между соседними кэш-линиями
	}
}

func (s *Sharded) shard(key string) *struct{ ... } {
	h := fnv32(key)
	return &s.shards[h%uint32(len(s.shards))]
}
```

В задании ты сделаешь такую карту. **False sharing**: два независимо изменяемых поля в одной кэш-линии (64 байта) заставляют ядра «перебрасывать» линию друг другу — выравнивай горячие поля/счётчики на разные линии.

## Приоритеты оптимизации

1. **Алгоритм и структуры данных** (O(n²) → O(n log n) важнее любой микрооптимизации).
2. **Ввод-вывод**: батчинг, кэш, индексы, параллельные запросы.
3. **Аллокации** и давление на GC.
4. **Конкуренция** (локи).
5. Микрооптимизации — в последнюю очередь и только на основе профиля.

> [!INTERVIEW] Типичный вопрос: «Сервис стал медленнее после релиза — что делаешь?» Ответ: смотрю метрики (RED), сравниваю версии, снимаю CPU/heap/goroutine профили на проблемном инстансе, ищу аллокации/блокировки/сетевые задержки, воспроизвожу бенчмарком, исправляю и подтверждаю benchstat, выкатываю канарейкой.

## Проверь себя

<details><summary>Как отличить CPU-bound проблему от I/O-bound?</summary>CPU-профиль показывает горячие функции — если время в вычислениях, это CPU; если ядра простаивают, а запросы медленные — смотри block/mutex профили, trace и внешние зависимости (БД, сеть).</details>

<details><summary>Что означает <code>allocs/op</code> и почему он важен?</summary>Число аллокаций на операцию; каждая — нагрузка на аллокатор и GC, поэтому снижение allocs/op часто даёт заметный выигрыш.</details>

<details><summary>Что такое false sharing?</summary>Независимые переменные в одной кэш-линии, которые изменяют разные ядра: линия постоянно инвалидируется, производительность падает; лечится выравниванием/паддингом.</details>

<details><summary>Зачем <code>benchstat</code>?</summary>Статистически сравнивает несколько прогонов до/после и показывает, значима ли разница, а не шум.</details>
