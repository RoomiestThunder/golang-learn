---
title: LRU-кэш (generic, потокобезопасный)
level: hard
---
@@ statement
Реализуй потокобезопасный **LRU-кэш** `LRU[K comparable, V any]` с операциями за O(1) на базе `container/list` + `map`.

- `NewLRU[K comparable, V any](capacity int, onEvict func(K, V)) *LRU[K, V]` — `capacity <= 0` → `panic`. `onEvict` (может быть `nil`) вызывается при **вытеснении** из-за переполнения (не при `Remove` и не при обновлении существующего ключа).
- `Get(key K) (V, bool)` — возвращает значение и помечает ключ самым свежим.
- `Put(key K, val V)` — добавляет/обновляет (обновление тоже делает ключ самым свежим); при переполнении вытесняет **наименее недавно использованный** элемент.
- `Remove(key K) bool`, `Len() int`.
- `Keys() []K` — ключи от самого свежего к самому старому.
- Должен безопасно работать из многих горутин (`sync.Mutex`).
@@ starter
package solution

type LRU[K comparable, V any] struct {
	// TODO
}

func NewLRU[K comparable, V any](capacity int, onEvict func(K, V)) *LRU[K, V] {
	return &LRU[K, V]{}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	var zero V
	return zero, false
}

func (c *LRU[K, V]) Put(key K, val V) {}

func (c *LRU[K, V]) Remove(key K) bool { return false }

func (c *LRU[K, V]) Len() int { return 0 }

func (c *LRU[K, V]) Keys() []K { return nil }
@@ tests
package solution

import (
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestLRUBasic(t *testing.T) {
	c := NewLRU[string, int](2, nil)
	c.Put("a", 1)
	c.Put("b", 2)
	if v, ok := c.Get("a"); !ok || v != 1 { // теперь "a" свежее "b"
		t.Fatalf("Get(a) = %d %v", v, ok)
	}
	c.Put("c", 3) // вытесняет "b"
	if _, ok := c.Get("b"); ok {
		t.Error("b должен быть вытеснен")
	}
	if !slices.Equal(c.Keys(), []string{"c", "a"}) {
		t.Errorf("Keys = %v", c.Keys())
	}
	c.Put("a", 10) // обновление делает "a" самым свежим и не меняет размер
	if v, _ := c.Get("a"); v != 10 || c.Len() != 2 || !slices.Equal(c.Keys(), []string{"a", "c"}) {
		t.Errorf("после обновления: %d %v %v", v, c.Len(), c.Keys())
	}
	if !c.Remove("a") || c.Remove("a") || c.Len() != 1 {
		t.Error("Remove")
	}
}

func TestEvictCallback(t *testing.T) {
	var evicted []string
	c := NewLRU[int, string](2, func(k int, v string) { evicted = append(evicted, fmt.Sprint(k, "=", v)) })
	c.Put(1, "x")
	c.Put(2, "y")
	c.Put(1, "z") // обновление — не вытеснение
	c.Remove(2)   // удаление — не вытеснение
	c.Put(3, "q")
	c.Put(4, "w") // вытесняет 1
	if !slices.Equal(evicted, []string{"1=z"}) {
		t.Errorf("evicted = %v", evicted)
	}
}

func TestCapacityPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("capacity <= 0 должен паниковать")
		}
	}()
	NewLRU[int, int](0, nil)
}

func TestLRUPerformanceAndConcurrency(t *testing.T) {
	c := NewLRU[int, int](1000, nil)
	start := time.Now()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50_000; i++ {
				k := (i*7 + g) % 3000
				c.Put(k, i)
				c.Get(k)
			}
		}()
	}
	wg.Wait()
	if c.Len() != 1000 {
		t.Errorf("Len = %d, want 1000", c.Len())
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("слишком медленно (%v) — все операции должны быть O(1)", d)
	}
}
@@ solution
package solution

import (
	"container/list"
	"sync"
)

type entry[K comparable, V any] struct {
	key K
	val V
}

type LRU[K comparable, V any] struct {
	mu      sync.Mutex
	cap     int
	ll      *list.List // front — самый свежий
	items   map[K]*list.Element
	onEvict func(K, V)
}

func NewLRU[K comparable, V any](capacity int, onEvict func(K, V)) *LRU[K, V] {
	if capacity <= 0 {
		panic("LRU: capacity must be positive")
	}
	return &LRU[K, V]{cap: capacity, ll: list.New(), items: make(map[K]*list.Element), onEvict: onEvict}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		return el.Value.(*entry[K, V]).val, true
	}
	var zero V
	return zero, false
}

func (c *LRU[K, V]) Put(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		el.Value.(*entry[K, V]).val = val
		c.ll.MoveToFront(el)
		return
	}
	c.items[key] = c.ll.PushFront(&entry[K, V]{key, val})
	if c.ll.Len() > c.cap {
		oldest := c.ll.Back()
		e := c.ll.Remove(oldest).(*entry[K, V])
		delete(c.items, e.key)
		if c.onEvict != nil {
			c.onEvict(e.key, e.val)
		}
	}
}

func (c *LRU[K, V]) Remove(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return false
	}
	c.ll.Remove(el)
	delete(c.items, key)
	return true
}

func (c *LRU[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *LRU[K, V]) Keys() []K {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]K, 0, c.ll.Len())
	for el := c.ll.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Value.(*entry[K, V]).key)
	}
	return keys
}
@@ hint
`map[K]*list.Element` даёт O(1) доступ к узлу, а `list.MoveToFront`/`Remove` работают за O(1). В элементе списка храни и ключ (нужен при вытеснении, чтобы удалить из map).
