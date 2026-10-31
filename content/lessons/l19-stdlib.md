---
id: l19
week: 4
day: 4
minutes: 120
title: Полезная стандартная библиотека
goals: slog, flag, сортировка, container/heap, regexp, hash/crypto/encoding, rand — инструменты на каждый день
lc: sort-colors | Sort Colors | Medium | Алгоритм голландского флага: три указателя за один проход
lc: largest-number | Largest Number | Medium | Своя сортировка: a раньше b, если a+b > b+a как строки
lc: k-closest-points-to-origin | K Closest Points to Origin | Medium | Сортировка по квадрату расстояния или куча на k элементов
lc: kth-largest-element-in-an-array | Kth Largest Element in an Array | Medium | Мин-куча размера k или Quickselect
---

## log/slog: структурное логирование

С Go 1.21 в стандартной библиотеке: уровни, атрибуты «ключ-значение», JSON/текстовые обработчики. Логи в production должны быть **машиночитаемыми**.

```go run
package main

import (
	"log/slog"
	"os"
)

func main() {
	// убираем время, чтобы вывод был стабильным в примере
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return a
		},
	}
	text := slog.New(slog.NewTextHandler(os.Stdout, opts))
	json := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	text.Info("пользователь вошёл", "user_id", 42, "ip", "10.0.0.1")
	text.Debug("детали", slog.Int("attempt", 2), slog.Group("req", slog.String("method", "GET"), slog.Int("status", 200)))
	text.Warn("медленный запрос", "ms", 1500)

	// With — логгер с постоянными атрибутами (например, request_id на весь запрос)
	reqLog := json.With("request_id", "abc-123")
	reqLog.Error("не удалось сохранить", "err", os.ErrPermission)

	slog.SetDefault(text) // пакетные slog.Info(...) пойдут через него
	slog.Info("через default")
}
```

Практика: логгер передавай как зависимость (поле структуры) или кладёт в контекст; ключи — snake_case; ошибки как атрибут `"err"`; **не логируй секреты**; для высоконагруженных путей — `LogAttrs` и ленивые значения (`slog.LogValuer`).

## flag: аргументы командной строки

```go
func main() {
	addr := flag.String("addr", ":8080", "адрес сервера")
	debug := flag.Bool("debug", false, "режим отладки")
	timeout := flag.Duration("timeout", 5*time.Second, "таймаут")
	flag.Parse()
	fmt.Println(*addr, *debug, *timeout, flag.Args()) // flag.Args() — остальные позиционные
}
// ./app -addr :9000 -debug -timeout 10s file.txt
```

Для сложных CLI используют `cobra`/`urfave/cli`; конфигурацию 12-factor приложений обычно берут из **переменных окружения** (`os.LookupEnv`) с понятными значениями по умолчанию.

## Сортировка

```go run
package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

type P struct {
	Name string
	Age  int
}

func main() {
	ints := []int{5, 2, 8}
	sort.Ints(ints)
	strs := []string{"b", "a"}
	sort.Strings(strs)
	fmt.Println(ints, strs, sort.SearchInts(ints, 5)) // бинарный поиск: индекс первого >= 5

	ps := []P{{"B", 30}, {"A", 30}, {"C", 25}}
	// sort.Slice — НЕстабильная; SliceStable сохраняет порядок равных
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].Age < ps[j].Age })
	fmt.Println(ps)

	// современный вариант — slices.SortFunc (быстрее, без рефлексии), компаратор возвращает int
	slices.SortFunc(ps, func(a, b P) int {
		if c := strings.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return a.Age - b.Age
	})
	fmt.Println(ps)

	i := sort.Search(100, func(i int) bool { return i*i >= 50 }) // общий бинарный поиск по монотонному предикату
	fmt.Println(i)
}
```

`sort.Search`/`slices.BinarySearch` — основа решений задач на бинарный поиск (неделя 7).

## container/heap, list, ring

`heap` — приоритетная очередь: реализуй `heap.Interface` (`sort.Interface` + `Push`/`Pop`).

```go run
package main

import (
	"container/heap"
	"container/list"
	"fmt"
)

type MinHeap []int

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func main() {
	h := &MinHeap{5, 2, 8}
	heap.Init(h)
	heap.Push(h, 1)
	for h.Len() > 0 {
		fmt.Print(heap.Pop(h), " ") // 1 2 5 8: всегда минимальный; Push/Pop — O(log n)
	}
	fmt.Println()

	l := list.New() // двусвязный список — основа LRU-кэша
	e := l.PushBack(1)
	l.PushFront(0)
	l.PushBack(2)
	l.MoveToFront(e)
	for x := l.Front(); x != nil; x = x.Next() {
		fmt.Print(x.Value, " ")
	}
	fmt.Println()
}
```

## regexp

Синтаксис RE2 (без обратных ссылок и lookahead), работает за линейное время — безопасен против ReDoS.

```go run
package main

import (
	"fmt"
	"regexp"
)

var (
	// компилируй ОДИН раз на уровне пакета; MustCompile паникует при ошибке в шаблоне
	emailRe = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)
	logRe   = regexp.MustCompile(`^(?P<level>[A-Z]+): (?P<msg>.*)$`)
)

func main() {
	text := "пиши на ann@example.com или bob.smith+go@mail.org"
	fmt.Println(emailRe.FindAllString(text, -1))
	fmt.Println(emailRe.MatchString("нет почты"))

	m := logRe.FindStringSubmatch("ERROR: диск заполнен")
	for i, name := range logRe.SubexpNames() {
		if name != "" {
			fmt.Println(name, "=", m[i])
		}
	}
	fmt.Println(regexp.MustCompile(`\d+`).ReplaceAllString("a1b22c333", "#"))
}
```

## hash, crypto, encoding, rand

```go run
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

func main() {
	sum := sha256.Sum256([]byte("hello"))
	fmt.Println(hex.EncodeToString(sum[:]))

	mac := hmac.New(sha256.New, []byte("secret-key"))
	mac.Write([]byte("message"))
	sig := mac.Sum(nil)
	// сравнивать подписи ТОЛЬКО константным временем
	fmt.Println(hmac.Equal(sig, sig), base64.URLEncoding.EncodeToString(sig)[:10])

	r := csv.NewReader(strings.NewReader("name,age\nann,30\n\"bob, jr\",25\n"))
	rows, _ := r.ReadAll()
	fmt.Println(rows)
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"x", "a,b", `q"q`})
	w.Flush()

	// math/rand/v2: не для безопасности! Для токенов и ключей — crypto/rand
	fmt.Println(rand.IntN(10) < 10, len(rand.Perm(5)))
}
```

**Безопасность:** `math/rand` предсказуем — для паролей, токенов, ID сессий используй `crypto/rand` (`rand.Read`, в Go 1.24 — `rand.Text()`). Хэширование паролей — не `sha256`, а bcrypt/scrypt/argon2 (`golang.org/x/crypto`).

## Другие полезные пакеты (знать, что существуют)

| Пакет | Для чего |
|---|---|
| `text/template`, `html/template` | Шаблоны (html — с автоэкранированием!) |
| `net/url`, `net/mail`, `net` | Парсинг URL/почты, TCP/UDP |
| `os/exec` | Запуск внешних команд (`exec.CommandContext`) |
| `os/signal` | Обработка SIGINT/SIGTERM (graceful shutdown) |
| `bytes` | Зеркало `strings` для `[]byte`, `bytes.Buffer` |
| `unicode/utf8`, `unicode` | Руны |
| `math`, `math/big`, `math/bits` | Математика, большие числа, битовые операции |
| `embed` | Вшивание файлов в бинарь |
| `reflect`, `unsafe` | Рефлексия и небезопасные операции — только когда без них нельзя |
| `runtime`, `runtime/pprof`, `runtime/debug` | Внутренности, профилирование |
| `compress/gzip`, `archive/zip`, `archive/tar` | Сжатие и архивы |
| `html`, `mime`, `net/http/httputil` | Утилиты веб |

## Проверь себя

<details><summary>Чем <code>sort.Slice</code> отличается от <code>sort.SliceStable</code>?</summary>Stable сохраняет относительный порядок равных элементов; Slice — нет (но быстрее).</details>

<details><summary>Как получить минимальный элемент из <code>container/heap</code>?</summary>После <code>heap.Init/Push</code> — <code>heap.Pop(h)</code> (корень — минимум по <code>Less</code>); сам слайс <code>h[0]</code> — тоже минимум.</details>

<details><summary>Почему regexp в Go безопасен от ReDoS?</summary>Движок RE2 гарантирует линейное время относительно длины входа, поэтому в нём нет backreferences и lookaround.</details>

<details><summary>Чем <code>crypto/rand</code> отличается от <code>math/rand</code>?</summary>crypto/rand — криптостойкий источник случайности (токены, ключи), math/rand — быстрый псевдослучайный генератор для симуляций.</details>
