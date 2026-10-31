---
title: Worker pool с порядком результатов
level: medium
---
@@ statement
Реализуй обобщённую функцию

```go
func RunPool[T, R any](items []T, workers int, f func(T) R) []R
```

- Обрабатывает `items` пулом из `workers` горутин (`workers < 1` → 1).
- Результаты возвращаются **в порядке входных элементов** (результат `items[i]` лежит в `out[i]`).
- Одновременно выполняется **не более `workers`** вызовов `f` (тест это измеряет).
- Должна работать для пустого входа.

Подсказка: раздай воркерам **индексы** через канал, а результат пиши в `out[i]` — гонки не будет, так как каждый индекс обрабатывает один воркер.
@@ starter
package solution

func RunPool[T, R any](items []T, workers int, f func(T) R) []R {
	out := make([]R, len(items))
	for i, it := range items {
		out[i] = f(it)
	}
	return out
}
@@ tests
package solution

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunPoolOrder(t *testing.T) {
	in := make([]int, 50)
	want := make([]string, 50)
	for i := range in {
		in[i] = i
		want[i] = string(rune('a' + i%26))
	}
	got := RunPool(in, 7, func(i int) string {
		time.Sleep(time.Duration(50-i) * 100 * time.Microsecond) // поздние заканчивают раньше
		return string(rune('a' + i%26))
	})
	if !slices.Equal(got, want) {
		t.Errorf("порядок результатов нарушен: %v", got)
	}
	if got := RunPool([]int{}, 3, func(i int) int { return i }); len(got) != 0 {
		t.Errorf("пустой вход: %v", got)
	}
	if got := RunPool([]int{1, 2}, 0, func(i int) int { return i * 2 }); !slices.Equal(got, []int{2, 4}) {
		t.Errorf("workers=0: %v", got)
	}
}

func TestRunPoolConcurrencyLimit(t *testing.T) {
	var cur, peak atomic.Int32
	f := func(int) int {
		n := cur.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		cur.Add(-1)
		return 0
	}
	RunPool(make([]int, 30), 4, f)
	if p := peak.Load(); p > 4 {
		t.Errorf("одновременно выполнялось %d вызовов, лимит 4", p)
	}
	if p := peak.Load(); p < 3 {
		t.Errorf("пул не параллелит: пик %d", p)
	}
}
@@ solution
package solution

import "sync"

func RunPool[T, R any](items []T, workers int, f func(T) R) []R {
	if workers < 1 {
		workers = 1
	}
	out := make([]R, len(items))
	idx := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < min(workers, max(len(items), 1)); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range idx {
				out[i] = f(items[i])
			}
		}()
	}
	for i := range items {
		idx <- i
	}
	close(idx)
	wg.Wait()
	return out
}
@@ hint
Запусти `workers` горутин, читающих индексы из канала `idx`; в главной горутине отправь все индексы, закрой канал и `wg.Wait()`.
