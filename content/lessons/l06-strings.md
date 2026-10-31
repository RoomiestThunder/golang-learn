---
id: l06
week: 2
day: 1
minutes: 120
title: Строки, байты и руны
goals: Понять, что строка — это неизменяемые байты в UTF-8; работать с рунами, пакетами strings, strconv, strings.Builder
lc: valid-anagram | Valid Anagram | Easy | Счётчики символов: [26]int или map[rune]int
lc: reverse-string | Reverse String | Easy | Два указателя на []byte — переворот на месте
lc: valid-palindrome | Valid Palindrome | Easy | Пропускай не-буквы двумя указателями; unicode.IsLetter/IsDigit
lc: longest-common-prefix | Longest Common Prefix | Easy | Вертикальное сканирование по столбцам
lc: is-subsequence | Is Subsequence | Easy | Два указателя: идём по t, двигаем i в s при совпадении
---

## Что такое string в Go

`string` — **неизменяемая последовательность байт** (по соглашению — UTF-8). Внутри это пара `(указатель, длина)`. Отсюда всё:

- `len(s)` — длина в **байтах**, а не в символах;
- `s[i]` — i-й **байт** (`byte`), а не символ;
- менять строку на месте нельзя — `s[0] = 'x'` не компилируется; подстрока `s[a:b]` делит память с исходной (без копирования);
- конкатенация создаёт новую строку.

```go run
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "Привет, 🙂"
	fmt.Println(len(s), utf8.RuneCountInString(s)) // байты и символы отличаются!

	fmt.Println(s[0], s[:2])  // байт 208 и первые 2 байта = «П» (кириллица занимает 2 байта)
	fmt.Printf("%x\n", s[:4]) // hex первых 4 байт

	for i, r := range s { // range по строке декодирует UTF-8: i — смещение в байтах, r — руна
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	runes := []rune(s)   // копия: по одной руне (int32) на символ
	bytes := []byte(s)   // копия: сырые байты
	runes[0] = 'п'
	fmt.Println(string(runes), len(runes), len(bytes))
}
```

**Руна** (`rune` = `int32`) — кодовая точка Unicode. Литерал `'я'` имеет тип rune. Один «видимый символ» (графема) может состоять из нескольких рун (эмодзи с модификаторами, буква+диакритика) — для таких случаев нужны внешние библиотеки.

> [!WARNING] Для переворота строки и подсчёта символов приводи к `[]rune`, иначе поломаешь многобайтовые символы. Для ASCII-данных `[]byte` быстрее.

## Литералы

```go
a := "строка с \n и \"кавычками\""   // интерпретируемая: работают escape-последовательности
b := `сырая строка: \n не интерпретируется,
можно писать на нескольких строках, удобно для SQL/JSON/regexp`
```

## Пакет strings

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "  Go is fun, Go is fast  "
	fmt.Println(strings.TrimSpace(s))
	fmt.Println(strings.Contains(s, "fun"), strings.HasPrefix(s, "  Go"), strings.HasSuffix(s, "fast"))
	fmt.Println(strings.Index(s, "Go"), strings.LastIndex(s, "Go"), strings.Count(s, "Go"))
	fmt.Println(strings.ToUpper("привет"), strings.ToLower("ПРИВЕТ"))
	fmt.Println(strings.Replace("aaaa", "a", "b", 2), strings.ReplaceAll("aaaa", "a", "b"))
	fmt.Printf("%q\n", strings.Split("a,b,,c", ","))   // ["a" "b" "" "c"]
	fmt.Printf("%q\n", strings.Fields(" a  b\tc\n"))   // делит по любым пробельным, без пустых
	fmt.Println(strings.Join([]string{"x", "y", "z"}, "-"))
	fmt.Println(strings.Repeat("ab", 3), strings.EqualFold("Go", "GO"))
	fmt.Println(strings.Trim("!!hello!!", "!"), strings.TrimLeft("xxabc", "x"), strings.TrimPrefix("prefix-body", "prefix-"))

	before, after, found := strings.Cut("key=value=x", "=") // Cut — удобнее Split для «ключ=значение»
	fmt.Println(before, after, found)

	fmt.Println(strings.Map(func(r rune) rune {
		if r == 'a' {
			return 'A'
		}
		return r
	}, "banana"))
}
```

## Эффективная сборка строк

`s += x` в цикле каждый раз выделяет новую строку: O(n²). Используй `strings.Builder` (внутри растущий `[]byte`, без лишнего копирования при `String()`):

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder // нулевое значение готово к работе
	sb.Grow(64)            // опционально: заранее выделить ёмкость
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&sb, "%d,", i) // Builder реализует io.Writer
	}
	sb.WriteString("конец")
	sb.WriteByte('!')
	sb.WriteRune('🙂')
	fmt.Println(sb.String(), sb.Len())
}
```

## strconv: числа ↔ строки

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("123")           // string → int
	f, _ := strconv.ParseFloat("3.14", 64)  // string → float64
	b, _ := strconv.ParseBool("true")
	i64, _ := strconv.ParseInt("ff", 16, 64) // основание 16, 64 бита
	_, err2 := strconv.Atoi("12a")
	fmt.Println(n, err, f, b, i64)
	fmt.Println(err2) // *strconv.NumError: strconv.Atoi: parsing "12a": invalid syntax

	fmt.Println(strconv.Itoa(99), strconv.FormatInt(255, 2), strconv.Quote("hi\n"), strconv.FormatFloat(1.5, 'f', 2, 64))
}
```

## Сравнение и unicode

- Строки сравниваются побайтово лексикографически: `"a" < "b"`, `==` работает.
- Пакет `unicode`: `IsLetter`, `IsDigit`, `IsSpace`, `IsUpper`, `ToUpper` для рун. `unicode/utf8`: `RuneCountInString`, `DecodeRuneInString`, `ValidString`.
- Регистронезависимое сравнение — `strings.EqualFold`, а не `ToLower(a) == ToLower(b)`.

## Подводные камни

1. `len("🙂") == 4`, а не 1.
2. `string(i)` для int даёт символ с кодом i; число в строку — `strconv.Itoa`.
3. Подстрока удерживает большую исходную строку в памяти — для долгоживущих кусков копируй: `strings.Clone`.
4. Для побайтовой обработки больших данных используй `bytes.Buffer`/`bufio`, а не строки.

## Проверь себя

<details><summary>Почему <code>len("привет")</code> равно 12?</summary>Кириллические буквы в UTF-8 занимают по 2 байта; <code>len</code> считает байты.</details>

<details><summary>Как перевернуть строку с юникодом?</summary>Преобразовать в <code>[]rune</code>, поменять порядок, вернуть в <code>string</code>.</details>

<details><summary>Что быстрее для сборки большой строки в цикле и почему?</summary><code>strings.Builder</code> — амортизированный append в байтовый буфер, без квадратичного копирования как у <code>+=</code>.</details>

<details><summary>Что возвращает range по строке?</summary>Пары (байтовое смещение начала руны, руна); некорректные байты дают U+FFFD.</details>
