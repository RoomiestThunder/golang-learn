---
title: Динамическое программирование
level: hard
---
@@ statement
Классические задачи ДП.

1. `CoinChange(coins []int, amount int) int` — минимальное число монет на сумму `amount` (каждый номинал неограниченно); `-1`, если невозможно; `amount == 0` → `0`.
2. `LIS(nums []int) int` — длина наибольшей строго возрастающей подпоследовательности. На 200 000 элементов должно работать быстро — нужен **O(n log n)** (массив «хвостов» + бинарный поиск `sort.SearchInts`).
3. `LCS(a, b string) int` — длина наибольшей общей подпоследовательности (по рунам).
4. `EditDistance(a, b string) int` — расстояние Левенштейна (вставка, удаление, замена — по 1) между строками (по рунам).
5. `CanPartition(nums []int) bool` — можно ли разбить массив положительных чисел на два подмножества с равной суммой (рюкзак 0/1, по возможности с одномерным массивом).
6. `Rob(nums []int) int` — максимальная сумма, которую можно взять, не беря два соседних элемента (house robber) за O(1) памяти.
@@ starter
package solution

func CoinChange(coins []int, amount int) int { return 0 }

func LIS(nums []int) int { return 0 }

func LCS(a, b string) int { return 0 }

func EditDistance(a, b string) int { return 0 }

func CanPartition(nums []int) bool { return false }

func Rob(nums []int) int { return 0 }
@@ tests
package solution

import (
	"math/rand/v2"
	"testing"
	"time"
)

func TestCoinChange(t *testing.T) {
	tests := []struct {
		coins  []int
		amount int
		want   int
	}{
		{[]int{1, 2, 5}, 11, 3}, {[]int{2}, 3, -1}, {[]int{1}, 0, 0}, {[]int{186, 419, 83, 408}, 6249, 20}, {nil, 5, -1}, {[]int{5, 10}, 15, 2},
	}
	for _, tt := range tests {
		if got := CoinChange(tt.coins, tt.amount); got != tt.want {
			t.Errorf("CoinChange(%v, %d) = %d, want %d", tt.coins, tt.amount, got, tt.want)
		}
	}
}

func TestLIS(t *testing.T) {
	tests := []struct {
		in   []int
		want int
	}{
		{nil, 0}, {[]int{10, 9, 2, 5, 3, 7, 101, 18}, 4}, {[]int{0, 1, 0, 3, 2, 3}, 4}, {[]int{7, 7, 7, 7}, 1}, {[]int{5, 4, 3}, 1},
	}
	for _, tt := range tests {
		if got := LIS(tt.in); got != tt.want {
			t.Errorf("LIS(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
	big := make([]int, 200_000)
	for i := range big {
		big[i] = i
	}
	rand.Shuffle(len(big), func(i, j int) { big[i], big[j] = big[j], big[i] })
	start := time.Now()
	LIS(big)
	asc := make([]int, 200_000)
	for i := range asc {
		asc[i] = i
	}
	if got := LIS(asc); got != 200_000 {
		t.Errorf("возрастающий: %d", got)
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("слишком медленно (нужен O(n log n)): %v", time.Since(start))
	}
}

func TestLCS(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{{"abcde", "ace", 3}, {"abc", "abc", 3}, {"abc", "def", 0}, {"", "abc", 0}, {"привет", "превед", 4}, {"AGGTAB", "GXTXAYB", 4}}
	for _, tt := range tests {
		if got := LCS(tt.a, tt.b); got != tt.want {
			t.Errorf("LCS(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestEditDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{{"horse", "ros", 3}, {"intention", "execution", 5}, {"", "abc", 3}, {"abc", "", 3}, {"same", "same", 0}, {"кот", "кит", 1}, {"kitten", "sitting", 3}}
	for _, tt := range tests {
		if got := EditDistance(tt.a, tt.b); got != tt.want {
			t.Errorf("EditDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCanPartition(t *testing.T) {
	tests := []struct {
		in   []int
		want bool
	}{
		{[]int{1, 5, 11, 5}, true}, {[]int{1, 2, 3, 5}, false}, {[]int{1}, false}, {[]int{2, 2}, true}, {[]int{100, 100, 100, 100, 100, 100}, true}, {nil, true},
	}
	for _, tt := range tests {
		if got := CanPartition(tt.in); got != tt.want {
			t.Errorf("CanPartition(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestRob(t *testing.T) {
	tests := []struct {
		in   []int
		want int
	}{{nil, 0}, {[]int{5}, 5}, {[]int{1, 2, 3, 1}, 4}, {[]int{2, 7, 9, 3, 1}, 12}, {[]int{2, 1, 1, 2}, 4}, {[]int{10, 1, 1, 10}, 20}}
	for _, tt := range tests {
		if got := Rob(tt.in); got != tt.want {
			t.Errorf("Rob(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
@@ solution
package solution

import "sort"

func CoinChange(coins []int, amount int) int {
	const inf = 1 << 30
	dp := make([]int, amount+1)
	for a := 1; a <= amount; a++ {
		dp[a] = inf
		for _, c := range coins {
			if c > 0 && c <= a && dp[a-c]+1 < dp[a] {
				dp[a] = dp[a-c] + 1
			}
		}
	}
	if dp[amount] >= inf {
		return -1
	}
	return dp[amount]
}

func LIS(nums []int) int {
	var tails []int // tails[i] — минимальный хвост возрастающей подпоследовательности длины i+1
	for _, x := range nums {
		i := sort.SearchInts(tails, x) // первый хвост >= x
		if i == len(tails) {
			tails = append(tails, x)
		} else {
			tails[i] = x
		}
	}
	return len(tails)
}

func LCS(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			if ra[i-1] == rb[j-1] {
				cur[j] = prev[j-1] + 1
			} else {
				cur[j] = max(prev[j], cur[j-1])
			}
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

func EditDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			if ra[i-1] == rb[j-1] {
				cur[j] = prev[j-1]
			} else {
				cur[j] = 1 + min(prev[j-1], prev[j], cur[j-1])
			}
		}
		prev = cur
	}
	return prev[len(rb)]
}

func CanPartition(nums []int) bool {
	total := 0
	for _, n := range nums {
		total += n
	}
	if total%2 != 0 {
		return false
	}
	half := total / 2
	dp := make([]bool, half+1)
	dp[0] = true
	for _, n := range nums {
		for s := half; s >= n; s-- { // обратный порядок: каждый предмет не более одного раза
			if dp[s-n] {
				dp[s] = true
			}
		}
	}
	return dp[half]
}

func Rob(nums []int) int {
	prev2, prev1 := 0, 0
	for _, n := range nums {
		prev2, prev1 = prev1, max(prev1, prev2+n)
	}
	return prev1
}
@@ hint
ДП: определи состояние, переход и базу. LIS за O(n log n): `tails` + `sort.SearchInts(tails, x)` (строго возрастающая → ищем первый хвост >= x). Рюкзак 0/1: внутренний цикл по сумме идёт от большего к меньшему.
