---
title: Своя обобщённая куча и слияние k списков
level: hard
---
@@ statement
Реализуй **без `container/heap`** обобщённую двоичную кучу на слайсе:

```go
func NewHeap[T any](less func(a, b T) bool) *Heap[T]   // less(a,b) == true, если a должен выйти раньше b
func (h *Heap[T]) Push(v T)
func (h *Heap[T]) Pop() (T, bool)     // вынуть «наименьший» по less; (нулевое значение, false), если пусто
func (h *Heap[T]) Peek() (T, bool)
func (h *Heap[T]) Len() int
```

Push/Pop — O(log n) (просеивание вверх `siftUp` и вниз `siftDown`). Для max-кучи достаточно передать `less` наоборот.

На её основе:

- `MergeK(lists [][]int) []int` — слей `k` отсортированных слайсов в один отсортированный за O(N log k) (куча хранит текущие «головы» списков).
- `KSmallestPairs`-подобная мини-задача: `TopKFrequent(nums []int, k int) []int` — `k` самых частых чисел (порядок: по убыванию частоты, при равенстве — по возрастанию значения). Используй кучу размера `k`.
@@ starter
package solution

type Heap[T any] struct {
	// TODO
}

func NewHeap[T any](less func(a, b T) bool) *Heap[T] { return &Heap[T]{} }

func (h *Heap[T]) Push(v T) {}

func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	return zero, false
}

func (h *Heap[T]) Peek() (T, bool) {
	var zero T
	return zero, false
}

func (h *Heap[T]) Len() int { return 0 }

func MergeK(lists [][]int) []int { return nil }

func TopKFrequent(nums []int, k int) []int { return nil }
@@ tests
package solution

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"
	"time"
)

func TestHeapSorts(t *testing.T) {
	h := NewHeap(func(a, b int) bool { return a < b })
	if _, ok := h.Pop(); ok {
		t.Error("Pop пустой кучи")
	}
	if _, ok := h.Peek(); ok {
		t.Error("Peek пустой кучи")
	}
	in := []int{5, 3, 9, 1, 1, 7, -2, 0}
	for _, x := range in {
		h.Push(x)
	}
	if v, _ := h.Peek(); v != -2 || h.Len() != len(in) {
		t.Errorf("Peek = %d, Len = %d", v, h.Len())
	}
	var out []int
	for h.Len() > 0 {
		v, ok := h.Pop()
		if !ok {
			t.Fatal("Pop вернул false при Len > 0")
		}
		out = append(out, v)
	}
	want := slices.Clone(in)
	slices.Sort(want)
	if !slices.Equal(out, want) {
		t.Errorf("got %v, want %v", out, want)
	}
}

func TestHeapMaxAndStructs(t *testing.T) {
	type task struct {
		prio int
		name string
	}
	h := NewHeap(func(a, b task) bool { return a.prio > b.prio }) // max-куча
	for i, n := range []string{"low", "high", "mid"} {
		h.Push(task{[]int{1, 10, 5}[i], n})
	}
	var names []string
	for h.Len() > 0 {
		v, _ := h.Pop()
		names = append(names, v.name)
	}
	if !slices.Equal(names, []string{"high", "mid", "low"}) {
		t.Errorf("got %v", names)
	}
}

func TestHeapBig(t *testing.T) {
	h := NewHeap(func(a, b int) bool { return a < b })
	nums := make([]int, 500_000)
	for i := range nums {
		nums[i] = rand.IntN(1 << 30)
		h.Push(nums[i])
	}
	sort.Ints(nums)
	start := time.Now()
	for i := 0; i < 1000; i++ {
		if v, _ := h.Pop(); v != nums[i] {
			t.Fatalf("на шаге %d: %d, want %d", i, v, nums[i])
		}
	}
	if time.Since(start) > time.Second {
		t.Error("Pop слишком медленный")
	}
}

func TestMergeK(t *testing.T) {
	got := MergeK([][]int{{1, 4, 7}, {}, {2, 5, 8}, {0, 3, 6, 9}, {5}})
	if !slices.Equal(got, []int{0, 1, 2, 3, 4, 5, 5, 6, 7, 8, 9}) {
		t.Errorf("got %v", got)
	}
	if len(MergeK(nil)) != 0 || len(MergeK([][]int{{}, {}})) != 0 {
		t.Error("пустые входы")
	}
	big := make([][]int, 200)
	var all []int
	for i := range big {
		s := make([]int, 2000)
		for j := range s {
			s[j] = rand.IntN(1_000_000)
		}
		slices.Sort(s)
		big[i] = s
		all = append(all, s...)
	}
	slices.Sort(all)
	start := time.Now()
	if !slices.Equal(MergeK(big), all) {
		t.Error("большое слияние неверно")
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("слишком медленно: %v", time.Since(start))
	}
}

func TestTopKFrequent(t *testing.T) {
	got := TopKFrequent([]int{1, 1, 1, 2, 2, 3, 4, 4, 4, 4}, 2)
	if !slices.Equal(got, []int{4, 1}) {
		t.Errorf("got %v", got)
	}
	if got := TopKFrequent([]int{5, 3, 3, 5, 9}, 3); !slices.Equal(got, []int{3, 5, 9}) {
		t.Errorf("ничьи по возрастанию значения: %v", got)
	}
	if got := TopKFrequent([]int{1}, 5); !slices.Equal(got, []int{1}) {
		t.Errorf("k больше числа различных: %v", got)
	}
}
@@ solution
package solution

type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

func NewHeap[T any](less func(a, b T) bool) *Heap[T] { return &Heap[T]{less: less} }

func (h *Heap[T]) Len() int { return len(h.data) }

func (h *Heap[T]) Push(v T) {
	h.data = append(h.data, v)
	i := len(h.data) - 1
	for i > 0 {
		p := (i - 1) / 2
		if !h.less(h.data[i], h.data[p]) {
			break
		}
		h.data[i], h.data[p] = h.data[p], h.data[i]
		i = p
	}
}

func (h *Heap[T]) Peek() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	return h.data[0], true
}

func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	n := len(h.data)
	if n == 0 {
		return zero, false
	}
	top := h.data[0]
	h.data[0] = h.data[n-1]
	h.data[n-1] = zero
	h.data = h.data[:n-1]
	h.siftDown(0)
	return top, true
}

func (h *Heap[T]) siftDown(i int) {
	n := len(h.data)
	for {
		l, r, smallest := 2*i+1, 2*i+2, i
		if l < n && h.less(h.data[l], h.data[smallest]) {
			smallest = l
		}
		if r < n && h.less(h.data[r], h.data[smallest]) {
			smallest = r
		}
		if smallest == i {
			return
		}
		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		i = smallest
	}
}

func MergeK(lists [][]int) []int {
	type head struct{ val, list, idx int }
	h := NewHeap(func(a, b head) bool { return a.val < b.val })
	total := 0
	for i, l := range lists {
		total += len(l)
		if len(l) > 0 {
			h.Push(head{l[0], i, 0})
		}
	}
	out := make([]int, 0, total)
	for h.Len() > 0 {
		x, _ := h.Pop()
		out = append(out, x.val)
		if next := x.idx + 1; next < len(lists[x.list]) {
			h.Push(head{lists[x.list][next], x.list, next})
		}
	}
	return out
}

func TopKFrequent(nums []int, k int) []int {
	freq := map[int]int{}
	for _, n := range nums {
		freq[n]++
	}
	type pair struct{ val, cnt int }
	// «худший» из кандидатов — в корне: меньше частота, при равенстве — больше значение
	h := NewHeap(func(a, b pair) bool {
		if a.cnt != b.cnt {
			return a.cnt < b.cnt
		}
		return a.val > b.val
	})
	for v, c := range freq {
		h.Push(pair{v, c})
		if h.Len() > k {
			h.Pop()
		}
	}
	out := make([]int, h.Len())
	for i := len(out) - 1; i >= 0; i-- {
		p, _ := h.Pop()
		out[i] = p.val
	}
	return out
}
@@ hint
Куча в слайсе: родитель `(i-1)/2`, дети `2i+1`, `2i+2`. `Push` — добавить в конец и `siftUp`; `Pop` — забрать корень, поставить на его место последний и `siftDown`. Для `MergeK` положи в кучу структуры `{значение, номер списка, индекс}`.
