---
id: l16
week: 4
day: 1
minutes: 120
title: io, bufio и файлы
goals: Понять Reader/Writer как главную абстракцию, потоковую обработку, буферизацию, работу с файлами
lc: ransom-note | Ransom Note | Easy | Счётчик символов журнала: [26]int
lc: isomorphic-strings | Isomorphic Strings | Easy | Два map: прямое и обратное отображение символов
lc: word-pattern | Word Pattern | Easy | Та же идея, что в isomorphic strings, но слова ↔ буквы
---

## Reader и Writer — сердце стандартной библиотеки

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```

Файл, сетевое соединение, HTTP-тело, gzip-поток, буфер в памяти, `os.Stdin/Stdout` — всё это `Reader`/`Writer`. Код, принимающий интерфейсы, работает с любым источником и **тестируется без файлов** (`strings.NewReader`, `bytes.Buffer`).

Контракт `Read`: читает до `len(p)` байт, возвращает число прочитанных и ошибку; `io.EOF` — штатный конец данных. **Обрабатывай `n > 0` до проверки ошибки** — данные и EOF могут прийти вместе.

```go run
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	r := strings.NewReader("Привет, потоки данных!\nвторая строка\n")

	// io.Copy: поток → поток, без загрузки всего в память
	n, _ := io.Copy(os.Stdout, io.LimitReader(r, 20)) // LimitReader ограничивает чтение 20 байтами
	fmt.Println("\nскопировано байт:", n)

	// io.ReadAll — всё сразу (осторожно с большими/недоверенными источниками)
	rest, _ := io.ReadAll(r)
	fmt.Printf("осталось: %q\n", rest)

	// TeeReader: читаем и одновременно пишем копию (логирование, хэширование)
	var copyBuf bytes.Buffer
	tee := io.TeeReader(strings.NewReader("tee data"), &copyBuf)
	io.ReadAll(tee)
	fmt.Println(copyBuf.String())

	// MultiWriter: запись в несколько мест сразу
	var a, b bytes.Buffer
	mw := io.MultiWriter(&a, &b, os.Stdout)
	fmt.Fprintln(mw, "в три места")
	fmt.Println(a.Len(), b.Len())

	// MultiReader склеивает потоки; Pipe — синхронный канал между Reader и Writer
	mr := io.MultiReader(strings.NewReader("раз "), strings.NewReader("два"))
	all, _ := io.ReadAll(mr)
	fmt.Println(string(all))
}
```

### Оборачивание: паттерн «декоратор»

Writer/Reader, принимающие другой Writer/Reader (`gzip.NewWriter(w)`, `bufio.NewReader(r)`, `io.LimitReader`), образуют **конвейеры**. Свой декоратор — структура, встраивающая интерфейс (см. урок 8). Именно так устроены `crypto/tls`, `compress/*`, `net/http`.

## bufio: буферизация и построчное чтение

Системные вызовы дороги: чтение по байту из файла — тормоз. `bufio` накапливает данные в буфере.

```go run
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	text := "first line\nsecond line\nthird"
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() { // по умолчанию — построчно; sc.Split(bufio.ScanWords) — по словам
		fmt.Printf("%q\n", sc.Text())
	}
	if err := sc.Err(); err != nil { // ошибку Scan нужно проверять отдельно!
		fmt.Println(err)
	}

	// Scanner ограничивает длину строки 64 КБ; для длинных: sc.Buffer(make([]byte, 1<<20), 1<<20)

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush() // ОБЯЗАТЕЛЬНО сбросить буфер, иначе данные потеряются
	for i := 0; i < 3; i++ {
		fmt.Fprintf(w, "строка %d\n", i)
	}
}
```

`bufio.Reader` даёт больше контроля: `ReadString('\n')`, `ReadLine`, `Peek`, `ReadRune`. Для быстрого чтения большого ввода в алгоритмических задачах: `bufio.NewReader(os.Stdin)` + `fmt.Fscan`.

## Работа с файлами

```go run
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	dir, _ := os.MkdirTemp("", "demo") // временная папка
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "notes.txt") // Join подставит правильный разделитель ОС

	// быстрые функции для небольших файлов целиком
	if err := os.WriteFile(path, []byte("hello\nworld\n"), 0o644); err != nil { // 0644: rw-r--r--
		panic(err)
	}
	data, _ := os.ReadFile(path)
	fmt.Printf("%q\n", data)

	// потоковая работа: открыть → defer Close → читать/писать
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644) // флаги: добавление в конец
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(f, "appended")
	if err := f.Close(); err != nil { // для записи ошибку Close проверяй!
		panic(err)
	}

	f, _ = os.Open(path) // только чтение
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
	}
	fmt.Println("строк:", n)

	// ошибки файловых операций — через errors.Is с sentinel из io/fs
	_, err = os.Open(filepath.Join(dir, "missing"))
	fmt.Println(errors.Is(err, fs.ErrNotExist), err != nil)

	entries, _ := os.ReadDir(dir) // отсортирован по имени
	for _, e := range entries {
		fmt.Println(e.Name(), e.IsDir())
	}
}
```

Полезно знать:

- `os.Args`, `os.Getenv/LookupEnv/Setenv`, `os.Exit(code)` (defer при этом **не выполняются**), `os.Stdin/Stdout/Stderr`.
- `filepath.Walk/WalkDir`, `filepath.Ext/Base/Dir/Glob`; для путей внутри `embed`/`fs.FS` — пакет `path`.
- `io/fs.FS` — абстракция файловой системы (`os.DirFS`, `embed.FS`, `fstest.MapFS` для тестов).
- `//go:embed templates/*` — вкладывает файлы прямо в бинарь (как эта платформа вшивает уроки).
- Для записи с гарантией целостности: пиши во временный файл в той же папке и `os.Rename`.

## Что запомнить

1. Принимай `io.Reader`/`io.Writer`, а не `*os.File` — код станет универсальным и тестируемым.
2. `defer f.Close()` сразу после успешного открытия; для записи проверяй ошибку Close.
3. Не читай большие данные через `ReadAll`/`ReadFile`, если можно потоково.
4. `Scanner` — для строк/слов, `bufio.Reader` — для бинарных/длинных данных.
5. Всегда `Flush()` у `bufio.Writer`.

## Проверь себя

<details><summary>Почему <code>io.EOF</code> — не ошибка в обычном смысле?</summary>Это сигнал нормального завершения потока; <code>Read</code> может вернуть <code>n &gt; 0</code> вместе с EOF, поэтому данные нужно обработать до проверки ошибки.</details>

<details><summary>Чем <code>bufio.Scanner</code> отличается от <code>bufio.Reader</code>?</summary>Scanner удобен для токенов (строки, слова) с ограничением размера токена; Reader даёт низкоуровневый контроль и не ограничивает длину строки.</details>

<details><summary>Что случится, если не вызвать <code>Flush</code> у <code>bufio.Writer</code>?</summary>Данные, оставшиеся в буфере, не будут записаны.</details>

<details><summary>Выполнится ли <code>defer</code> при <code>os.Exit(1)</code>?</summary>Нет, процесс завершается немедленно.</details>
