---
title: HTTP-клиент: FetchJSON и ретраи
level: medium
---
@@ statement
Напиши надёжный клиент.

1. `FetchJSON(ctx, c *http.Client, url string, out any) error`:
   - запрос с контекстом (`NewRequestWithContext`), заголовок `Accept: application/json`;
   - статус не `2xx` → ошибка `*StatusError{Code int, Body string}` (в `Body` — первые **1024 байта** тела ответа);
   - иначе декодируй тело в `out`; не забывай закрывать `Body`.
2. `GetWithRetry(ctx, c *http.Client, url string, attempts int, out any) error` — вызывает `FetchJSON` до `attempts` раз.
   - Повторяй при сетевых ошибках и ответах `429`, `500`, `502`, `503`, `504`. **Не** повторяй при остальных 4xx и при ошибках декодирования.
   - Пауза между попытками: `5ms`, `10ms`, `20ms`... (удваивается); учитывай отмену контекста (вернуть `ctx.Err()`).
   - После исчерпания попыток верни последнюю ошибку (`errors.As` на `*StatusError` должен продолжать работать).
@@ starter
package solution

import (
	"context"
	"fmt"
	"net/http"
)

type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string { return fmt.Sprintf("статус %d: %s", e.Code, e.Body) }

func FetchJSON(ctx context.Context, c *http.Client, url string, out any) error {
	return nil
}

func GetWithRetry(ctx context.Context, c *http.Client, url string, attempts int, out any) error {
	return nil
}
@@ tests
package solution

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type payload struct {
	N int `json:"n"`
}

func TestFetchJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			http.Error(w, "need Accept", 406)
			return
		}
		switch r.URL.Path {
		case "/ok":
			w.Write([]byte(`{"n":42}`))
		case "/bad":
			http.Error(w, strings.Repeat("x", 5000), 400)
		case "/garbage":
			w.Write([]byte(`<html>`))
		}
	}))
	defer srv.Close()

	var p payload
	if err := FetchJSON(context.Background(), srv.Client(), srv.URL+"/ok", &p); err != nil || p.N != 42 {
		t.Fatalf("ok: %v %+v", err, p)
	}
	err := FetchJSON(context.Background(), srv.Client(), srv.URL+"/bad", &p)
	var se *StatusError
	if !errors.As(err, &se) || se.Code != 400 || len(se.Body) != 1024 {
		t.Fatalf("400: %v (тело ограничено 1024 байтами)", err)
	}
	if err := FetchJSON(context.Background(), srv.Client(), srv.URL+"/garbage", &p); err == nil || errors.As(err, &se) {
		t.Errorf("мусор: нужна ошибка декодирования, got %v", err)
	}
}

func TestRetry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		switch r.URL.Path {
		case "/flaky":
			if n < 3 {
				w.WriteHeader(503)
				return
			}
			w.Write([]byte(`{"n":7}`))
		case "/always500":
			w.WriteHeader(500)
		case "/notfound":
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	ctx := context.Background()

	var p payload
	if err := GetWithRetry(ctx, srv.Client(), srv.URL+"/flaky", 5, &p); err != nil || p.N != 7 || calls.Load() != 3 {
		t.Errorf("flaky: err=%v p=%+v calls=%d", err, p, calls.Load())
	}

	calls.Store(0)
	err := GetWithRetry(ctx, srv.Client(), srv.URL+"/always500", 4, &p)
	var se *StatusError
	if !errors.As(err, &se) || se.Code != 500 || calls.Load() != 4 {
		t.Errorf("always500: err=%v calls=%d", err, calls.Load())
	}

	calls.Store(0)
	err = GetWithRetry(ctx, srv.Client(), srv.URL+"/notfound", 5, &p)
	if !errors.As(err, &se) || se.Code != 404 || calls.Load() != 1 {
		t.Errorf("404 не повторяется: err=%v calls=%d", err, calls.Load())
	}
}

func TestRetryNetworkErrorAndCancel(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close() // порт закрыт → сетевая ошибка

	var p payload
	start := time.Now()
	if err := GetWithRetry(context.Background(), http.DefaultClient, url, 3, &p); err == nil {
		t.Error("ожидалась сетевая ошибка")
	}
	if time.Since(start) < 10*time.Millisecond {
		t.Error("между попытками должны быть паузы (5ms, 10ms)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer srv2.Close()
	err := GetWithRetry(ctx, srv2.Client(), srv2.URL, 1000, &p)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("отмена контекста: %v", err)
	}
}
@@ solution
package solution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string { return fmt.Sprintf("статус %d: %s", e.Code, e.Body) }

type decodeError struct{ err error }

func (e *decodeError) Error() string { return "декодирование: " + e.err.Error() }
func (e *decodeError) Unwrap() error { return e.err }

func FetchJSON(ctx context.Context, c *http.Client, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &StatusError{Code: resp.StatusCode, Body: string(b)}
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return &decodeError{err}
	}
	return nil
}

func retryable(err error) bool {
	var se *StatusError
	if errors.As(err, &se) {
		switch se.Code {
		case 429, 500, 502, 503, 504:
			return true
		}
		return false
	}
	var de *decodeError
	return !errors.As(err, &de) // сетевые ошибки повторяем
}

func GetWithRetry(ctx context.Context, c *http.Client, url string, attempts int, out any) error {
	if attempts < 1 {
		attempts = 1
	}
	delay := 5 * time.Millisecond
	var err error
	for i := 0; i < attempts; i++ {
		if err = FetchJSON(ctx, c, url, out); err == nil {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if !retryable(err) || i == attempts-1 {
			return err
		}
		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
		delay *= 2
	}
	return err
}
@@ hint
Сетевая ошибка приходит из `c.Do` (не `*StatusError`); статус не-2xx — это уже твоя ошибка. Для «не повторять декодирование» заведи отдельный тип ошибки.
