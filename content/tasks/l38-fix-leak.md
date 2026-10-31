---
title: Найди утечку горутин
level: medium
---
@@ statement
Функция `FirstNonEmpty` запускает все `fetchers` параллельно и возвращает **первый непустой** результат по времени завершения (если все вернули `""` — пустую строку). Работает, но **течёт**: когда победитель найден, остальные горутины навсегда блокируются на отправке в канал.

Исправь функцию так, чтобы после возврата ни одна горутина не оставалась висеть (после того как медленные fetcher'ы закончат работу). Поведение должно остаться прежним.
@@ starter
package solution

func FirstNonEmpty(fetchers []func() string) string {
	ch := make(chan string) // небуферизованный
	for _, f := range fetchers {
		go func() { ch <- f() }()
	}
	for range fetchers {
		if s := <-ch; s != "" {
			return s
		}
	}
	return ""
}
@@ tests
package solution

import (
	"runtime"
	"testing"
	"time"
)

func fetch(d time.Duration, v string) func() string {
	return func() string { time.Sleep(d); return v }
}

func TestBehavior(t *testing.T) {
	got := FirstNonEmpty([]func() string{
		fetch(10*time.Millisecond, ""), fetch(60*time.Millisecond, "slow"), fetch(30*time.Millisecond, "fast"), fetch(5*time.Millisecond, ""),
	})
	if got != "fast" {
		t.Errorf("got %q, want fast", got)
	}
	if got := FirstNonEmpty([]func() string{fetch(time.Millisecond, ""), fetch(time.Millisecond, "")}); got != "" {
		t.Errorf("все пусты: %q", got)
	}
	if got := FirstNonEmpty(nil); got != "" {
		t.Errorf("nil: %q", got)
	}
}

func TestNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		FirstNonEmpty([]func() string{fetch(time.Millisecond, "x"), fetch(80*time.Millisecond, "y"), fetch(90*time.Millisecond, "z")})
	}
	time.Sleep(300 * time.Millisecond) // даём медленным fetcher'ам закончить
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("утекло %d горутин", after-before)
	}
}
@@ solution
package solution

func FirstNonEmpty(fetchers []func() string) string {
	ch := make(chan string, len(fetchers)) // буфер: отправитель никогда не блокируется
	for _, f := range fetchers {
		go func() { ch <- f() }()
	}
	for range fetchers {
		if s := <-ch; s != "" {
			return s
		}
	}
	return ""
}
@@ hint
Достаточно сделать канал буферизованным на `len(fetchers)` элементов.
