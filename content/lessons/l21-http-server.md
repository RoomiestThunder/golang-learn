---
id: l21
week: 5
day: 1
minutes: 120
title: HTTP-сервер: net/http, маршрутизация, middleware
goals: Handler, ServeMux с шаблонами Go 1.22, запросы и ответы, middleware-цепочки, настройка Server
lc: implement-trie-prefix-tree | Implement Trie (Prefix Tree) | Medium | Узел с map[rune]*Node (или [26]*Node) и флагом конца слова; так устроены роутеры
lc: design-add-and-search-words-data-structure | Design Add and Search Words Data Structure | Medium | Trie + DFS для символа '.'
---

## Минимальный сервер

В стандартной библиотеке есть production-готовый HTTP-сервер: HTTP/1.1, HTTP/2 (при TLS), keep-alive, graceful shutdown. Для большинства задач внешний фреймворк не нужен.

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Главный контракт — интерфейс:

```go
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
// http.HandlerFunc — адаптер: функция с нужной сигнатурой становится Handler'ом
```

Каждый запрос обрабатывается **в отдельной горутине** — общее состояние должно быть потокобезопасным.

## Запуск сервера внутри программы (песочница)

`httptest.NewServer` поднимает настоящий сервер на случайном порту — отличный способ экспериментировать и тестировать прямо здесь:

```go run
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	// Go 1.22+: метод и wildcard в шаблоне
	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Привет, %s!", r.PathValue("name"))
	})
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q") // ?q=go
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"query": q, "ua": r.UserAgent()})
	})
	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20)) // ВСЕГДА ограничивай размер тела
		if err != nil {
			http.Error(w, "слишком большое тело", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusCreated) // статус — до первой записи тела!
		w.Write(body)
	})
	mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) { // {path...} — «остаток пути»
		fmt.Fprint(w, "файл: ", r.PathValue("path"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	get := func(method, path, body string) {
		req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s %s → %d %q\n", method, path, resp.StatusCode, b)
	}
	get("GET", "/hello/Gopher", "")
	get("GET", "/search?q=go", "")
	get("POST", "/echo", "данные")
	get("GET", "/files/a/b/c.txt", "")
	get("POST", "/hello/x", "") // 405 Method Not Allowed — мультиплексор сам различает метод
	get("GET", "/nope", "")     // 404
}
```

### Шаблоны маршрутов (Go 1.22)

| Шаблон | Значение |
|---|---|
| `GET /users` | только GET (HEAD разрешён автоматически) |
| `/users/{id}` | сегмент пути в `r.PathValue("id")` |
| `/files/{path...}` | остаток пути |
| `/posts/{$}` | только точное совпадение со слэшем |
| `example.com/api/` | с учётом хоста |

Более специфичный шаблон побеждает; конфликты обнаруживаются при регистрации (паника на старте). Раньше для этого приходилось подключать `gorilla/mux`, `chi` — они по-прежнему популярны (middleware-группы, удобства), но для многих API стандартного мультиплексора хватает.

## Работа с запросом и ответом

```go
func handler(w http.ResponseWriter, r *http.Request) {
	r.Method; r.URL.Path; r.URL.Query(); r.Header.Get("Authorization")
	r.Context()                          // отмена при разрыве соединения
	r.FormValue("name")                  // формы (query + body)
	r.PathValue("id")

	w.Header().Set("Content-Type", "application/json") // заголовки — ДО WriteHeader/Write
	w.WriteHeader(http.StatusOK)         // если не вызвать — будет 200 при первом Write
	w.Write(data)
}
```

Помощники: `http.Error(w, msg, code)`, `http.Redirect`, `http.NotFound`, `http.ServeFile`, `http.FileServer(http.FS(fsys))` (раздача статики, в том числе из `embed.FS`).

> [!WARNING] После `WriteHeader`/первого `Write` заголовки менять поздно. После `http.Error` нужен `return`. Нельзя писать в `w` после возврата из обработчика (и из другой горутины — безопасно только до возврата).

## Middleware

Middleware — функция, оборачивающая `http.Handler` другим `Handler`: логирование, аутентификация, recover, CORS, метрики, таймауты.

```go run
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Chain: первый в списке — самый внешний
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// statusRecorder запоминает код ответа (ResponseWriter его не отдаёт)
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

func Logging(l *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)
			l.Printf("%s %s -> %d", r.Method, r.URL.Path, rec.status)
			_ = start
		})
	}
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "fine") })
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) { panic("ой") })

	logger := log.New(io.Discard, "", 0) // в песочнице тихо; подставь os.Stdout, чтобы видеть
	srv := httptest.NewServer(Chain(mux, Recover, Logging(logger)))
	defer srv.Close()

	for _, p := range []string{"/ok", "/boom", "/missing"} {
		resp, _ := http.Get(srv.URL + p)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(p, resp.StatusCode, string(b))
	}
}
```

Типичный набор: `RequestID` → `Logging` → `Recover` → `CORS` → `Auth` → `RateLimit` → обработчик.

## Настройка http.Server

`http.ListenAndServe` создаёт сервер **без таймаутов** — в production это уязвимость (медленные клиенты исчерпают соединения). Настраивай явно:

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           handler,
	ReadHeaderTimeout: 5 * time.Second,   // защита от Slowloris — задавай ВСЕГДА
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
	MaxHeaderBytes:    1 << 20,
}
log.Fatal(srv.ListenAndServe())
```

Дальше: HTTPS — `ListenAndServeTLS`; HTTP/2 включается автоматически с TLS; профили — `import _ "net/http/pprof"` (**только на внутреннем порту!**).

## Запомни

1. Обработчик получает `r.Context()` — прокидывай его во все вызовы (БД, клиенты).
2. Всегда ограничивай размер тела (`MaxBytesReader`), валидируй вход.
3. Один обработчик — одна ответственность; общие вещи — в middleware.
4. Ошибки клиенту — понятный код и безопасный текст, детали — в лог.
5. Не храни состояние запроса в глобальных переменных.

## Проверь себя

<details><summary>Что такое <code>http.HandlerFunc</code>?</summary>Тип-адаптер: функция с сигнатурой <code>func(w, r)</code>, у которой есть метод <code>ServeHTTP</code>, поэтому она удовлетворяет <code>http.Handler</code>.</details>

<details><summary>Зачем <code>ReadHeaderTimeout</code>?</summary>Чтобы клиент, медленно присылающий заголовки (Slowloris), не удерживал соединение бесконечно.</details>

<details><summary>Можно ли менять заголовки после <code>w.Write</code>?</summary>Нет — заголовки уже отправлены вместе со статусом (по умолчанию 200).</details>

<details><summary>Как в Go 1.22 получить параметр пути?</summary>Шаблон <code>/users/{id}</code> и <code>r.PathValue("id")</code>.</details>
