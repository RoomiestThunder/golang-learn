---
title: Шардированная map против contention
level: medium
---
@@ statement
Реализуй `ShardedMap[V any]` — потокобезопасную map со строковыми ключами, которая делит данные на **шарды**, каждый со своим мьютексом (так снижают конкуренцию за один общий лок).

- `NewShardedMap[V any](shards int) *ShardedMap[V]` — число шардов (`< 1` → 1).
- `Set(key string, v V)`, `Get(key string) (V, bool)`, `Delete(key string)`, `Len() int` (сумма по шардам).
- `Update(key string, f func(old V, exists bool) V)` — **атомарное** чтение-изменение-запись для ключа (нужно для счётчиков без гонок на «прочитал → изменил → записал»). Блокируется только шард этого ключа.
- `Range(f func(key string, v V) bool)` — обход всех элементов (порядок не определён); `false` из `f` прекращает обход. Не держи несколько локов одновременно.

Шард выбирай по хэшу ключа (`hash/fnv`): один и тот же ключ **всегда** попадает в один шард.
@@ starter
package solution

type ShardedMap[V any] struct {
	// TODO
}

func NewShardedMap[V any](shards int) *ShardedMap[V] { return &ShardedMap[V]{} }

func (m *ShardedMap[V]) Set(key string, v V) {}

func (m *ShardedMap[V]) Get(key string) (V, bool) {
	var zero V
	return zero, false
}

func (m *ShardedMap[V]) Delete(key string) {}

func (m *ShardedMap[V]) Len() int { return 0 }

func (m *ShardedMap[V]) Update(key string, f func(old V, exists bool) V) {}

func (m *ShardedMap[V]) Range(f func(key string, v V) bool) {}
@@ tests
package solution

import (
	"fmt"
	"sync"
	"testing"
)

func TestBasic(t *testing.T) {
	m := NewShardedMap[int](8)
	for i := 0; i < 100; i++ {
		m.Set(fmt.Sprint("k", i), i)
	}
	if m.Len() != 100 {
		t.Fatalf("Len = %d", m.Len())
	}
	if v, ok := m.Get("k42"); !ok || v != 42 {
		t.Errorf("Get = %d %v", v, ok)
	}
	if _, ok := m.Get("missing"); ok {
		t.Error("несуществующий ключ")
	}
	m.Delete("k42")
	m.Delete("nope")
	if _, ok := m.Get("k42"); ok || m.Len() != 99 {
		t.Error("Delete")
	}
	seen := 0
	m.Range(func(k string, v int) bool { seen++; return true })
	if seen != 99 {
		t.Errorf("Range обошёл %d", seen)
	}
	stopped := 0
	m.Range(func(k string, v int) bool { stopped++; return stopped < 5 })
	if stopped != 5 {
		t.Errorf("Range не остановился: %d", stopped)
	}
}

func TestSingleShardAndZero(t *testing.T) {
	for _, n := range []int{0, 1, -3} {
		m := NewShardedMap[string](n)
		m.Set("a", "x")
		if v, ok := m.Get("a"); !ok || v != "x" {
			t.Errorf("shards=%d: %q %v", n, v, ok)
		}
	}
}

func TestUpdateAtomic(t *testing.T) {
	m := NewShardedMap[int](16)
	var wg sync.WaitGroup
	keys := []string{"a", "b", "c", "d", "e"}
	for g := 0; g < 100; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				m.Update(keys[(g+i)%len(keys)], func(old int, exists bool) int { return old + 1 })
			}
		}()
	}
	wg.Wait()
	total := 0
	m.Range(func(k string, v int) bool { total += v; return true })
	if total != 100*500 {
		t.Errorf("сумма = %d, want %d (потерянные обновления)", total, 100*500)
	}
}

func TestUpdateExists(t *testing.T) {
	m := NewShardedMap[string](4)
	m.Update("x", func(old string, exists bool) string {
		if exists || old != "" {
			t.Errorf("первый Update: exists=%v old=%q", exists, old)
		}
		return "first"
	})
	m.Update("x", func(old string, exists bool) string {
		if !exists || old != "first" {
			t.Errorf("второй Update: exists=%v old=%q", exists, old)
		}
		return old + "+second"
	})
	if v, _ := m.Get("x"); v != "first+second" {
		t.Errorf("v = %q", v)
	}
}

func TestConcurrentMixed(t *testing.T) {
	m := NewShardedMap[int](32)
	var wg sync.WaitGroup
	for g := 0; g < 64; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2000; i++ {
				k := fmt.Sprint("key", (g*31+i)%500)
				m.Set(k, i)
				m.Get(k)
				if i%7 == 0 {
					m.Delete(k)
				}
				if i%501 == 0 {
					m.Range(func(string, int) bool { return true })
				}
			}
		}()
	}
	wg.Wait()
	if m.Len() > 500 {
		t.Errorf("Len = %d > 500", m.Len())
	}
}
@@ solution
package solution

import (
	"hash/fnv"
	"sync"
)

type shard[V any] struct {
	mu sync.RWMutex
	m  map[string]V
}

type ShardedMap[V any] struct {
	shards []shard[V]
}

func NewShardedMap[V any](shards int) *ShardedMap[V] {
	if shards < 1 {
		shards = 1
	}
	s := &ShardedMap[V]{shards: make([]shard[V], shards)}
	for i := range s.shards {
		s.shards[i].m = make(map[string]V)
	}
	return s
}

func (m *ShardedMap[V]) shardFor(key string) *shard[V] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return &m.shards[h.Sum32()%uint32(len(m.shards))]
}

func (m *ShardedMap[V]) Set(key string, v V) {
	s := m.shardFor(key)
	s.mu.Lock()
	s.m[key] = v
	s.mu.Unlock()
}

func (m *ShardedMap[V]) Get(key string) (V, bool) {
	s := m.shardFor(key)
	s.mu.RLock()
	v, ok := s.m[key]
	s.mu.RUnlock()
	return v, ok
}

func (m *ShardedMap[V]) Delete(key string) {
	s := m.shardFor(key)
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

func (m *ShardedMap[V]) Len() int {
	n := 0
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		n += len(s.m)
		s.mu.RUnlock()
	}
	return n
}

func (m *ShardedMap[V]) Update(key string, f func(old V, exists bool) V) {
	s := m.shardFor(key)
	s.mu.Lock()
	old, ok := s.m[key]
	s.m[key] = f(old, ok)
	s.mu.Unlock()
}

func (m *ShardedMap[V]) Range(f func(key string, v V) bool) {
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		// копируем под локом, вызываем f без лока — чтобы f мог обращаться к карте
		snapshot := make(map[string]V, len(s.m))
		for k, v := range s.m {
			snapshot[k] = v
		}
		s.mu.RUnlock()
		for k, v := range snapshot {
			if !f(k, v) {
				return
			}
		}
	}
}
@@ hint
Структура шарда: `mu sync.RWMutex` + `m map[string]V`. Для `Range` скопируй шард под `RLock` и вызывай `f` уже без блокировки.
