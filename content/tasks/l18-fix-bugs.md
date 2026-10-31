---
title: Найди баги по падающим тестам
level: medium
---
@@ statement
В коде ниже **пять багов**. Таблицы тестов уже написаны (можно открыть кнопкой «🧪 Тесты»): запусти проверку, прочитай сообщения `got/want`, найди причину и исправь. Именно так выглядит цикл «красный тест → причина → зелёный».

Что должны делать функции:

- `IsLeap(year)` — високосный год по григорианскому календарю.
- `Median(nums)` — медиана; для чётного количества — среднее двух центральных; для пустого слайса — `0, false`. Входной слайс **не должен меняться**.
- `Clamp(x, lo, hi)` — ограничивает `x` отрезком `[lo, hi]` (если `lo > hi` — поменять их местами).
- `WordCount(s)` — число слов (последовательности не-пробельных символов).
- `Chunk(s, size)` — разбивает строку на куски по `size` **рун**; `size <= 0` → `nil`.
@@ starter
package solution

import "sort"

func IsLeap(year int) bool {
	return year%4 == 0
}

func Median(nums []int) (float64, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	sort.Ints(nums)
	return float64(nums[len(nums)/2]), true
}

func Clamp(x, lo, hi int) int {
	if x < lo {
		return hi
	}
	if x > hi {
		return lo
	}
	return x
}

func WordCount(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			n++
		}
	}
	return n + 1
}

func Chunk(s string, size int) []string {
	if size <= 0 {
		return nil
	}
	var out []string
	for i := 0; i < len(s); i += size {
		out = append(out, s[i:min(i+size, len(s))])
	}
	return out
}
@@ tests
package solution

import (
	"fmt"
	"slices"
	"testing"
)

func TestIsLeap(t *testing.T) {
	tests := []struct {
		year int
		want bool
	}{{2024, true}, {2023, false}, {1900, false}, {2000, true}, {2100, false}, {1996, true}}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.year), func(t *testing.T) {
			if got := IsLeap(tt.year); got != tt.want {
				t.Errorf("IsLeap(%d) = %v, want %v", tt.year, got, tt.want)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	in := []int{5, 1, 4, 2}
	got, ok := Median(in)
	if !ok || got != 3 {
		t.Errorf("Median(even) = %v, %v; want 3", got, ok)
	}
	if !slices.Equal(in, []int{5, 1, 4, 2}) {
		t.Errorf("Median изменила входной слайс: %v", in)
	}
	if got, _ := Median([]int{9, 1, 5}); got != 5 {
		t.Errorf("Median(odd) = %v, want 5", got)
	}
	if got, ok := Median(nil); ok || got != 0 {
		t.Errorf("Median(nil) = %v, %v", got, ok)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct{ x, lo, hi, want int }{
		{5, 0, 10, 5}, {-3, 0, 10, 0}, {42, 0, 10, 10}, {5, 10, 0, 5}, {-1, 10, 0, 0},
	}
	for _, tt := range tests {
		if got := Clamp(tt.x, tt.lo, tt.hi); got != tt.want {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.x, tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestWordCount(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{{"", 0}, {"hello", 1}, {"hello world", 2}, {"  many   spaces here  ", 3}, {"tab\tand\nnewline", 3}}
	for _, tt := range tests {
		if got := WordCount(tt.in); got != tt.want {
			t.Errorf("WordCount(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestChunk(t *testing.T) {
	if got := Chunk("abcdefg", 3); !slices.Equal(got, []string{"abc", "def", "g"}) {
		t.Errorf("ascii: %q", got)
	}
	if got := Chunk("приветмир", 4); !slices.Equal(got, []string{"прив", "етми", "р"}) {
		t.Errorf("кириллица (куски по рунам, а не по байтам): %q", got)
	}
	if got := Chunk("abc", 0); got != nil {
		t.Errorf("size=0: %q", got)
	}
	if got := Chunk("", 2); len(got) != 0 {
		t.Errorf("пустая строка: %q", got)
	}
}
@@ solution
package solution

import (
	"slices"
	"strings"
)

func IsLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func Median(nums []int) (float64, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	s := slices.Clone(nums)
	slices.Sort(s)
	mid := len(s) / 2
	if len(s)%2 == 1 {
		return float64(s[mid]), true
	}
	return float64(s[mid-1]+s[mid]) / 2, true
}

func Clamp(x, lo, hi int) int {
	if lo > hi {
		lo, hi = hi, lo
	}
	return min(max(x, lo), hi)
}

func WordCount(s string) int { return len(strings.Fields(s)) }

func Chunk(s string, size int) []string {
	if size <= 0 {
		return nil
	}
	var out []string
	r := []rune(s)
	for i := 0; i < len(r); i += size {
		out = append(out, string(r[i:min(i+size, len(r))]))
	}
	return out
}
@@ hint
Баги: (1) правило 100/400, (2) сортировка входа + берётся один элемент при чётной длине, (3) перепутаны границы, (4) `strings.Fields` вместо подсчёта пробелов, (5) срез строки по байтам вместо рун.
