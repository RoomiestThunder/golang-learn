---
title: Map / Filter / Reduce / GroupBy
level: medium
---
@@ statement
Напиши обобщённые функции:

- `Map[T, U any](xs []T, f func(T) U) []U`
- `Filter[T any](xs []T, pred func(T) bool) []T`
- `Reduce[T, A any](xs []T, init A, f func(A, T) A) A`
- `GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T` (порядок внутри групп сохраняется)
- `Chunk[T any](xs []T, size int) [][]T` — разбить на куски по `size` (последний может быть короче); `size <= 0` → `panic`.
- `Keys[K comparable, V any](m map[K]V) []K` — ключи в произвольном порядке.

Входные слайсы не изменяй. Для пустого входа `Map`/`Filter` возвращают пустой результат.
@@ starter
package solution

func Map[T, U any](xs []T, f func(T) U) []U { return nil }

func Filter[T any](xs []T, pred func(T) bool) []T { return nil }

func Reduce[T, A any](xs []T, init A, f func(A, T) A) A { return init }

func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T { return nil }

func Chunk[T any](xs []T, size int) [][]T { return nil }

func Keys[K comparable, V any](m map[K]V) []K { return nil }
@@ tests
package solution

import (
	"reflect"
	"slices"
	"strconv"
	"testing"
)

func TestMapFilterReduce(t *testing.T) {
	xs := []int{1, 2, 3, 4}
	if got := Map(xs, strconv.Itoa); !slices.Equal(got, []string{"1", "2", "3", "4"}) {
		t.Errorf("Map = %v", got)
	}
	if got := Filter(xs, func(x int) bool { return x%2 == 0 }); !slices.Equal(got, []int{2, 4}) {
		t.Errorf("Filter = %v", got)
	}
	if got := Reduce(xs, "", func(acc string, x int) string { return acc + strconv.Itoa(x) }); got != "1234" {
		t.Errorf("Reduce = %q", got)
	}
	if got := Reduce([]int(nil), 7, func(a, x int) int { return a + x }); got != 7 {
		t.Errorf("Reduce(nil) = %d", got)
	}
	if len(Map([]int{}, strconv.Itoa)) != 0 || len(Filter([]int(nil), func(int) bool { return true })) != 0 {
		t.Error("пустой вход → пустой результат")
	}
}

func TestGroupBy(t *testing.T) {
	words := []string{"apple", "avocado", "banana", "blueberry", "cherry"}
	got := GroupBy(words, func(s string) byte { return s[0] })
	want := map[byte][]string{'a': {"apple", "avocado"}, 'b': {"banana", "blueberry"}, 'c': {"cherry"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v", got)
	}
}

func TestChunk(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4, 5}, 2)
	if !reflect.DeepEqual(got, [][]int{{1, 2}, {3, 4}, {5}}) {
		t.Errorf("Chunk = %v", got)
	}
	if len(Chunk([]int{}, 3)) != 0 {
		t.Error("Chunk пустого")
	}
	defer func() {
		if recover() == nil {
			t.Error("size=0 должен паниковать")
		}
	}()
	Chunk([]int{1}, 0)
}

func TestKeys(t *testing.T) {
	got := Keys(map[string]int{"a": 1, "b": 2, "c": 3})
	slices.Sort(got)
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("Keys = %v", got)
	}
}
@@ solution
package solution

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, pred func(T) bool) []T {
	out := make([]T, 0)
	for _, x := range xs {
		if pred(x) {
			out = append(out, x)
		}
	}
	return out
}

func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, x := range xs {
		k := key(x)
		out[k] = append(out[k], x)
	}
	return out
}

func Chunk[T any](xs []T, size int) [][]T {
	if size <= 0 {
		panic("Chunk: size must be positive")
	}
	var out [][]T
	for i := 0; i < len(xs); i += size {
		out = append(out, xs[i:min(i+size, len(xs))])
	}
	return out
}

func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
@@ hint
`Chunk`: шаг `i += size`, конец куска `min(i+size, len(xs))`.
