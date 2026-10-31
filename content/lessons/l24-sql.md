---
id: l24
week: 5
day: 4
minutes: 120
title: Базы данных: database/sql, транзакции, миграции
goals: Работа с PostgreSQL из Go: пул, запросы, NULL, транзакции, защита от инъекций, паттерн репозиторий
lc: second-highest-salary | Second Highest Salary | Medium | SQL (PostgreSQL/MySQL): LIMIT/OFFSET или подзапрос с MAX; не забудь про NULL
lc: employees-earning-more-than-their-managers | Employees Earning More Than Their Managers | Easy | SQL: self join таблицы Employee
lc: duplicate-emails | Duplicate Emails | Easy | SQL: GROUP BY email HAVING COUNT(*) > 1
lc: rank-scores | Rank Scores | Medium | SQL: оконная функция DENSE_RANK() OVER (ORDER BY score DESC)
---

> [!NOTE] Задачи LeetCode этого урока — **на SQL** (раздел Database на leetcode.com, выбирай PostgreSQL). Для бэкенд-разработчика уровня middle SQL — обязательный навык, поэтому он в программе.

## database/sql: общая картина

`database/sql` — **единый интерфейс** к реляционным БД. Драйвер подключается импортом «ради побочного эффекта»:

```go
import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер PostgreSQL (регистрирует себя)
)

db, err := sql.Open("pgx", "postgres://user:pass@localhost:5432/app?sslmode=disable")
if err != nil { ... }              // Open НЕ подключается, лишь валидирует параметры
defer db.Close()
if err := db.PingContext(ctx); err != nil { ... } // реальная проверка соединения
```

`*sql.DB` — это **пул соединений**, потокобезопасный; создаётся **один на приложение**:

```go
db.SetMaxOpenConns(25)                  // верхняя граница соединений (по умолчанию — без лимита!)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(30 * time.Minute) // периодически пересоздавать (балансировщики, failover)
db.SetConnMaxIdleTime(5 * time.Minute)
```

Для PostgreSQL часто используют `pgx` напрямую (`pgxpool`): быстрее, больше возможностей (COPY, LISTEN/NOTIFY, нативные типы).

## Запросы

```go
// Exec — INSERT/UPDATE/DELETE без строк результата
res, err := db.ExecContext(ctx, `UPDATE users SET name = $1 WHERE id = $2`, name, id)
n, _ := res.RowsAffected()

// QueryRow — ровно одна строка
var u User
err = db.QueryRowContext(ctx, `SELECT id, name, email FROM users WHERE id = $1`, id).
	Scan(&u.ID, &u.Name, &u.Email)
if errors.Is(err, sql.ErrNoRows) { return User{}, ErrNotFound } // «нет строки» — ожидаемый случай

// Query — много строк
rows, err := db.QueryContext(ctx, `SELECT id, name FROM users WHERE active = $1 ORDER BY id LIMIT $2`, true, 100)
if err != nil { return nil, err }
defer rows.Close()                    // обязательно: иначе соединение «зависнет» в пуле
var out []User
for rows.Next() {
	var u User
	if err := rows.Scan(&u.ID, &u.Name); err != nil { return nil, err }
	out = append(out, u)
}
if err := rows.Err(); err != nil { return nil, err } // ошибка во время итерации
```

Всегда используй варианты с `Context`.

### ⚠️ SQL-инъекции

**Никогда** не собирай SQL конкатенацией/`Sprintf` из пользовательских данных. Значения передавай **параметрами** (`$1, $2` в PostgreSQL, `?` в MySQL/SQLite) — драйвер/БД экранируют их сами.

```go
// УЯЗВИМО: name = "x'; DROP TABLE users; --"
db.Query("SELECT * FROM users WHERE name = '" + name + "'")

// ВЕРНО
db.Query("SELECT * FROM users WHERE name = $1", name)
```

Параметры нельзя подставлять вместо **идентификаторов** (имён таблиц/колонок, `ORDER BY`) — их проверяют по белому списку.

## NULL

```go
var middle sql.NullString        // {String string; Valid bool}
var age sql.NullInt64
var deleted *time.Time           // или указатель: nil == NULL
err := row.Scan(&middle, &age, &deleted)
if middle.Valid { ... }
```

Альтернативы: `COALESCE` в SQL, проектирование схемы с `NOT NULL DEFAULT`.

## Транзакции

Нескольких операций, которые должны выполниться **все или ни одной** (ACID).

```go
func (s *Store) Transfer(ctx context.Context, from, to int, amount int64) (err error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil { return err }
	defer func() {
		if err != nil {
			_ = tx.Rollback()           // откат при любой ошибке/панике (после Commit безвреден)
		}
	}()

	res, err := tx.ExecContext(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1`, amount, from)
	if err != nil { return err }
	if n, _ := res.RowsAffected(); n == 0 { return ErrInsufficientFunds }

	if _, err = tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, to); err != nil {
		return err
	}
	return tx.Commit()
}
```

Все запросы внутри — через `tx`, а не `db`. Держи транзакции **короткими** (без сетевых вызовов внутри). Паттерн «unit of work» — функция `WithTx(ctx, func(tx) error) error`, которая сама делает Begin/Commit/Rollback — так тестируется и не забывается (ты напишешь его в задании).

Для конкурентных изменений: атомарные `UPDATE ... SET x = x + 1`, `SELECT ... FOR UPDATE` (блокировка строк), оптимистичные блокировки (колонка `version`), уровни изоляции (Read Committed по умолчанию в PostgreSQL; Serializable требует повторов при `40001`).

## Подготовленные запросы

`db.PrepareContext` компилирует запрос один раз для многократного выполнения; pgx делает это автоматически (statement cache). Не злоупотребляй ручным Prepare без необходимости.

## Миграции

Схему меняют **версионированными миграциями** в репозитории, а не руками на сервере: `golang-migrate`, `goose`, `atlas`. Файлы вида `0001_create_users.up.sql / .down.sql`, применяются при деплое отдельным шагом. Правила: миграции неизменяемы после выкатки; обратно совместимые изменения (добавить колонку → задеплоить код → потом удалять старое — «expand/contract»); долгие `ALTER` на больших таблицах — осторожно.

## Индексы и производительность (минимум для middle)

- Индекс — на колонки из `WHERE`/`JOIN`/`ORDER BY`; составной индекс — по порядку выборочности; `EXPLAIN (ANALYZE)` показывает реальный план.
- **N+1**: цикл с запросом на каждый элемент → замени на `WHERE id = ANY($1)` / `JOIN`.
- Не `SELECT *`; пагинация keyset; `LIMIT` всегда.
- `ON CONFLICT DO NOTHING/UPDATE` для идемпотентных вставок (upsert).

## Инструменты доступа к данным

| Подход | Плюсы | Минусы |
|---|---|---|
| `database/sql` + `pgx` | контроль, стандарт | много рутины со `Scan` |
| `sqlx` | `StructScan`, меньше кода | всё ещё ручной SQL |
| **`sqlc`** | пишешь SQL → генерирует типобезопасный Go-код | шаг генерации |
| ORM (`gorm`, `ent`) | быстрый старт, связи | скрытые запросы, сложная отладка |

Популярный выбор в продакшене: **pgx + sqlc + goose/migrate**.

## Тестирование кода с БД

Юнит-тесты — на фейках интерфейса репозитория; интеграционные — на реальной БД в контейнере (`testcontainers-go`) с миграциями; `go-sqlmock` — для проверки текста запросов (хрупко). 

## Паттерн Repository

```go
type UserRepository interface {                 // объявляется там, где используется (service)
	Get(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, u *User) error
}
// реализация: postgres.UserRepo{db *sql.DB}; для тестов: memory.UserRepo
```

Ошибки БД транслируй в доменные (`sql.ErrNoRows` → `ErrNotFound`; нарушение уникальности `23505` → `ErrAlreadyExists`) — слой сервиса не должен знать про `database/sql`.

## Проверь себя

<details><summary>Что делает <code>sql.Open</code>?</summary>Создаёт пул (ленивый), проверяет параметры; подключение происходит при первом запросе, проверить сразу — <code>PingContext</code>.</details>

<details><summary>Почему нужно закрывать <code>rows</code> и проверять <code>rows.Err()</code>?</summary>Незакрытые rows удерживают соединение; ошибка итерации (обрыв сети) не видна в <code>Next()</code> — только через <code>Err()</code>.</details>

<details><summary>Как защититься от SQL-инъекций?</summary>Параметризованные запросы (плейсхолдеры), белые списки для идентификаторов, права БД по минимуму.</details>

<details><summary>Что такое N+1 и как его избежать?</summary>Один запрос списка и N запросов за связанными данными; решается JOIN или батчингом (<code>= ANY($1)</code>).</details>
