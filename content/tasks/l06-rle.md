---
title: Сжатие строки (RLE)
level: medium
---
@@ statement
Реализуй простое сжатие повторов (run-length encoding) для строк из **букв** (любых, включая кириллицу):

- `Compress(s string) string`: подряд идущие одинаковые символы заменяются на символ и количество, если оно больше 1. `"aaabcc"` → `"a3bc2"`, `"abc"` → `"abc"`, `"ууууу"` → `"у5"`, `""` → `""`.
- `Decompress(s string) (string, error)`: обратная операция. Число после символа может быть многозначным (`"x12"` → 12 раз `x`). Ошибка, если строка начинается с цифры или записанное число повторов меньше 2 (например `a0`, `a1`).

Используй `strings.Builder` и `strconv`. Работай с рунами, а не с байтами.
@@ starter
package solution

func Compress(s string) string {
	return s
}

func Decompress(s string) (string, error) {
	return s, nil
}
@@ tests
package solution

import "testing"

func TestCompress(t *testing.T) {
	tests := map[string]string{
		"": "", "a": "a", "abc": "abc", "aaabcc": "a3bc2", "ууууу": "у5", "aabbbbbbbbbbbb": "a2b12", "ababab": "ababab",
	}
	for in, want := range tests {
		if got := Compress(in); got != want {
			t.Errorf("Compress(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDecompress(t *testing.T) {
	tests := map[string]string{"": "", "a3bc2": "aaabcc", "у5": "ууууу", "x12": "xxxxxxxxxxxx", "abc": "abc"}
	for in, want := range tests {
		got, err := Decompress(in)
		if err != nil || got != want {
			t.Errorf("Decompress(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"3a", "a0", "12"} {
		if _, err := Decompress(bad); err == nil {
			t.Errorf("Decompress(%q) должна вернуть ошибку", bad)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, s := range []string{"aaaabbbcca", "яяяяяяяяяяяя", "zzzzzzzzzzzzzzzzzzzzzzzzzq"} {
		got, err := Decompress(Compress(s))
		if err != nil || got != s {
			t.Errorf("round trip %q → %q, %v", s, got, err)
		}
	}
}
@@ solution
package solution

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

func Compress(s string) string {
	var sb strings.Builder
	rs := []rune(s)
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && rs[j] == rs[i] {
			j++
		}
		sb.WriteRune(rs[i])
		if n := j - i; n > 1 {
			sb.WriteString(strconv.Itoa(n))
		}
		i = j
	}
	return sb.String()
}

func Decompress(s string) (string, error) {
	var sb strings.Builder
	rs := []rune(s)
	for i := 0; i < len(rs); {
		r := rs[i]
		if unicode.IsDigit(r) {
			return "", errors.New("число без символа")
		}
		i++
		j := i
		for j < len(rs) && unicode.IsDigit(rs[j]) {
			j++
		}
		count := 1
		if j > i {
			n, err := strconv.Atoi(string(rs[i:j]))
			if err != nil || n < 2 {
				return "", errors.New("некорректное число повторов")
			}
			count = n
		}
		sb.WriteString(strings.Repeat(string(r), count))
		i = j
	}
	return sb.String(), nil
}
@@ hint
Для каждой группы: найди конец серии одинаковых рун (`j`), запиши руну и, если длина больше 1, число.
