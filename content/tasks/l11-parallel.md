---
title: Параллельная сумма и Gather
level: medium
---
@@ statement
1. `ParallelSum(nums []int, workers int) int` — раздели слайс на `workers` примерно равных кусков, посчитай суммы кусков в отдельных горутинах (каждая пишет в **свой** элемент слайса результатов — без мьютекса) и сложи. `workers <= 0` считай как 1; воркеров не больше, чем элементов.
2. `Gather(fns []func() int) []int` — запусти **все** функции одновременно и верни результаты **в порядке исходного слайса**. Тест проверяет одновременность: каждая функция ждёт, пока все остальные стартуют.

Используй `sync.WaitGroup`. Никаких общих переменных без синхронизации.
@@ starter
package solution

func ParallelSum(nums []int, workers int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

func Gather(fns []func() int) []int {
	out := make([]int, len(fns))
	for i, f := range fns {
		out[i] = f() // TODO: параллельно
	}
	return out
}
@@ tests
package solution

import (
	"slices"
	"sync"
	"testing"
	"time"
)

func TestParallelSum(t *testing.T) {
	nums := make([]int, 100_001)
	want := 0
	for i := range nums {
		nums[i] = i
		want += i
	}
	for _, w := range []int{-3, 0, 1, 4, 7, 100_001, 500_000} {
		if got := ParallelSum(nums, w); got != want {
			t.Errorf("workers=%d: got %d, want %d", w, got, want)
		}
	}
	if ParallelSum(nil, 4) != 0 || ParallelSum([]int{5}, 8) != 5 {
		t.Error("крайние случаи")
	}
}

func TestGatherIsConcurrent(t *testing.T) {
	const n = 8
	var started sync.WaitGroup
	started.Add(n)
	all := make(chan struct{})
	go func() { started.Wait(); close(all) }()

	fns := make([]func() int, n)
	for i := range fns {
		fns[i] = func() int {
			started.Done()
			select {
			case <-all: // дождались старта всех остальных
				return i * 10
			case <-time.After(500 * time.Millisecond):
				return -1 // кто-то не стартовал — выполнение последовательное
			}
		}
	}
	got := Gather(fns)
	want := []int{0, 10, 20, 30, 40, 50, 60, 70}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v (функции должны выполняться одновременно, результат — в исходном порядке)", got, want)
	}
}
@@ solution
package solution

import "sync"

func ParallelSum(nums []int, workers int) int {
	if workers <= 0 {
		workers = 1
	}
	workers = min(workers, len(nums))
	if workers == 0 {
		return 0
	}
	partial := make([]int, workers)
	chunk := (len(nums) + workers - 1) / workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		lo, hi := w*chunk, min((w+1)*chunk, len(nums))
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, n := range nums[lo:hi] {
				partial[w] += n
			}
		}()
	}
	wg.Wait()
	total := 0
	for _, p := range partial {
		total += p
	}
	return total
}

func Gather(fns []func() int) []int {
	out := make([]int, len(fns))
	var wg sync.WaitGroup
	for i, f := range fns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = f()
		}()
	}
	wg.Wait()
	return out
}
@@ hint
Размер куска: `chunk := (len(nums)+workers-1)/workers`. Результат горутины `i` кладётся в `out[i]` — разные элементы слайса безопасно писать параллельно.
