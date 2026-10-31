---
id: l07
week: 2
day: 2
minutes: 120
title: Структуры, методы, указатели, встраивание
goals: Описывать данные структурами, выбирать value/pointer receiver, использовать композицию через embedding
lc: implement-queue-using-stacks | Implement Queue using Stacks | Easy | Две стопки: входная и выходная, переливаем лениво
lc: min-stack | Min Stack | Medium | Храни рядом с каждым элементом текущий минимум
lc: design-hashmap | Design HashMap | Easy | Массив бакетов + цепочки; тренирует структуры и методы
---

## Структуры

Структура — набор именованных полей. Это основной способ описывать данные в Go (классов нет).

```go run
package main

import "fmt"

type Address struct {
	City, Street string
}

type User struct {
	ID    int
	Name  string
	Email string
	Addr  Address
	Tags  []string
}

func main() {
	u1 := User{ID: 1, Name: "Ann"}                 // именованные поля — предпочтительный вариант
	u2 := User{2, "Bob", "b@x.io", Address{"Рим", "Via"}, nil} // позиционный — хрупкий, только для малых типов
	var u3 User                                     // zero value: все поля нулевые
	u3.Addr.City = "Киев"                           // вложенное поле

	fmt.Printf("%+v\n%+v\n%+v\n", u1, u2, u3)

	p := &u1          // указатель на структуру
	p.Name = "Anna"   // (*p).Name = ... — Go разыменовывает автоматически
	fmt.Println(u1.Name) // изменение через указатель видно в исходной переменной

	a, b := Address{"Рим", "Via"}, Address{"Рим", "Via"}
	fmt.Println(a == b) // структуры сравнимы через ==, если сравнимы все поля
	// u1 == u2 не скомпилируется: в User есть слайс Tags (слайсы, map и функции несравнимы)
}
```

**Анонимные структуры** удобны для табличных тестов и разовых данных: `point := struct{ X, Y int }{1, 2}`.

**Теги полей** — метаданные для рефлексии (JSON, БД, валидация):

```go
type Product struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price,omitempty"` // не включать, если ноль
	Secret string `json:"-"`               // не сериализовать
}
```

## Методы

Метод — функция с **получателем (receiver)**:

```go run
package main

import (
	"fmt"
	"math"
)

type Circle struct{ R float64 }

// value receiver: получает КОПИЮ; читает, но не меняет оригинал
func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

// pointer receiver: получает указатель; может менять состояние и не копирует структуру
func (c *Circle) Scale(k float64) { c.R *= k }

func main() {
	c := Circle{R: 2}
	c.Scale(1.5)          // Go сам сделает (&c).Scale(1.5), если c адресуема
	fmt.Printf("%.2f %.2f\n", c.R, c.Area())

	f := c.Area           // method value: связан с копией c на момент взятия
	c.R = 100
	fmt.Printf("%.2f\n", f())
}
```

### Value или pointer receiver?

Используй **pointer receiver**, если:
1. метод меняет получателя;
2. структура большая (дешевле передать указатель);
3. структура содержит `sync.Mutex` и подобное (копировать нельзя!);
4. **для согласованности** — если хоть один метод типа с указателем, пусть все будут с указателем.

Иначе — value receiver (маленькие неизменяемые типы вроде `time.Time`, `Point`).

### Наборы методов (method sets)

- Для значения `T` в набор входят методы с получателем `T`.
- Для `*T` — методы и `T`, и `*T`.

Поэтому `T` не удовлетворяет интерфейсу, если нужные методы объявлены на `*T` (подробнее в уроке 8).

### Конструкторы и идиома «полезный zero value»

В Go нет конструкторов; по соглашению пишут функцию `NewX`. Но лучше проектировать тип так, чтобы `var x X` уже работал:

```go run
package main

import "fmt"

type Stack struct {
	items []int // nil-слайс — валидное начальное состояние
}

func (s *Stack) Push(v int) { s.items = append(s.items, v) }

func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func main() {
	var s Stack // готов к использованию без конструктора
	s.Push(1)
	s.Push(2)
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
}
```

Когда инварианты нужно навязать (валидация, значения по умолчанию) — `NewX(...) (*X, error)`; поля делают неэкспортируемыми, чтобы пользователь не мог создать невалидное состояние.

## Встраивание (embedding)

Встраивание — это **композиция** с автоматическим «продвижением» (promotion) полей и методов. Это не наследование: нет виртуальных вызовов, внешний тип не «является» внутренним для системы типов.

```go run
package main

import "fmt"

type Animal struct{ Name string }

func (a Animal) Describe() string { return "Животное " + a.Name }
func (a *Animal) Rename(n string)  { a.Name = n }

type Dog struct {
	Animal        // встроенное поле: Dog.Name, Dog.Describe() работают напрямую
	Breed string
}

// переопределение: метод внешнего типа затеняет продвинутый
func (d Dog) Describe() string { return "Собака " + d.Name + " (" + d.Breed + "), " + d.Animal.Describe() }

func main() {
	d := Dog{Animal{"Рекс"}, "овчарка"}
	d.Rename("Шарик") // метод *Animal доступен через d
	fmt.Println(d.Describe())
	fmt.Println(d.Animal.Describe())
}
```

Типичное применение: встроить `sync.Mutex`, чтобы получить `Lock/Unlock`; расширить поведение стандартного типа; встроить интерфейс, чтобы переопределить лишь часть методов (декоратор).

> [!WARNING] Встраивание *экспортирует* методы встроенного типа в API внешнего. Если это нежелательно — используй обычное поле (`mu sync.Mutex`) и делегирование.

## Порядок полей и память

Компилятор выравнивает поля, поэтому порядок влияет на размер структуры: `struct{a bool; b int64; c bool}` занимает 24 байта, а `struct{b int64; a, c bool}` — 16. Для обычного кода это неважно, для миллионов объектов — важно (`unsafe.Sizeof`, линтер `fieldalignment`).

## Паттерн Functional Options

Гибкие конструкторы без десятка параметров:

```go
type Server struct { addr string; timeout time.Duration }
type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{addr: addr, timeout: 30 * time.Second}
	for _, o := range opts { o(s) }
	return s
}
// NewServer(":8080", WithTimeout(5*time.Second))
```

## Проверь себя

<details><summary>Чем отличаются value и pointer receiver?</summary>Value получает копию (изменения не видны вызывающему), pointer — адрес (можно менять и не копировать большие данные).</details>

<details><summary>Можно ли вызвать pointer-метод у значения в map?</summary>Нет: элементы map не адресуемы (<code>m[k].PtrMethod()</code> не скомпилируется). Нужно хранить указатели или вынимать в переменную.</details>

<details><summary>Является ли embedding наследованием?</summary>Нет. Это композиция с продвижением методов; внутренний метод не знает о внешнем типе (нет виртуальной диспетчеризации).</details>

<details><summary>Зачем нужен тег <code>json:"name,omitempty"</code>?</summary>Задаёт имя поля в JSON и пропускает нулевые значения при сериализации.</details>
