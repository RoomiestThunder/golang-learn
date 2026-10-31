---
id: l17
week: 4
day: 2
minutes: 120
title: encoding/json и time
goals: Сериализация JSON, теги, свои Marshal/Unmarshal, потоки; корректная работа со временем и таймерами
lc: day-of-the-week | Day of the Week | Easy | Можно решить через пакет time — потренируй API
lc: day-of-the-year | Day of the Year | Easy | Номер дня в году: накопительные дни месяцев, не забудь про високосный год
lc: number-of-days-between-two-dates | Number of Days Between Two Dates | Easy | Преобразуй обе даты в «дни от эпохи» и вычти
---

## encoding/json

```go run
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Address struct {
	City string `json:"city"`
}

type User struct {
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Email    string            `json:"email,omitempty"` // пропустить, если пусто
	Password string            `json:"-"`               // никогда не сериализуется
	Age      *int              `json:"age"`             // указатель: различаем «нет» (null) и 0
	Tags     []string          `json:"tags"`
	Meta     map[string]any    `json:"meta,omitempty"`
	Addr     Address           `json:"address"`
	internal string            // неэкспортируемые поля игнорируются
}

func main() {
	u := User{ID: 1, Name: "Ann", Password: "secret", Tags: []string{"go"}, Addr: Address{"Рим"}}

	b, err := json.Marshal(u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	pretty, _ := json.MarshalIndent(map[string]any{"ok": true, "n": 1.5}, "", "  ")
	fmt.Println(string(pretty))

	// Unmarshal: в структуру (поля сопоставляются по тегу, регистронезависимо)
	var got User
	in := `{"id": 7, "name": "Bob", "age": 30, "tags": ["a","b"], "address": {"city": "Киев"}, "extra": 1}`
	if err := json.Unmarshal([]byte(in), &got); err != nil { // обязательно указатель!
		log.Fatal(err)
	}
	fmt.Println(got.ID, got.Name, *got.Age, got.Tags, got.Addr.City) // неизвестное поле "extra" проигнорировано

	// неизвестная структура: map[string]any, числа становятся float64!
	var anyv map[string]any
	json.Unmarshal([]byte(`{"a": 1, "b": [true, null], "c": {"d": "x"}}`), &anyv)
	fmt.Printf("%v %T\n", anyv["a"], anyv["a"])

	// ошибки синтаксиса и типов
	err = json.Unmarshal([]byte(`{"id": "not a number"}`), &got)
	fmt.Println(err)

	// строгий разбор потока: Decoder
	dec := json.NewDecoder(strings.NewReader(`{"id":1,"nope":2}`))
	dec.DisallowUnknownFields()
	fmt.Println(dec.Decode(&got))
}
```

Что важно:

- Поля должны быть **экспортируемыми**, иначе `json` их не видит.
- `omitempty` пропускает нулевые значения (0, "", nil, пустые слайс/map); для структур не работает (в Go 1.24 есть `omitzero`).
- `nil`-слайс → `null`, пустой `[]string{}` → `[]`. Для API-ответов лучше инициализировать, чтобы отдавать `[]`.
- Числа в `any` превращаются в `float64`: большие `int64` (ID) теряют точность. Решения: `Decoder.UseNumber()`, `json.Number` или типизированная структура.
- `json.RawMessage` откладывает разбор части документа (полиморфные поля).
- Для потоков и HTTP: `json.NewEncoder(w).Encode(v)` и `json.NewDecoder(r.Body).Decode(&v)` — без промежуточных `[]byte`.

### Собственная (де)сериализация

```go run
package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Level int

const (
	Debug Level = iota
	Info
)

var names = []string{"debug", "info"}

// MarshalText/UnmarshalText (encoding.TextMarshaler) работают и для JSON-значений, и для ключей map
func (l Level) MarshalText() ([]byte, error) { return []byte(names[l]), nil }

func (l *Level) UnmarshalText(b []byte) error {
	for i, n := range names {
		if strings.EqualFold(n, string(b)) {
			*l = Level(i)
			return nil
		}
	}
	return fmt.Errorf("неизвестный уровень %q", b)
}

type Entry struct {
	Level Level  `json:"level"`
	Msg   string `json:"msg"`
}

func main() {
	b, _ := json.Marshal(Entry{Info, "старт"})
	fmt.Println(string(b))

	var e Entry
	fmt.Println(json.Unmarshal([]byte(`{"level":"DEBUG","msg":"x"}`), &e), e)
	fmt.Println(json.Unmarshal([]byte(`{"level":"fatal"}`), &e))
}
```

Есть и `json.Marshaler`/`json.Unmarshaler` (`MarshalJSON() ([]byte, error)`) для полного контроля.

> [!NOTE] Стандартный `encoding/json` медлителен из-за рефлексии. Для горячих путей применяют кодогенерацию/альтернативы (easyjson, sonic, jsoniter); в Go 1.25 доступен экспериментальный `encoding/json/v2` (`GOEXPERIMENT=jsonv2`).

## Пакет time

### Основные типы

- `time.Time` — момент времени (с локацией и монотонными часами); **передаётся по значению**, сравнивай `t1.Equal(t2)` / `Before` / `After`, а не `==`.
- `time.Duration` — `int64` наносекунд: `500*time.Millisecond`, `d.Seconds()`, `d.String()`.
- `time.Location` — часовой пояс.

```go run
package main

import (
	"fmt"
	"time"
)

func main() {
	// Формат задаётся примером «Mon Jan 2 15:04:05 MST 2006» (1 2 3 4 5 6 7) — запомни эту дату
	t := time.Date(2025, time.March, 9, 14, 30, 5, 0, time.UTC)
	fmt.Println(t.Format("2006-01-02 15:04:05"), t.Format(time.RFC3339), t.Format("02.01.06 3:04PM"))
	fmt.Println(t.Weekday(), t.YearDay(), t.Unix(), t.Month() == time.March)

	p, err := time.Parse("2006-01-02", "2024-02-29")
	fmt.Println(p, err)
	_, err = time.Parse("2006-01-02", "2024-02-30")
	fmt.Println(err)

	// арифметика
	week := t.Add(7 * 24 * time.Hour)
	fmt.Println(week.Sub(t), week.After(t))
	fmt.Println(t.AddDate(0, 1, 25).Format("2006-01-02")) // календарная арифметика: +1 мес +25 дней
	fmt.Println(t.Truncate(time.Hour).Format(time.TimeOnly), time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Format(time.DateOnly))

	// часовые пояса
	loc, err := time.LoadLocation("Europe/Moscow")
	if err == nil {
		fmt.Println(t.In(loc).Format(time.RFC1123))
	}

	d, _ := time.ParseDuration("1h15m30s")
	fmt.Println(d, d.Minutes(), d.Round(time.Hour), 90*time.Second)
}
```

Правила работы со временем:

1. Храни и передавай время в **UTC** (`time.Now().UTC()`), переводи в локаль только для показа.
2. Для измерения интервалов используй `time.Since(start)` — монотонные часы защищают от перевода системного времени.
3. Не сравнивай `Time` через `==` (учитывается локация и монотонная часть).
4. `AddDate` работает по календарю (`31 января + 1 месяц` → 3 марта: Go нормализует переполнение).
5. Для тестируемости **не вызывай `time.Now()` глубоко в коде** — передавай «часы» (`type Clock interface{ Now() time.Time }`) или функцию `now func() time.Time`. Для деньговых/юридических расчётов используй календарные даты, а не «+24 часа».
6. Парсинг дат из внешних систем: `time.Parse` + явная локация (`time.ParseInLocation`).

### Таймеры и тикеры

```go
timer := time.NewTimer(2 * time.Second) // сработает один раз
defer timer.Stop()
select {
case <-timer.C:
case <-ctx.Done():
}

ticker := time.NewTicker(time.Second) // периодически
defer ticker.Stop() // иначе утечёт
for range ticker.C { ... }

time.AfterFunc(time.Minute, func() { ... }) // вызов функции по таймеру
```

Начиная с Go 1.23 неостановленные таймеры и тикеры без ссылок собираются GC, но `Stop()` по-прежнему хороший тон.

## Проверь себя

<details><summary>Почему <code>float64</code> при разборе JSON в <code>any</code> — проблема?</summary>Все числа становятся float64; целые больше 2^53 (например, ID) теряют точность. Используй типизированные структуры или <code>UseNumber</code>.</details>

<details><summary>Как отличить в JSON отсутствие поля от нулевого значения?</summary>Сделать поле указателем (<code>*int</code>): nil — поля нет/null, иначе значение (даже 0).</details>

<details><summary>Как записывается шаблон формата времени в Go?</summary>Через эталонную дату 2006-01-02 15:04:05 (Mon Jan 2 15:04:05 MST 2006), а не через %Y-%m-%d.</details>

<details><summary>Почему не стоит сравнивать <code>time.Time</code> через ==?</summary>Структура включает локацию и монотонное время; одинаковые моменты могут оказаться неравными. Используй <code>Equal</code>.</details>
