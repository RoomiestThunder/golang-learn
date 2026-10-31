---
title: Конвейер на каналах
level: medium
---
@@ statement
Построй конвейер из трёх функций. Все возвращаемые каналы **закрываются** отправителем, когда данные кончились.

- `Gen(nums ...int) <-chan int` — отправляет числа и закрывается.
- `Square(in <-chan int) <-chan int` — читает `in`, отправляет квадраты, закрывается после закрытия `in`.
- `Merge(cs ...<-chan int) <-chan int` — fan-in: объединяет несколько каналов в один; закрывается, когда **все** входные закрыты (порядок значений не определён). `Merge()` без аргументов возвращает уже закрытый канал.

Используй горутины и `sync.WaitGroup`.
@@ starter
package solution

func Gen(nums ...int) <-chan int {
	return nil
}

func Square(in <-chan int) <-chan int {
	return nil
}

func Merge(cs ...<-chan int) <-chan int {
	return nil
}
@@ tests
package solution

import (
	"slices"
	"testing"
	"time"
)

// drain читает канал до закрытия, но не дольше 2 секунд (зависший канал → ошибка, а не вечное ожидание)
func drain(t *testing.T, c <-chan int) []int {
	t.Helper()
	var out []int
	timeout := time.After(2 * time.Second)
	for {
		select {
		case v, ok := <-c:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-timeout:
			t.Fatal("канал не закрылся за 2 секунды")
			return nil
		}
	}
}

func TestGenSquare(t *testing.T) {
	got := drain(t, Square(Gen(1, 2, 3, 4)))
	if !slices.Equal(got, []int{1, 4, 9, 16}) {
		t.Errorf("got %v", got)
	}
	if got := drain(t, Gen()); len(got) != 0 {
		t.Errorf("пустой Gen: %v", got)
	}
}

func TestMerge(t *testing.T) {
	got := drain(t, Merge(Gen(1, 2, 3), Square(Gen(4, 5)), Gen(), Gen(100)))
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3, 16, 25, 100}) {
		t.Errorf("got %v", got)
	}
	if got := drain(t, Merge()); len(got) != 0 {
		t.Errorf("Merge() = %v", got)
	}
}
@@ solution
package solution

import "sync"

func Gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func Square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func Merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
@@ hint
В `Merge` на каждый входной канал — своя горутина; ещё одна ждёт `wg.Wait()` и закрывает `out`.
