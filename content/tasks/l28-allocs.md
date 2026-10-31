---
title: Снижаем аллокации
level: hard
---
@@ statement
Реализуй функции так, чтобы они укладывались в бюджет аллокаций (тест проверяет через `testing.AllocsPerRun`).

1. `JoinInts(nums []int, sep string) string` — склеивает числа через разделитель (`[]int{1,2,3}`, `", "` → `"1, 2, 3"`). Бюджет: **не более 2 аллокаций** на вызов для 1000 чисел. (Намёк: `strings.Builder`/`strconv.AppendInt` с предварительным `Grow`.) Не используй `fmt.Sprint`/`Sprintf` и `strings.Join` от промежуточного `[]string`.
2. `FilterInPlace(nums []int, keep func(int) bool) []int` — оставляет только элементы, для которых `keep` вернул `true`, **переиспользуя тот же массив** и сохраняя порядок. Бюджет: **0 аллокаций**. Результат — срез начала исходного массива.
3. `CompactSorted(nums []int) []int` — удаляет подряд идущие дубликаты в отсортированном срезе **на месте**. Бюджет: **0 аллокаций**.
@@ starter
package solution

import (
	"fmt"
	"strings"
)

func JoinInts(nums []int, sep string) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = fmt.Sprint(n)
	}
	return strings.Join(parts, sep)
}

func FilterInPlace(nums []int, keep func(int) bool) []int {
	var out []int
	for _, n := range nums {
		if keep(n) {
			out = append(out, n)
		}
	}
	return out
}

func CompactSorted(nums []int) []int {
	var out []int
	for i, n := range nums {
		if i == 0 || n != nums[i-1] {
			out = append(out, n)
		}
	}
	return out
}
@@ tests
package solution

import (
	"slices"
	"strings"
	"testing"
)

func TestJoinInts(t *testing.T) {
	tests := []struct {
		in   []int
		sep  string
		want string
	}{
		{nil, ",", ""}, {[]int{7}, ",", "7"}, {[]int{1, 2, 3}, ", ", "1, 2, 3"}, {[]int{-1, 0, 1000000}, "|", "-1|0|1000000"},
	}
	for _, tt := range tests {
		if got := JoinInts(tt.in, tt.sep); got != tt.want {
			t.Errorf("JoinInts(%v, %q) = %q, want %q", tt.in, tt.sep, got, tt.want)
		}
	}
}

func TestJoinIntsAllocs(t *testing.T) {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i * 37
	}
	if got := JoinInts(nums, ","); strings.Count(got, ",") != 999 {
		t.Fatalf("результат неверен")
	}
	allocs := testing.AllocsPerRun(50, func() { _ = JoinInts(nums, ",") })
	if allocs > 2 {
		t.Errorf("JoinInts: %.0f аллокаций, бюджет 2", allocs)
	}
}

func TestFilterInPlace(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6}
	got := FilterInPlace(in, func(n int) bool { return n%2 == 0 })
	if !slices.Equal(got, []int{2, 4, 6}) {
		t.Errorf("got %v", got)
	}
	if &got[0] != &in[0] {
		t.Error("результат должен использовать тот же массив")
	}
	if got := FilterInPlace([]int{1, 3}, func(n int) bool { return n%2 == 0 }); len(got) != 0 {
		t.Errorf("все отфильтрованы: %v", got)
	}
}

func TestFilterInPlaceAllocs(t *testing.T) {
	data := make([]int, 1000)
	keep := func(n int) bool { return n%3 == 0 }
	allocs := testing.AllocsPerRun(50, func() {
		for i := range data {
			data[i] = i
		}
		_ = FilterInPlace(data, keep)
	})
	if allocs != 0 {
		t.Errorf("FilterInPlace: %.0f аллокаций, бюджет 0", allocs)
	}
}

func TestCompactSorted(t *testing.T) {
	in := []int{1, 1, 2, 3, 3, 3, 4}
	got := CompactSorted(in)
	if !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("got %v", got)
	}
	if len(CompactSorted(nil)) != 0 || !slices.Equal(CompactSorted([]int{5}), []int{5}) {
		t.Error("крайние случаи")
	}
	data := make([]int, 1000)
	allocs := testing.AllocsPerRun(50, func() {
		for i := range data {
			data[i] = i / 4
		}
		_ = CompactSorted(data)
	})
	if allocs != 0 {
		t.Errorf("CompactSorted: %.0f аллокаций, бюджет 0", allocs)
	}
}
@@ solution
package solution

import "strconv"

func JoinInts(nums []int, sep string) string {
	if len(nums) == 0 {
		return ""
	}
	buf := make([]byte, 0, len(nums)*(len(sep)+8))
	for i, n := range nums {
		if i > 0 {
			buf = append(buf, sep...)
		}
		buf = strconv.AppendInt(buf, int64(n), 10)
	}
	return string(buf)
}

func FilterInPlace(nums []int, keep func(int) bool) []int {
	out := nums[:0]
	for _, n := range nums {
		if keep(n) {
			out = append(out, n)
		}
	}
	return out
}

func CompactSorted(nums []int) []int {
	if len(nums) == 0 {
		return nums
	}
	w := 1
	for r := 1; r < len(nums); r++ {
		if nums[r] != nums[w-1] {
			nums[w] = nums[r]
			w++
		}
	}
	return nums[:w]
}
@@ hint
`out := nums[:0]` — «пустой» срез поверх того же массива; `append` в него перезаписывает уже прочитанные элементы. Для строки: один буфер `[]byte` с запасом + `strconv.AppendInt`, затем одно `string(buf)`.
