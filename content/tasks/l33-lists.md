---
title: Связные списки и монотонный стек
level: medium
---
@@ statement
Типы и функции для односвязного списка (`ListNode` уже объявлен):

- `ReverseList(head *ListNode) *ListNode` — развернуть список **на месте** (без выделения новых узлов), вернуть новую голову.
- `MergeSorted(a, b *ListNode) *ListNode` — слить два отсортированных списка в один отсортированный, **переиспользуя узлы** (стабильно: при равенстве первым идёт узел из `a`).
- `HasCycle(head *ListNode) bool` — есть ли цикл (алгоритм Флойда, O(1) памяти).
- `MiddleNode(head *ListNode) *ListNode` — средний узел; для чётной длины — **второй** из двух средних.
- `RemoveNthFromEnd(head *ListNode, n int) *ListNode` — удалить n-й узел с конца (`1 <= n <= длина`) за один проход (подсказка: dummy-голова и два указателя).

И задача на монотонный стек:
- `DailyTemperatures(t []int) []int` — для каждого дня через сколько дней будет теплее (0, если не будет). На входе 100 000 элементов — нужно O(n).
@@ starter
package solution

type ListNode struct {
	Val  int
	Next *ListNode
}

func ReverseList(head *ListNode) *ListNode { return head }

func MergeSorted(a, b *ListNode) *ListNode { return a }

func HasCycle(head *ListNode) bool { return false }

func MiddleNode(head *ListNode) *ListNode { return head }

func RemoveNthFromEnd(head *ListNode, n int) *ListNode { return head }

func DailyTemperatures(t []int) []int { return nil }
@@ tests
package solution

import (
	"slices"
	"testing"
	"time"
)

func build(xs ...int) *ListNode {
	dummy := &ListNode{}
	tail := dummy
	for _, x := range xs {
		tail.Next = &ListNode{Val: x}
		tail = tail.Next
	}
	return dummy.Next
}

func toSlice(l *ListNode) []int {
	out := []int{}
	for i := 0; l != nil && i < 1_000_000; i++ {
		out = append(out, l.Val)
		l = l.Next
	}
	return out
}

func TestReverse(t *testing.T) {
	head := build(1, 2, 3, 4)
	nodes := []*ListNode{head, head.Next, head.Next.Next, head.Next.Next.Next}
	got := ReverseList(head)
	if !slices.Equal(toSlice(got), []int{4, 3, 2, 1}) {
		t.Errorf("got %v", toSlice(got))
	}
	if got != nodes[3] || nodes[0].Next != nil {
		t.Error("список нужно разворачивать на месте, переиспользуя узлы")
	}
	if ReverseList(nil) != nil || ReverseList(build(7)).Val != 7 {
		t.Error("крайние случаи")
	}
}

func TestMerge(t *testing.T) {
	tests := []struct{ a, b, want []int }{
		{[]int{1, 3, 5}, []int{2, 4, 6}, []int{1, 2, 3, 4, 5, 6}},
		{[]int{}, []int{1}, []int{1}}, {[]int{1, 1}, []int{}, []int{1, 1}}, {[]int{}, []int{}, []int{}}, {[]int{1, 2, 2}, []int{2, 3}, []int{1, 2, 2, 2, 3}},
	}
	for _, tt := range tests {
		if got := toSlice(MergeSorted(build(tt.a...), build(tt.b...))); !slices.Equal(got, tt.want) {
			t.Errorf("Merge(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
	a := &ListNode{Val: 1}
	b := &ListNode{Val: 1}
	if got := MergeSorted(a, b); got != a || got.Next != b {
		t.Error("стабильность: при равенстве сначала узел из a, узлы переиспользуются")
	}
}

func TestCycleAndMiddle(t *testing.T) {
	l := build(1, 2, 3, 4, 5)
	if HasCycle(l) || HasCycle(nil) {
		t.Error("цикла нет")
	}
	tail := l
	for tail.Next != nil {
		tail = tail.Next
	}
	tail.Next = l.Next // цикл на второй узел
	if !HasCycle(l) {
		t.Error("цикл есть")
	}
	if m := MiddleNode(build(1, 2, 3, 4, 5)); m.Val != 3 {
		t.Errorf("middle(5) = %d", m.Val)
	}
	if m := MiddleNode(build(1, 2, 3, 4)); m.Val != 3 {
		t.Errorf("middle(4) = %d, нужен второй средний", m.Val)
	}
	if m := MiddleNode(build(9)); m.Val != 9 {
		t.Error("один узел")
	}
}

func TestRemoveNth(t *testing.T) {
	tests := []struct {
		in   []int
		n    int
		want []int
	}{
		{[]int{1, 2, 3, 4, 5}, 2, []int{1, 2, 3, 5}}, {[]int{1}, 1, []int{}}, {[]int{1, 2}, 1, []int{1}}, {[]int{1, 2}, 2, []int{2}},
	}
	for _, tt := range tests {
		if got := toSlice(RemoveNthFromEnd(build(tt.in...), tt.n)); !slices.Equal(got, tt.want) {
			t.Errorf("Remove(%v, %d) = %v, want %v", tt.in, tt.n, got, tt.want)
		}
	}
}

func TestDailyTemperatures(t *testing.T) {
	got := DailyTemperatures([]int{73, 74, 75, 71, 69, 72, 76, 73})
	if !slices.Equal(got, []int{1, 1, 4, 2, 1, 1, 0, 0}) {
		t.Errorf("got %v", got)
	}
	if got := DailyTemperatures([]int{30, 40, 50, 60}); !slices.Equal(got, []int{1, 1, 1, 0}) {
		t.Errorf("got %v", got)
	}
	big := make([]int, 100_000)
	for i := range big {
		big[i] = 100_000 - i // убывающие: наивное решение даёт O(n²)
	}
	start := time.Now()
	res := DailyTemperatures(big)
	if time.Since(start) > time.Second || res[0] != 0 {
		t.Errorf("слишком медленно или неверно: %v", time.Since(start))
	}
}
@@ solution
package solution

type ListNode struct {
	Val  int
	Next *ListNode
}

func ReverseList(head *ListNode) *ListNode {
	var prev *ListNode
	for cur := head; cur != nil; {
		next := cur.Next
		cur.Next = prev
		prev, cur = cur, next
	}
	return prev
}

func MergeSorted(a, b *ListNode) *ListNode {
	dummy := &ListNode{}
	tail := dummy
	for a != nil && b != nil {
		if a.Val <= b.Val {
			tail.Next, a = a, a.Next
		} else {
			tail.Next, b = b, b.Next
		}
		tail = tail.Next
	}
	if a != nil {
		tail.Next = a
	} else {
		tail.Next = b
	}
	return dummy.Next
}

func HasCycle(head *ListNode) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}

func MiddleNode(head *ListNode) *ListNode {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
	}
	return slow
}

func RemoveNthFromEnd(head *ListNode, n int) *ListNode {
	dummy := &ListNode{Next: head}
	fast, slow := dummy, dummy
	for i := 0; i < n; i++ {
		fast = fast.Next
	}
	for fast.Next != nil {
		fast, slow = fast.Next, slow.Next
	}
	slow.Next = slow.Next.Next
	return dummy.Next
}

func DailyTemperatures(t []int) []int {
	res := make([]int, len(t))
	var st []int
	for i, cur := range t {
		for len(st) > 0 && t[st[len(st)-1]] < cur {
			j := st[len(st)-1]
			st = st[:len(st)-1]
			res[j] = i - j
		}
		st = append(st, i)
	}
	return res
}
@@ hint
Dummy-голова (`dummy := &ListNode{Next: head}`) убирает особые случаи. Для `RemoveNthFromEnd`: быстрый указатель уходит на `n` шагов вперёд, затем оба движутся до конца.
