---
title: Обобщённое множество Set[T]
level: medium
---
@@ statement
Сделай `Set[T comparable]` на основе `map[T]struct{}`.

- `NewSet[T comparable](items ...T) Set[T]`
- методы: `Add(v T)`, `Remove(v T)`, `Has(v T) bool`, `Len() int`
- `Items() []T` — элементы в произвольном порядке
- операции возвращают **новое** множество, не изменяя аргументы: `Union(other)`, `Intersect(other)`, `Difference(other)` (элементы `s`, которых нет в `other`)
- `Equal(other Set[T]) bool`
@@ starter
package solution

type Set[T comparable] struct {
	// TODO
}

func NewSet[T comparable](items ...T) Set[T] { return Set[T]{} }

func (s Set[T]) Add(v T)                   {}
func (s Set[T]) Remove(v T)                {}
func (s Set[T]) Has(v T) bool              { return false }
func (s Set[T]) Len() int                  { return 0 }
func (s Set[T]) Items() []T                { return nil }
func (s Set[T]) Union(o Set[T]) Set[T]     { return s }
func (s Set[T]) Intersect(o Set[T]) Set[T] { return s }
func (s Set[T]) Difference(o Set[T]) Set[T] { return s }
func (s Set[T]) Equal(o Set[T]) bool       { return false }
@@ tests
package solution

import (
	"slices"
	"testing"
)

func sorted[T ~int | ~string](s Set[T]) []T {
	items := s.Items()
	slices.Sort(items)
	return items
}

func TestSetBasics(t *testing.T) {
	s := NewSet(1, 2, 2, 3)
	if s.Len() != 3 || !s.Has(2) || s.Has(9) {
		t.Fatalf("len=%d", s.Len())
	}
	s.Add(9)
	s.Remove(1)
	s.Remove(100) // удаление отсутствующего безопасно
	if !slices.Equal(sorted(s), []int{2, 3, 9}) {
		t.Errorf("Items = %v", sorted(s))
	}
}

func TestSetOps(t *testing.T) {
	a, b := NewSet("x", "y", "z"), NewSet("y", "z", "w")
	if got := sorted(a.Union(b)); !slices.Equal(got, []string{"w", "x", "y", "z"}) {
		t.Errorf("Union = %v", got)
	}
	if got := sorted(a.Intersect(b)); !slices.Equal(got, []string{"y", "z"}) {
		t.Errorf("Intersect = %v", got)
	}
	if got := sorted(a.Difference(b)); !slices.Equal(got, []string{"x"}) {
		t.Errorf("Difference = %v", got)
	}
	if a.Len() != 3 || b.Len() != 3 {
		t.Error("операции не должны менять исходные множества")
	}
	if !a.Equal(NewSet("z", "y", "x")) || a.Equal(b) || a.Equal(NewSet("x", "y")) {
		t.Error("Equal работает неверно")
	}
}

func TestZeroSetUsable(t *testing.T) {
	var s Set[int] // нулевое значение: Has/Len/Items не должны паниковать
	if s.Has(1) || s.Len() != 0 || len(s.Items()) != 0 {
		t.Error("нулевой Set должен быть пустым")
	}
}
@@ solution
package solution

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, it := range items {
		s.Add(it)
	}
	return s
}

func (s Set[T]) Add(v T)      { s[v] = struct{}{} }
func (s Set[T]) Remove(v T)   { delete(s, v) }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }
func (s Set[T]) Len() int     { return len(s) }

func (s Set[T]) Items() []T {
	out := make([]T, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func (s Set[T]) Union(o Set[T]) Set[T] {
	out := NewSet(s.Items()...)
	for k := range o {
		out.Add(k)
	}
	return out
}

func (s Set[T]) Intersect(o Set[T]) Set[T] {
	out := NewSet[T]()
	for k := range s {
		if o.Has(k) {
			out.Add(k)
		}
	}
	return out
}

func (s Set[T]) Difference(o Set[T]) Set[T] {
	out := NewSet[T]()
	for k := range s {
		if !o.Has(k) {
			out.Add(k)
		}
	}
	return out
}

func (s Set[T]) Equal(o Set[T]) bool {
	if len(s) != len(o) {
		return false
	}
	for k := range s {
		if !o.Has(k) {
			return false
		}
	}
	return true
}
@@ hint
Самый простой путь: `type Set[T comparable] map[T]struct{}`. Нулевой (nil) map безопасен для чтения.
