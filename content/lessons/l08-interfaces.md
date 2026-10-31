---
id: l08
week: 2
day: 3
minutes: 120
title: Интерфейсы
goals: Неявная реализация, маленькие интерфейсы, type assertion/switch, ловушка nil-интерфейса, мокирование
lc: design-browser-history | Design Browser History | Medium | Слайс + текущий индекс; при visit обрезаем «будущее»
lc: implement-stack-using-queues | Implement Stack using Queues | Easy | Очередь на слайсе; при push прокручиваем элементы
lc: design-underground-system | Design Underground System | Medium | Два map: текущие поездки и агрегаты (сумма, количество) по маршрутам
---

## Идея

Интерфейс описывает **поведение** — набор методов. Тип реализует интерфейс **неявно**: достаточно иметь нужные методы, никаких `implements`. Это позволяет объявлять интерфейс там, где он **используется**, а не там, где реализован.

```go run
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }

func (r Rect) Area() float64        { return r.W * r.H }
func (r Rect) Perimeter() float64   { return 2 * (r.W + r.H) }
func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

func describe(s Shape) {
	fmt.Printf("%T: площадь %.2f, периметр %.2f\n", s, s.Area(), s.Perimeter())
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}}
	for _, s := range shapes {
		describe(s)
	}
}
```

## Правила хорошего тона

1. **Маленькие интерфейсы** (1–3 метода): `io.Reader`, `io.Writer`, `fmt.Stringer`, `error`, `sort.Interface`. «Чем больше интерфейс, тем слабее абстракция» — Роб Пайк.
2. **«Принимай интерфейсы, возвращай структуры»**: параметры функций — интерфейсы (гибкость и тестируемость), результаты — конкретные типы.
3. Определяй интерфейс **на стороне потребителя**, не заранее у производителя.
4. Не создавай интерфейс ради интерфейса — пока нет двух реализаций или потребности в подмене в тестах, он не нужен.

Стандартные интерфейсы, которые нужно знать:

| Интерфейс | Метод |
|---|---|
| `fmt.Stringer` | `String() string` |
| `error` | `Error() string` |
| `io.Reader` | `Read(p []byte) (n int, err error)` |
| `io.Writer` | `Write(p []byte) (n int, err error)` |
| `io.Closer` | `Close() error` |
| `sort.Interface` | `Len`, `Less(i, j)`, `Swap(i, j)` |
| `http.Handler` | `ServeHTTP(w, r)` |
| `context.Context`, `json.Marshaler`, `encoding.TextMarshaler`... | |

## Как устроен интерфейс внутри

Значение интерфейса — пара **(динамический тип, динамическое значение)**. Интерфейс равен `nil`, только когда **оба** поля пусты.

### Ловушка: nil-указатель в интерфейсе ≠ nil-интерфейс

```go run
package main

import "fmt"

type MyErr struct{}

func (*MyErr) Error() string { return "my error" }

func mayFail(fail bool) error {
	var e *MyErr // nil-указатель типа *MyErr
	if fail {
		e = &MyErr{}
	}
	return e // ОШИБКА: возвращаем интерфейс (тип=*MyErr, значение=nil) — он НЕ nil
}

func mayFailOK(fail bool) error {
	if fail {
		return &MyErr{}
	}
	return nil // явный nil-интерфейс
}

func main() {
	fmt.Println(mayFail(false) == nil)   // false! (классическая ловушка на собеседованиях)
	fmt.Println(mayFailOK(false) == nil) // true
}
```

**Правило:** функции, возвращающие `error`, должны возвращать литеральный `nil`, а не nil-указатель конкретного типа.

## Type assertion и type switch

```go run
package main

import "fmt"

type Stringer interface{ String() string }
type ID int

func (i ID) String() string { return fmt.Sprintf("ID-%d", int(i)) }

func describe(v any) string { // any = interface{} — «что угодно»
	switch x := v.(type) { // x принимает конкретный тип внутри каждой ветки
	case nil:
		return "nil"
	case int, int64:
		return fmt.Sprintf("целое %v", x)
	case string:
		return "строка длины " + fmt.Sprint(len(x))
	case []int:
		return fmt.Sprintf("слайс из %d int", len(x))
	case Stringer:
		return "Stringer: " + x.String()
	case error:
		return "ошибка: " + x.Error()
	default:
		return fmt.Sprintf("неизвестный тип %T", x)
	}
}

func main() {
	for _, v := range []any{nil, 42, "hey", []int{1, 2}, ID(7), fmt.Errorf("oops"), 3.14} {
		fmt.Println(describe(v))
	}

	var v any = "text"
	s, ok := v.(string) // форма с ok — безопасная
	n, ok2 := v.(int)
	fmt.Println(s, ok, n, ok2)
	// _ = v.(int)  // без ok: panic: interface conversion
}
```

Пустой интерфейс `any` теряет информацию о типе — используй редко (JSON с неизвестной структурой, универсальные контейнеры до дженериков). С Go 1.18 для большинства случаев лучше дженерики.

## Встраивание интерфейсов и расширение

```go
type ReadWriter interface {
	io.Reader
	io.Writer
}
```

Встраивание интерфейса в **структуру** — приём «декоратор»: переопределяем один метод, остальные делегируются:

```go run
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type countingWriter struct {
	io.Writer // встроенный интерфейс: Write делегируется ему, если не переопределён
	n int
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.n += len(p)
	return c.Writer.Write(p)
}

func main() {
	cw := &countingWriter{Writer: os.Stdout}
	io.Copy(cw, strings.NewReader("привет, интерфейсы\n"))
	fmt.Println("записано байт:", cw.n)
}
```

## Method sets и pointer receiver

```go
type Speaker interface{ Speak() string }
type Cat struct{}
func (c *Cat) Speak() string { return "мяу" } // метод на *Cat

var _ Speaker = &Cat{} // ОК
// var _ Speaker = Cat{} // ошибка: Cat does not implement Speaker (Speak has pointer receiver)
```

Приём **проверки на этапе компиляции**: `var _ Speaker = (*Cat)(nil)` — упадёт сборка, если контракт нарушен.

## Подмена зависимостей в тестах

Интерфейсы — основа тестируемого кода: сервис зависит от `interface{ Get(id) (User, error) }`, в проде — настоящая БД, в тесте — крошечный fake:

```go
type UserStore interface{ Get(id int) (User, error) }

type Service struct{ store UserStore }

type fakeStore map[int]User
func (f fakeStore) Get(id int) (User, error) { /* ... */ }
```

## Сравнение интерфейсов

Два интерфейса равны, если равны динамические типы и значения. Сравнение интерфейсов с несравнимыми динамическими значениями (слайсы, map) вызывает **panic во время выполнения**.

## Проверь себя

<details><summary>Как тип реализует интерфейс в Go?</summary>Неявно — достаточно набора нужных методов (с учётом method set: для pointer receiver реализует только <code>*T</code>).</details>

<details><summary>Почему <code>err != nil</code> бывает true, хотя функция вернула «nil»?</summary>Вернули nil-указатель конкретного типа, завёрнутый в интерфейс: у него есть динамический тип, значит интерфейс не nil.</details>

<details><summary>В чём разница type assertion и type switch?</summary>Assertion проверяет/извлекает один тип (с ok — безопасно), switch — разбор по многим типам.</details>

<details><summary>Почему советуют «принимать интерфейсы, возвращать структуры»?</summary>Параметр-интерфейс упрощает подмену и расширение; возвращаемая структура не прячет возможности и не навязывает абстракцию потребителю.</details>
