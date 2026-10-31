---
title: Потокобезопасные счётчики
level: medium
---
@@ statement
Сделай тип `Counters` — потокобезопасный набор счётчиков по строковым ключам.

- `NewCounters() *Counters`
- `Inc(key string) int` — увеличить на 1 и вернуть новое значение
- `Add(key string, n int)`
- `Get(key string) int` (нет ключа → 0)
- `Snapshot() map[string]int` — **копия** текущего состояния (изменение копии не влияет на `Counters`)

Используй `sync.RWMutex`: `Get` и `Snapshot` читают (`RLock`), остальные пишут. Тест запускает сотни горутин; обычный `map` без защиты упадёт с `fatal error: concurrent map writes`.
@@ starter
package solution

type Counters struct {
	m map[string]int
}

func NewCounters() *Counters { return &Counters{m: map[string]int{}} }

func (c *Counters) Inc(key string) int      { c.m[key]++; return c.m[key] }
func (c *Counters) Add(key string, n int)   { c.m[key] += n }
func (c *Counters) Get(key string) int      { return c.m[key] }
func (c *Counters) Snapshot() map[string]int { return c.m }
@@ tests
package solution

import (
	"sync"
	"testing"
)

func TestCountersConcurrent(t *testing.T) {
	c := NewCounters()
	var wg sync.WaitGroup
	keys := []string{"a", "b", "c"}
	for g := 0; g < 300; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				c.Inc(keys[(g+i)%3])
				if i%100 == 0 {
					_ = c.Get("a")
					_ = c.Snapshot()
				}
			}
		}()
	}
	wg.Wait()
	total := 0
	for _, k := range keys {
		total += c.Get(k)
	}
	if total != 300*1000 {
		t.Errorf("сумма %d, ожидалось %d (потерянные обновления)", total, 300*1000)
	}
}

func TestCountersSemantics(t *testing.T) {
	c := NewCounters()
	if c.Inc("x") != 1 || c.Inc("x") != 2 {
		t.Error("Inc возвращает новое значение")
	}
	c.Add("x", 10)
	c.Add("y", -3)
	if c.Get("x") != 12 || c.Get("y") != -3 || c.Get("none") != 0 {
		t.Errorf("значения: %d %d", c.Get("x"), c.Get("y"))
	}
	snap := c.Snapshot()
	snap["x"] = 999
	delete(snap, "y")
	if c.Get("x") != 12 || c.Get("y") != -3 {
		t.Error("Snapshot должен возвращать копию")
	}
}
@@ solution
package solution

import "sync"

type Counters struct {
	mu sync.RWMutex
	m  map[string]int
}

func NewCounters() *Counters { return &Counters{m: map[string]int{}} }

func (c *Counters) Inc(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key]++
	return c.m[key]
}

func (c *Counters) Add(key string, n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] += n
}

func (c *Counters) Get(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m[key]
}

func (c *Counters) Snapshot() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int, len(c.m))
	for k, v := range c.m {
		out[k] = v
	}
	return out
}
@@ hint
Добавь в структуру поле `mu sync.RWMutex`. `Snapshot` копирует map под `RLock` (или `maps.Clone`).
