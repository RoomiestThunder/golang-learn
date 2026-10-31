---
id: l14
week: 3
day: 4
minutes: 120
title: Пакет context
goals: Отмена, дедлайны и значения запроса; проброс контекста через слои; утечки и ошибки использования
lc: time-based-key-value-store | Time Based Key-Value Store | Medium | Для каждого ключа — слайс (timestamp, value); бинарный поиск sort.Search
lc: design-underground-system | Design Underground System | Medium | Если не делал в уроке 8 — самое время
lc: moving-average-from-data-stream | Moving Average from Data Stream | Easy | Очередь + текущая сумма (аналог скользящего окна)
---

## Зачем

В серверах у каждого запроса есть время жизни: клиент отключился, истёк таймаут, сработала отмена выше по стеку. Все порождённые операции (запросы в БД, вызовы других сервисов, горутины) должны **вовремя остановиться** и освободить ресурсы. `context.Context` — стандартный способ нести **отмену, дедлайн и значения** через границы API.

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}  // закрывается при отмене/истечении
	Err() error             // nil | context.Canceled | context.DeadlineExceeded
	Value(key any) any
}
```

## Создание контекстов

```go run
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func worker(ctx context.Context, id int) error {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done(): // единый способ узнать об отмене
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
			// полезная работа
		}
	}
}

func main() {
	// корень дерева: Background (main, тесты, верхний уровень) / TODO (пока непонятно какой)
	root := context.Background()

	// WithCancel: ручная отмена
	ctx, cancel := context.WithCancel(root)
	done := make(chan error)
	go func() { done <- worker(ctx, 1) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	fmt.Println("отмена:", <-done)

	// WithTimeout: отмена по таймеру (WithDeadline — по абсолютному времени)
	ctx2, cancel2 := context.WithTimeout(root, 60*time.Millisecond)
	defer cancel2() // ВСЕГДА вызывай cancel, даже если таймаут сработает: освобождает ресурсы
	err := worker(ctx2, 2)
	fmt.Println(err, errors.Is(err, context.DeadlineExceeded))
}
```

Контексты образуют **дерево**: отмена родителя отменяет всех потомков; отмена потомка родителя не затрагивает. Дедлайн потомка не может быть позже дедлайна родителя.

## Правила использования

1. **Первый параметр** функции, называется `ctx`: `func Do(ctx context.Context, ...)`.
2. **Не храни** контекст в структуре (исключения — редки); передавай явно по цепочке вызовов.
3. **Всегда** вызывай `cancel` (обычно `defer cancel()` сразу после создания) — иначе утекут таймеры и дочерние ресурсы. `go vet` проверяет это (lostcancel).
4. Не передавай `nil` вместо контекста — используй `context.TODO()`.
5. Блокирующие операции должны уважать `ctx.Done()`: `select` с `case <-ctx.Done()`, либо использовать API, принимающие контекст (`http.NewRequestWithContext`, `db.QueryContext`, `exec.CommandContext`).
6. Проверка отмены в долгих циклах: `if err := ctx.Err(); err != nil { return err }`.

## Значения в контексте

`WithValue` — для **данных, привязанных к запросу и пересекающих API** (request ID, пользователь, trace id). Не для опциональных параметров функций и не для зависимостей (логгер, БД — в структуры).

```go run
package main

import (
	"context"
	"fmt"
)

type ctxKey int // собственный непубличный тип ключа — исключает коллизии между пакетами

const requestIDKey ctxKey = 0

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

func handler(ctx context.Context) {
	id, _ := RequestID(ctx)
	fmt.Println("обрабатываю запрос", id)
}

func main() {
	handler(WithRequestID(context.Background(), "req-42"))
}
```

## Причины и дополнительные возможности

- `context.WithCancelCause(parent)` + `context.Cause(ctx)` (Go 1.20) — передать *почему* отменили.
- `context.WithoutCancel(ctx)` (1.21) — сохранить значения, но не отмену (для «дожать» фоновую работу после ответа клиенту).
- `context.AfterFunc(ctx, f)` — выполнить функцию после отмены.
- `context.WithTimeoutCause`, `WithDeadlineCause`.

## Контекст в HTTP-сервере

```go
func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // отменится, когда клиент закроет соединение или завершится обработчик
	user, err := s.repo.GetUser(ctx, 42)
	if errors.Is(err, context.Canceled) { return } // клиент ушёл — не пишем ответ
	...
}
```

## Типичные ошибки

- Забыть `defer cancel()` → утечка.
- Игнорировать `ctx` в долгой операции → отмена не работает, горутины висят.
- Использовать `context.Background()` внутри обработчика вместо `r.Context()` → потеря отмены.
- Класть в контекст всё подряд (тип `map[string]any`) — скрытые зависимости.
- Создавать `WithTimeout` на каждый вызов в горячем цикле без нужды.

> [!INTERVIEW] «Как реализовать таймаут вызова функции без поддержки контекста?» — запустить её в горутине, результат отправить в буферизованный канал (cap=1, чтобы горутина не зависла), а ждать через `select` с `ctx.Done()`. Саму функцию остановить нельзя — горутина дорабатывает «в фоне», поэтому внешние API должны принимать `ctx`.

## Проверь себя

<details><summary>Что вернёт <code>ctx.Err()</code> до отмены и после?</summary>До отмены — nil; после — <code>context.Canceled</code> (ручная отмена) или <code>context.DeadlineExceeded</code> (таймаут).</details>

<details><summary>Зачем вызывать <code>cancel()</code>, если таймаут всё равно сработает?</summary>Чтобы немедленно освободить таймер и ресурсы дочерних контекстов, не дожидаясь срабатывания.</details>

<details><summary>Что можно хранить в context.Value?</summary>Данные, относящиеся к запросу и пересекающие границы API (request-id, auth-информация); не зависимости и не опциональные параметры.</details>

<details><summary>Почему тип ключа делают собственным, а не <code>string</code>?</summary>Чтобы ключи разных пакетов не конфликтовали (одинаковые строки разных типов — разные ключи).</details>
