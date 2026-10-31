---
id: l33
week: 7
day: 3
minutes: 120
title: Алгоритмы III: стек, очередь, связные списки, куча
goals: Монотонный стек, операции со списками через указатели, своя куча, слияние отсортированных последовательностей
lc: reverse-linked-list | Reverse Linked List | Easy | Итеративно: prev, cur, next; затем рекурсивно
lc: merge-two-sorted-lists | Merge Two Sorted Lists | Easy | Фиктивная голова (dummy) избавляет от особых случаев
lc: remove-nth-node-from-end-of-list | Remove Nth Node From End of List | Medium | Быстрый указатель опережает на n шагов
lc: reorder-list | Reorder List | Medium | Найти середину → развернуть вторую половину → слить по очереди
lc: add-two-numbers | Add Two Numbers | Medium | Сложение с переносом по узлам
lc: daily-temperatures | Daily Temperatures | Medium | Монотонный стек индексов
lc: largest-rectangle-in-histogram | Largest Rectangle in Histogram | Hard | Монотонный стек возрастающих высот
lc: merge-k-sorted-lists | Merge k Sorted Lists | Hard | Куча из голов списков (container/heap)
---

## Стек и очередь

В Go отдельного типа нет — используют слайс:

```go
stack := []int{}
stack = append(stack, x)          // push
top := stack[len(stack)-1]        // peek
stack = stack[:len(stack)-1]      // pop

queue := []int{}
queue = append(queue, x)          // enqueue
front := queue[0]                 // peek
queue = queue[1:]                 // dequeue (память освободится при следующем росте; для очень длинных очередей — индекс head)
```

Классика на стек: скобки, вычисление выражений (RPN), «отмена», обход в глубину без рекурсии.

### Монотонный стек

Хранит индексы с **монотонными значениями**; позволяет для каждого элемента найти «следующий больший/меньший» за O(n) (каждый элемент кладётся и снимается раз).

```go run
package main

import "fmt"

// для каждого дня — через сколько дней будет теплее
func dailyTemperatures(t []int) []int {
	res := make([]int, len(t))
	var st []int // индексы с убывающими температурами
	for i, cur := range t {
		for len(st) > 0 && t[st[len(st)-1]] < cur {
			j := st[len(st)-1]
			st = st[:len(st)-1]
			res[j] = i - j // нашли для j следующий более тёплый день
		}
		st = append(st, i)
	}
	return res
}

func main() {
	fmt.Println(dailyTemperatures([]int{73, 74, 75, 71, 69, 72, 76, 73}))
}
```

Применения: next greater element, ближайший меньший, площадь гистограммы, «дождевая вода», максимум в скользящем окне (дек).

## Связный список в Go

```go
type ListNode struct {
	Val  int
	Next *ListNode
}
```

Приёмы, которые решают 90% задач:

1. **Dummy head** (`dummy := &ListNode{}`): единая обработка вставки/удаления первого узла; возвращай `dummy.Next`.
2. **Два указателя**: быстрый/медленный (середина, цикл), отстающий на n (n-й с конца).
3. **Разворот**: `prev, cur = cur, cur.Next` с перестановкой ссылок.
4. Осторожно с `nil`: проверяй `cur != nil && cur.Next != nil`.

```go run
package main

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

func fromSlice(xs ...int) *ListNode {
	dummy := &ListNode{}
	tail := dummy
	for _, x := range xs {
		tail.Next = &ListNode{Val: x}
		tail = tail.Next
	}
	return dummy.Next
}

func (l *ListNode) String() string {
	s := ""
	for ; l != nil; l = l.Next {
		s += fmt.Sprint(l.Val, " ")
	}
	return "[ " + s + "]"
}

func reverse(head *ListNode) *ListNode {
	var prev *ListNode
	for cur := head; cur != nil; {
		next := cur.Next // запомнить, пока не потеряли
		cur.Next = prev  // развернуть ссылку
		prev, cur = cur, next
	}
	return prev
}

func middle(head *ListNode) *ListNode {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
	}
	return slow
}

func hasCycle(head *ListNode) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}

func merge(a, b *ListNode) *ListNode {
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

func main() {
	l := fromSlice(1, 2, 3, 4, 5)
	fmt.Println(reverse(l), middle(fromSlice(1, 2, 3, 4, 5)).Val, hasCycle(l))
	fmt.Println(merge(fromSlice(1, 4, 6), fromSlice(2, 3, 7)))
}
```

## Куча (приоритетная очередь)

Двоичная куча в массиве: родитель `i`, дети `2i+1` и `2i+2`. `Push`/`Pop` — O(log n), `Peek` — O(1). В Go есть `container/heap` (интерфейс с `Push/Pop/Len/Less/Swap`), но удобнее иметь собственную обобщённую версию — ты напишешь её в задании.

Паттерны:
- **Top-K**: мин-куча размера k (хранит k наибольших; корень — k-й).
- **Слияние k отсортированных** последовательностей: куча из текущих голов.
- **Медиана потока**: две кучи (max-куча нижней половины + min-куча верхней).
- **Планирование задач** по приоритету/времени, алгоритм Дейкстры.

```go
// min-куча по умолчанию; для max — инвертируй сравнение
h := &IntHeap{}
heap.Push(h, 5)
min := heap.Pop(h).(int)
```

## Проверь себя

<details><summary>Как найти середину связного списка за один проход?</summary>Два указателя: быстрый шагает по 2, медленный по 1; когда быстрый у конца, медленный в середине.</details>

<details><summary>Зачем фиктивная голова?</summary>Чтобы вставка/удаление первого узла не требовали отдельной ветки кода; результат — <code>dummy.Next</code>.</details>

<details><summary>Какую сложность имеет Top-K через мин-кучу?</summary>O(n log k) по времени и O(k) по памяти.</details>

<details><summary>Что хранит монотонный стек?</summary>Индексы (или значения) в монотонном порядке; при приходе нового элемента снимаются все нарушающие порядок — на них и даётся ответ.</details>
