---
id: l03
week: 1
day: 3
minutes: 120
title: Управление потоком: if, for, switch
goals: Единственный цикл for, range, switch без fallthrough, метки, break/continue
lc: count-primes | Count Primes | Medium | Решето Эратосфена: вычёркиваем кратные начиная с p*p
lc: happy-number | Happy Number | Easy | Цикл + map[int]bool для обнаружения цикла (или два указателя Флойда)
lc: power-of-two | Power of Two | Easy | Решение в одну строку через n & (n-1)
---

## if

Скобки у условия не нужны, фигурные у тела — обязательны. Перед условием можно выполнить **краткую инструкцию**; её переменные видны только внутри if/else — главный идиоматичный приём для обработки ошибок:

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	if n, err := strconv.Atoi("42"); err != nil {
		fmt.Println("ошибка:", err)
	} else if n > 40 {
		fmt.Println("большое число", n) // n и err видны во всех ветках
	} else {
		fmt.Println("маленькое", n)
	}
	// здесь n и err уже не существуют
}
```

Идиома **раннего возврата**: сначала обрабатываем ошибки и крайние случаи и выходим, «счастливый путь» остаётся без вложенности (не пиши `else` после `return`).

## for — единственный цикл

```go run
package main

import "fmt"

func main() {
	// 1. классический
	for i := 0; i < 3; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// 2. как while
	n := 1
	for n < 100 {
		n *= 3
	}
	fmt.Println(n)

	// 3. бесконечный + break
	c := 0
	for {
		c++
		if c == 5 {
			break
		}
	}

	// 4. range по слайсу: индекс и значение (значение — КОПИЯ элемента)
	for i, v := range []string{"a", "b", "c"} {
		fmt.Print(i, "=", v, " ")
	}
	fmt.Println()

	// 5. range по числу (Go 1.22+)
	for i := range 3 {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// 6. range по строке — по рунам (подробности в уроке про строки)
	for i, r := range "Go🙂" {
		fmt.Print(i, ":", string(r), " ")
	}
	fmt.Println()

	// 7. range по map — порядок СЛУЧАЙНЫЙ
	m := map[string]int{"x": 1}
	for k, v := range m {
		fmt.Println(k, v)
	}

	// 8. range по каналу — до закрытия (урок про каналы); range по функции-итератору (Go 1.23)
}
```

### Переменная цикла: каждая итерация — своя копия (с Go 1.22)

До 1.22 все замыкания внутри цикла разделяли **одну** переменную `i` — знаменитый баг с горутинами. Теперь `for i := ...` создаёт новую `i` на каждой итерации, и такой код безопасен:

```go run
package main

import "fmt"

func main() {
	var fns []func()
	for i := 0; i < 3; i++ {
		fns = append(fns, func() { fmt.Print(i, " ") })
	}
	for _, f := range fns {
		f()
	}
	fmt.Println("// до Go 1.22 было бы 3 3 3")
}
```

> [!INTERVIEW] Про loopvar спрашивают часто: «что напечатает код и что изменилось в Go 1.22?». Ответ: поведение зависит от директивы `go` в go.mod — для модулей с `go 1.22+` каждая итерация получает новую переменную.

## switch

Работает без `break` (по умолчанию выходит после выбранной ветки), условия могут быть любыми выражениями, `fallthrough` нужен редко и пишется явно:

```go run
package main

import "fmt"

func grade(score int) string {
	switch { // switch без выражения = цепочка if/else if
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	default:
		return "C"
	}
}

func main() {
	switch day := 3; day {
	case 1, 7: // несколько значений в одной ветке
		fmt.Println("выходной")
	case 2, 3, 4, 5, 6:
		fmt.Println("будни")
		fallthrough // принудительно провалиться в следующую ветку
	default:
		fmt.Println("(default после fallthrough)")
	}
	fmt.Println(grade(80), grade(95), grade(10))
}
```

Есть ещё `switch v := x.(type)` — type switch для интерфейсов (урок 8) и `select` для каналов (урок 12).

## break, continue и метки

```go run
package main

import "fmt"

func main() {
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 2 {
				continue outer // следующая итерация ВНЕШНЕГО цикла
			}
			if i == 2 {
				break outer // выход из обоих циклов
			}
			fmt.Print(i, j, " ")
		}
	}
	fmt.Println()
}
```

> [!WARNING] `break` внутри `switch`/`select` внутри `for` выходит только из `switch`/`select`, а не из цикла. Для выхода из цикла нужна метка.

## Практические приёмы

- Счётчик + условие: `for i := len(s) - 1; i >= 0; i--` — обход с конца.
- Изменение коллекции во время `range` по слайсу — range вычисляет выражение один раз (длина зафиксирована).
- `for range ch { }` и `for range 5 { }` — когда значение не нужно.
- `goto` существует, но почти не используется.

## Проверь себя

<details><summary>Сколько в Go видов циклов?</summary>Один — <code>for</code>, у него несколько форм (3-компонентный, while-подобный, бесконечный, range).</details>

<details><summary>Как switch ведёт себя после совпавшей ветки?</summary>Выходит из switch сам (без <code>break</code>); для провала в следующую ветку нужен <code>fallthrough</code>.</details>

<details><summary>Одинаков ли порядок обхода map?</summary>Нет, он намеренно рандомизирован. Нужен порядок — собери ключи в слайс и отсортируй.</details>

<details><summary>Что поменялось в семантике переменной цикла в Go 1.22?</summary>Каждая итерация имеет свою копию переменной, поэтому замыкания/горутины видят «правильное» значение.</details>
