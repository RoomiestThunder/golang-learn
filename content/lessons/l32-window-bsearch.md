---
id: l32
week: 7
day: 2
minutes: 120
title: Алгоритмы II: скользящее окно, префиксные суммы, бинарный поиск
goals: Узнавать задачи на окно и бинарный поиск по ответу, писать их без off-by-one
lc: best-time-to-buy-and-sell-stock | Best Time to Buy and Sell Stock | Easy | Один проход: минимальная цена слева и максимальная прибыль
lc: longest-substring-without-repeating-characters | Longest Substring Without Repeating Characters | Medium | Окно + map последнего индекса символа
lc: longest-repeating-character-replacement | Longest Repeating Character Replacement | Medium | Окно с условием: длина − max частота ≤ k
lc: minimum-window-substring | Minimum Window Substring | Hard | Окно с двумя счётчиками «нужно / есть»
lc: subarray-sum-equals-k | Subarray Sum Equals K | Medium | Префиксные суммы + map частот
lc: binary-search | Binary Search | Easy | Базовый шаблон — отработай до автоматизма
lc: search-in-rotated-sorted-array | Search in Rotated Sorted Array | Medium | Определяй, какая половина отсортирована
lc: koko-eating-bananas | Koko Eating Bananas | Medium | Бинарный поиск по ОТВЕТУ (скорость) с проверкой за O(n)
---

## Скользящее окно

**Когда:** задача про **непрерывный подмассив/подстроку** с условием (макс. длина, мин. длина, сумма, число различных элементов).

**Идея:** поддерживай окно `[l, r]`; расширяй `r`, а когда условие нарушено (или выполнено) — сдвигай `l`. Каждый элемент входит и выходит максимум по разу → O(n).

```go run
package main

import "fmt"

// Самая длинная подстрока без повторяющихся символов
func lengthOfLongestSubstring(s string) int {
	last := map[byte]int{} // символ → индекс, где его видели последний раз
	best, l := 0, 0
	for r := 0; r < len(s); r++ {
		if i, ok := last[s[r]]; ok && i >= l {
			l = i + 1 // перепрыгиваем повтор
		}
		last[s[r]] = r
		best = max(best, r-l+1)
	}
	return best
}

// Минимальная длина подмассива с суммой >= target (только неотрицательные числа)
func minSubArrayLen(target int, nums []int) int {
	best, sum, l := len(nums)+1, 0, 0
	for r, x := range nums {
		sum += x
		for sum >= target { // окно «достаточно» — пробуем сузить
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

func main() {
	fmt.Println(lengthOfLongestSubstring("abcabcbb"), lengthOfLongestSubstring("pwwkew"))
	fmt.Println(minSubArrayLen(7, []int{2, 3, 1, 2, 4, 3}))
}
```

Две формы:
- **Окно фиксированной длины k**: добавил `a[r]`, убрал `a[r-k]`, обновил ответ.
- **Окно переменной длины**: «растянуть, пока можно; сжать, пока нужно» (внутренний `for` или `if`).

Структуры внутри окна: счётчик частот (`map` или `[256]int`), сумма, дек (`[]int` с индексами) для максимума окна — «монотонная очередь».

## Префиксные суммы

`prefix[i] = a[0] + ... + a[i-1]` — сумма на отрезке `[l, r)` = `prefix[r] - prefix[l]` за O(1).

```go run
package main

import "fmt"

type NumArray struct{ prefix []int }

func NewNumArray(nums []int) *NumArray {
	p := make([]int, len(nums)+1)
	for i, x := range nums {
		p[i+1] = p[i] + x
	}
	return &NumArray{p}
}

// сумма nums[l..r] включительно
func (n *NumArray) SumRange(l, r int) int { return n.prefix[r+1] - n.prefix[l] }

func main() {
	a := NewNumArray([]int{-2, 0, 3, -5, 2, -1})
	fmt.Println(a.SumRange(0, 2), a.SumRange(2, 5), a.SumRange(0, 5))
}
```

Расширения: 2D-префиксные суммы, «разностный массив» (массовые прибавления на отрезках), префиксные суммы + map (подмассивы с суммой k — см. прошлый урок), префиксные XOR/произведения.

## Бинарный поиск

Работает, когда есть **монотонность**: предикат `f(i)` сначала `false`, потом `true` (или массив отсортирован). Шаблон «первый индекс, где предикат истинен» — самый универсальный, избавляет от путаницы с границами:

```go run
package main

import (
	"fmt"
	"sort"
)

// lowerBound — первый индекс i, где a[i] >= target (или len(a))
func lowerBound(a []int, target int) int {
	lo, hi := 0, len(a) // ответ в [lo, hi]; hi = len(a) означает «нет такого»
	for lo < hi {
		mid := lo + (hi-lo)/2 // не (lo+hi)/2 — переполнение на больших числах
		if a[mid] >= target {
			hi = mid // mid может быть ответом — оставляем его
		} else {
			lo = mid + 1 // mid точно не подходит
		}
	}
	return lo
}

// Бинарный поиск по ОТВЕТУ: минимальная скорость k, чтобы съесть все кучи за h часов
func minEatingSpeed(piles []int, h int) int {
	canFinish := func(k int) bool {
		hours := 0
		for _, p := range piles {
			hours += (p + k - 1) / k // потолок деления
		}
		return hours <= h
	}
	lo, hi := 1, 0
	for _, p := range piles {
		hi = max(hi, p)
	}
	for lo < hi {
		mid := lo + (hi-lo)/2
		if canFinish(mid) {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func main() {
	a := []int{1, 3, 3, 5, 8}
	fmt.Println(lowerBound(a, 3), lowerBound(a, 4), lowerBound(a, 100))
	// стандартная библиотека делает то же самое:
	fmt.Println(sort.SearchInts(a, 3), sort.Search(len(a), func(i int) bool { return a[i] >= 4 }))
	fmt.Println(minEatingSpeed([]int{3, 6, 7, 11}, 8))
}
```

Алгоритм-«скелет»:
1. Определи диапазон ответа `[lo, hi]`.
2. Сформулируй предикат `ok(mid)`, монотонный по `mid`.
3. `for lo < hi { mid := lo + (hi-lo)/2; if ok(mid) { hi = mid } else { lo = mid + 1 } }`.
4. Ответ — `lo` (проверь, что он валиден).

Примеры «поиска по ответу»: минимальная ёмкость корабля, максимальное минимальное расстояние, корень числа, k-й элемент в отсортированной матрице, разделение массива на k частей.

Для **вращённого** массива: на каждом шаге одна из половин отсортирована — проверь, лежит ли цель в ней.

## Типичные ошибки (off-by-one)

- Смешивание «закрытого» `[lo, hi]` и «полуоткрытого» `[lo, hi)` интервалов — выбери один и придерживайся.
- Бесконечный цикл: `lo = mid` при `mid == lo`. Используй `lo = mid + 1` или округление вверх.
- Забыть про `len == 0`.
- `mid := (lo + hi) / 2` на больших значениях.

## Проверь себя

<details><summary>Как понять, что задача на скользящее окно?</summary>Ищется непрерывный подотрезок, а добавление/удаление элемента на границах позволяет пересчитать характеристику за O(1).</details>

<details><summary>Как за O(1) получить сумму на отрезке?</summary>Через префиксные суммы: <code>prefix[r+1] - prefix[l]</code>.</details>

<details><summary>Что такое бинарный поиск по ответу?</summary>Поиск минимального/максимального значения параметра, для которого монотонный предикат «подходит ли?» становится истинным; проверка — обычно O(n).</details>
