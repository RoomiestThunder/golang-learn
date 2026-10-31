---
id: l12
week: 3
day: 2
minutes: 120
title: Каналы и select
goals: Небуферизованные и буферизованные каналы, close, range, направления, select, таймауты
lc: lru-cache | LRU Cache | Medium | Map + двусвязный список (container/list); все операции O(1)
lc: kth-largest-element-in-a-stream | Kth Largest Element in a Stream | Easy | Мин-куча размера k (container/heap)
lc: design-twitter | Design Twitter | Medium | Слияние лент через кучу, map подписок
---

> «Не общайтесь через разделяемую память — делитесь памятью через общение.» Каналы — основной способ передавать **владение данными** между горутинами.

## Основы

Канал — типизированная очередь с синхронизацией. `ch := make(chan T)` — **небуферизованный**: отправка и приём происходят одновременно (rendezvous); `make(chan T, n)` — **буферизованный** на n элементов.

```go run
package main

import "fmt"

func main() {
	ch := make(chan string) // небуферизованный

	go func() {
		ch <- "ping" // блокируется, пока кто-то не примет
	}()

	msg := <-ch // блокируется, пока кто-то не отправит
	fmt.Println(msg)

	buf := make(chan int, 2) // буфер на 2
	buf <- 1
	buf <- 2 // не блокируется, пока буфер не полон
	fmt.Println(len(buf), cap(buf), <-buf, <-buf)
}
```

Таблица поведения:

| Операция | nil-канал | открытый | закрытый |
|---|---|---|---|
| отправка `ch <- v` | блок навсегда | блок, если нет получателя/места | **panic** |
| приём `<-ch` | блок навсегда | блок, если пусто | пустой буфер → нулевое значение, `ok=false` |
| `close(ch)` | **panic** | ок | **panic** |

## close и range

`close(ch)` сообщает «больше данных не будет». Закрывает **отправитель** (владелец) и только один раз; получатели не закрывают. Приём из закрытого канала не блокируется: сначала отдаёт остаток буфера, затем нулевые значения с `ok == false`.

```go run
package main

import "fmt"

func producer(n int) <-chan int { // <-chan int — только для чтения
	out := make(chan int)
	go func() {
		defer close(out) // сигнал завершения
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func main() {
	for v := range producer(5) { // range читает до закрытия канала
		fmt.Print(v, " ")
	}
	fmt.Println()

	ch := make(chan int, 1)
	ch <- 7
	close(ch)
	v, ok := <-ch
	v2, ok2 := <-ch
	fmt.Println(v, ok, v2, ok2)
}
```

**Направленные каналы**: `chan<- T` только отправка, `<-chan T` только приём — фиксируют контракт в сигнатуре.

## select — ожидание нескольких операций

```go run
package main

import (
	"fmt"
	"time"
)

func main() {
	a := make(chan string)
	b := make(chan string)
	go func() { time.Sleep(20 * time.Millisecond); a <- "от A" }()
	go func() { time.Sleep(60 * time.Millisecond); b <- "от B" }()

	for i := 0; i < 2; i++ {
		select { // выбирает ту ветку, что готова; если готовы несколько — случайную
		case m := <-a:
			fmt.Println(m)
		case m := <-b:
			fmt.Println(m)
		case <-time.After(200 * time.Millisecond): // таймаут
			fmt.Println("таймаут")
		}
	}

	// default — неблокирующая операция
	c := make(chan int, 1)
	select {
	case v := <-c:
		fmt.Println("получили", v)
	default:
		fmt.Println("пусто, не ждём")
	}
}
```

### Типичные идиомы

```go
// 1. сигнал остановки: канал пустых структур, закрытие = broadcast всем читателям
done := make(chan struct{})
go func() {
	for {
		select {
		case <-done:
			return
		case job := <-jobs:
			process(job)
		}
	}
}()
close(done)

// 2. таймаут одной операции — лучше через context (урок 14)

// 3. отключение ветки select: присвоить nil-канал (чтение из nil блокируется навсегда)
```

## Ошибки и deadlock

- `fatal error: all goroutines are asleep - deadlock!` — все горутины заблокированы, runtime это обнаруживает (но только когда зависли **все**).
- Отправка в небуферизованный канал без получателя — классическая блокировка (и утечка, если в горутине).
- Закрытие канала дважды / отправка в закрытый — panic. Если несколько отправителей — закрывает координатор (через `WaitGroup`: «последний гасит свет»).
- `time.After` в цикле создаёт таймеры, не освобождающиеся до срабатывания (до Go 1.23 — утечка); используй `time.NewTimer` + `Reset`/`Stop` в горячих циклах.

```go run
package main

import (
	"fmt"
	"sync"
)

// Правило "один закрывающий": merge закрывает out, когда все входы исчерпаны
func merge(cs ...<-chan int) <-chan int {
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
	go func() { wg.Wait(); close(out) }()
	return out
}

func gen(vals ...int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for _, v := range vals {
			c <- v
		}
	}()
	return c
}

func main() {
	sum := 0
	for v := range merge(gen(1, 2, 3), gen(10, 20), gen(100)) {
		sum += v
	}
	fmt.Println("сумма:", sum)
}
```

## Канал или мьютекс?

- **Каналы**: передача данных/владения, конвейеры, координация и сигналы, распределение работы.
- **Мьютекс**: защита общего состояния (кэш, счётчик, структура данных), когда горутины читают/пишут одно и то же.

Эмпирика: если описываешь «поток данных» — канал, если «защищаю поле структуры» — `sync.Mutex`.

## Проверь себя

<details><summary>Чем отличается небуферизованный канал от буферизованного?</summary>Небуферизованный синхронизирует отправителя и получателя (оба должны быть готовы), буферизованный позволяет отправителю продолжить, пока буфер не заполнен.</details>

<details><summary>Что вернёт чтение из закрытого канала?</summary>Оставшиеся элементы буфера, затем нулевое значение типа и <code>ok == false</code>; не блокируется.</details>

<details><summary>Кто должен закрывать канал?</summary>Отправляющая сторона (владелец), один раз; получатель канал не закрывает.</details>

<details><summary>Что делает select, если готовы сразу несколько case?</summary>Выбирает один из готовых случайным образом.</details>

<details><summary>Как «выключить» ветку select?</summary>Присвоить соответствующему каналу nil — операции с nil-каналом блокируются навсегда и ветка не выбирается.</details>
