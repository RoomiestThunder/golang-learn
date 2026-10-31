---
title: Top-K через container/heap
level: medium
---
@@ statement
1. `TopK(nums []int, k int) []int` — `k` наибольших элементов по **убыванию**. Используй `container/heap` с **мин-кучей размера k** (сложность O(n log k), для миллиона элементов должно работать быстро). `k <= 0` → пустой результат; `k > len(nums)` → все элементы по убыванию. Входной слайс не менять.
2. `KthLargest` в стиле LeetCode:
   - `NewKthLargest(k int, nums []int) *KthLargest`
   - `(*KthLargest).Add(v int) int` — добавляет значение и возвращает текущий k-й по величине элемент потока.

   Гарантируется `k >= 1`. Если элементов пока меньше k, `Add` вернёт `0` (в тестах такой случай не встречается).
@@ starter
package solution

func TopK(nums []int, k int) []int {
	return nil
}

type KthLargest struct{}

func NewKthLargest(k int, nums []int) *KthLargest { return &KthLargest{} }

func (kl *KthLargest) Add(v int) int { return 0 }
@@ tests
package solution

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"
)

func TestTopK(t *testing.T) {
	in := []int{5, 1, 9, 3, 7, 9, 2}
	snapshot := slices.Clone(in)
	if got := TopK(in, 3); !slices.Equal(got, []int{9, 9, 7}) {
		t.Errorf("TopK(3) = %v", got)
	}
	if !slices.Equal(in, snapshot) {
		t.Error("входной слайс изменён")
	}
	if got := TopK(in, 100); !slices.Equal(got, []int{9, 9, 7, 5, 3, 2, 1}) {
		t.Errorf("TopK(100) = %v", got)
	}
	if got := TopK(in, 0); len(got) != 0 {
		t.Errorf("TopK(0) = %v", got)
	}
	if got := TopK(nil, 3); len(got) != 0 {
		t.Errorf("TopK(nil) = %v", got)
	}
}

func TestTopKBig(t *testing.T) {
	nums := make([]int, 1_000_000)
	for i := range nums {
		nums[i] = rand.IntN(1 << 30)
	}
	want := slices.Clone(nums)
	slices.SortFunc(want, func(a, b int) int { return b - a })
	start := time.Now()
	got := TopK(nums, 10)
	if d := time.Since(start); d > time.Second {
		t.Errorf("слишком медленно: %v", d)
	}
	if !slices.Equal(got, want[:10]) {
		t.Errorf("got %v, want %v", got, want[:10])
	}
}

func TestKthLargest(t *testing.T) {
	kl := NewKthLargest(3, []int{4, 5, 8, 2})
	for _, tc := range []struct{ add, want int }{{3, 4}, {5, 5}, {10, 5}, {9, 8}, {4, 8}} {
		if got := kl.Add(tc.add); got != tc.want {
			t.Errorf("Add(%d) = %d, want %d", tc.add, got, tc.want)
		}
	}
	kl2 := NewKthLargest(1, nil)
	if kl2.Add(-1) != -1 || kl2.Add(5) != 5 || kl2.Add(2) != 5 {
		t.Error("k=1 работает как максимум")
	}
}
@@ solution
package solution

import (
	"container/heap"
	"slices"
)

type minHeap []int

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *minHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func TopK(nums []int, k int) []int {
	if k <= 0 {
		return nil
	}
	h := &minHeap{}
	for _, n := range nums {
		if h.Len() < k {
			heap.Push(h, n)
		} else if n > (*h)[0] {
			(*h)[0] = n
			heap.Fix(h, 0)
		}
	}
	out := []int(*h)
	out = slices.Clone(out)
	slices.SortFunc(out, func(a, b int) int { return b - a })
	return out
}

type KthLargest struct {
	k int
	h *minHeap
}

func NewKthLargest(k int, nums []int) *KthLargest {
	kl := &KthLargest{k: k, h: &minHeap{}}
	for _, n := range nums {
		kl.Add(n)
	}
	return kl
}

func (kl *KthLargest) Add(v int) int {
	if kl.h.Len() < kl.k {
		heap.Push(kl.h, v)
	} else if v > (*kl.h)[0] {
		(*kl.h)[0] = v
		heap.Fix(kl.h, 0)
	}
	if kl.h.Len() < kl.k {
		return 0
	}
	return (*kl.h)[0]
}
@@ hint
Мин-куча размера k: если элемент больше корня — замени корень и вызови `heap.Fix(h, 0)`. В конце корень — k-й по величине; для результата `TopK` отсортируй копию содержимого кучи по убыванию.
