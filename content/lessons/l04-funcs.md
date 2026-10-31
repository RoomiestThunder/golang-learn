---
id: l04
week: 1
day: 4
minutes: 120
title: Функции, замыкания, defer, panic
goals: Множественные возвраты, variadic, функции как значения, замыкания, defer/panic/recover
lc: fibonacci-number | Fibonacci Number | Easy | Рекурсия → мемоизация → два числа без массива
lc: pascals-triangle | Pascal's Triangle | Easy | Каждая строка строится из предыдущей; слайс слайсов
lc: climbing-stairs | Climbing Stairs | Easy | Это снова Фибоначчи — первый шаг к динамическому программированию
---

## Объявление и множественные возвращаемые значения

```go run
package main

import (
	"errors"
	"fmt"
)

// параметры одного типа можно группировать: (a, b int)
func divmod(a, b int) (int, int, error) {
	if b == 0 {
		return 0, 0, errors.New("деление на ноль")
	}
	return a / b, a % b, nil
}

// именованные результаты: инициализированы нулями, голый return возвращает их
func minmax(xs []int) (lo, hi int) {
	if len(xs) == 0 {
		return
	}
	lo, hi = xs[0], xs[0]
	for _, x := range xs[1:] {
		lo, hi = min(lo, x), max(hi, x)
	}
	return
}

func main() {
	q, r, err := divmod(17, 5)
	fmt.Println(q, r, err)
	_, _, err = divmod(1, 0)
	fmt.Println(err)
	fmt.Println(minmax([]int{4, -2, 9}))
}
```

Конвенция: `error` — **последний** результат. Именованные результаты применяй, когда они документируют смысл (`lo, hi`) или нужны `defer`; голый `return` в длинных функциях ухудшает читаемость.

## Variadic-функции

```go run
package main

import "fmt"

func sum(prefix string, nums ...int) int { // nums внутри — []int
	total := 0
	for _, n := range nums {
		total += n
	}
	fmt.Println(prefix, len(nums), "аргументов")
	return total
}

func main() {
	fmt.Println(sum("a"))                // nums == nil
	fmt.Println(sum("b", 1, 2, 3))
	xs := []int{4, 5}
	fmt.Println(sum("c", xs...))         // распаковка слайса; слайс НЕ копируется
}
```

## Функции — значения первого класса

Функцию можно положить в переменную, передать аргументом, вернуть. Тип функции: `func(int, int) int`.

```go run
package main

import (
	"fmt"
	"sort"
	"strings"
)

type Op func(float64, float64) float64 // именованный функциональный тип

func apply(op Op, a, b float64) float64 { return op(a, b) }

func main() {
	add := func(a, b float64) float64 { return a + b } // анонимная функция
	fmt.Println(apply(add, 2, 3))

	words := strings.Fields("пирог яблоко арбуз")
	sort.Slice(words, func(i, j int) bool { return words[i] < words[j] }) // типичное применение
	fmt.Println(words)

	func() { fmt.Println("IIFE: вызов сразу") }()
}
```

## Замыкания

Замыкание «захватывает» переменные внешней области **по ссылке** — они живут, пока на них ссылается функция:

```go run
package main

import "fmt"

func counter() func() int {
	n := 0 // переживёт выход из counter
	return func() int {
		n++
		return n
	}
}

func main() {
	a, b := counter(), counter() // у каждого свой n
	fmt.Println(a(), a(), a(), b())
}
```

Применения: генераторы, декораторы/middleware, ленивые вычисления, колбэки с состоянием, мемоизация.

## defer

`defer f(args)` откладывает вызов до выхода из **окружающей функции**. Правила:

1. Аргументы вычисляются **в момент `defer`**, а вызов — позже.
2. Несколько `defer` выполняются в порядке **LIFO** (стек).
3. `defer` видит и может менять **именованные** результаты.
4. Выполняется и при `panic`.

```go run
package main

import "fmt"

func demo() (result int) {
	defer func() { result *= 2 }() // изменит возвращаемое значение
	for i := 0; i < 3; i++ {
		defer fmt.Print("defer ", i, "; ") // LIFO: 2, 1, 0
	}
	x := 1
	defer fmt.Println("аргумент вычислен сразу: x =", x)
	x = 100
	return 21
}

func main() {
	fmt.Println("\nрезультат:", demo())
}
```

Типичное применение — освобождение ресурсов сразу после получения:

```go
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close()          // гарантированно закроем на любом пути выхода
mu.Lock(); defer mu.Unlock()
```

> [!WARNING] `defer` в цикле накапливается до конца **функции**, а не итерации. Для обработки файлов в цикле выноси тело в отдельную функцию.

## panic и recover

`panic` — аварийная остановка: размотка стека с выполнением defer'ов. Используется для **невосстановимых программных ошибок** (nil-разыменование, выход за границы слайса, деление на 0 — паникуют сами). Для ожидаемых ошибок — возвращай `error`.

`recover()` работает **только внутри defer** и останавливает панику:

```go run
package main

import "fmt"

func safe(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("перехвачена паника: %v", r)
		}
	}()
	f()
	return nil
}

func main() {
	fmt.Println(safe(func() { var m map[string]int; m["a"] = 1 }))   // запись в nil-map
	fmt.Println(safe(func() { s := []int{1}; _ = s[5] }))            // индекс вне границ
	fmt.Println(safe(func() { panic("вручную") }))
	fmt.Println(safe(func() {}))
}
```

Легитимно: `recover` на границе горутины/HTTP-обработчика, чтобы одна паника не уронила весь сервис (в `net/http` это уже сделано для каждого запроса). Паника в горутине без recover **убивает весь процесс**.

## Рекурсия

Go не оптимизирует хвостовую рекурсию, но стеки горутин растут динамически (до 1 ГБ по умолчанию), так что глубокая рекурсия возможна. Для обходов часто лучше явный стек/очередь.

## Проверь себя

<details><summary>Порядок выполнения нескольких defer?</summary>LIFO — последний отложенный выполнится первым.</details>

<details><summary>Когда вычисляются аргументы отложенного вызова?</summary>В момент выполнения самого оператора <code>defer</code>, а не при вызове.</details>

<details><summary>Где работает recover?</summary>Только напрямую внутри отложенной функции (defer) во время паники; вне defer возвращает nil.</details>

<details><summary>Что значит «замыкание захватывает по ссылке»?</summary>Замкнутая переменная общая для всех, кто её видит; изменения видны и снаружи, и в других замыканиях.</details>
