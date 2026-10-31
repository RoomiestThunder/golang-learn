---
id: l27
week: 6
day: 2
minutes: 120
title: Инфраструктура: Redis, очереди, gRPC, наблюдаемость
goals: Понимать окружение, в котором живут Go-сервисы: кэш, брокеры сообщений, RPC, метрики/логи/трейсы, Kubernetes
lc: find-median-from-data-stream | Find Median from Data Stream | Hard | Две кучи: max-куча для нижней половины и min-куча для верхней (стриминговые метрики)
lc: range-sum-query-immutable | Range Sum Query - Immutable | Easy | Префиксные суммы — основа агрегаций по окнам
---

Этот урок — карта территории. Реализовывать всё сам не нужно, но на собеседовании middle ждут, что ты знаешь **зачем** это нужно, **какие у него гарантии** и **как подключается из Go**.

## Docker и docker-compose

Сервис упаковывают в образ (multi-stage сборка из урока 20), локальное окружение — `docker-compose.yml`:

```yaml
services:
  api:
    build: .
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://app:app@db:5432/app?sslmode=disable
      REDIS_ADDR: redis:6379
    depends_on: { db: { condition: service_healthy }, redis: { condition: service_started } }
  db:
    image: postgres:16
    environment: { POSTGRES_USER: app, POSTGRES_PASSWORD: app, POSTGRES_DB: app }
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U app"], interval: 5s, retries: 10 }
    volumes: [pgdata:/var/lib/postgresql/data]
  redis:
    image: redis:7
volumes: { pgdata: {} }
```

Запуск: `docker compose up --build`. Для CI/тестов — `testcontainers-go` поднимает зависимости из кода теста.

## Redis и кэширование

Redis — in-memory хранилище «ключ → структура» (строки, хэши, списки, множества, sorted sets, streams). Клиент: `github.com/redis/go-redis/v9`.

```go
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
err := rdb.Set(ctx, "user:42", payload, 10*time.Minute).Err() // TTL обязателен для кэша
val, err := rdb.Get(ctx, "user:42").Result()
if errors.Is(err, redis.Nil) { /* промах кэша */ }
```

Паттерны:

- **Cache-aside** (самый частый): читаем из кэша → промах → читаем из БД → кладём в кэш с TTL. Запись: обновили БД → **инвалидировали** ключ.
- **Write-through / write-behind** — запись через кэш (реже, сложнее с консистентностью).
- **Cache stampede**: тысячи запросов одновременно промахиваются по «остывшему» ключу и убивают БД → `singleflight`, случайный джиттер TTL, «мягкий» TTL с фоновым обновлением.
- **Негативное кэширование**: кэшируй и «не найдено» (короткий TTL).
- Другие применения: rate limiting (`INCR` + `EXPIRE`), распределённые блокировки (осторожно: Redlock спорен), очереди/стримы, счётчики, лидерборды (sorted sets), pub/sub.
- Кэш — оптимизация, а не источник истины: сервис должен работать (медленнее) при недоступном кэше.

> [!INTERVIEW] «Две самые сложные проблемы в программировании: инвалидация кэша и именование.» Будь готов обсудить, как не получить устаревшие данные.

## Очереди сообщений

Асинхронное взаимодействие: производитель кладёт сообщение, потребитель обрабатывает позже. Развязывает сервисы, сглаживает пики, позволяет повторы.

| Система | Модель | Когда |
|---|---|---|
| **Kafka** | распределённый лог, партиции, consumer groups, хранение истории, replay | потоки событий, высокая пропускная способность, event sourcing |
| **RabbitMQ** | брокер с очередями/обменниками (AMQP), маршрутизация, ack | задачи, сложная маршрутизация, RPC-паттерны |
| **NATS** | лёгкий pub/sub, JetStream для персистентности | микросервисы, низкая латентность |
| **SQS/PubSub** | управляемые облачные очереди | когда не хочется администрировать |

Ключевые понятия, которые спрашивают:

- **Гарантии доставки**: at-most-once (могут потеряться), **at-least-once** (могут дублироваться — самый частый), exactly-once (на практике = at-least-once + идемпотентный обработчик).
- Поэтому потребитель обязан быть **идемпотентным**: хранит id обработанных сообщений или делает операцию, безопасную к повтору (`INSERT ... ON CONFLICT DO NOTHING`).
- **Ack/nack, retry, dead-letter queue (DLQ)** для «ядовитых» сообщений.
- **Порядок** гарантируется только внутри партиции/очереди; ключ партиционирования = ключ порядка.
- **Backpressure**: ограничивай параллелизм потребителя (worker pool), prefetch.
- **Transactional outbox**: чтобы атомарно «записать в БД и отправить событие», событие пишут в таблицу `outbox` в той же транзакции, а отдельный процесс публикует его в брокер.

Go-клиенты: `segmentio/kafka-go`, `confluent-kafka-go`, `rabbitmq/amqp091-go`, `nats-io/nats.go`.

## gRPC и Protocol Buffers

gRPC — RPC поверх HTTP/2 с контрактом в `.proto`; из него генерируется типизированный код клиента и сервера.

```proto
syntax = "proto3";
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc Watch(WatchRequest) returns (stream Event);   // стриминг
}
message GetUserRequest { int64 id = 1; }
message User { int64 id = 1; string name = 2; }
```

| | REST/JSON | gRPC |
|---|---|---|
| Контракт | OpenAPI (опционально) | `.proto` (обязательно) |
| Формат | текст, больше размер | бинарный protobuf, компактный и быстрый |
| Стриминг | SSE/WebSocket | встроен (серверный, клиентский, двунаправленный) |
| Браузер | нативно | через grpc-web/gateway |
| Подходит | публичные API | межсервисное общение |

Особенности: дедлайны и отмена — через `context`; статусы gRPC (`codes.NotFound`, `codes.Unavailable`...); интерцепторы — аналог middleware; балансировка на клиенте; обратная совместимость — не переиспользуй номера полей.

## Наблюдаемость (observability)

Три столпа:

1. **Логи** — события (структурные JSON, `slog`), корреляция по `request_id`/`trace_id`.
2. **Метрики** — числа во времени (Prometheus): RPS, ошибки, задержки. Методики: **RED** (Rate, Errors, Duration — для сервисов) и **USE** (Utilization, Saturation, Errors — для ресурсов). Типы: counter, gauge, histogram (для перцентилей p95/p99).
3. **Трейсы** — путь запроса через сервисы (OpenTelemetry): span'ы, контекст пробрасывается заголовком `traceparent`.

```go
// Prometheus: github.com/prometheus/client_golang
var reqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_duration_seconds", Buckets: prometheus.DefBuckets,
}, []string{"method", "route", "status"})

http.Handle("/metrics", promhttp.Handler()) // Prometheus сам приходит и забирает метрики (pull)
```

Метки (labels) с высокой кардинальностью (user_id, полный URL) убивают Prometheus — используй шаблон маршрута (`/users/{id}`), а не фактический путь.

Рядом: **алертинг** (по SLO, не по каждому «чиху»), **профилирование в проде** (pprof, continuous profiling — Pyroscope), health-эндпоинты.

## Go в Kubernetes: что знать

- **Probes**: `livenessProbe` (перезапуск), `readinessProbe` (в балансировку), `startupProbe`.
- **Ресурсы**: `requests/limits` CPU и памяти. Go-runtime по умолчанию не знает о cgroup-лимитах CPU (до Go 1.25): ставь `GOMAXPROCS` равным лимиту (`go.uber.org/automaxprocs`; в Go 1.25 учитывается автоматически). Память — `GOMEMLIMIT` чуть ниже лимита контейнера (≈ 90%), чтобы GC успевал до OOM-kill.
- **Graceful shutdown**: на SIGTERM — завершить запросы; `preStop` задержка, `terminationGracePeriodSeconds`.
- **Конфигурация**: ConfigMap/Secret → переменные окружения.
- **Стратегии выкладки**: rolling update, canary, blue-green; миграции БД — совместимые с двумя версиями кода.
- Горизонтальное масштабирование (HPA) требует stateless сервиса: состояние — в БД/кэше, не в памяти процесса.

## CI/CD

Минимальный конвейер: `go vet` + линтер → `go test -race ./...` → сборка образа → публикация → деплой (GitOps: Argo CD/Flux). Кэшируй модули и build cache. Секреты — из хранилища CI.

## Проверь себя

<details><summary>Что такое cache stampede и как с ним бороться?</summary>Массовый одновременный промах по одному ключу, перегружающий источник; лечится singleflight, джиттером TTL, фоновым обновлением.</details>

<details><summary>Почему потребитель очереди должен быть идемпотентным?</summary>Доставка at-least-once допускает повторы сообщений; повторная обработка не должна менять результат.</details>

<details><summary>Чем RED отличается от USE?</summary>RED — метрики запросов сервиса (частота, ошибки, длительность); USE — метрики ресурса (загрузка, насыщение, ошибки).</details>

<details><summary>Зачем нужен transactional outbox?</summary>Чтобы запись в БД и публикация события были атомарны: событие пишется в БД в той же транзакции, а публикуется отдельным процессом.</details>
