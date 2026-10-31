---
title: REST API: список дел
level: hard
---
@@ statement
Реализуй `NewTodoAPI() http.Handler` — потокобезопасный REST API в памяти. Все ответы — JSON (`Content-Type: application/json`), ошибки — `{"error":"<текст>"}`.

Ресурс `Todo`: `{"id": 1, "title": "...", "done": false}`.

| Запрос | Успех | Ошибки |
|---|---|---|
| `POST /todos` тело `{"title":"..."}` | `201`, заголовок `Location: /todos/<id>`, тело — созданный Todo (id с 1, по порядку) | `400` — невалидный JSON / неизвестные поля; `422` — пустой title (после `TrimSpace`) |
| `GET /todos` | `200`, массив по возрастанию id (пустой — `[]`, не `null`); фильтр `?done=true/false` | `400` — если `done` не булево |
| `GET /todos/{id}` | `200`, Todo | `400` — id не число; `404` — нет такого |
| `PATCH /todos/{id}` тело `{"title":"...","done":true}` — поля необязательные | `200`, обновлённый Todo | `400`, `404`; `422` — если передан пустой title |
| `DELETE /todos/{id}` | `204` без тела | `404` |

Подсказки: поля PATCH — указатели (`*string`, `*bool`), чтобы отличать «не передано» от «нулевое значение»; `DisallowUnknownFields`; `sync.Mutex`.
@@ starter
package solution

import "net/http"

func NewTodoAPI() http.Handler {
	return http.NewServeMux()
}
@@ tests
package solution

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type item struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func call(t *testing.T, h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeTodo(t *testing.T, rec *httptest.ResponseRecorder) item {
	t.Helper()
	var td item
	if err := json.Unmarshal(rec.Body.Bytes(), &td); err != nil {
		t.Fatalf("тело %q не JSON Todo: %v", rec.Body.String(), err)
	}
	return td
}

func TestCreateAndGet(t *testing.T) {
	h := NewTodoAPI()
	rec := call(t, h, "POST", "/todos", `{"title":"выучить Go"}`)
	if rec.Code != 201 || rec.Header().Get("Location") != "/todos/1" {
		t.Fatalf("POST: %d, Location=%q, тело %s", rec.Code, rec.Header().Get("Location"), rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	if td := decodeTodo(t, rec); td != (item{1, "выучить Go", false}) {
		t.Errorf("создано %+v", td)
	}
	if td := decodeTodo(t, call(t, h, "GET", "/todos/1", "")); td.Title != "выучить Go" {
		t.Errorf("GET: %+v", td)
	}
	if call(t, h, "GET", "/todos/2", "").Code != 404 {
		t.Error("несуществующий → 404")
	}
	if call(t, h, "GET", "/todos/abc", "").Code != 400 {
		t.Error("нечисловой id → 400")
	}
}

func TestValidation(t *testing.T) {
	h := NewTodoAPI()
	tests := []struct {
		body string
		code int
	}{
		{`{"title":"   "}`, 422}, {`{}`, 422}, {`{"title":"x","oops":1}`, 400}, {`not json`, 400}, {``, 400},
	}
	for _, tt := range tests {
		rec := call(t, h, "POST", "/todos", tt.body)
		if rec.Code != tt.code {
			t.Errorf("POST %q: код %d, want %d", tt.body, rec.Code, tt.code)
		}
		if !strings.Contains(rec.Body.String(), `"error"`) {
			t.Errorf("POST %q: нет поля error в %q", tt.body, rec.Body.String())
		}
	}
}

func TestListAndFilter(t *testing.T) {
	h := NewTodoAPI()
	if body := strings.TrimSpace(call(t, h, "GET", "/todos", "").Body.String()); body != "[]" {
		t.Fatalf("пустой список должен быть [], а не %q", body)
	}
	for _, title := range []string{"a", "b", "c"} {
		call(t, h, "POST", "/todos", `{"title":"`+title+`"}`)
	}
	call(t, h, "PATCH", "/todos/2", `{"done":true}`)

	var all, done, open []item
	json.Unmarshal(call(t, h, "GET", "/todos", "").Body.Bytes(), &all)
	json.Unmarshal(call(t, h, "GET", "/todos?done=true", "").Body.Bytes(), &done)
	json.Unmarshal(call(t, h, "GET", "/todos?done=false", "").Body.Bytes(), &open)
	if len(all) != 3 || all[0].ID != 1 || all[2].ID != 3 {
		t.Errorf("all = %+v", all)
	}
	if len(done) != 1 || done[0].ID != 2 || len(open) != 2 {
		t.Errorf("done=%+v open=%+v", done, open)
	}
	if call(t, h, "GET", "/todos?done=maybe", "").Code != 400 {
		t.Error("плохой фильтр → 400")
	}
}

func TestPatchDelete(t *testing.T) {
	h := NewTodoAPI()
	call(t, h, "POST", "/todos", `{"title":"old"}`)

	rec := call(t, h, "PATCH", "/todos/1", `{"title":"new"}`)
	if td := decodeTodo(t, rec); rec.Code != 200 || td != (item{1, "new", false}) {
		t.Errorf("PATCH title: %d %+v", rec.Code, td)
	}
	rec = call(t, h, "PATCH", "/todos/1", `{"done":true}`)
	if td := decodeTodo(t, rec); td != (item{1, "new", true}) {
		t.Errorf("PATCH done должен сохранить title: %+v", td)
	}
	if call(t, h, "PATCH", "/todos/1", `{"title":""}`).Code != 422 {
		t.Error("пустой title в PATCH → 422")
	}
	if call(t, h, "PATCH", "/todos/9", `{"done":true}`).Code != 404 {
		t.Error("PATCH несуществующего → 404")
	}

	rec = call(t, h, "DELETE", "/todos/1", "")
	if rec.Code != 204 || rec.Body.Len() != 0 {
		t.Errorf("DELETE: %d %q", rec.Code, rec.Body)
	}
	if call(t, h, "DELETE", "/todos/1", "").Code != 404 || call(t, h, "GET", "/todos/1", "").Code != 404 {
		t.Error("после удаления — 404")
	}
}

func TestConcurrentCreate(t *testing.T) {
	h := NewTodoAPI()
	var wg sync.WaitGroup
	ids := make(chan int, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := call(t, h, "POST", "/todos", `{"title":"x"}`)
			ids <- decodeTodo(t, rec).ID
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[int]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("повторяющийся id %d", id)
		}
		seen[id] = true
	}
	if len(seen) != 100 {
		t.Errorf("создано %d", len(seen))
	}
}
@@ solution
package solution

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type api struct {
	mu     sync.Mutex
	nextID int
	items  map[int]todo
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readBody(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("лишние данные")
	}
	return nil
}

func NewTodoAPI() http.Handler {
	a := &api{nextID: 1, items: map[int]todo{}}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Title string `json:"title"` }
		if err := readBody(w, r, &in); err != nil {
			writeErr(w, 400, "некорректный JSON")
			return
		}
		if strings.TrimSpace(in.Title) == "" {
			writeErr(w, 422, "title обязателен")
			return
		}
		a.mu.Lock()
		t := todo{ID: a.nextID, Title: in.Title}
		a.items[t.ID] = t
		a.nextID++
		a.mu.Unlock()
		w.Header().Set("Location", "/todos/"+strconv.Itoa(t.ID))
		writeJSON(w, 201, t)
	})

	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		var filter *bool
		if v := r.URL.Query().Get("done"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil || (v != "true" && v != "false") {
				writeErr(w, 400, "done должен быть true или false")
				return
			}
			filter = &b
		}
		a.mu.Lock()
		out := make([]todo, 0, len(a.items))
		for _, t := range a.items {
			if filter == nil || t.Done == *filter {
				out = append(out, t)
			}
		}
		a.mu.Unlock()
		sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
		writeJSON(w, 200, out)
	})

	id := func(w http.ResponseWriter, r *http.Request) (int, bool) {
		n, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeErr(w, 400, "id должен быть числом")
			return 0, false
		}
		return n, true
	}

	mux.HandleFunc("GET /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, ok := id(w, r)
		if !ok {
			return
		}
		a.mu.Lock()
		t, found := a.items[n]
		a.mu.Unlock()
		if !found {
			writeErr(w, 404, "не найдено")
			return
		}
		writeJSON(w, 200, t)
	})

	mux.HandleFunc("PATCH /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, ok := id(w, r)
		if !ok {
			return
		}
		var in struct {
			Title *string `json:"title"`
			Done  *bool   `json:"done"`
		}
		if err := readBody(w, r, &in); err != nil {
			writeErr(w, 400, "некорректный JSON")
			return
		}
		if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
			writeErr(w, 422, "title не может быть пустым")
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		t, found := a.items[n]
		if !found {
			writeErr(w, 404, "не найдено")
			return
		}
		if in.Title != nil {
			t.Title = *in.Title
		}
		if in.Done != nil {
			t.Done = *in.Done
		}
		a.items[n] = t
		writeJSON(w, 200, t)
	})

	mux.HandleFunc("DELETE /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		n, ok := id(w, r)
		if !ok {
			return
		}
		a.mu.Lock()
		_, found := a.items[n]
		delete(a.items, n)
		a.mu.Unlock()
		if !found {
			writeErr(w, 404, "не найдено")
			return
		}
		w.WriteHeader(204)
	})
	return mux
}
@@ hint
Начни с общих функций `writeJSON`/`writeErr`. Для PATCH используй структуру с указателями. Пустой список: `make([]todo, 0)`, а не `var out []todo` (иначе будет `null`).
