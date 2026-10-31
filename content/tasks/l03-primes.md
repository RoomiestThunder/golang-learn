---
title: Простые числа (решето)
level: medium
---
@@ statement
Реализуй `Primes(n int) []int` — все простые числа **до n включительно** в порядке возрастания (решето Эратосфена). Для `n < 2` верни пустой (не обязательно nil) слайс.

Проверка идёт и на скорость: `Primes(2_000_000)` должен отработать быстро (наивная проверка делением каждого числа не уложится).
@@ starter
package solution

// Primes возвращает все простые числа от 2 до n включительно.
func Primes(n int) []int {
	return nil
}
@@ tests
package solution

import (
	"slices"
	"testing"
	"time"
)

func TestPrimesSmall(t *testing.T) {
	tests := []struct {
		n    int
		want []int
	}{
		{0, nil}, {1, nil}, {2, []int{2}}, {3, []int{2, 3}},
		{10, []int{2, 3, 5, 7}}, {30, []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}},
	}
	for _, tt := range tests {
		got := Primes(tt.n)
		if len(got) == 0 && len(tt.want) == 0 {
			continue
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("Primes(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestPrimesBig(t *testing.T) {
	start := time.Now()
	got := Primes(2_000_000)
	if len(got) != 148933 {
		t.Fatalf("простых до 2e6 должно быть 148933, got %d", len(got))
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("слишком медленно: %v", d)
	}
}
@@ solution
package solution

func Primes(n int) []int {
	if n < 2 {
		return nil
	}
	composite := make([]bool, n+1)
	var primes []int
	for i := 2; i <= n; i++ {
		if composite[i] {
			continue
		}
		primes = append(primes, i)
		for j := i * i; j <= n; j += i {
			composite[j] = true
		}
	}
	return primes
}
@@ hint
Заведи `[]bool` длины n+1. Нашёл простое `i` — вычёркивай кратные, начиная с `i*i`.
