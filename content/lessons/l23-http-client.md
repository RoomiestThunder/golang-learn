---
id: l23
week: 5
day: 3
minutes: 120
title: HTTP-клиент, таймауты и graceful shutdown
goals: Правильно ходить в чужие сервисы, ретраить, тестировать клиент, корректно останавливать сервер
lc: number-of-recent-calls | Number of Recent Calls | Easy | Очередь: выбрасывай запросы старше 3000 мс
lc: design-circular-deque | Design Circular Deque | Medium | Кольцевой буфер с двумя концами
---

## http.Client: типичные ошибки

`http.Get(url)` использует `http.DefaultClient` **без таймаута** — запрос может висеть вечно. В production клиент настраивают всегда:

```go
client := &http.Client{
	Timeout: 10 * time.Second, // общий дедлайн на запрос целиком (включая чтение тела)
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,   // по умолчанию всего 2! — для высокой нагрузки на один хост мало
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}
```

**Клиент создают один раз и переиспользуют** (внутри пул соединений, потокобезопасен). Создание клиента на каждый запрос = утечка и потеря keep-alive.

### Запрос с контекстом

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil { return err }
req.Header.Set("Accept", "application/json")

resp, err := client.Do(req)
if err != nil { return fmt.Errorf("запрос: %w", err) } // сетевая ошибка/таймаут
defer resp.Body.Close()                                // ОБЯЗАТЕЛЬНО, иначе утечка соединения

if resp.StatusCode != http.StatusOK {                  // 4xx/5xx — НЕ ошибка err!
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	return fmt.Errorf("статус %d: %s", resp.StatusCode, body)
}
var out Result
if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { ... }
```

Три правила клиента: **закрывай `Body`**, **проверяй статус**, **ограничивай читаемый размер**. Чтобы соединение вернулось в пул для повторного использования, тело нужно дочитать до конца (`io.Copy(io.Discard, resp.Body)`) перед закрытием.

## Ретраи

Повторять имеет смысл **временные** ошибки: сетевые, 429, 502/503/504 — и только **идемпотентные** запросы (GET, PUT, DELETE; POST — лишь с idempotency key). Алгоритм: экспоненциальный backoff **с джиттером** (случайной добавкой, чтобы клиенты не «били» сервер синхронно), ограничение числа попыток, уважение `ctx` и заголовка `Retry-After`. Не повторяй 4xx (кроме 408/429).

## Тестирование клиента: httptest

```go run
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type StatusError struct{ Code int }

func (e *StatusError) Error() string { return fmt.Sprintf("неожиданный статус %d", e.Code) }

func fetch(ctx context.Context, c *http.Client, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &StatusError{resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func main() {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			json.NewEncoder(w).Encode(map[string]int{"n": 42})
		case "/flaky":
			if calls.Add(1) < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			json.NewEncoder(w).Encode(map[string]int{"n": 1})
		case "/slow":
			time.Sleep(300 * time.Millisecond)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	ctx := context.Background()

	var out map[string]int
	fmt.Println(fetch(ctx, client, srv.URL+"/ok", &out), out)

	var se *StatusError
	err := fetch(ctx, client, srv.URL+"/missing", &out)
	fmt.Println(errors.As(err, &se), se)

	err = fetch(ctx, client, srv.URL+"/slow", &out) // сработает таймаут клиента
	fmt.Println(err != nil)

	// ретрай до успеха
	for attempt := 1; attempt <= 5; attempt++ {
		if err := fetch(ctx, client, srv.URL+"/flaky", &out); err == nil {
			fmt.Println("успех с попытки", attempt, out)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
```

Для подмены транспорта без сети реализуй `http.RoundTripper` (`type roundTripFunc func(*http.Request) (*http.Response, error)`) — удобно для юнит-тестов клиентов.

## Graceful shutdown сервера

Цель: при SIGTERM (деплой, Kubernetes) **перестать принимать новые соединения, дать завершиться текущим запросам** и закрыть ресурсы.

```go
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8080", Handler: newRouter(), ReadHeaderTimeout: 5 * time.Second}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:                // сервер упал сам (например, порт занят)
		return err
	case <-ctx.Done():                  // получен сигнал
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil { // ждёт активные запросы, не прерывая их
		return fmt.Errorf("shutdown: %w", err)
	}
	// здесь закрываем БД, очереди, флашим логи/метрики
	return nil
}
```

Нюансы:

- `ListenAndServe` после `Shutdown` возвращает `http.ErrServerClosed` — это **не ошибка**.
- Долгоживущие соединения (WebSocket, SSE) `Shutdown` не закрывает — используй `srv.RegisterOnShutdown`.
- В Kubernetes: readiness-проба должна начать отдавать «не готов» **до** остановки, чтобы балансировщик перестал слать трафик; `terminationGracePeriodSeconds` ≥ таймаута shutdown.
- Фоновые воркеры останавливай через общий `context` и жди их `WaitGroup`.
- Порядок остановки обратен запуску: сначала принимаем меньше трафика, потом закрываем зависимости.

## Health-эндпоинты

- `/healthz` (liveness) — процесс жив, без проверки зависимостей.
- `/readyz` (readiness) — готов принимать трафик (БД доступна, миграции выполнены).

## Проверь себя

<details><summary>Что произойдёт, если не закрыть <code>resp.Body</code>?</summary>Соединение не вернётся в пул и будет утекать (горутины и файловые дескрипторы); число соединений быстро исчерпается.</details>

<details><summary>Возвращает ли <code>client.Do</code> ошибку для статуса 500?</summary>Нет, ошибка только при сетевых проблемах; статус нужно проверять вручную.</details>

<details><summary>Чем <code>Client.Timeout</code> отличается от таймаута в контексте?</summary>Client.Timeout — общий лимит на весь запрос для всех вызовов клиента; контекст — на конкретный вызов и поддерживает ручную отмену. Часто используют оба.</details>

<details><summary>Что вернёт <code>ListenAndServe</code> после <code>Shutdown</code>?</summary><code>http.ErrServerClosed</code> — штатная ситуация, её нужно отфильтровать.</details>
