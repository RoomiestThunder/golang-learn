---
title: Кэш с единственной загрузкой
level: hard
---
@@ statement
Реализуй `Cache` с методом `GetOrLoad(key string, load func(string) (string, error)) (string, error)`.

Требования:
1. Значение найдено в кэше — вернуть без вызова `load`.
2. Если ключа нет — вызвать `load(key)` и сохранить результат.
3. **Если сто горутин одновременно просят один и тот же отсутствующий ключ, `load` вызывается ровно один раз**, остальные ждут и получают тот же результат (защита от «cache stampede», так работает `singleflight`).
4. Ошибки **не кэшируются**: после ошибки следующий вызов снова пробует загрузить. Все ожидавшие одну загрузку получают её ошибку.
5. Разные ключи грузятся независимо и параллельно (не блокируй весь кэш на время `load`).

Нулевое значение `Cache` должно быть пригодно к использованию (инициализируй map лениво).
@@ starter
package solution

type Cache struct {
	// TODO
}

func (c *Cache) GetOrLoad(key string, load func(string) (string, error)) (string, error) {
	return load(key)
}
@@ tests
package solution

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadOnce(t *testing.T) {
	var c Cache
	var calls atomic.Int32
	load := func(k string) (string, error) {
		calls.Add(1)
		time.Sleep(50 * time.Millisecond)
		return "value-" + k, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := c.GetOrLoad("k", load)
			if err != nil || v != "value-k" {
				t.Errorf("got %q, %v", v, err)
			}
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Errorf("load вызвана %d раз, ожидался 1", n)
	}
	// из кэша
	c.GetOrLoad("k", load)
	if calls.Load() != 1 {
		t.Error("повторный запрос должен браться из кэша")
	}
}

func TestErrorsNotCached(t *testing.T) {
	var c Cache
	var calls atomic.Int32
	boom := errors.New("boom")
	load := func(k string) (string, error) {
		if calls.Add(1) == 1 {
			return "", boom
		}
		return "ok", nil
	}
	if _, err := c.GetOrLoad("k", load); !errors.Is(err, boom) {
		t.Fatalf("первая ошибка: %v", err)
	}
	if v, err := c.GetOrLoad("k", load); err != nil || v != "ok" {
		t.Fatalf("после ошибки ожидали повторную загрузку: %q %v", v, err)
	}
}

func TestDifferentKeysParallel(t *testing.T) {
	var c Cache
	load := func(k string) (string, error) { time.Sleep(200 * time.Millisecond); return k, nil }
	start := time.Now()
	var wg sync.WaitGroup
	for _, k := range []string{"a", "b", "c", "d"} {
		wg.Add(1)
		go func() { defer wg.Done(); c.GetOrLoad(k, load) }()
	}
	wg.Wait()
	if d := time.Since(start); d > 600*time.Millisecond {
		t.Errorf("разные ключи грузятся последовательно: %v", d)
	}
}
@@ solution
package solution

import "sync"

type call struct {
	wg  sync.WaitGroup
	val string
	err error
}

type Cache struct {
	mu       sync.Mutex
	data     map[string]string
	inflight map[string]*call
}

func (c *Cache) GetOrLoad(key string, load func(string) (string, error)) (string, error) {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[string]string{}
		c.inflight = map[string]*call{}
	}
	if v, ok := c.data[key]; ok {
		c.mu.Unlock()
		return v, nil
	}
	if cl, ok := c.inflight[key]; ok { // кто-то уже грузит — ждём его
		c.mu.Unlock()
		cl.wg.Wait()
		return cl.val, cl.err
	}
	cl := &call{}
	cl.wg.Add(1)
	c.inflight[key] = cl
	c.mu.Unlock()

	cl.val, cl.err = load(key) // вне мьютекса

	c.mu.Lock()
	if cl.err == nil {
		c.data[key] = cl.val
	}
	delete(c.inflight, key)
	c.mu.Unlock()
	cl.wg.Done()
	return cl.val, cl.err
}
@@ hint
Хранить нужно два map: готовые значения и «загрузки в процессе» (`*call` с `WaitGroup`). Мьютекс держи только на проверку/запись map, не на время `load`.
