---
title: Руны: переворот и палиндром
level: easy
---
@@ statement
1. `Reverse(s string) string` — переверни строку **по символам** (рунам), чтобы кириллица и эмодзи не ломались.
2. `IsPalindrome(s string) bool` — палиндром, если учитывать только буквы и цифры (`unicode.IsLetter`/`IsDigit`) и не различать регистр. Пустая строка — палиндром.
3. `CountRunes(s string) (letters, digits, others int)` — посчитай буквы, цифры и всё остальное (пробелы, знаки, эмодзи).
@@ starter
package solution

func Reverse(s string) string {
	return s
}

func IsPalindrome(s string) bool {
	return false
}

func CountRunes(s string) (letters, digits, others int) {
	return
}
@@ tests
package solution

import "testing"

func TestReverse(t *testing.T) {
	tests := map[string]string{"": "", "abc": "cba", "Привет": "тевирП", "a🙂b": "b🙂a", "héllo": "olléh"}
	for in, want := range tests {
		if got := Reverse(in); got != want {
			t.Errorf("Reverse(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true}, {"A man, a plan, a canal: Panama", true}, {"race a car", false},
		{"Аргентина манит негра", true}, {"12321", true}, {"12 3", false}, {".,", true},
	}
	for _, tt := range tests {
		if got := IsPalindrome(tt.in); got != tt.want {
			t.Errorf("IsPalindrome(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestCountRunes(t *testing.T) {
	l, d, o := CountRunes("Привет, мир 2024! 🙂")
	if l != 9 || d != 4 || o != 6 {
		t.Errorf("got %d %d %d, want 9 4 6", l, d, o)
	}
}
@@ solution
package solution

import (
	"unicode"
)

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func IsPalindrome(s string) bool {
	var clean []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			clean = append(clean, unicode.ToLower(r))
		}
	}
	for i, j := 0, len(clean)-1; i < j; i, j = i+1, j-1 {
		if clean[i] != clean[j] {
			return false
		}
	}
	return true
}

func CountRunes(s string) (letters, digits, others int) {
	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			letters++
		case unicode.IsDigit(r):
			digits++
		default:
			others++
		}
	}
	return
}
@@ hint
Перевод в `[]rune`, два индекса навстречу друг другу, обмен через `r[i], r[j] = r[j], r[i]`.
