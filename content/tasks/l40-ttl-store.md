---
title: Экзамен 1: хранилище с TTL
level: hard
---
@@ statement
Реализуй обобщённое потокобезопасное хранилище с временем жизни записей. Время — **внедряемое** (`now func() time.Time`), поэтому тесты «перематывают» часы.

```go
func NewStore[V any](now func() time.Time) *Store[V]
func (s *Store[V]) Set(key string, v V, ttl time.Duration) // ttl <= 0 → запись не истекает
func (s *Store[V]) Get(key string) (V, bool)               // истёкшая запись → (нулевое значение, false)
func (s *Store[V]) Delete(key string) bool                 // true, если живая запись была
func (s *Store[V]) Len() int                               // только ЖИВЫЕ записи
func (s *Store[V]) TTL(key string) (time.Duration, bool)   // сколько осталось; для вечных — (0, true); нет/истекла — (0, false)
func (s *Store[V]) Cleanup() int                           // физически удаляет истёкшие, возвращает их число
func (s *Store[V]) StartJanitor(ctx context.Context, interval time.Duration) <-chan struct{}
```

- Запись истекает, когда `now() >= expiresAt` (ровно в момент истечения — уже мертва).
- `Set` на существующий ключ заменяет значение и срок.
- `StartJanitor` запускает **фоновую горутину**, которая каждые `interval` (реальный `time.Ticker`) вызывает `Cleanup`, пока `ctx` не отменён. Возвращает канал, который закрывается, когда горутина завершилась. Тикер нужно остановить.
@@ starter
package solution

import (
	"context"
	"time"
)

type Store[V any] struct {
	// TODO
}

func NewStore[V any](now func() time.Time) *Store[V] { return &Store[V]{} }

func (s *Store[V]) Set(key string, v V, ttl time.Duration) {}

func (s *Store[V]) Get(key string) (V, bool) {
	var zero V
	return zero, false
}

func (s *Store[V]) Delete(key string) bool { return false }

func (s *Store[V]) Len() int { return 0 }

func (s *Store[V]) TTL(key string) (time.Duration, bool) { return 0, false }

func (s *Store[V]) Cleanup() int { return 0 }

func (s *Store[V]) StartJanitor(ctx context.Context, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}
@@ tests
package solution

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock { return &clock{t: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)} }

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func TestSetGetExpire(t *testing.T) {
	clk := newClock()
	s := NewStore[string](clk.Now)
	s.Set("a", "one", time.Minute)
	s.Set("forever", "∞", 0)
	if v, ok := s.Get("a"); !ok || v != "one" {
		t.Fatalf("Get(a) = %q %v", v, ok)
	}
	clk.Advance(59 * time.Second)
	if _, ok := s.Get("a"); !ok {
		t.Error("за секунду до истечения запись жива")
	}
	clk.Advance(time.Second) // ровно TTL
	if _, ok := s.Get("a"); ok {
		t.Error("в момент now == expiresAt запись уже мертва")
	}
	clk.Advance(24 * time.Hour)
	if v, ok := s.Get("forever"); !ok || v != "∞" {
		t.Errorf("вечная запись: %q %v", v, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Error("неизвестный ключ")
	}
}

func TestOverwriteLenTTL(t *testing.T) {
	clk := newClock()
	s := NewStore[int](clk.Now)
	s.Set("k", 1, time.Minute)
	clk.Advance(30 * time.Second)
	if d, ok := s.TTL("k"); !ok || d != 30*time.Second {
		t.Errorf("TTL = %v %v", d, ok)
	}
	s.Set("k", 2, time.Minute) // обновляет и значение, и срок
	if d, _ := s.TTL("k"); d != time.Minute {
		t.Errorf("TTL после перезаписи = %v", d)
	}
	if v, _ := s.Get("k"); v != 2 {
		t.Errorf("значение = %d", v)
	}
	s.Set("p", 3, 0)
	if d, ok := s.TTL("p"); !ok || d != 0 {
		t.Errorf("TTL вечной = %v %v", d, ok)
	}
	if _, ok := s.TTL("nope"); ok {
		t.Error("TTL несуществующей")
	}
	if s.Len() != 2 {
		t.Errorf("Len = %d", s.Len())
	}
	clk.Advance(2 * time.Minute)
	if s.Len() != 1 {
		t.Errorf("Len считает только живые: %d", s.Len())
	}
	if _, ok := s.TTL("k"); ok {
		t.Error("TTL истёкшей")
	}
}

func TestDeleteCleanup(t *testing.T) {
	clk := newClock()
	s := NewStore[int](clk.Now)
	for i := 0; i < 10; i++ {
		s.Set(fmt.Sprint("k", i), i, time.Duration(i+1)*time.Second)
	}
	if !s.Delete("k0") || s.Delete("k0") || s.Delete("nope") {
		t.Error("Delete")
	}
	clk.Advance(5 * time.Second) // ключ ki живёт i+1 секунд: k1..k4 уже истекли
	s.Set("dead", 0, time.Second)
	clk.Advance(2 * time.Second)
	if _, ok := s.Get("dead"); ok {
		t.Fatal("dead должна истечь")
	}
	if n := s.Cleanup(); n != 7 { // k1..k6 (ключ ki истекает через i+1 с) + dead
		t.Errorf("Cleanup удалил %d, ожидалось 7", n)
	}
	if n := s.Cleanup(); n != 0 {
		t.Errorf("повторный Cleanup: %d", n)
	}
	if s.Delete("k2") {
		t.Error("Delete истёкшей записи должен вернуть false")
	}
}

func TestJanitor(t *testing.T) {
	clk := newClock()
	s := NewStore[int](clk.Now)
	for i := 0; i < 5; i++ {
		s.Set(fmt.Sprint(i), i, time.Minute)
	}
	clk.Advance(2 * time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	done := s.StartJanitor(ctx, 10*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	if n := s.Cleanup(); n != 0 {
		t.Errorf("janitor уже должен был всё убрать, но Cleanup нашёл %d", n)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("janitor не остановился после отмены контекста")
	}
}

func TestConcurrentAccess(t *testing.T) {
	clk := newClock()
	s := NewStore[int](clk.Now)
	ctx, cancel := context.WithCancel(context.Background())
	done := s.StartJanitor(ctx, time.Millisecond)
	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2000; i++ {
				k := fmt.Sprint("k", (g+i)%50)
				s.Set(k, i, time.Duration(i%5)*time.Millisecond)
				s.Get(k)
				s.Len()
				if i%100 == 0 {
					clk.Advance(time.Millisecond)
				}
			}
		}()
	}
	wg.Wait()
	cancel()
	<-done
}
@@ solution
package solution

import (
	"context"
	"sync"
	"time"
)

type item[V any] struct {
	val     V
	expires time.Time // нулевое время — не истекает
}

type Store[V any] struct {
	mu    sync.Mutex
	now   func() time.Time
	items map[string]item[V]
}

func NewStore[V any](now func() time.Time) *Store[V] {
	return &Store[V]{now: now, items: make(map[string]item[V])}
}

func (it item[V]) expired(now time.Time) bool {
	return !it.expires.IsZero() && !now.Before(it.expires)
}

func (s *Store[V]) Set(key string, v V, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it := item[V]{val: v}
	if ttl > 0 {
		it.expires = s.now().Add(ttl)
	}
	s.items[key] = it
}

func (s *Store[V]) Get(key string) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[key]
	if !ok || it.expired(s.now()) {
		var zero V
		return zero, false
	}
	return it.val, true
}

func (s *Store[V]) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[key]
	delete(s.items, key)
	return ok && !it.expired(s.now())
}

func (s *Store[V]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now, n := s.now(), 0
	for _, it := range s.items {
		if !it.expired(now) {
			n++
		}
	}
	return n
}

func (s *Store[V]) TTL(key string) (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	it, ok := s.items[key]
	if !ok || it.expired(now) {
		return 0, false
	}
	if it.expires.IsZero() {
		return 0, true
	}
	return it.expires.Sub(now), true
}

func (s *Store[V]) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now, removed := s.now(), 0
	for k, it := range s.items {
		if it.expired(now) {
			delete(s.items, k)
			removed++
		}
	}
	return removed
}

func (s *Store[V]) StartJanitor(ctx context.Context, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.Cleanup()
			}
		}
	}()
	return done
}
@@ hint
В записи храни `expires time.Time` (нулевое значение = вечная). «Истекла» — это `!now.Before(expires)`. Janitor: `time.NewTicker` + `select` на `ctx.Done()` и `ticker.C`, `defer close(done)` и `defer ticker.Stop()`.
