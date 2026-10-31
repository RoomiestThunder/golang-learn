---
title: Окно и бинарный поиск
level: medium
---
@@ statement
Пять функций (на скорость: наивные решения не пройдут по времени).

1. `LongestUniqueSubstring(s string) int` — длина самой длинной подстроки без повторяющихся символов (символы — **руны**).
2. `MinSubArrayLen(target int, nums []int) int` — минимальная длина непрерывного подмассива положительных чисел с суммой `>= target`; `0`, если такого нет.
3. `LowerBound(nums []int, target int) int` — для отсортированного по возрастанию массива первый индекс `i`, где `nums[i] >= target` (или `len(nums)`). За O(log n).
4. `MinEatingSpeed(piles []int, h int) int` — минимальная скорость `k` (бананов в час), чтобы съесть все кучи за `h` часов; за час съедается одна куча или её часть, не больше `k`: часов на кучу `p` нужно `ceil(p/k)`. (`len(piles) <= h`.)
5. `SearchRotated(nums []int, target int) int` — индекс `target` в **циклически сдвинутом** отсортированном массиве уникальных чисел (`[4,5,6,7,0,1,2]`), или `-1`. За O(log n).
@@ starter
package solution

func LongestUniqueSubstring(s string) int { return 0 }

func MinSubArrayLen(target int, nums []int) int { return 0 }

func LowerBound(nums []int, target int) int { return 0 }

func MinEatingSpeed(piles []int, h int) int { return 0 }

func SearchRotated(nums []int, target int) int { return -1 }
@@ tests
package solution

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"
	"time"
)

func TestLongestUniqueSubstring(t *testing.T) {
	tests := map[string]int{"": 0, "a": 1, "abcabcbb": 3, "bbbbb": 1, "pwwkew": 3, "abba": 2, "dvdf": 3, "привет мир": 8, "🙂🙂a🙂b": 3}
	for in, want := range tests {
		if got := LongestUniqueSubstring(in); got != want {
			t.Errorf("LongestUniqueSubstring(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestMinSubArrayLen(t *testing.T) {
	tests := []struct {
		target int
		nums   []int
		want   int
	}{
		{7, []int{2, 3, 1, 2, 4, 3}, 2}, {4, []int{1, 4, 4}, 1}, {11, []int{1, 1, 1, 1, 1, 1, 1, 1}, 0}, {15, []int{1, 2, 3, 4, 5}, 5}, {3, nil, 0},
	}
	for _, tt := range tests {
		if got := MinSubArrayLen(tt.target, tt.nums); got != tt.want {
			t.Errorf("MinSubArrayLen(%d, %v) = %d, want %d", tt.target, tt.nums, got, tt.want)
		}
	}
	big := make([]int, 1_000_000)
	for i := range big {
		big[i] = 1 + i%3
	}
	start := time.Now()
	if got := MinSubArrayLen(1_500_000, big); got == 0 {
		t.Error("big: должно быть найдено")
	}
	if time.Since(start) > time.Second {
		t.Errorf("слишком медленно: %v", time.Since(start))
	}
}

func TestLowerBound(t *testing.T) {
	a := []int{1, 3, 3, 3, 5, 8}
	tests := map[int]int{0: 0, 1: 0, 2: 1, 3: 1, 4: 4, 8: 5, 9: 6}
	for target, want := range tests {
		if got := LowerBound(a, target); got != want {
			t.Errorf("LowerBound(%d) = %d, want %d", target, got, want)
		}
	}
	if LowerBound(nil, 5) != 0 {
		t.Error("пустой массив")
	}
	big := make([]int, 1<<20)
	for i := range big {
		big[i] = i * 2
	}
	start := time.Now()
	for q := 0; q < 300_000; q++ {
		x := rand.IntN(1 << 21)
		if LowerBound(big, x) != sort.SearchInts(big, x) {
			t.Fatalf("расхождение с sort.SearchInts для %d", x)
		}
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("слишком медленно (нужен O(log n)): %v", time.Since(start))
	}
}

func TestMinEatingSpeed(t *testing.T) {
	tests := []struct {
		piles []int
		h     int
		want  int
	}{
		{[]int{3, 6, 7, 11}, 8, 4}, {[]int{30, 11, 23, 4, 20}, 5, 30}, {[]int{30, 11, 23, 4, 20}, 6, 23}, {[]int{1000000000}, 2, 500000000}, {[]int{1}, 1, 1},
	}
	for _, tt := range tests {
		if got := MinEatingSpeed(tt.piles, tt.h); got != tt.want {
			t.Errorf("MinEatingSpeed(%v, %d) = %d, want %d", tt.piles, tt.h, got, tt.want)
		}
	}
}

func TestSearchRotated(t *testing.T) {
	base := []int{0, 1, 2, 4, 5, 6, 7}
	for shift := 0; shift < len(base); shift++ {
		arr := append(slices.Clone(base[shift:]), base[:shift]...)
		for i, v := range arr {
			if got := SearchRotated(arr, v); got != i {
				t.Errorf("SearchRotated(%v, %d) = %d, want %d", arr, v, got, i)
			}
		}
		if got := SearchRotated(arr, 3); got != -1 {
			t.Errorf("SearchRotated(%v, 3) = %d, want -1", arr, got)
		}
	}
	if SearchRotated(nil, 1) != -1 || SearchRotated([]int{1}, 1) != 0 {
		t.Error("крайние случаи")
	}
	big := make([]int, 1<<20)
	for i := range big {
		big[i] = i
	}
	rot := append(slices.Clone(big[777_777:]), big[:777_777]...)
	start := time.Now()
	for q := 0; q < 200_000; q++ {
		x := rand.IntN(len(big))
		if rot[SearchRotated(rot, x)] != x {
			t.Fatalf("не нашли %d", x)
		}
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("слишком медленно (нужен O(log n)): %v", time.Since(start))
	}
}
@@ solution
package solution

func LongestUniqueSubstring(s string) int {
	last := map[rune]int{}
	best, l := 0, 0
	rs := []rune(s)
	for r, c := range rs {
		if i, ok := last[c]; ok && i >= l {
			l = i + 1
		}
		last[c] = r
		best = max(best, r-l+1)
	}
	return best
}

func MinSubArrayLen(target int, nums []int) int {
	best, sum, l := len(nums)+1, 0, 0
	for r, x := range nums {
		sum += x
		for sum >= target {
			best = min(best, r-l+1)
			sum -= nums[l]
			l++
		}
	}
	if best > len(nums) {
		return 0
	}
	return best
}

func LowerBound(nums []int, target int) int {
	lo, hi := 0, len(nums)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if nums[mid] >= target {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func MinEatingSpeed(piles []int, h int) int {
	lo, hi := 1, 1
	for _, p := range piles {
		hi = max(hi, p)
	}
	for lo < hi {
		mid := lo + (hi-lo)/2
		hours := 0
		for _, p := range piles {
			hours += (p + mid - 1) / mid
		}
		if hours <= h {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func SearchRotated(nums []int, target int) int {
	lo, hi := 0, len(nums)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if nums[mid] == target {
			return mid
		}
		if nums[lo] <= nums[mid] { // левая половина отсортирована
			if nums[lo] <= target && target < nums[mid] {
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		} else { // правая половина отсортирована
			if nums[mid] < target && target <= nums[hi] {
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}
	}
	return -1
}
@@ hint
Окно: левая граница `l` сдвигается, пока условие нарушено/выполнено. Бинарный поиск: шаблон `for lo < hi { mid := lo + (hi-lo)/2; if ok(mid) { hi = mid } else { lo = mid + 1 } }`. Для вращённого массива: определи, какая из половин отсортирована.
