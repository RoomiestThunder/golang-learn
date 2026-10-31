---
title: Гипотеза Коллатца
level: easy
---
@@ statement
`CollatzSteps(n int) (int, error)` — число шагов, за которое `n` приходит к 1, если:

- чётное `n` делим на 2,
- нечётное заменяем на `3n + 1`.

`CollatzSteps(1) == 0`, `CollatzSteps(6) == 8` (6→3→10→5→16→8→4→2→1).
Для `n <= 0` верни ошибку (последовательность не определена) и `0`.
@@ starter
package solution

func CollatzSteps(n int) (int, error) {
	return 0, nil
}
@@ tests
package solution

import "testing"

func TestCollatz(t *testing.T) {
	tests := []struct{ n, want int }{{1, 0}, {2, 1}, {6, 8}, {7, 16}, {27, 111}, {97, 118}}
	for _, tt := range tests {
		got, err := CollatzSteps(tt.n)
		if err != nil || got != tt.want {
			t.Errorf("CollatzSteps(%d) = %d, %v; want %d", tt.n, got, err, tt.want)
		}
	}
}

func TestCollatzInvalid(t *testing.T) {
	for _, n := range []int{0, -5} {
		got, err := CollatzSteps(n)
		if err == nil || got != 0 {
			t.Errorf("CollatzSteps(%d) = %d, %v; нужна ошибка и 0", n, got, err)
		}
	}
}
@@ solution
package solution

import "errors"

func CollatzSteps(n int) (int, error) {
	if n <= 0 {
		return 0, errors.New("n должно быть положительным")
	}
	steps := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		steps++
	}
	return steps, nil
}
@@ hint
Цикл `for n != 1 { ... }` и счётчик шагов.
