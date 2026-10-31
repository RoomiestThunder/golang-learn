---
id: l39
week: 8
day: 4
minutes: 150
title: Итоговый проект, часть 1: сокращатель ссылок — ядро
goals: Спроектировать сервис, написать кодировщик base62, хранилище и HTTP-слой по тестам
lc: encode-and-decode-tinyurl | Encode and Decode TinyURL | Medium | Это и есть ядро твоего проекта: решай своим кодировщиком
lc: longest-palindromic-substring | Longest Palindromic Substring | Medium | Расширение от центра за O(n²)
lc: decode-ways | Decode Ways | Medium | Одномерное ДП по префиксам
---

Два последних дня курса — **единый проект**, в котором собирается всё изученное: HTTP, JSON, конкурентность, интерфейсы, тесты, конфигурация, graceful shutdown. Цель — получить **репозиторий, который не стыдно показать** на собеседовании.

## Что строим: `shrink` — сервис сокращения ссылок

| Функция | Как |
|---|---|
| Сократить | `POST /shorten` `{"url":"https://example.com/very/long"}` → `201` `{"code":"1Ab","short_url":"http://localhost:8080/1Ab"}` |
| Перейти | `GET /{code}` → `302 Found` + `Location` |
| Статистика | `GET /stats/{code}` → `{"code":"1Ab","url":"...","clicks":42}` |
| Здоровье | `GET /healthz` → `200 ok` |
| Дедупликация | один и тот же URL → один и тот же код |
| Срок жизни (день 5) | ссылка может иметь TTL |
| Защита (день 5) | rate limit по IP, ограничение размера тела, таймауты сервера |

## Архитектура

```text
cmd/shrink/main.go           ← конфиг (env), сборка зависимостей, graceful shutdown
internal/
  base62/        Encode/Decode          ← задание 1
  shortener/     Service + Store        ← задание 2 (бизнес-логика + хранилище)
  httpapi/       handlers, middleware   ← задание 3
  ratelimit/     token bucket           ← из урока 36
```

Правила (из недели 5): сервис зависит от **интерфейса** хранилища; HTTP-слой — от интерфейса сервиса; всё собирается в `main`; время и ID — внедряемые зависимости; ошибки доменные (`ErrNotFound`, `ErrInvalidURL`) и мапятся на статусы в одном месте.

## Дизайн-решения (и почему)

- **Код = base62(счётчик)**: уникальность без коллизий и без проверок. Недостаток — перечислимость: чтобы не было очевидных `1,2,3`, счётчик стартует с большого смещения (например 125 → код из 2 символов; или 62³ → 4 символа). Для скрытности можно перемешать алфавит.
- **Дедупликация** через `map[url]code`: повторная отправка не раздувает базу.
- **302, а не 301**: 301 кэшируется браузером навсегда — не сможешь считать клики и менять цель.
- **Валидация URL**: схема `http`/`https`, непустой хост, длина ≤ 2048; запрет `localhost`/приватных адресов (SSRF) — по желанию.
- **Потокобезопасность**: каждый запрос в своей горутине — хранилище под `sync.RWMutex`; счётчик кликов — `atomic` либо под тем же локом.

## План работы (2 + 2 часа)

1. **`l39-base62`** — кодировщик (быстрая разминка, 20 мин).
2. **`l39-store`** — хранилище + сервис с дедупликацией, кликами и валидацией (60 мин).
3. **`l39-http`** — HTTP-обработчики на `ServeMux`, тесты с `httptest` (60 мин).
4. Создай настоящий репозиторий и скопируй свои решения в структуру выше (+30 мин), добавь `main.go`.

## Каркас `main.go`

Так будет выглядеть сборка (после выполнения заданий можешь использовать свои реализации):

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/shrink/internal/httpapi"
	"example.com/shrink/internal/shortener"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	addr := getenv("ADDR", ":8080")
	baseURL := getenv("BASE_URL", "http://localhost:8080")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := shortener.NewMemoryStore()
	svc := shortener.NewService(store)
	handler := httpapi.New(svc, baseURL, logger)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("слушаю", "addr", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("останавливаюсь")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func getenv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}
```

## Критерии «готово»

- `go vet ./...`, `go test -race ./...` — зелёные.
- Покрытие ключевой логики тестами (кодек, дедупликация, 404, валидация).
- README: что это, как запустить, примеры `curl`, дизайн-решения, что можно улучшить.
- Dockerfile (multi-stage) — урок 20; `docker run -p 8080:8080 shrink` работает.

## Проверь себя

<details><summary>Почему используем 302, а не 301?</summary>301 браузеры кэшируют бессрочно: клики не попадут на сервер, и ссылку нельзя изменить/удалить. 302 — временный редирект.</details>

<details><summary>Как кодировать число в base62?</summary>Деление на 62 с остатком, остатки — индексы в алфавите; результат собирается в обратном порядке (старшая цифра первой).</details>

<details><summary>Какие риски у сокращателя ссылок?</summary>Фишинг и вредоносные цели, SSRF при проверке ссылок, перебор кодов, злоупотребление (нужен rate limit), утечка приватных ссылок.</details>
