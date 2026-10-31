---
title: Middleware: Chain, Recover, RequestID, BasicAuth
level: medium
---
@@ statement
Реализуй набор middleware (тип `Middleware = func(http.Handler) http.Handler`).

1. `Chain(h http.Handler, mws ...Middleware) http.Handler` — первый в списке выполняется **первым** (самый внешний).
2. `Recover(next http.Handler) http.Handler` — перехватывает панику обработчика, отвечает `500` и телом `internal error`; сервис не падает.
3. `RequestID(next http.Handler) http.Handler` — если у запроса есть заголовок `X-Request-ID`, использует его, иначе генерирует новый (непустой, разный для разных запросов); записывает его в заголовок **ответа** `X-Request-ID` и в контекст запроса. `RequestIDFrom(ctx context.Context) string` достаёт значение (пустая строка, если нет).
4. `BasicAuth(user, pass string) Middleware` — требует HTTP Basic Auth. При неверных данных: `401` и заголовок `WWW-Authenticate: Basic realm="restricted"`; сравнение **за константное время** (`crypto/subtle`), чтобы не раскрывать пароль по времени ответа.
@@ starter
package solution

import (
	"context"
	"net/http"
)

type Middleware = func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler { return h }

func Recover(next http.Handler) http.Handler { return next }

func RequestID(next http.Handler) http.Handler { return next }

func RequestIDFrom(ctx context.Context) string { return "" }

func BasicAuth(user, pass string) Middleware {
	return func(next http.Handler) http.Handler { return next }
}
@@ tests
package solution

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestChainOrder(t *testing.T) {
	var trace []string
	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, name+":in")
				next.ServeHTTP(w, r)
				trace = append(trace, name+":out")
			})
		}
	}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { trace = append(trace, "handler") }),
		mw("A"), mw("B"), mw("C"))
	serve(h, httptest.NewRequest("GET", "/", nil))
	if got := strings.Join(trace, ","); got != "A:in,B:in,C:in,handler,C:out,B:out,A:out" {
		t.Errorf("порядок: %s", got)
	}
	if serve(Chain(http.NotFoundHandler()), httptest.NewRequest("GET", "/", nil)).Code != 404 {
		t.Error("пустой Chain должен возвращать исходный handler")
	}
}

func TestRecover(t *testing.T) {
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") }))
	rec := serve(h, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 || !strings.Contains(rec.Body.String(), "internal error") {
		t.Errorf("код %d, тело %q", rec.Code, rec.Body.String())
	}
	ok := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("fine")) }))
	if rec := serve(ok, httptest.NewRequest("GET", "/", nil)); rec.Code != 200 || rec.Body.String() != "fine" {
		t.Errorf("нормальный запрос: %d %q", rec.Code, rec.Body.String())
	}
}

func TestRequestID(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = RequestIDFrom(r.Context()) }))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	rec := serve(h, req)
	if rec.Header().Get("X-Request-ID") != "abc-123" || seen != "abc-123" {
		t.Errorf("входной id не сохранён: header=%q ctx=%q", rec.Header().Get("X-Request-ID"), seen)
	}

	r1 := serve(h, httptest.NewRequest("GET", "/", nil)).Header().Get("X-Request-ID")
	r2 := serve(h, httptest.NewRequest("GET", "/", nil)).Header().Get("X-Request-ID")
	if r1 == "" || r2 == "" || r1 == r2 {
		t.Errorf("сгенерированные id: %q %q", r1, r2)
	}
	if RequestIDFrom(httptest.NewRequest("GET", "/", nil).Context()) != "" {
		t.Error("без middleware id должен быть пустым")
	}
}

func TestBasicAuth(t *testing.T) {
	h := BasicAuth("admin", "s3cret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("secret data")) }))

	rec := serve(h, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 401 || !strings.HasPrefix(rec.Header().Get("WWW-Authenticate"), "Basic") {
		t.Errorf("без данных: %d %q", rec.Code, rec.Header().Get("WWW-Authenticate"))
	}
	bad := httptest.NewRequest("GET", "/", nil)
	bad.SetBasicAuth("admin", "wrong")
	if rec := serve(h, bad); rec.Code != 401 || strings.Contains(rec.Body.String(), "secret data") {
		t.Errorf("неверный пароль: %d", rec.Code)
	}
	good := httptest.NewRequest("GET", "/", nil)
	good.SetBasicAuth("admin", "s3cret")
	if rec := serve(h, good); rec.Code != 200 || rec.Body.String() != "secret data" {
		t.Errorf("верные данные: %d %q", rec.Code, rec.Body.String())
	}
}
@@ solution
package solution

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

type Middleware = func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
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

type ctxKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, id)))
	})
}

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func BasicAuth(user, pass string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			userOK := subtle.ConstantTimeCompare([]byte(u), []byte(user)) == 1
			passOK := subtle.ConstantTimeCompare([]byte(p), []byte(pass)) == 1
			if !ok || !userOK || !passOK {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
@@ hint
`r.WithContext(context.WithValue(r.Context(), key, id))` — создаёт копию запроса с новым контекстом. Для `Chain` оборачивай с конца списка.
