---
title: Хэш-таблицы и два указателя
level: medium
---
@@ statement
Три классические задачи (аналоги LeetCode — реши их потом и там).

1. `GroupAnagrams(words []string) [][]string` — сгруппируй слова, являющиеся анаграммами друг друга (слова из строчных латинских букв `a–z`). Порядок групп и слов внутри групп любой.
2. `ThreeSum(nums []int) [][]int` — все **уникальные** тройки `[a, b, c]` с `a + b + c == 0`; внутри тройки числа по возрастанию, порядок троек любой. Входной слайс не должен меняться. Сложность — O(n²) (на 3000 элементов должно работать быстро).
3. `LongestConsecutive(nums []int) int` — длина самой длинной последовательности подряд идущих чисел (порядок в массиве не важен, например `[100,4,200,1,3,2]` → `4`). Сложность O(n) — на 300 000 элементов должно работать быстро (подсказка: множество на `map`, счёт начинай только с чисел без левого соседа).
@@ starter
package solution

func GroupAnagrams(words []string) [][]string {
	return nil
}

func ThreeSum(nums []int) [][]int {
	return nil
}

func LongestConsecutive(nums []int) int {
	return 0
}
@@ tests
package solution

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"sort"
	"testing"
	"time"
)

func canonGroups(g [][]string) []string {
	out := make([]string, 0, len(g))
	for _, grp := range g {
		c := slices.Clone(grp)
		slices.Sort(c)
		out = append(out, fmt.Sprint(c))
	}
	sort.Strings(out)
	return out
}

func TestGroupAnagrams(t *testing.T) {
	got := canonGroups(GroupAnagrams([]string{"eat", "tea", "tan", "ate", "nat", "bat"}))
	want := []string{"[ate eat tea]", "[bat]", "[nat tan]"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := canonGroups(GroupAnagrams([]string{""})); !slices.Equal(got, []string{"[]"}) {
		t.Errorf("пустая строка: %v", got)
	}
	if len(GroupAnagrams(nil)) != 0 {
		t.Error("nil → пусто")
	}
}

func canonTriples(ts [][]int) []string {
	out := make([]string, 0, len(ts))
	for _, tr := range ts {
		out = append(out, fmt.Sprint(tr))
	}
	sort.Strings(out)
	return out
}

func TestThreeSum(t *testing.T) {
	in := []int{-1, 0, 1, 2, -1, -4}
	snapshot := slices.Clone(in)
	got := canonTriples(ThreeSum(in))
	want := canonTriples([][]int{{-1, -1, 2}, {-1, 0, 1}})
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if !slices.Equal(in, snapshot) {
		t.Error("входной слайс изменён")
	}
	if got := ThreeSum([]int{0, 0, 0, 0}); len(got) != 1 {
		t.Errorf("нули: %v", got)
	}
	if got := ThreeSum([]int{1, 2, 3}); len(got) != 0 {
		t.Errorf("нет решений: %v", got)
	}
	if got := ThreeSum(nil); len(got) != 0 {
		t.Errorf("nil: %v", got)
	}
}

func TestThreeSumBig(t *testing.T) {
	nums := make([]int, 3000)
	for i := range nums {
		nums[i] = rand.IntN(2001) - 1000
	}
	start := time.Now()
	res := ThreeSum(nums)
	if time.Since(start) > 3*time.Second {
		t.Errorf("слишком медленно: %v", time.Since(start))
	}
	seen := map[string]bool{}
	for _, tr := range res {
		if len(tr) != 3 || tr[0]+tr[1]+tr[2] != 0 || !(tr[0] <= tr[1] && tr[1] <= tr[2]) {
			t.Fatalf("плохая тройка %v", tr)
		}
		k := fmt.Sprint(tr)
		if seen[k] {
			t.Fatalf("дубликат %v", tr)
		}
		seen[k] = true
	}
}

func TestLongestConsecutive(t *testing.T) {
	tests := []struct {
		in   []int
		want int
	}{{nil, 0}, {[]int{100, 4, 200, 1, 3, 2}, 4}, {[]int{0, 3, 7, 2, 5, 8, 4, 6, 0, 1}, 9}, {[]int{1, 2, 0, 1}, 3}, {[]int{-1, -2, -3, 10}, 3}}
	for _, tt := range tests {
		if got := LongestConsecutive(tt.in); got != tt.want {
			t.Errorf("LongestConsecutive(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
	big := make([]int, 300_000)
	for i := range big {
		big[i] = i * 2
	}
	rand.Shuffle(len(big), func(i, j int) { big[i], big[j] = big[j], big[i] })
	big = append(big, 1, 3, 5, 7)
	start := time.Now()
	if got := LongestConsecutive(big); got != 9 { // 0..8
		t.Errorf("big = %d", got)
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("слишком медленно: %v", time.Since(start))
	}
}
@@ solution
package solution

import (
	"slices"
	"sort"
)

func GroupAnagrams(words []string) [][]string {
	groups := map[[26]int][]string{}
	for _, w := range words {
		var key [26]int
		for i := 0; i < len(w); i++ {
			key[w[i]-'a']++
		}
		groups[key] = append(groups[key], w)
	}
	out := make([][]string, 0, len(groups))
	for _, g := range groups {
		out = append(out, g)
	}
	return out
}

func ThreeSum(nums []int) [][]int {
	a := slices.Clone(nums)
	sort.Ints(a)
	var res [][]int
	for i := 0; i < len(a)-2; i++ {
		if a[i] > 0 {
			break
		}
		if i > 0 && a[i] == a[i-1] {
			continue
		}
		l, r := i+1, len(a)-1
		for l < r {
			switch s := a[i] + a[l] + a[r]; {
			case s < 0:
				l++
			case s > 0:
				r--
			default:
				res = append(res, []int{a[i], a[l], a[r]})
				for l < r && a[l] == a[l+1] {
					l++
				}
				for l < r && a[r] == a[r-1] {
					r--
				}
				l++
				r--
			}
		}
	}
	return res
}

func LongestConsecutive(nums []int) int {
	set := make(map[int]struct{}, len(nums))
	for _, n := range nums {
		set[n] = struct{}{}
	}
	best := 0
	for n := range set {
		if _, hasLeft := set[n-1]; hasLeft {
			continue
		}
		length := 1
		for {
			if _, ok := set[n+length]; !ok {
				break
			}
			length++
		}
		best = max(best, length)
	}
	return best
}
@@ hint
GroupAnagrams: ключ map — `[26]int`. ThreeSum: сортировка, зафиксируй `i`, двигай `l` и `r`, пропускай дубликаты. LongestConsecutive: считай только от чисел, у которых нет `n-1` во множестве.
