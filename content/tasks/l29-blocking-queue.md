---
title: Блокирующая очередь на sync.Cond
level: hard
---
@@ statement
Реализуй ограниченную блокирующую очередь `Queue[T any]` **без каналов** — на `sync.Mutex` и `sync.Cond` (так устроены внутренности каналов и пулов).

- `NewQueue[T any](capacity int) *Queue[T]` (`capacity <= 0` → `panic`).
- `Put(v T) bool` — кладёт элемент; если очередь **полна**, блокируется до появления места. Возвращает `false` (и ничего не кладёт), если очередь закрыта (в том числе пока ждал).
- `Get() (T, bool)` — берёт самый старый элемент (FIFO); если очередь **пуста**, блокируется. После `Close` — сначала отдаёт оставшиеся элементы, затем возвращает `(нулевое значение, false)`.
- `Close()` — закрывает очередь; будит всех ожидающих. Повторный `Close` безопасен.
- `Len() int`.

Ожидание на `Cond` всегда в цикле `for условие { cond.Wait() }`.
@@ starter
package solution

type Queue[T any] struct {
	// TODO
}

func NewQueue[T any](capacity int) *Queue[T] { return &Queue[T]{} }

func (q *Queue[T]) Put(v T) bool { return false }

func (q *Queue[T]) Get() (T, bool) {
	var zero T
	return zero, false
}

func (q *Queue[T]) Close() {}

func (q *Queue[T]) Len() int { return 0 }
@@ tests
package solution

import (
	"sync"
	"testing"
	"time"
)

func TestFIFO(t *testing.T) {
	q := NewQueue[int](3)
	for i := 1; i <= 3; i++ {
		if !q.Put(i) {
			t.Fatal("Put вернул false")
		}
	}
	if q.Len() != 3 {
		t.Errorf("Len = %d", q.Len())
	}
	for want := 1; want <= 3; want++ {
		if v, ok := q.Get(); !ok || v != want {
			t.Errorf("Get = %d %v, want %d", v, ok, want)
		}
	}
}

func TestPutBlocksWhenFull(t *testing.T) {
	q := NewQueue[string](1)
	q.Put("a")
	done := make(chan struct{})
	go func() { q.Put("b"); close(done) }()
	select {
	case <-done:
		t.Fatal("Put не заблокировался на полной очереди")
	case <-time.After(100 * time.Millisecond):
	}
	if v, _ := q.Get(); v != "a" {
		t.Errorf("v = %q", v)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Put не разблокировался после освобождения места")
	}
	if v, _ := q.Get(); v != "b" {
		t.Errorf("v = %q", v)
	}
}

func TestGetBlocksWhenEmpty(t *testing.T) {
	q := NewQueue[int](2)
	got := make(chan int)
	go func() { v, _ := q.Get(); got <- v }()
	select {
	case <-got:
		t.Fatal("Get не заблокировался на пустой очереди")
	case <-time.After(100 * time.Millisecond):
	}
	q.Put(42)
	select {
	case v := <-got:
		if v != 42 {
			t.Errorf("v = %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("Get не проснулся")
	}
}

func TestClose(t *testing.T) {
	q := NewQueue[int](5)
	q.Put(1)
	q.Put(2)
	q.Close()
	q.Close() // повторно — безопасно
	if q.Put(3) {
		t.Error("Put после Close должен вернуть false")
	}
	for _, want := range []int{1, 2} {
		if v, ok := q.Get(); !ok || v != want {
			t.Errorf("остаток: %d %v", v, ok)
		}
	}
	if v, ok := q.Get(); ok || v != 0 {
		t.Errorf("после дренажа: %d %v", v, ok)
	}
}

func TestCloseWakesWaiters(t *testing.T) {
	q := NewQueue[int](1)
	var wg sync.WaitGroup
	results := make(chan bool, 3)
	for i := 0; i < 3; i++ { // три читателя ждут на пустой очереди
		wg.Add(1)
		go func() { defer wg.Done(); _, ok := q.Get(); results <- ok }()
	}
	time.Sleep(50 * time.Millisecond)
	q.Close()
	waited := make(chan struct{})
	go func() { wg.Wait(); close(waited) }()
	select {
	case <-waited:
	case <-time.After(time.Second):
		t.Fatal("Close не разбудил читателей")
	}
	for i := 0; i < 3; i++ {
		if <-results {
			t.Error("после Close читатель должен получить ok=false")
		}
	}

	q2 := NewQueue[int](1)
	q2.Put(1)
	res := make(chan bool)
	go func() { res <- q2.Put(2) }() // ждёт места
	time.Sleep(50 * time.Millisecond)
	q2.Close()
	select {
	case ok := <-res:
		if ok {
			t.Error("заблокированный Put после Close должен вернуть false")
		}
	case <-time.After(time.Second):
		t.Fatal("Close не разбудил писателя")
	}
}

func TestProducersConsumers(t *testing.T) {
	q := NewQueue[int](4)
	const producers, perProducer = 4, 2000
	var wgP, wgC sync.WaitGroup
	sums := make(chan int, 3)
	for c := 0; c < 3; c++ {
		wgC.Add(1)
		go func() {
			defer wgC.Done()
			s := 0
			for {
				v, ok := q.Get()
				if !ok {
					break
				}
				s += v
			}
			sums <- s
		}()
	}
	for p := 0; p < producers; p++ {
		wgP.Add(1)
		go func() {
			defer wgP.Done()
			for i := 1; i <= perProducer; i++ {
				q.Put(i)
			}
		}()
	}
	wgP.Wait()
	q.Close()
	wgC.Wait()
	close(sums)
	total := 0
	for s := range sums {
		total += s
	}
	if want := producers * perProducer * (perProducer + 1) / 2; total != want {
		t.Errorf("сумма %d, want %d (потеряны или задвоены элементы)", total, want)
	}
}
@@ solution
package solution

import "sync"

type Queue[T any] struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	items    []T
	cap      int
	closed   bool
}

func NewQueue[T any](capacity int) *Queue[T] {
	if capacity <= 0 {
		panic("Queue: capacity must be positive")
	}
	q := &Queue[T]{cap: capacity}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

func (q *Queue[T]) Put(v T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == q.cap && !q.closed {
		q.notFull.Wait()
	}
	if q.closed {
		return false
	}
	q.items = append(q.items, v)
	q.notEmpty.Signal()
	return true
}

func (q *Queue[T]) Get() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed {
		q.notEmpty.Wait()
	}
	var zero T
	if len(q.items) == 0 { // закрыта и пуста
		return zero, false
	}
	v := q.items[0]
	q.items[0] = zero // не удерживаем ссылку
	q.items = q.items[1:]
	q.notFull.Signal()
	return v, true
}

func (q *Queue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
}

func (q *Queue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
@@ hint
Две условные переменные над одним мьютексом: `notEmpty` (ждут читатели) и `notFull` (ждут писатели). `Close` вызывает `Broadcast` на обеих.
