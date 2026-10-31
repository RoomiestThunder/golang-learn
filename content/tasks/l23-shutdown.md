---
title: Graceful shutdown сервера
level: medium
---
@@ statement
Реализуй `Serve(ctx context.Context, ln net.Listener, h http.Handler, grace time.Duration) error`.

- Запускает `http.Server` на переданном listener'е (с `ReadHeaderTimeout` 5 секунд).
- Когда `ctx` отменяется — вызывает `srv.Shutdown` с таймаутом `grace`: новые соединения не принимаются, а уже начатые запросы **доводятся до конца**.
- Возвращает `nil` при штатной остановке (`http.ErrServerClosed` — не ошибка!). Если сервер упал сам по другой причине — вернуть эту ошибку. Если `Shutdown` не уложился в `grace` — вернуть ошибку.
- Функция блокирующая: возвращается после полной остановки.
@@ starter
package solution

import (
	"context"
	"net"
	"net/http"
	"time"
)

func Serve(ctx context.Context, ln net.Listener, h http.Handler, grace time.Duration) error {
	return http.Serve(ln, h)
}
@@ tests
package solution

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestGracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond) // «долгий» запрос
		w.Write([]byte("done"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- Serve(ctx, ln, h, 5*time.Second) }()

	type result struct {
		body string
		err  error
	}
	res := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			res <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b := make([]byte, 16)
		n, _ := resp.Body.Read(b)
		res <- result{body: string(b[:n])}
	}()

	<-started
	cancel() // запрос ещё выполняется

	select {
	case r := <-res:
		if r.err != nil || r.body != "done" {
			t.Fatalf("активный запрос должен завершиться успешно: %+v", r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("запрос не завершился")
	}
	select {
	case err := <-served:
		if err != nil {
			t.Errorf("Serve вернула %v, ожидался nil", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve не вернулась после отмены контекста")
	}
	if _, err := http.Get("http://" + ln.Addr().String()); err == nil {
		t.Error("после остановки сервер не должен принимать соединения")
	}
}

func TestGraceExceeded(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	started := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(time.Second)
	})
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- Serve(ctx, ln, h, 50*time.Millisecond) }()
	go http.Get("http://" + ln.Addr().String())
	<-started
	cancel()
	select {
	case err := <-served:
		if err == nil {
			t.Error("при превышении grace должна вернуться ошибка")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve не вернулась")
	}
}
@@ solution
package solution

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func Serve(ctx context.Context, ln net.Listener, h http.Handler, grace time.Duration) error {
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
@@ hint
`srv.Serve(ln)` блокируется — запусти его в горутине. После `Shutdown` он вернёт `http.ErrServerClosed`: это успех.
