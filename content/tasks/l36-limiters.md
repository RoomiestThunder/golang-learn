---
title: Rate limiters: token bucket и sliding window
level: hard
---
@@ statement
Реализуй два ограничителя частоты с **внедряемыми часами** (`now func() time.Time`) — как в реальном коде, чтобы тесты не зависели от `time.Sleep`. Оба потокобезопасны.

**`Limiter` — token bucket**
- `NewLimiter(rate float64, burst int, now func() time.Time) *Limiter`: `rate` — токенов в секунду, `burst` — ёмкость ведра. В начале ведро **полное**.
- `Allow() bool` = `AllowN(1)`.
- `AllowN(n int) bool` — забирает `n` токенов, если они есть (возвращает `true`), иначе `false` и **ничего не списывает**. `n > burst` — всегда `false`. Пополнение вычисляется лениво при вызове: `tokens = min(burst, tokens + elapsed.Seconds()*rate)`.
- `Tokens() float64` — текущее число токенов (после пополнения).

**`SlidingWindow` — лог запросов по ключу**
- `NewSlidingWindow(limit int, window time.Duration, now func() time.Time) *SlidingWindow`.
- `Allow(key string) bool` — разрешает не более `limit` запросов на ключ за скользящее окно: учитываются запросы с временем `> now - window` (запрос ровно `window` назад уже не считается). Отклонённые запросы в лог не попадают. Ключи не влияют друг на друга. Устаревшие записи удаляй (не копи память).
@@ starter
package solution

import "time"

type Limiter struct {
	// TODO
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter { return &Limiter{} }

func (l *Limiter) Allow() bool { return false }

func (l *Limiter) AllowN(n int) bool { return false }

func (l *Limiter) Tokens() float64 { return 0 }

type SlidingWindow struct {
	// TODO
}

func NewSlidingWindow(limit int, window time.Duration, now func() time.Time) *SlidingWindow {
	return &SlidingWindow{}
}

func (s *SlidingWindow) Allow(key string) bool { return false }
@@ tests
package solution

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *fakeClock { return &fakeClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func TestTokenBucket(t *testing.T) {
	clk := newClock()
	l := NewLimiter(10, 5, clk.Now) // 10 токенов/с, ёмкость 5
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("запрос %d из начального всплеска должен пройти", i+1)
		}
	}
	if l.Allow() {
		t.Error("шестой запрос должен быть отклонён")
	}
	clk.Advance(100 * time.Millisecond) // +1 токен
	if !l.Allow() || l.Allow() {
		t.Error("за 100 мс накапливается ровно один токен")
	}
	clk.Advance(time.Hour) // переполнение ведра не копится
	if tk := l.Tokens(); math.Abs(tk-5) > 1e-9 {
		t.Errorf("Tokens = %v, want 5 (потолок burst)", tk)
	}
}

func TestAllowN(t *testing.T) {
	clk := newClock()
	l := NewLimiter(1, 5, clk.Now)
	if !l.AllowN(3) {
		t.Fatal("AllowN(3) из 5")
	}
	if l.AllowN(3) {
		t.Error("осталось 2 токена — AllowN(3) должен отказать")
	}
	if tk := l.Tokens(); math.Abs(tk-2) > 1e-9 {
		t.Errorf("отказ не должен списывать токены: Tokens = %v", tk)
	}
	if !l.AllowN(2) {
		t.Error("AllowN(2)")
	}
	if l.AllowN(6) || l.AllowN(100) {
		t.Error("n > burst — всегда false")
	}
	clk.Advance(1500 * time.Millisecond)
	if tk := l.Tokens(); math.Abs(tk-1.5) > 1e-9 {
		t.Errorf("дробные токены: %v", tk)
	}
}

func TestLimiterConcurrent(t *testing.T) {
	clk := newClock()
	l := NewLimiter(0, 50, clk.Now)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 50 {
		t.Errorf("разрешено %d, должно быть ровно 50", allowed.Load())
	}
}

func TestSlidingWindow(t *testing.T) {
	clk := newClock()
	w := NewSlidingWindow(3, time.Second, clk.Now)
	for i := 0; i < 3; i++ {
		if !w.Allow("a") {
			t.Fatalf("запрос %d должен пройти", i+1)
		}
	}
	if w.Allow("a") {
		t.Error("четвёртый запрос в окне — отказ")
	}
	if !w.Allow("b") {
		t.Error("другой ключ независим")
	}
	clk.Advance(999 * time.Millisecond)
	if w.Allow("a") {
		t.Error("через 999 мс первые запросы ещё в окне")
	}
	clk.Advance(1 * time.Millisecond) // ровно 1 с: запросы с t=0 уже вне окна
	if !w.Allow("a") {
		t.Error("ровно через окно старые запросы не считаются")
	}
}

func TestSlidingWindowRollingAndRejectsNotLogged(t *testing.T) {
	clk := newClock()
	w := NewSlidingWindow(2, 10*time.Second, clk.Now)
	w.Allow("k") // t=0
	clk.Advance(4 * time.Second)
	w.Allow("k")                  // t=4
	if w.Allow("k") {             // отказ в t=4 — не должен попасть в лог
		t.Fatal("лимит 2")
	}
	clk.Advance(6 * time.Second) // t=10: запрос t=0 ушёл из окна, t=4 остался
	if !w.Allow("k") {
		t.Error("после t=10 освободился один слот")
	}
	if w.Allow("k") {
		t.Error("снова лимит (запросы t=4 и t=10)")
	}
	clk.Advance(4 * time.Second) // t=14: запрос t=4 ушёл; отказ из t=4 не должен продлевать блок
	if !w.Allow("k") {
		t.Error("отклонённые запросы не должны занимать слоты")
	}
}

func TestSlidingWindowConcurrent(t *testing.T) {
	clk := newClock()
	w := NewSlidingWindow(10, time.Minute, clk.Now)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if w.Allow("same") {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 10 {
		t.Errorf("разрешено %d, должно быть 10", allowed.Load())
	}
}
@@ solution
package solution

import (
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	rate   float64
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{rate: rate, burst: float64(burst), tokens: float64(burst), last: now(), now: now}
}

func (l *Limiter) refill() {
	t := l.now()
	if elapsed := t.Sub(l.last).Seconds(); elapsed > 0 {
		l.tokens = min(l.burst, l.tokens+elapsed*l.rate)
	}
	l.last = t
}

func (l *Limiter) Allow() bool { return l.AllowN(1) }

func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if float64(n) > l.burst || l.tokens < float64(n) {
		return false
	}
	l.tokens -= float64(n)
	return true
}

func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}

type SlidingWindow struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	log    map[string][]time.Time
}

func NewSlidingWindow(limit int, window time.Duration, now func() time.Time) *SlidingWindow {
	return &SlidingWindow{limit: limit, window: window, now: now, log: map[string][]time.Time{}}
}

func (s *SlidingWindow) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.now()
	cutoff := t.Add(-s.window)
	entries := s.log[key]
	i := 0
	for i < len(entries) && !entries[i].After(cutoff) { // записи <= cutoff уже вне окна
		i++
	}
	entries = entries[i:]
	if len(entries) >= s.limit {
		s.log[key] = entries
		return false
	}
	s.log[key] = append(entries, t)
	return true
}
@@ hint
Token bucket хранит `tokens` и время последнего пополнения; при каждом вызове сначала «досыпь» токены за прошедшее время (с потолком `burst`). Sliding window: на ключ — слайс меток времени; отбрось записи `<= now-window`, потом сравни длину с лимитом.
