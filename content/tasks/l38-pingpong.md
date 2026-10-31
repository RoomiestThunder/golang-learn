---
title: Ping-pong и обедающие философы
level: hard
---
@@ statement
Две классические concurrency-задачи.

1. `PingPong(n int) []string` — запусти **две горутины**, которые поочерёдно «играют» `n` раундов: первая добавляет в результат `"ping"`, вторая — `"pong"`. Итог — слайс из `2n` элементов `ping, pong, ping, pong, …`. Порядок обеспечивай **только синхронизацией** (каналами) — без `time.Sleep`. Запись в общий слайс должна быть безопасной за счёт передачи «хода». Все горутины должны завершиться. `n <= 0` → пустой результат.

2. `Dine(n, meals int, eat func(id int))` — задача об **обедающих философах**: `n` философов (`n >= 2`) сидят за круглым столом, между соседними лежит вилка (всего `n`); философ `id` использует вилки `id` и `(id+1) % n`. Каждый философ должен поесть ровно `meals` раз: берёт обе вилки, вызывает `eat(id)`, кладёт вилки. Два **соседа** никогда не должны есть одновременно. Решение обязано быть свободным от deadlock (подсказка: единый порядок захвата — сначала вилка с меньшим номером). Функция возвращается, когда все наелись.
@@ starter
package solution

func PingPong(n int) []string {
	return nil
}

func Dine(n, meals int, eat func(id int)) {
}
@@ tests
package solution

import (
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestPingPong(t *testing.T) {
	before := runtime.NumGoroutine()
	got := PingPong(1000)
	if len(got) != 2000 {
		t.Fatalf("len = %d", len(got))
	}
	for i, s := range got {
		want := "ping"
		if i%2 == 1 {
			want = "pong"
		}
		if s != want {
			t.Fatalf("на позиции %d: %q, want %q", i, s, want)
		}
	}
	if len(PingPong(0)) != 0 || !slices.Equal(PingPong(1), []string{"ping", "pong"}) {
		t.Error("крайние случаи")
	}
	time.Sleep(50 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("горутины не завершились: было %d, стало %d", before, after)
	}
}

func TestDine(t *testing.T) {
	const n, meals = 5, 50
	eating := make([]atomic.Bool, n)
	eaten := make([]atomic.Int32, n)
	var violations atomic.Int32

	done := make(chan struct{})
	go func() {
		Dine(n, meals, func(id int) {
			left, right := (id+n-1)%n, (id+1)%n
			if eating[left].Load() || eating[right].Load() {
				violations.Add(1) // сосед ест одновременно
			}
			eating[id].Store(true)
			time.Sleep(200 * time.Microsecond)
			eaten[id].Add(1)
			eating[id].Store(false)
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock: философы не наелись за 10 секунд")
	}
	for i := 0; i < n; i++ {
		if got := eaten[i].Load(); got != meals {
			t.Errorf("философ %d поел %d раз, ожидалось %d", i, got, meals)
		}
	}
	if v := violations.Load(); v != 0 {
		t.Errorf("соседи ели одновременно %d раз", v)
	}
}

func TestDineTwo(t *testing.T) {
	var count atomic.Int32
	done := make(chan struct{})
	go func() { Dine(2, 100, func(int) { count.Add(1) }); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock при n=2")
	}
	if count.Load() != 200 {
		t.Errorf("count = %d", count.Load())
	}
}
@@ solution
package solution

import "sync"

func PingPong(n int) []string {
	if n <= 0 {
		return []string{}
	}
	pingTurn, pongTurn := make(chan struct{}), make(chan struct{})
	out := make([]string, 0, 2*n)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			<-pingTurn
			out = append(out, "ping")
			pongTurn <- struct{}{}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			<-pongTurn
			out = append(out, "pong")
			if i < n-1 {
				pingTurn <- struct{}{}
			}
		}
	}()
	pingTurn <- struct{}{} // первый ход
	wg.Wait()
	return out
}

func Dine(n, meals int, eat func(id int)) {
	forks := make([]sync.Mutex, n)
	var wg sync.WaitGroup
	for id := 0; id < n; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			first, second := id, (id+1)%n
			if first > second {
				first, second = second, first // единый порядок захвата — нет циклического ожидания
			}
			for m := 0; m < meals; m++ {
				forks[first].Lock()
				forks[second].Lock()
				eat(id)
				forks[second].Unlock()
				forks[first].Unlock()
			}
		}()
	}
	wg.Wait()
}
@@ hint
PingPong: два канала-«хода»; главная горутина отдаёт первый ход. Философы: мьютекс на вилку, захват всегда в порядке возрастания номеров вилок.
