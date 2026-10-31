---
id: l35
week: 7
day: 5
minutes: 120
title: Алгоритмы V: перебор с возвратом и динамическое программирование
goals: Шаблон backtracking без багов копирования слайсов; мышление ДП: состояние, переход, база, порядок
lc: subsets | Subsets | Medium | Шаблон: на каждом шаге включить/не включить элемент
lc: permutations | Permutations | Medium | Массив used[] или обмен элементов
lc: combination-sum | Combination Sum | Medium | Элемент можно брать повторно: рекурсия с тем же индексом
lc: word-search | Word Search | Medium | DFS по сетке с откатом пометки посещения
lc: n-queens | N-Queens | Hard | Три множества: столбцы и две диагонали
lc: house-robber | House Robber | Medium | dp[i] = max(dp[i-1], dp[i-2] + a[i]) → две переменные
lc: coin-change | Coin Change | Medium | dp[a] = 1 + min(dp[a - coin]); неограниченный рюкзак
lc: longest-increasing-subsequence | Longest Increasing Subsequence | Medium | O(n²) ДП, затем O(n log n) с бинарным поиском по tails
lc: unique-paths | Unique Paths | Medium | Сетка: dp[i][j] = dp[i-1][j] + dp[i][j-1]
lc: word-break | Word Break | Medium | dp[i] — можно ли разбить префикс длины i
lc: longest-common-subsequence | Longest Common Subsequence | Medium | Двумерное ДП по префиксам двух строк
lc: partition-equal-subset-sum | Partition Equal Subset Sum | Medium | Рюкзак 0/1: достижима ли сумма total/2
lc: edit-distance | Edit Distance | Medium | dp[i][j] = 1 + min(вставка, удаление, замена)
---

## Backtracking (перебор с возвратом)

**Идея:** строим решение по шагам; на каждом шаге делаем **выбор**, рекурсивно идём дальше, затем **отменяем выбор** (backtrack) и пробуем следующий. Это DFS по «дереву решений». Сложность обычно экспоненциальная, поэтому важно **отсекать** заведомо плохие ветки (pruning).

Универсальный шаблон:

```go
func solve(state, choices) {
	if isComplete(state) { save(copyOf(state)); return }
	for _, c := range choices {
		if !valid(state, c) { continue }   // отсечение
		apply(state, c)                    // выбор
		solve(state, nextChoices)
		undo(state, c)                     // откат
	}
}
```

### ⚠️ Главная ловушка Go: копирование слайса при сохранении

```go run
package main

import "fmt"

func subsets(nums []int) [][]int {
	var res [][]int
	var cur []int
	var dfs func(start int)
	dfs = func(start int) {
		// ВАЖНО: сохраняем КОПИЮ. cur — общий буфер, который мы продолжим менять
		res = append(res, append([]int(nil), cur...))
		for i := start; i < len(nums); i++ {
			cur = append(cur, nums[i]) // выбор
			dfs(i + 1)
			cur = cur[:len(cur)-1]     // откат
		}
	}
	dfs(0)
	return res
}

func permutations(nums []int) [][]int {
	var res [][]int
	used := make([]bool, len(nums))
	var cur []int
	var dfs func()
	dfs = func() {
		if len(cur) == len(nums) {
			res = append(res, append([]int(nil), cur...))
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

func combinationSum(cands []int, target int) [][]int {
	var res [][]int
	var cur []int
	var dfs func(start, remain int)
	dfs = func(start, remain int) {
		if remain == 0 {
			res = append(res, append([]int(nil), cur...))
			return
		}
		for i := start; i < len(cands) && cands[i] <= remain; i++ { // cands отсортирован → отсечение
			cur = append(cur, cands[i])
			dfs(i, remain-cands[i]) // i, а не i+1: число можно использовать повторно
			cur = cur[:len(cur)-1]
		}
	}
	dfs(0, target)
	return res
}

func main() {
	fmt.Println(subsets([]int{1, 2, 3}))
	fmt.Println(len(permutations([]int{1, 2, 3, 4})))
	fmt.Println(combinationSum([]int{2, 3, 6, 7}, 7))
}
```

Если сохранить `cur` без копии, в результате все элементы окажутся «одним и тем же» слайсом, меняющимся на глазах (разделяемый массив — урок 5!).

Дубликаты во входе: отсортируй и пропускай `if i > start && nums[i] == nums[i-1] { continue }`.

Классика: подмножества, перестановки, сочетания, расстановка ферзей, судоку, поиск слова в сетке, разбиение палиндромов, генерация скобок.

## Динамическое программирование

**ДП = рекурсия + запоминание** (или заполнение таблицы снизу вверх). Применимо, когда задача состоит из **перекрывающихся подзадач** с **оптимальной подструктурой**.

Рецепт:
1. **Состояние**: что однозначно описывает подзадачу? (`dp[i]`, `dp[i][j]`, `dp[i][w]`)
2. **Переход**: как выразить состояние через меньшие?
3. **База**: значения для самых маленьких подзадач.
4. **Порядок** вычисления (чтобы зависимости уже были посчитаны) и **ответ** (какая клетка таблицы).
5. **Оптимизация памяти**: если зависимость только от предыдущей строки/двух значений — храни их.

### Сверху вниз (мемоизация) и снизу вверх (таблица)

```go run
package main

import "fmt"

// Fibonacci: мемоизация
func fibMemo(n int, memo map[int]int) int {
	if n < 2 {
		return n
	}
	if v, ok := memo[n]; ok {
		return v
	}
	memo[n] = fibMemo(n-1, memo) + fibMemo(n-2, memo)
	return memo[n]
}

// Coin Change: минимальное число монет для суммы amount (неограниченный рюкзак)
func coinChange(coins []int, amount int) int {
	const inf = 1 << 30
	dp := make([]int, amount+1) // dp[a] — минимум монет для суммы a
	for a := 1; a <= amount; a++ {
		dp[a] = inf
		for _, c := range coins {
			if c <= a && dp[a-c]+1 < dp[a] {
				dp[a] = dp[a-c] + 1
			}
		}
	}
	if dp[amount] >= inf {
		return -1
	}
	return dp[amount]
}

// LCS: наибольшая общая подпоследовательность (двумерное ДП)
func lcs(a, b string) int {
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp[len(a)][len(b)]
}

func main() {
	fmt.Println(fibMemo(50, map[int]int{}))
	fmt.Println(coinChange([]int{1, 2, 5}, 11), coinChange([]int{2}, 3))
	fmt.Println(lcs("abcde", "ace"))
}
```

### Семейства задач ДП

| Семейство | Состояние | Примеры |
|---|---|---|
| Линейное 1D | `dp[i]` | climbing stairs, house robber, max subarray, decode ways |
| Рюкзак 0/1 | `dp[w]` (цикл по весу **в обратном порядке**) | partition equal subset sum, target sum |
| Неограниченный рюкзак | `dp[w]` (по возрастанию) | coin change, combination sum IV |
| Сетка | `dp[i][j]` | unique paths, min path sum |
| Две последовательности | `dp[i][j]` по префиксам | LCS, edit distance |
| Подпоследовательность | `dp[i]` = лучший, заканчивающийся в `i` | LIS (есть O(n log n) через бинарный поиск) |
| Интервалы | `dp[l][r]` | burst balloons, palindromic substrings |
| Разбиение строки | `dp[i]` — префикс длины i | word break |

> [!TIP] Если не видишь ДП, нарисуй **полное дерево рекурсии** для малого входа: повторяющиеся поддеревья — сигнал к мемоизации.

## Жадные алгоритмы

Иногда локально оптимальный выбор даёт глобальный оптимум (интервалы: сортировка по концу; Jump Game; Gas Station). Доказательство корректности — спрашивают; если не уверен — ДП безопаснее.

## Проверь себя

<details><summary>Почему в backtracking нужно сохранять копию текущего слайса?</summary>Общий буфер продолжает изменяться при откатах; без копии все сохранённые «результаты» будут ссылаться на один и тот же изменяющийся массив.</details>

<details><summary>Чем мемоизация отличается от табличного ДП?</summary>Мемоизация — рекурсия сверху вниз с кэшем (считает только нужные состояния), табличное — итеративное заполнение снизу вверх (обычно быстрее и без рекурсии).</details>

<details><summary>Почему в рюкзаке 0/1 внутренний цикл идёт по весам в обратном порядке?</summary>Чтобы каждый предмет использовался не более одного раза (не «видеть» уже обновлённые значения текущего предмета).</details>

<details><summary>Каковы признаки ДП-задачи?</summary>Оптимизация/подсчёт, перекрывающиеся подзадачи, возможность описать состояние и переход.</details>
