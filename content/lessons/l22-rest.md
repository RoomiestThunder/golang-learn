---
id: l22
week: 5
day: 2
minutes: 120
title: REST API: проектирование и реализация
goals: Ресурсы и методы, коды статусов, JSON-ошибки, валидация, пагинация, структура обработчиков
lc: design-circular-queue | Design Circular Queue | Medium | Кольцевой буфер на слайсе: head, count, индексы по модулю
lc: design-linked-list | Design Linked List | Medium | Односвязный список с фиктивной головой; аккуратно с индексами
lc: snapshot-array | Snapshot Array | Medium | Для каждого индекса храни историю (snap_id, value) и ищи бинарным поиском
---

## Принципы REST (практика, а не догма)

- **Ресурсы** — существительные в множественном числе: `/todos`, `/todos/42`, `/users/7/orders`.
- **Методы** выражают действие: `GET` (чтение, безопасный, идемпотентный), `POST` (создание/действие), `PUT` (полная замена, идемпотентный), `PATCH` (частичное изменение), `DELETE` (идемпотентный).
- **Статусы** несут смысл. Клиент должен отличать «ты ошибся» (4xx) от «я сломался» (5xx).
- **Формат** — JSON; `Content-Type: application/json`.
- **Идемпотентность**: повторный `PUT`/`DELETE` не должен менять результат; для `POST` (платежи) вводят `Idempotency-Key`.

| Код | Когда |
|---|---|
| 200 OK | успешное чтение/обновление |
| 201 Created | создан ресурс (+ заголовок `Location`) |
| 204 No Content | успех без тела (DELETE) |
| 400 Bad Request | невалидный JSON/параметры |
| 401 Unauthorized | нет/плохая аутентификация |
| 403 Forbidden | аутентифицирован, но нельзя |
| 404 Not Found | ресурса нет |
| 405 / 415 | метод / тип содержимого не поддерживается |
| 409 Conflict | конфликт состояния (дубликат, версия) |
| 422 Unprocessable Entity | синтаксис верный, но данные не проходят валидацию |
| 429 Too Many Requests | лимит запросов |
| 500 / 502 / 503 / 504 | внутренняя ошибка / плохой шлюз / недоступен / таймаут шлюза |

## Единый формат ошибок

Клиенту проще, когда ошибки всегда одинаковы (или стандарт RFC 9457 `application/problem+json`):

```json
{"error": {"code": "validation_failed", "message": "title не должен быть пустым", "fields": {"title": "required"}}}
```

## Каркас обработчиков

Вся повторяющаяся работа выносится в маленькие функции: чтение JSON, запись JSON, запись ошибки.

```go run
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type store struct {
	mu     sync.RWMutex
	nextID int
	items  map[int]Todo
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) { // после объекта — ничего лишнего
		return errors.New("тело должно содержать один JSON-объект")
	}
	return nil
}

func newAPI() http.Handler {
	s := &store{nextID: 1, items: map[int]Todo{}}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Title string `json:"title"` }
		if err := readJSON(w, r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "некорректный JSON: "+err.Error())
			return
		}
		if strings.TrimSpace(in.Title) == "" {
			writeError(w, http.StatusUnprocessableEntity, "title обязателен")
			return
		}
		s.mu.Lock()
		t := Todo{ID: s.nextID, Title: in.Title}
		s.items[t.ID] = t
		s.nextID++
		s.mu.Unlock()
		w.Header().Set("Location", "/todos/"+strconv.Itoa(t.ID))
		writeJSON(w, http.StatusCreated, t)
	})

	mux.HandleFunc("GET /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "id должен быть числом")
			return
		}
		s.mu.RLock()
		t, ok := s.items[id]
		s.mu.RUnlock()
		if !ok {
			writeError(w, http.StatusNotFound, "не найдено")
			return
		}
		writeJSON(w, http.StatusOK, t)
	})
	return mux
}

func main() {
	srv := httptest.NewServer(newAPI())
	defer srv.Close()

	do := func(method, path, body string) {
		req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("%-5s %-10s %d %s", method, path, resp.StatusCode, b)
	}
	do("POST", "/todos", `{"title":"выучить Go"}`)
	do("POST", "/todos", `{"title":"  "}`)
	do("POST", "/todos", `{"oops":1}`)
	do("GET", "/todos/1", "")
	do("GET", "/todos/99", "")
	do("GET", "/todos/abc", "")
}
```

## Хорошая структура кода

```text
handler (HTTP: разбор запроса, статусы)  →  service (бизнес-логика, правила)  →  repository (хранилище)
```

Обработчик не знает о SQL; сервис не знает о `http.Request`. Соответствие ошибок статусам — в одном месте:

```go
func httpStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):       return 404
	case errors.Is(err, ErrAlreadyExists):  return 409
	case errors.As(err, new(*ValidationError)): return 422
	default:                                return 500
	}
}
```

## Пагинация, фильтры, сортировка

- **Offset/limit**: `GET /todos?limit=20&offset=40` — просто, но медленно на больших смещениях и «дрейфует» при вставках.
- **Cursor (keyset)**: `GET /todos?after=<id>&limit=20` → `WHERE id > $1 ORDER BY id LIMIT 20` — стабильно и быстро. В ответ добавляй `next_cursor`.
- Всегда ограничивай `limit` сверху (например, 100).

## Версионирование и совместимость

`/v1/...` в пути (самое простое) или заголовок. Добавление полей — совместимо, удаление/переименование — нет. Клиенты должны игнорировать неизвестные поля.

## Документация

Описывай API в **OpenAPI** (swagger): из спецификации генерируются клиенты/серверные заглушки (`oapi-codegen`), документация и валидация.

## Безопасность и надёжность API (минимум)

- Ограничения размеров, таймауты, лимиты запросов (429).
- CORS только для нужных источников.
- Не раскрывай внутренние ошибки в ответах; логируй с `request_id`.
- Аутентификация — следующая неделя; всегда HTTPS на периметре.
- Проверяй `Content-Type`, не доверяй данным клиента (в том числе ID владельца — берётся из токена).

## Проверь себя

<details><summary>Чем PUT отличается от PATCH?</summary>PUT заменяет ресурс целиком (идемпотентен), PATCH применяет частичные изменения.</details>

<details><summary>401 или 403?</summary>401 — не удалось определить, кто ты (нет/плохой токен); 403 — личность известна, но прав нет.</details>

<details><summary>Почему keyset-пагинация лучше offset?</summary>Не просматривает пропускаемые строки (быстро на больших страницах) и не «плывёт» при вставке/удалении записей.</details>

<details><summary>Какой статус вернуть при создании ресурса?</summary>201 Created с заголовком Location и телом созданного ресурса.</details>
