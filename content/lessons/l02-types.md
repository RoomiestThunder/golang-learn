---
id: l02
week: 1
day: 2
minutes: 120
title: Переменные, типы и константы
goals: Объявления, нулевые значения, числовые типы, преобразования, const и iota, пользовательские типы
lc: palindrome-number | Palindrome Number | Easy | Можно без строки: переверни половину числа
lc: plus-one | Plus One | Easy | Перенос разряда; если все девятки — слайс придётся удлинить
lc: roman-to-integer | Roman to Integer | Easy | map[byte]int и правило «меньшее перед большим — вычитаем»
---

## Объявление переменных

```go run
package main

import "fmt"

var global = "пакетная переменная" // вне функций := нельзя, только var

func main() {
	var a int          // нулевое значение: 0
	var b = 3.5        // тип выведен: float64
	var c, d = 1, "x"  // несколько сразу
	e := true          // короткая форма := только внутри функций
	f, g := 10, "hi"   // хотя бы одна новая переменная слева обязательна
	var (
		h int    = 7
		i string = "блок"
	)
	_ = global // _ — пустой идентификатор: значение выбрасывается

	fmt.Println(a, b, c, d, e, f, g, h, i)
	fmt.Printf("%T %T %T %T\n", b, c, d, e)
}
```

### Нулевые значения — ключевая особенность Go

Неинициализированная переменная **всегда** имеет нулевое значение, «мусора» нет:

| Тип | Zero value |
|---|---|
| `int`, `float64`, ... | `0` |
| `bool` | `false` |
| `string` | `""` |
| указатель, `slice`, `map`, `chan`, `func`, `interface` | `nil` |
| `struct` | каждое поле в zero value |

Хороший Go-код проектируют так, чтобы **нулевое значение было полезным** (`sync.Mutex`, `bytes.Buffer`, `var sb strings.Builder` работают сразу).

### Затенение (shadowing)

```go
x := 1
if true {
	x := 2 // НОВАЯ переменная, внешняя x не меняется
	_ = x
}
```

Частый источник багов: `err := ...` во внутреннем блоке «затеняет» внешнюю `err`. `go vet -vettool=shadow` ловит такие случаи.

## Базовые типы

| Группа | Типы |
|---|---|
| Целые со знаком | `int8 int16 int32 int64`, `int` (32/64 бит по платформе) |
| Беззнаковые | `uint8 uint16 uint32 uint64`, `uint`, `uintptr` |
| Алиасы | `byte` = `uint8`, `rune` = `int32` (кодовая точка Unicode) |
| Вещественные | `float32`, `float64` |
| Комплексные | `complex64`, `complex128` |
| Прочее | `bool`, `string` |

В повседневном коде: `int` для индексов и счётчиков, `float64` для дробных, `int64`/`uint64` когда важна ширина (ID, время, файлы), `byte`/`rune` для строк.

### Явные преобразования — неявных нет

```go run
package main

import (
	"fmt"
	"math"
	"strconv"
)

func main() {
	var i int = 42
	var f float64 = float64(i) / 5 // без float64(i) не скомпилируется: int / float64 нельзя
	var u8 uint8 = 250
	u8 += 10 // переполнение целых «заворачивается» молча: 260 % 256 = 4
	fmt.Println(f, u8)

	big := int64(math.MaxInt32) + 1
	fmt.Println(int32(big)) // усечение: -2147483648

	// число <-> строка: только через strconv (string(65) даст "A", а не "65"!)
	s := strconv.Itoa(65)
	n, err := strconv.Atoi("123x")
	fmt.Println(s, string(rune(65)), n, err)

	fmt.Println(7/2, -7/2, 7%3, -7%3, 7.0/2) // целочисленное деление усекает к нулю
}
```

> [!WARNING] Переполнение целых в Go не паникует. Для денег используй целые копейки (`int64`) или `math/big`, **никогда** `float64`: `0.1 + 0.2 != 0.3`.

## Константы и iota

Константы вычисляются на этапе компиляции. **Нетипизированные** константы (`const Pi = 3.14`) гибки: подходят к любому совместимому типу и имеют произвольную точность.

```go run
package main

import "fmt"

const Pi = 3.14159265358979 // untyped
const Name string = "typed"

type Weekday int

const (
	Sunday Weekday = iota // iota — счётчик внутри const-блока, начинается с 0
	Monday
	Tuesday
)

type Size uint64

const (
	_       = iota // пропускаем 0
	KB Size = 1 << (10 * iota)
	MB
	GB
)

func main() {
	var r float32 = Pi // untyped-константа подстроилась под float32
	fmt.Println(r, Name, Sunday, Monday, Tuesday)
	fmt.Println(KB, MB, GB)
}
```

## Свои типы: `type`

```go
type Celsius float64        // НОВЫЙ тип с базовым float64; Celsius и float64 несовместимы без приведения
type Celsius2 = float64     // псевдоним (alias) — тот же самый тип

type UserID int64           // защита от путаницы: UserID и OrderID нельзя случайно перепутать
```

Зачем: безопасность типов (`UserID` ≠ `OrderID`) и **методы** — на пользовательский тип можно повесить функции. Краткий предпросмотр (подробно — в уроке про структуры):

```go run
package main

import "fmt"

type Celsius float64

// String — метод. Если тип имеет String() string, fmt использует его при печати (интерфейс fmt.Stringer).
func (c Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

func main() {
	t := Celsius(36.6)
	fmt.Println(t)
	fmt.Printf("%v | %s | %.2f\n", t, t, float64(t))
}
```

## Указатели — короткое введение

```go run
package main

import "fmt"

func inc(p *int) { *p++ } // *p — разыменование: читаем/пишем значение по адресу

func main() {
	x := 10
	p := &x // & — взять адрес
	inc(p)
	inc(&x)
	fmt.Println(x, *p, p == &x)

	var np *int // nil-указатель; разыменование nil → panic
	fmt.Println(np == nil)
}
```

В Go **всё передаётся по значению** (копируется). Чтобы функция изменила переменную, передают указатель. Арифметики указателей нет.

## Проверь себя

<details><summary>Чему равна нулевая строка / указатель / слайс?</summary><code>""</code> / <code>nil</code> / <code>nil</code> (nil-слайс безопасно использовать с <code>len</code>, <code>append</code>, <code>range</code>).</details>

<details><summary>Почему <code>var f float64 = i</code> не компилируется, если <code>i int</code>?</summary>В Go нет неявных преобразований между разными типами, нужно <code>float64(i)</code>.</details>

<details><summary>Что делает <code>iota</code>?</summary>Автоинкремент внутри <code>const</code>-блока: с 0, +1 на каждую строку; используют для перечислений и степеней двойки.</details>

<details><summary>В чём разница <code>type A int</code> и <code>type A = int</code>?</summary>Первое — новый тип (можно навесить методы, нужны приведения), второе — псевдоним того же типа.</details>

<details><summary>Что вернёт <code>string(65)</code> и как правильно получить "65"?</summary>"A" (код символа); правильно <code>strconv.Itoa(65)</code>.</details>
