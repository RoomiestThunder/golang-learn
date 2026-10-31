---
title: wc: считаем поток
level: medium
---
@@ statement
Реализуй `Count(r io.Reader) (Stats, error)` — аналог утилиты `wc`:

```go
type Stats struct{ Lines, Words, Bytes int }
```

- `Bytes` — общее число байт;
- `Words` — число «слов»: последовательностей байт без ASCII-пробельных символов (`' ' \t \n \r \v \f`);
- `Lines` — число строк: количество `\n`, **плюс одна**, если данные не пусты и последний символ — не `\n` (то есть последняя строка без перевода строки тоже считается).

Требования: читай **потоково** фиксированным буфером (в тесте строка длиной 200 КБ и источник, отдающий по 1 байту за раз — `bufio.Scanner` с настройками по умолчанию тут не подходит). Ошибку чтения (кроме `io.EOF`) верни вместе с тем, что уже посчитано.
@@ starter
package solution

import "io"

type Stats struct{ Lines, Words, Bytes int }

func Count(r io.Reader) (Stats, error) {
	return Stats{}, nil
}
@@ tests
package solution

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Stats
	}{
		{"empty", "", Stats{0, 0, 0}},
		{"simple", "hello world\nfoo\n", Stats{2, 3, 16}},
		{"no trailing newline", "one two\nthree", Stats{2, 3, 13}},
		{"only newlines", "\n\n", Stats{2, 0, 2}},
		{"tabs and cr", "a\tb\r\nc  d", Stats{2, 4, 9}},
		{"unicode bytes", "привет мир\n", Stats{1, 2, 20}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Count(strings.NewReader(tt.in))
			if err != nil || got != tt.want {
				t.Errorf("Count(%q) = %+v, %v; want %+v", tt.in, got, err, tt.want)
			}
		})
	}
}

func TestCountPartialReads(t *testing.T) {
	long := strings.Repeat("ab ", 70_000) // строка ~210 КБ
	got, err := Count(iotest.OneByteReader(strings.NewReader(long + "\nlast")))
	want := Stats{Lines: 2, Words: 70_001, Bytes: len(long) + 5}
	if err != nil || got != want {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}
}

func TestCountError(t *testing.T) {
	boom := errors.New("boom")
	_, err := Count(iotest.ErrReader(boom))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
}
@@ solution
package solution

import "io"

type Stats struct{ Lines, Words, Bytes int }

func Count(r io.Reader) (Stats, error) {
	var st Stats
	buf := make([]byte, 32*1024)
	inWord := false
	last := byte('\n')
	for {
		n, err := r.Read(buf)
		for _, c := range buf[:n] {
			st.Bytes++
			if c == '\n' {
				st.Lines++
			}
			if c == ' ' || (c >= '\t' && c <= '\r') {
				inWord = false
			} else if !inWord {
				inWord = true
				st.Words++
			}
			last = c
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return st, err
		}
	}
	if st.Bytes > 0 && last != '\n' {
		st.Lines++
	}
	return st, nil
}
@@ hint
Читай в цикле `n, err := r.Read(buf)`; сначала обработай `buf[:n]`, потом смотри на `err`. Запоминай последний байт, чтобы решить, нужна ли «добавочная» строка.
