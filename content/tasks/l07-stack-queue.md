---
title: Стек и очередь на структурах
level: easy
---
@@ statement
Реализуй два типа с методами. Оба должны работать как **нулевое значение** (`var s Stack` готов к использованию без конструктора).

`Stack`: `Push(v int)`, `Pop() (int, bool)`, `Peek() (int, bool)`, `Len() int`.

`Queue` (FIFO): `Enqueue(v int)`, `Dequeue() (int, bool)`, `Len() int`. `Dequeue` должен работать за амортизированное O(1) — не сдвигай весь слайс на каждый вызов (храни индекс начала `head` и иногда сжимай).

Пустые `Pop`/`Peek`/`Dequeue` возвращают `0, false`.
@@ starter
package solution

type Stack struct {
	// TODO: поля
}

func (s *Stack) Push(v int)         {}
func (s *Stack) Pop() (int, bool)   { return 0, false }
func (s *Stack) Peek() (int, bool)  { return 0, false }
func (s *Stack) Len() int           { return 0 }

type Queue struct {
	// TODO: поля
}

func (q *Queue) Enqueue(v int)       {}
func (q *Queue) Dequeue() (int, bool) { return 0, false }
func (q *Queue) Len() int             { return 0 }
@@ tests
package solution

import "testing"

func TestStack(t *testing.T) {
	var s Stack
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop пустого стека")
	}
	if _, ok := s.Peek(); ok {
		t.Fatal("Peek пустого стека")
	}
	for i := 1; i <= 3; i++ {
		s.Push(i)
	}
	if v, ok := s.Peek(); !ok || v != 3 || s.Len() != 3 {
		t.Fatalf("Peek = %d %v, Len = %d", v, ok, s.Len())
	}
	for want := 3; want >= 1; want-- {
		if v, ok := s.Pop(); !ok || v != want {
			t.Fatalf("Pop = %d %v, want %d", v, ok, want)
		}
	}
	if s.Len() != 0 {
		t.Error("стек должен быть пуст")
	}
}

func TestQueue(t *testing.T) {
	var q Queue
	if _, ok := q.Dequeue(); ok {
		t.Fatal("Dequeue пустой очереди")
	}
	for i := 1; i <= 3; i++ {
		q.Enqueue(i)
	}
	if v, _ := q.Dequeue(); v != 1 {
		t.Fatalf("первый = %d", v)
	}
	q.Enqueue(4)
	for _, want := range []int{2, 3, 4} {
		if v, ok := q.Dequeue(); !ok || v != want {
			t.Fatalf("got %d %v, want %d", v, ok, want)
		}
	}
	if q.Len() != 0 {
		t.Error("очередь должна быть пуста")
	}
}

func TestQueueMany(t *testing.T) {
	var q Queue
	const n = 200_000
	for i := 0; i < n; i++ {
		q.Enqueue(i)
	}
	for i := 0; i < n; i++ {
		if v, ok := q.Dequeue(); !ok || v != i {
			t.Fatalf("на шаге %d получили %d %v", i, v, ok)
		}
	}
}
@@ solution
package solution

type Stack struct{ items []int }

func (s *Stack) Push(v int) { s.items = append(s.items, v) }

func (s *Stack) Pop() (int, bool) {
	v, ok := s.Peek()
	if ok {
		s.items = s.items[:len(s.items)-1]
	}
	return v, ok
}

func (s *Stack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack) Len() int { return len(s.items) }

type Queue struct {
	items []int
	head  int
}

func (q *Queue) Enqueue(v int) { q.items = append(q.items, v) }

func (q *Queue) Dequeue() (int, bool) {
	if q.head >= len(q.items) {
		return 0, false
	}
	v := q.items[q.head]
	q.head++
	if q.head > 64 && q.head*2 >= len(q.items) { // сжимаем, когда пустая часть велика
		q.items = append([]int(nil), q.items[q.head:]...)
		q.head = 0
	}
	return v, true
}

func (q *Queue) Len() int { return len(q.items) - q.head }
@@ hint
Очередь: `items []int` и `head int`. `Dequeue` читает `items[head]` и увеличивает `head`; когда `head` велик — копируй хвост в новый слайс.
