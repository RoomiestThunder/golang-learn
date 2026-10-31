---
title: Слайсы: уникальные, вставка, удаление
level: medium
---
@@ statement
Реализуй три функции. **Входной слайс менять нельзя** — возвращай новый (тесты проверяют, что исходник остался прежним).

- `Unique(nums []int) []int` — убрать дубликаты, оставив первое вхождение каждого; порядок сохранить.
- `Insert(s []int, i, v int) []int` — вставить `v` на позицию `i` (`0 <= i <= len(s)`, иначе `panic`).
- `Remove(s []int, i int) []int` — удалить элемент `i` (`0 <= i < len(s)`, иначе `panic`).

Тема урока: разделяемая память слайсов и поведение `append`.
@@ starter
package solution

func Unique(nums []int) []int {
	return nil
}

func Insert(s []int, i, v int) []int {
	return nil
}

func Remove(s []int, i int) []int {
	return nil
}
@@ tests
package solution

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	in := []int{3, 1, 3, 2, 1, 4}
	snapshot := slices.Clone(in)
	got := Unique(in)
	if !slices.Equal(got, []int{3, 1, 2, 4}) {
		t.Errorf("Unique = %v", got)
	}
	if !slices.Equal(in, snapshot) {
		t.Errorf("исходный слайс изменён: %v", in)
	}
	if got := Unique(nil); len(got) != 0 {
		t.Errorf("Unique(nil) = %v", got)
	}
}

func TestInsert(t *testing.T) {
	base := make([]int, 3, 10) // есть запас cap: наивный append затрёт общий массив
	copy(base, []int{1, 2, 3})
	snapshot := slices.Clone(base)
	for _, tc := range []struct {
		i    int
		want []int
	}{{0, []int{9, 1, 2, 3}}, {2, []int{1, 2, 9, 3}}, {3, []int{1, 2, 3, 9}}} {
		got := Insert(base, tc.i, 9)
		if !slices.Equal(got, tc.want) {
			t.Errorf("Insert(i=%d) = %v, want %v", tc.i, got, tc.want)
		}
		if !slices.Equal(base, snapshot) || !slices.Equal(base[:cap(base)][:3], snapshot) {
			t.Fatalf("исходный слайс изменён: %v", base)
		}
	}
	mustPanic(t, func() { Insert(base, 4, 1) })
	mustPanic(t, func() { Insert(base, -1, 1) })
}

func TestRemove(t *testing.T) {
	in := []int{1, 2, 3, 4}
	snapshot := slices.Clone(in)
	for _, tc := range []struct {
		i    int
		want []int
	}{{0, []int{2, 3, 4}}, {2, []int{1, 2, 4}}, {3, []int{1, 2, 3}}} {
		if got := Remove(in, tc.i); !slices.Equal(got, tc.want) {
			t.Errorf("Remove(i=%d) = %v, want %v", tc.i, got, tc.want)
		}
		if !slices.Equal(in, snapshot) {
			t.Fatalf("исходный слайс изменён: %v", in)
		}
	}
	mustPanic(t, func() { Remove(in, 4) })
}

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("ожидалась паника")
		}
	}()
	f()
}
@@ solution
package solution

func Unique(nums []int) []int {
	seen := make(map[int]struct{}, len(nums))
	out := make([]int, 0, len(nums))
	for _, n := range nums {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func Insert(s []int, i, v int) []int {
	if i < 0 || i > len(s) {
		panic("index out of range")
	}
	out := make([]int, 0, len(s)+1)
	out = append(out, s[:i]...)
	out = append(out, v)
	return append(out, s[i:]...)
}

func Remove(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		panic("index out of range")
	}
	out := make([]int, 0, len(s)-1)
	out = append(out, s[:i]...)
	return append(out, s[i+1:]...)
}
@@ hint
`append(s[:i], ...)` пишет в общий массив входного слайса. Сначала создай новый слайс через `make`, затем копируй в него части.
