---
id: l10
week: 2
day: 5
minutes: 120
title: Дженерики и итераторы
goals: Параметры типов, ограничения, обобщённые структуры, пакеты slices/maps/cmp, range-over-func
lc: group-anagrams | Group Anagrams | Medium | Ключ — отсортированная строка или [26]int (массивы годятся как ключ map)
lc: top-k-frequent-elements | Top K Frequent Elements | Medium | Частоты в map, затем сортировка или bucket sort за O(n)
lc: intersection-of-two-arrays | Intersection of Two Arrays | Easy | Множество на map[int]struct{}
---

## Зачем

До Go 1.18 функцию `Max` для `int` и `float64` приходилось писать дважды или терять типобезопасность через `any`. Дженерики решают это: функция/тип **параметризуются типом**, проверка идёт на этапе компиляции.

```go run
package main

import "fmt"

// [T any] — параметр типа; any = без ограничений
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, pred func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if pred(x) {
			out = append(out, x)
		}
	}
	return out
}

func main() {
	nums := []int{1, 2, 3, 4, 5}
	sq := Map(nums, func(x int) int { return x * x })       // T и U выведены компилятором
	strs := Map(nums, func(x int) string { return fmt.Sprint("#", x) })
	even := Filter(nums, func(x int) bool { return x%2 == 0 })
	fmt.Println(sq, strs, even)
	fmt.Println(Map[int, float64](nums, func(x int) float64 { return float64(x) / 2 })) // явная подстановка
}
```

## Ограничения (constraints)

Ограничение — интерфейс, задающий **множество допустимых типов** и/или методов.

```go run
package main

import (
	"cmp"
	"fmt"
)

// встроенные: any (всё), comparable (можно ==, != — ключи map)
func Index[T comparable](xs []T, v T) int {
	for i, x := range xs {
		if x == v {
			return i
		}
	}
	return -1
}

// cmp.Ordered: все типы с <, <=, >, >= (числа и строки)
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// собственное ограничение: ~int значит «любой тип с базовым int» (в том числе type ID int)
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](xs ...T) T {
	var s T // zero value типа T
	for _, x := range xs {
		s += x
	}
	return s
}

type Meters float64

func main() {
	fmt.Println(Index([]string{"a", "b"}, "b"), Max(3, 7), Max("a", "b"), Max(2.5, 1))
	fmt.Println(Sum(1, 2, 3), Sum(Meters(1.5), Meters(2)))
}
```

## Обобщённые типы

```go run
package main

import (
	"errors"
	"fmt"
)

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, error) {
	var zero T
	if len(s.items) == 0 {
		return zero, errors.New("стек пуст")
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, nil
}

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

// Set на базе map — классический пример
type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T)           { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool      { _, ok := s[v]; return ok }

func main() {
	var s Stack[string]
	s.Push("a")
	s.Push("b")
	v, _ := s.Pop()
	fmt.Println(v)

	p := Pair[string, int]{"age", 30}
	fmt.Printf("%+v\n", p)

	set := Set[int]{}
	set.Add(5)
	fmt.Println(set.Has(5), set.Has(6))
}
```

Ограничения дженериков в Go:

- **У методов нет собственных параметров типа** (`func (s *Stack[T]) Map[U any](...)` — нельзя). Пиши обычную функцию.
- Нельзя вызвать метод, которого нет в ограничении; нельзя использовать `+`, `<` без подходящего constraint.
- Нет специализации и перегрузки.

## Когда использовать

✅ Контейнеры (Set, Stack, Cache[K,V]), универсальные алгоритмы (Map/Filter/Reduce, Max), утилиты над слайсами и map.

❌ Когда достаточно интерфейса (`io.Reader`), когда нужна только одна конкретная реализация, «на всякий случай». Правило: **сначала напиши конкретный код, обобщай при повторении**.

## Стандартные пакеты slices, maps, cmp

Это замена самодельным утилитам:

```go run
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	people := []Person{{"Борис", 30}, {"Анна", 25}, {"Вика", 30}}

	// сортировка по нескольким ключам
	slices.SortFunc(people, func(a, b Person) int {
		return cmp.Or(
			cmp.Compare(b.Age, a.Age),       // по возрасту убыв.
			strings.Compare(a.Name, b.Name), // затем по имени
		)
	})
	fmt.Println(people)

	idx := slices.IndexFunc(people, func(p Person) bool { return p.Age < 30 })
	fmt.Println(idx, slices.ContainsFunc(people, func(p Person) bool { return p.Name == "Вика" }))

	xs := []int{3, 1, 2, 3}
	slices.Sort(xs)
	xs = slices.Compact(xs) // убирает подряд идущие дубликаты
	fmt.Println(xs, slices.Max(xs))

	m := map[string]int{"b": 2, "a": 1}
	keys := slices.Sorted(maps.Keys(m)) // Keys возвращает итератор; Sorted собирает и сортирует
	fmt.Println(keys, maps.Clone(m))
}
```

## Итераторы: range over func (Go 1.23)

Итератор — функция, принимающая `yield`. `for range` умеет по ним ходить:

```go run
package main

import (
	"fmt"
	"iter"
	"slices"
)

// iter.Seq[V] = func(yield func(V) bool)
func Count(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i++ {
			if !yield(i) { // false => потребитель сделал break
				return
			}
		}
	}
}

func Filter[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if pred(v) && !yield(v) {
				return
			}
		}
	}
}

func main() {
	for v := range Filter(Count(100), func(x int) bool { return x%7 == 0 }) {
		if v > 30 {
			break
		}
		fmt.Print(v, " ")
	}
	fmt.Println()

	// iter.Seq2[K,V] — пары; slices.All, maps.All, slices.Values, slices.Collect...
	for i, s := range slices.All([]string{"x", "y"}) {
		fmt.Print(i, s, " ")
	}
	fmt.Println(slices.Collect(Count(4)))
}
```

Итераторы позволяют строить ленивые конвейеры без промежуточных слайсов; в стандартной библиотеке: `slices.Values/All/Collect/Sorted`, `maps.Keys/Values/All`, `strings.SplitSeq` (Go 1.24).

## Проверь себя

<details><summary>Чем <code>comparable</code> отличается от <code>any</code>?</summary>comparable допускает только типы, которые можно сравнивать через == (нужно для ключей map и поиска), any — вообще любые.</details>

<details><summary>Что значит <code>~int</code> в ограничении?</summary>Любой тип, у которого базовый тип int (в том числе пользовательские <code>type ID int</code>).</details>

<details><summary>Почему у методов нет собственных type-параметров?</summary>Ограничение дизайна Go: параметры типа допустимы у функций и типов; методы используют параметры своего типа-получателя.</details>

<details><summary>Как вернуть «нулевое значение» в обобщённой функции?</summary><code>var zero T; return zero</code>.</details>
