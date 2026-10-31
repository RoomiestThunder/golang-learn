---
title: MinMax и variadic
level: easy
---
@@ statement
Напиши `MinMax(nums ...int) (lo, hi int, err error)` с **именованными** результатами.

- Для пустого вызова верни ошибку (`errors.New`), `lo` и `hi` — нули.
- Иначе минимум и максимум всех аргументов.
- Должно работать и как `MinMax(3, 1, 2)`, и как `MinMax(xs...)`.

Дополнительно реализуй `Apply(nums []int, fns ...func(int) int) []int` — применяет все функции по порядку к каждому элементу и возвращает **новый** слайс (исходный не меняется).
@@ starter
package solution

func MinMax(nums ...int) (lo, hi int, err error) {
	return
}

func Apply(nums []int, fns ...func(int) int) []int {
	return nil
}
@@ tests
package solution

import (
	"slices"
	"testing"
)

func TestMinMax(t *testing.T) {
	lo, hi, err := MinMax(3, -1, 7, 2)
	if err != nil || lo != -1 || hi != 7 {
		t.Errorf("got %d %d %v", lo, hi, err)
	}
	xs := []int{5}
	lo, hi, err = MinMax(xs...)
	if err != nil || lo != 5 || hi != 5 {
		t.Errorf("single: %d %d %v", lo, hi, err)
	}
	if _, _, err = MinMax(); err == nil {
		t.Error("для пустого вызова нужна ошибка")
	}
}

func TestApply(t *testing.T) {
	in := []int{1, 2, 3}
	got := Apply(in, func(x int) int { return x + 1 }, func(x int) int { return x * 10 })
	if !slices.Equal(got, []int{20, 30, 40}) {
		t.Errorf("got %v", got)
	}
	if !slices.Equal(in, []int{1, 2, 3}) {
		t.Errorf("исходный слайс изменён: %v", in)
	}
	if got := Apply(in); !slices.Equal(got, in) {
		t.Errorf("без функций результат должен равняться входу, got %v", got)
	}
}
@@ solution
package solution

import "errors"

func MinMax(nums ...int) (lo, hi int, err error) {
	if len(nums) == 0 {
		err = errors.New("нет аргументов")
		return
	}
	lo, hi = nums[0], nums[0]
	for _, n := range nums[1:] {
		lo, hi = min(lo, n), max(hi, n)
	}
	return
}

func Apply(nums []int, fns ...func(int) int) []int {
	out := make([]int, len(nums))
	for i, n := range nums {
		for _, f := range fns {
			n = f(n)
		}
		out[i] = n
	}
	return out
}
@@ hint
Для `Apply` создай `out := make([]int, len(nums))` и для каждого элемента прогони цепочку функций.
