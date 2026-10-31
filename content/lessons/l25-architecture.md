---
id: l25
week: 5
day: 5
minutes: 120
title: Архитектура сервиса: слои, DI, конфигурация
goals: Собрать сервис из слоёв с внедрением зависимостей, конфигурацией и тестируемостью; не переусложнить
lc: design-authentication-manager | Design Authentication Manager | Medium | Map токен → время истечения; renew только для неистёкших
lc: design-parking-system | Design Parking System | Easy | Простой счётчик по типам мест — разминка на структуры
lc: design-food-rating-system | Design a Food Rating System | Medium | Map + куча/сортированная структура на кухню (ленивое удаление)
---

## Зачем думать об архитектуре

Цель — чтобы код можно было **менять и тестировать** без каскада правок. В Go это достигается не фреймворками, а простыми приёмами: маленькие пакеты, интерфейсы у потребителя, явные зависимости.

## Слои и направление зависимостей

```text
cmd/api/main.go          ← composition root: читает конфиг, создаёт и «склеивает» всё
   │
transport/http           ← handler: разбор запроса, вызов сервиса, формирование ответа
   │  (зависит от)
service                  ← бизнес-логика, правила, транзакционные границы
   │  (зависит от интерфейса)
repository (интерфейс)   ← объявлен в service ──► реализован в storage/postgres
```

Правила: домен/сервис **не импортируют** `net/http` и `database/sql`; детали (HTTP, SQL, Kafka) — на краях; зависимости направлены к центру. Это «чистая/гексагональная» архитектура в упрощённом виде.

> [!WARNING] Не строй пять слоёв для CRUD на три таблицы. Начинай с простого (handler → store) и вводи слой, когда появляется реальная причина: сложные правила, несколько источников данных, потребность в тестах без БД.

## Внедрение зависимостей (DI) — просто конструкторами

Никакого контейнера не нужно: объект получает всё необходимое **в конструкторе** и хранит интерфейсы.

```go run
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEmailTaken = errors.New("email уже занят")
	ErrInvalid    = errors.New("некорректные данные")
)

type User struct {
	ID        int
	Email     string
	PassHash  string
	CreatedAt time.Time
}

// Интерфейсы объявлены ЗДЕСЬ, у потребителя — минимальные, под нужды сервиса
type UserRepo interface {
	ByEmail(ctx context.Context, email string) (User, bool)
	Create(ctx context.Context, u *User) error
}
type Hasher interface{ Hash(pw string) string }
type Clock interface{ Now() time.Time }

type UserService struct {
	repo   UserRepo
	hasher Hasher
	clock  Clock
}

func NewUserService(r UserRepo, h Hasher, c Clock) *UserService {
	return &UserService{repo: r, hasher: h, clock: c}
}

func (s *UserService) Register(ctx context.Context, email, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") || len(password) < 8 {
		return User{}, ErrInvalid
	}
	if _, exists := s.repo.ByEmail(ctx, email); exists {
		return User{}, fmt.Errorf("регистрация %s: %w", email, ErrEmailTaken)
	}
	u := User{Email: email, PassHash: s.hasher.Hash(password), CreatedAt: s.clock.Now()}
	if err := s.repo.Create(ctx, &u); err != nil {
		return User{}, fmt.Errorf("сохранение пользователя: %w", err)
	}
	return u, nil
}

// ---- реализации для примера / тестов ----
type memRepo struct{ users map[string]User }

func (m *memRepo) ByEmail(_ context.Context, e string) (User, bool) { u, ok := m.users[e]; return u, ok }
func (m *memRepo) Create(_ context.Context, u *User) error {
	u.ID = len(m.users) + 1
	m.users[u.Email] = *u
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(pw string) string { return "hash(" + pw + ")" }

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func main() {
	svc := NewUserService(
		&memRepo{users: map[string]User{}},
		fakeHasher{},
		fixedClock{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	ctx := context.Background()
	u, err := svc.Register(ctx, " Ann@Example.com ", "longpassword")
	fmt.Println(u, err)
	_, err = svc.Register(ctx, "ann@example.com", "another-pass")
	fmt.Println(errors.Is(err, ErrEmailTaken), err)
	_, err = svc.Register(ctx, "bad", "x")
	fmt.Println(err)
}
```

Выгоды очевидны: сервис тестируется без БД, времени и криптографии. Тот же приём работает для почты, платёжных шлюзов, очередей.

Если зависимостей десятки, подойдут генераторы (`google/wire`) или контейнеры (`uber/fx`), но ручная сборка в `main` — самый прозрачный вариант.

## Конфигурация (12-factor)

Конфигурация — из **окружения**, с проверкой и значениями по умолчанию, в одной структуре:

```go
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    slog.Level
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		ShutdownTimeout: 15 * time.Second,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL обязателен")
	}
	return cfg, nil
}
```

Секреты (пароли, ключи) — не в репозитории: переменные окружения, секреты Kubernetes, Vault. Конфигурацию валидируй **при старте** и падай сразу («fail fast»).

## Composition root

```go
func main() {
	if err := run(); err != nil {
		slog.Error("остановка с ошибкой", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	...
	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	repo := postgres.NewUserRepo(db)
	svc := service.NewUserService(repo, bcryptHasher{}, realClock{})
	h := httpapi.New(svc, logger)
	return httpapi.Serve(ctx, cfg.HTTPAddr, h) // graceful shutdown внутри
}
```

Вся «склейка» в одном месте, `main` остаётся тонким.

## Сквозные вопросы

- **Ошибки**: сервис возвращает доменные ошибки; транспорт переводит их в статусы; БД-ошибки транслирует репозиторий.
- **Контекст**: первым параметром во всех методах слоёв.
- **Логирование**: логгер как зависимость; один лог на ошибку, на верхнем уровне; `request_id`.
- **Транзакции**: границы — в сервисе (`repo.WithTx`), чтобы бизнес-операция была атомарной.
- **DTO vs доменные модели**: структуры запросов/ответов API отделяй от доменных (поля, теги JSON, безопасность — не отдавай `PassHash`).
- **Пакеты по фичам**: для больших систем удобнее группировать по предметной области (`user/`, `order/`), а не по техническим слоям.
- **Наблюдаемость** (метрики, трейсы) — в middleware и обёртках-декораторах интерфейсов.

## Антипаттерны

- «Интерфейс на каждую структуру» заранее, моки на всё → хрупкие тесты.
- Пакеты `utils`, `common`, `models` на весь проект.
- Глобальные переменные (`var DB *sql.DB`) вместо зависимостей.
- Логика в обработчиках; SQL в сервисах; `http.Request` в домене.
- Преждевременные микросервисы: начни с **модульного монолита**.

## Проверь себя

<details><summary>Где объявлять интерфейс репозитория — рядом с реализацией или у потребителя?</summary>У потребителя (в пакете сервиса): интерфейс описывает ровно то, что нужно сервису, и реализации не зависят от него.</details>

<details><summary>Зачем внедрять часы (<code>Clock</code>) вместо <code>time.Now()</code>?</summary>Чтобы тесты были детерминированными: подставляется фиксированное или управляемое время.</details>

<details><summary>Что такое composition root?</summary>Единственное место (обычно <code>main</code>), где создаются конкретные реализации и связываются в граф зависимостей.</details>

<details><summary>Почему конфиг валидируют при старте?</summary>«Fail fast»: ошибка настройки обнаруживается сразу при деплое, а не через часы под нагрузкой.</details>
