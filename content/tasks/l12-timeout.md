---
title: Таймаут и «кто быстрее» через select
level: medium
---
@@ statement
1. `WithTimeout(f func() int, d time.Duration) (int, bool)` — запусти `f` в горутине; если она вернула результат за время `d` — `(результат, true)`, иначе `(0, false)`. **Важно:** после таймаута горутина не должна зависнуть навсегда на отправке результата (используй буферизованный канал).
2. `FirstOf(fns ...func() int) int` — запусти все функции одновременно и верни результат той, что закончит первой. Остальные горутины не должны утечь (результат в буфер). Для пустого набора — `panic`.

Подсказка: `select` с `time.After`.
@@ starter
package solution

import "time"

func WithTimeout(f func() int, d time.Duration) (int, bool) {
	return f(), true
}

func FirstOf(fns ...func() int) int {
	return fns[0]()
}
@@ tests
package solution

import (
	"runtime"
	"testing"
	"time"
)

func sleeper(d time.Duration, v int) func() int {
	return func() int { time.Sleep(d); return v }
}

func TestWithTimeout(t *testing.T) {
	if v, ok := WithTimeout(sleeper(10*time.Millisecond, 7), time.Second); !ok || v != 7 {
		t.Errorf("быстрая: %d %v", v, ok)
	}
	start := time.Now()
	v, ok := WithTimeout(sleeper(500*time.Millisecond, 1), 50*time.Millisecond)
	if ok || v != 0 {
		t.Errorf("медленная: %d %v", v, ok)
	}
	if d := time.Since(start); d > 300*time.Millisecond {
		t.Errorf("вернулась слишком поздно: %v", d)
	}
}

func TestFirstOf(t *testing.T) {
	got := FirstOf(sleeper(300*time.Millisecond, 1), sleeper(20*time.Millisecond, 2), sleeper(200*time.Millisecond, 3))
	if got != 2 {
		t.Errorf("FirstOf = %d, want 2", got)
	}
	defer func() {
		if recover() == nil {
			t.Error("FirstOf() без аргументов должна паниковать")
		}
	}()
	FirstOf()
}

func TestNoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	WithTimeout(sleeper(150*time.Millisecond, 1), 10*time.Millisecond)
	FirstOf(sleeper(10*time.Millisecond, 1), sleeper(100*time.Millisecond, 2), sleeper(120*time.Millisecond, 3))
	time.Sleep(500 * time.Millisecond) // даём медленным функциям закончить
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("утечка горутин: было %d, стало %d (отправляй результат в буферизованный канал)", before, after)
	}
}
@@ solution
package solution

import "time"

func WithTimeout(f func() int, d time.Duration) (int, bool) {
	ch := make(chan int, 1) // буфер: горутина не зависнет, если мы уже ушли по таймауту
	go func() { ch <- f() }()
	select {
	case v := <-ch:
		return v, true
	case <-time.After(d):
		return 0, false
	}
}

func FirstOf(fns ...func() int) int {
	if len(fns) == 0 {
		panic("FirstOf: нет функций")
	}
	ch := make(chan int, len(fns))
	for _, f := range fns {
		go func() { ch <- f() }()
	}
	return <-ch
}
@@ hint
`ch := make(chan int, 1)` — отправитель никогда не блокируется, даже если читателя уже нет.
