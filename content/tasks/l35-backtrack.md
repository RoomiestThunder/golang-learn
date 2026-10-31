---
title: Backtracking: подмножества, перестановки, комбинации
level: medium
---
@@ statement
Реализуй перебор с возвратом. **Каждый результат должен быть независимой копией** (не делить массив с рабочим буфером!).

1. `Subsets(nums []int) [][]int` — все подмножества (включая пустое) массива **уникальных** чисел; элементы внутри подмножества — в порядке исходного массива. Порядок подмножеств любой. Для `n` чисел — `2ⁿ` подмножеств.
2. `Permutations(nums []int) [][]int` — все перестановки массива уникальных чисел (`n!` штук), порядок любой.
3. `CombinationSum(candidates []int, target int) [][]int` — все **уникальные** комбинации чисел из `candidates` (положительные, без повторов в массиве), дающие `target`; каждое число можно брать **неограниченно** раз; внутри комбинации — по неубыванию.
4. `SubsetsWithDup(nums []int) [][]int` — подмножества массива **с повторяющимися** числами без дубликатов среди результатов (внутри подмножества — по неубыванию).
@@ starter
package solution

func Subsets(nums []int) [][]int { return nil }

func Permutations(nums []int) [][]int { return nil }

func CombinationSum(candidates []int, target int) [][]int { return nil }

func SubsetsWithDup(nums []int) [][]int { return nil }
@@ tests
package solution

import (
	"fmt"
	"slices"
	"sort"
	"testing"
)

func canon(res [][]int) []string {
	out := make([]string, len(res))
	for i, r := range res {
		out[i] = fmt.Sprint(r)
	}
	sort.Strings(out)
	return out
}

func eq(t *testing.T, name string, got, want [][]int) {
	t.Helper()
	g, w := canon(got), canon(want)
	if !slices.Equal(g, w) {
		t.Errorf("%s:\n got  %v\n want %v", name, g, w)
	}
}

func TestSubsets(t *testing.T) {
	eq(t, "Subsets([1 2 3])", Subsets([]int{1, 2, 3}), [][]int{{}, {1}, {2}, {3}, {1, 2}, {1, 3}, {2, 3}, {1, 2, 3}})
	eq(t, "Subsets([])", Subsets(nil), [][]int{{}})
	if got := Subsets(make([]int, 0)); len(got) != 1 {
		t.Errorf("пустой вход → одно пустое подмножество, got %v", got)
	}
	// независимость копий
	res := Subsets([]int{1, 2, 3, 4})
	if len(res) != 16 {
		t.Fatalf("len = %d", len(res))
	}
	for _, r := range res {
		if len(r) > 0 {
			r[0] = -99
		}
	}
	for _, r := range res {
		if len(r) > 1 && r[1] == -99 {
			t.Fatal("результаты делят память — сохраняй копии")
		}
	}
}

func TestPermutations(t *testing.T) {
	eq(t, "Permutations([1 2 3])", Permutations([]int{1, 2, 3}), [][]int{{1, 2, 3}, {1, 3, 2}, {2, 1, 3}, {2, 3, 1}, {3, 1, 2}, {3, 2, 1}})
	if len(Permutations([]int{1, 2, 3, 4, 5, 6, 7})) != 5040 {
		t.Error("7! = 5040")
	}
	eq(t, "Permutations([1])", Permutations([]int{1}), [][]int{{1}})
	in := []int{3, 1, 2}
	Permutations(in)
	if !slices.Equal(in, []int{3, 1, 2}) {
		t.Error("входной слайс изменён")
	}
}

func TestCombinationSum(t *testing.T) {
	eq(t, "CombinationSum([2 3 6 7], 7)", CombinationSum([]int{2, 3, 6, 7}, 7), [][]int{{2, 2, 3}, {7}})
	eq(t, "CombinationSum([2 3 5], 8)", CombinationSum([]int{2, 3, 5}, 8), [][]int{{2, 2, 2, 2}, {2, 3, 3}, {3, 5}})
	if got := CombinationSum([]int{2}, 1); len(got) != 0 {
		t.Errorf("нет решений: %v", got)
	}
	eq(t, "несортированный вход", CombinationSum([]int{7, 3, 2}, 7), [][]int{{2, 2, 3}, {7}})
}

func TestSubsetsWithDup(t *testing.T) {
	eq(t, "SubsetsWithDup([1 2 2])", SubsetsWithDup([]int{1, 2, 2}), [][]int{{}, {1}, {2}, {1, 2}, {2, 2}, {1, 2, 2}})
	eq(t, "SubsetsWithDup([2 1 2])", SubsetsWithDup([]int{2, 1, 2}), [][]int{{}, {1}, {2}, {1, 2}, {2, 2}, {1, 2, 2}})
	eq(t, "SubsetsWithDup([0])", SubsetsWithDup([]int{0}), [][]int{{}, {0}})
}
@@ solution
package solution

import "slices"

func Subsets(nums []int) [][]int {
	res := [][]int{}
	var cur []int
	var dfs func(start int)
	dfs = func(start int) {
		res = append(res, slices.Clone(append([]int{}, cur...)))
		for i := start; i < len(nums); i++ {
			cur = append(cur, nums[i])
			dfs(i + 1)
			cur = cur[:len(cur)-1]
		}
	}
	dfs(0)
	return res
}

func Permutations(nums []int) [][]int {
	var res [][]int
	used := make([]bool, len(nums))
	var cur []int
	var dfs func()
	dfs = func() {
		if len(cur) == len(nums) {
			res = append(res, slices.Clone(cur))
			return
		}
		for i, x := range nums {
			if used[i] {
				continue
			}
			used[i] = true
			cur = append(cur, x)
			dfs()
			cur = cur[:len(cur)-1]
			used[i] = false
		}
	}
	dfs()
	return res
}

func CombinationSum(candidates []int, target int) [][]int {
	cands := slices.Clone(candidates)
	slices.Sort(cands)
	var res [][]int
	var cur []int
	var dfs func(start, remain int)
	dfs = func(start, remain int) {
		if remain == 0 {
			res = append(res, slices.Clone(cur))
			return
		}
		for i := start; i < len(cands) && cands[i] <= remain; i++ {
			cur = append(cur, cands[i])
			dfs(i, remain-cands[i])
			cur = cur[:len(cur)-1]
		}
	}
	dfs(0, target)
	return res
}

func SubsetsWithDup(nums []int) [][]int {
	a := slices.Clone(nums)
	slices.Sort(a)
	res := [][]int{}
	var cur []int
	var dfs func(start int)
	dfs = func(start int) {
		res = append(res, slices.Clone(append([]int{}, cur...)))
		for i := start; i < len(a); i++ {
			if i > start && a[i] == a[i-1] {
				continue // пропускаем повтор на том же уровне
			}
			cur = append(cur, a[i])
			dfs(i + 1)
			cur = cur[:len(cur)-1]
		}
	}
	dfs(0)
	return res
}
@@ hint
Шаблон: `res = append(res, slices.Clone(cur))` на каждом решении; выбор `cur = append(cur, x)`, рекурсия, откат `cur = cur[:len(cur)-1]`. Для дубликатов: отсортируй и пропускай `if i > start && a[i] == a[i-1]`.
