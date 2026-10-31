---
title: Замыкания и defer
level: medium
---
@@ statement
Три небольшие задачи на замыкания и `defer`.

1. `NewCounter(start, step int) func() int` — каждый вызов возвращённой функции возвращает текущее значение и сдвигает его на `step` (первый вызов вернёт `start`). Счётчики независимы.
2. `Memoize(f func(int) int) func(int) int` — кэширует результаты `f` по аргументу (повторный вызов с тем же аргументом не вызывает `f`).
3. `SafeDiv(a, b int) (res int, err error)` — делит `a` на `b`. При `b == 0` **не возвращает ошибку явно**: пусть сработает паника деления, а ты перехватишь её через `defer` + `recover()` и превратишь в `err` (res при этом 0).
@@ starter
package solution

func NewCounter(start, step int) func() int {
	return func() int { return 0 }
}

func Memoize(f func(int) int) func(int) int {
	return f
}

func SafeDiv(a, b int) (res int, err error) {
	return a / b, nil
}
@@ tests
package solution

import "testing"

func TestCounter(t *testing.T) {
	c1, c2 := NewCounter(10, 5), NewCounter(0, 1)
	got := []int{c1(), c1(), c1(), c2(), c2()}
	want := []int{10, 15, 20, 0, 1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestMemoize(t *testing.T) {
	calls := 0
	sq := Memoize(func(x int) int { calls++; return x * x })
	if sq(4) != 16 || sq(4) != 16 || sq(5) != 25 || sq(4) != 16 {
		t.Fatal("неверные значения")
	}
	if calls != 2 {
		t.Errorf("f вызывалась %d раз, ожидалось 2", calls)
	}
}

func TestSafeDiv(t *testing.T) {
	if r, err := SafeDiv(7, 2); r != 3 || err != nil {
		t.Errorf("SafeDiv(7,2) = %d, %v", r, err)
	}
	r, err := SafeDiv(1, 0)
	if err == nil || r != 0 {
		t.Errorf("SafeDiv(1,0) = %d, %v; нужна ошибка и 0", r, err)
	}
}
@@ solution
package solution

import "fmt"

func NewCounter(start, step int) func() int {
	cur := start
	return func() int {
		v := cur
		cur += step
		return v
	}
}

func Memoize(f func(int) int) func(int) int {
	cache := map[int]int{}
	return func(x int) int {
		if v, ok := cache[x]; ok {
			return v
		}
		v := f(x)
		cache[x] = v
		return v
	}
}

func SafeDiv(a, b int) (res int, err error) {
	defer func() {
		if r := recover(); r != nil {
			res, err = 0, fmt.Errorf("паника: %v", r)
		}
	}()
	return a / b, nil
}
@@ hint
Для `SafeDiv` результаты должны быть именованными — только так `defer` сможет их изменить после паники.
