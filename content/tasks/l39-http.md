---
title: Проект: HTTP-слой сокращателя
level: medium
---
@@ statement
Реализуй `NewHandler(svc Service, baseURL string) http.Handler` — HTTP-интерфейс сервиса. Сервис приходит как **интерфейс** (в тестах — фейк):

```go
type Service interface {
	Shorten(rawURL string) (code string, err error)  // ErrInvalidURL для плохого URL
	Resolve(code string) (url string, err error)     // ErrNotFound для неизвестного кода
}
```

Маршруты (шаблоны `ServeMux` Go 1.22):

| Запрос | Поведение |
|---|---|
| `POST /shorten` тело `{"url":"..."}` | успех: `201`, `Content-Type: application/json`, тело `{"code":"<code>","short_url":"<baseURL>/<code>"}` (у `baseURL` убери завершающий `/`). Битый JSON / неизвестные поля / пустое тело → `400`. `errors.Is(err, ErrInvalidURL)` → `422`. Любая другая ошибка сервиса → `500` и тело `{"error":"internal error"}` (детали ошибки клиенту **не** отдавать). |
| `GET /{code}` | успех: `302 Found` + заголовок `Location`. `ErrNotFound` → `404`. Иная ошибка → `500`. |
| `GET /healthz` | `200`, тело `ok` |

Все JSON-ошибки имеют вид `{"error":"<текст>"}`. Размер тела запроса ограничь 64 КиБ.
@@ starter
package solution

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound   = errors.New("ссылка не найдена")
	ErrInvalidURL = errors.New("некорректный URL")
)

type Service interface {
	Shorten(rawURL string) (string, error)
	Resolve(code string) (string, error)
}

func NewHandler(svc Service, baseURL string) http.Handler {
	return http.NewServeMux()
}
@@ tests
package solution

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeSvc struct {
	shortenErr error
	resolveErr error
	gotURL     string
}

func (f *fakeSvc) Shorten(u string) (string, error) {
	f.gotURL = u
	if f.shortenErr != nil {
		return "", f.shortenErr
	}
	return "abc", nil
}

func (f *fakeSvc) Resolve(code string) (string, error) {
	if f.resolveErr != nil {
		return "", f.resolveErr
	}
	if code != "abc" {
		return "", fmt.Errorf("lookup %s: %w", code, ErrNotFound)
	}
	return "https://example.com/target", nil
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestShorten(t *testing.T) {
	svc := &fakeSvc{}
	h := NewHandler(svc, "http://short.io/")
	rec := do(h, "POST", "/shorten", `{"url":"https://example.com/long"}`)
	if rec.Code != 201 {
		t.Fatalf("код %d, тело %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	var resp struct {
		Code     string `json:"code"`
		ShortURL string `json:"short_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.Code != "abc" || resp.ShortURL != "http://short.io/abc" {
		t.Errorf("тело %s (%v)", rec.Body, err)
	}
	if svc.gotURL != "https://example.com/long" {
		t.Errorf("в сервис передан %q", svc.gotURL)
	}
}

func TestShortenErrors(t *testing.T) {
	tests := []struct {
		name string
		svc  *fakeSvc
		body string
		code int
	}{
		{"битый JSON", &fakeSvc{}, `{"url":`, 400},
		{"неизвестное поле", &fakeSvc{}, `{"url":"x","extra":1}`, 400},
		{"пустое тело", &fakeSvc{}, ``, 400},
		{"невалидный URL", &fakeSvc{shortenErr: fmt.Errorf("проверка: %w", ErrInvalidURL)}, `{"url":"nope"}`, 422},
		{"сбой сервиса", &fakeSvc{shortenErr: errors.New("db password=secret leaked")}, `{"url":"https://a.com"}`, 500},
	}
	for _, tt := range tests {
		rec := do(NewHandler(tt.svc, "http://x"), "POST", "/shorten", tt.body)
		if rec.Code != tt.code {
			t.Errorf("%s: код %d, want %d", tt.name, rec.Code, tt.code)
		}
		if !strings.Contains(rec.Body.String(), `"error"`) {
			t.Errorf("%s: тело %q не содержит error", tt.name, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "secret") {
			t.Errorf("%s: детали внутренней ошибки утекли клиенту: %s", tt.name, rec.Body)
		}
	}
}

func TestRedirect(t *testing.T) {
	h := NewHandler(&fakeSvc{}, "http://x")
	rec := do(h, "GET", "/abc", "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://example.com/target" {
		t.Errorf("код %d, Location %q", rec.Code, rec.Header().Get("Location"))
	}
	if rec := do(h, "GET", "/unknown", ""); rec.Code != 404 || !strings.Contains(rec.Body.String(), `"error"`) {
		t.Errorf("неизвестный код: %d %s", rec.Code, rec.Body)
	}
	broken := NewHandler(&fakeSvc{resolveErr: errors.New("boom")}, "http://x")
	if rec := do(broken, "GET", "/abc", ""); rec.Code != 500 {
		t.Errorf("сбой: %d", rec.Code)
	}
}

func TestHealthAndMethods(t *testing.T) {
	h := NewHandler(&fakeSvc{}, "http://x")
	if rec := do(h, "GET", "/healthz", ""); rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Errorf("healthz: %d %q", rec.Code, rec.Body)
	}
	if rec := do(h, "GET", "/shorten", ""); rec.Code != 405 && rec.Code != 404 {
		t.Errorf("GET /shorten: %d", rec.Code)
	}
	if rec := do(h, "POST", "/abc", ""); rec.Code != 405 {
		t.Errorf("POST /abc: %d, ожидался 405", rec.Code)
	}
}

func TestBodyLimit(t *testing.T) {
	h := NewHandler(&fakeSvc{}, "http://x")
	big := `{"url":"https://example.com/` + strings.Repeat("a", 100_000) + `"}`
	if rec := do(h, "POST", "/shorten", big); rec.Code < 400 || rec.Code >= 500 {
		t.Errorf("слишком большое тело должно отклоняться 4xx, получили %d", rec.Code)
	}
}
@@ solution
package solution

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	ErrNotFound   = errors.New("ссылка не найдена")
	ErrInvalidURL = errors.New("некорректный URL")
)

type Service interface {
	Shorten(rawURL string) (string, error)
	Resolve(code string) (string, error)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func NewHandler(svc Service, baseURL string) http.Handler {
	baseURL = strings.TrimRight(baseURL, "/")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			URL string `json:"url"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			writeErr(w, http.StatusBadRequest, "некорректный JSON")
			return
		}
		if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			writeErr(w, http.StatusBadRequest, "лишние данные в теле")
			return
		}
		code, err := svc.Shorten(in.URL)
		switch {
		case errors.Is(err, ErrInvalidURL):
			writeErr(w, http.StatusUnprocessableEntity, "некорректный URL")
		case err != nil:
			writeErr(w, http.StatusInternalServerError, "internal error")
		default:
			writeJSON(w, http.StatusCreated, map[string]string{"code": code, "short_url": fmt.Sprintf("%s/%s", baseURL, code)})
		}
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		target, err := svc.Resolve(r.PathValue("code"))
		switch {
		case errors.Is(err, ErrNotFound):
			writeErr(w, http.StatusNotFound, "ссылка не найдена")
		case err != nil:
			writeErr(w, http.StatusInternalServerError, "internal error")
		default:
			http.Redirect(w, r, target, http.StatusFound)
		}
	})
	return mux
}
@@ hint
Шаблон `GET /{code}` не конфликтует с `GET /healthz`: более специфичный (литеральный) побеждает. Ошибки сервиса различай через `errors.Is`, а клиенту для неизвестных ошибок отдавай общий текст.
