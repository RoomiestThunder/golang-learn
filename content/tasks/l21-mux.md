---
title: Маршруты на ServeMux
level: easy
---
@@ statement
Реализуй `NewMux() http.Handler` с маршрутами (шаблоны Go 1.22 с методами):

| Запрос | Ответ |
|---|---|
| `GET /health` | `200`, тело `ok` |
| `GET /hello/{name}` | `200`, тело `Hello, <name>!` |
| `GET /sum?a=1&b=2` | `200`, JSON `{"sum":3}` с `Content-Type: application/json`. Если `a` или `b` не целые числа — `400` и JSON `{"error":"..."}` |
| `POST /upper` | `200`, тело запроса в верхнем регистре (`strings.ToUpper`). Если тело больше 1024 байт — `413` |

Остальное (404 для неизвестных путей, 405 для неверного метода) должен обеспечить сам `ServeMux`. Используй `r.PathValue`, `r.URL.Query()`, `http.MaxBytesReader`.
@@ starter
package solution

import "net/http"

func NewMux() http.Handler {
	mux := http.NewServeMux()
	// TODO: зарегистрируй маршруты
	return mux
}
@@ tests
package solution

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func do(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRoutes(t *testing.T) {
	h := NewMux()
	tests := []struct {
		method, target, body string
		code                 int
		wantBody             string
	}{
		{"GET", "/health", "", 200, "ok"},
		{"GET", "/hello/Gopher", "", 200, "Hello, Gopher!"},
		{"GET", "/sum?a=40&b=2", "", 200, `{"sum":42}`},
		{"GET", "/sum?a=-5&b=2", "", 200, `{"sum":-3}`},
		{"POST", "/upper", "go lang", 200, "GO LANG"},
		{"GET", "/nope", "", 404, ""},
		{"POST", "/health", "", 405, ""},
		{"GET", "/upper", "", 405, ""},
	}
	for _, tt := range tests {
		rec := do(h, tt.method, tt.target, tt.body)
		if rec.Code != tt.code {
			t.Errorf("%s %s: код %d, want %d", tt.method, tt.target, rec.Code, tt.code)
			continue
		}
		if tt.wantBody != "" && strings.TrimSpace(rec.Body.String()) != tt.wantBody {
			t.Errorf("%s %s: тело %q, want %q", tt.method, tt.target, rec.Body.String(), tt.wantBody)
		}
	}
}

func TestSumErrors(t *testing.T) {
	h := NewMux()
	for _, target := range []string{"/sum?a=x&b=1", "/sum?a=1", "/sum"} {
		rec := do(h, "GET", target, "")
		if rec.Code != 400 {
			t.Errorf("%s: код %d, want 400", target, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("%s: Content-Type = %q", target, ct)
		}
		if !strings.Contains(rec.Body.String(), `"error"`) {
			t.Errorf("%s: тело %q без поля error", target, rec.Body.String())
		}
	}
	if ct := do(h, "GET", "/sum?a=1&b=1", "").Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("успешный /sum: Content-Type = %q", ct)
	}
}

func TestUpperTooBig(t *testing.T) {
	rec := do(NewMux(), "POST", "/upper", strings.Repeat("x", 2000))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("код %d, want 413", rec.Code)
	}
	if rec := do(NewMux(), "POST", "/upper", strings.Repeat("x", 1024)); rec.Code != 200 {
		t.Errorf("ровно 1024 байта должны проходить, код %d", rec.Code)
	}
}
@@ solution
package solution

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func NewMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!", r.PathValue("name"))
	})

	mux.HandleFunc("GET /sum", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		a, errA := strconv.Atoi(r.URL.Query().Get("a"))
		b, errB := strconv.Atoi(r.URL.Query().Get("b"))
		if errA != nil || errB != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "a и b должны быть целыми числами"})
			return
		}
		json.NewEncoder(w).Encode(map[string]int{"sum": a + b})
	})

	mux.HandleFunc("POST /upper", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
		if err != nil {
			http.Error(w, "тело слишком большое", http.StatusRequestEntityTooLarge)
			return
		}
		fmt.Fprint(w, strings.ToUpper(string(body)))
	})
	return mux
}
@@ hint
Заголовок `Content-Type` ставь **до** `WriteHeader`. `io.ReadAll(http.MaxBytesReader(...))` вернёт ошибку, если тело превышает лимит.
