---
title: Приветствие
level: easy
---
@@ statement
Напиши функцию `Greet(name string) string`.

- Для непустого имени вернуть `"Привет, <name>!"`.
- Для пустой строки (или строки из одних пробелов) — `"Привет, незнакомец!"`.

Подсказка по пакетам: `strings.TrimSpace`, `fmt.Sprintf`.
@@ starter
package solution

// Greet возвращает приветствие.
func Greet(name string) string {
	return ""
}
@@ tests
package solution

import "testing"

func TestGreet(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Гоша", "Привет, Гоша!"},
		{"  Анна ", "Привет, Анна!"},
		{"", "Привет, незнакомец!"},
		{"   ", "Привет, незнакомец!"},
	}
	for _, tt := range tests {
		if got := Greet(tt.in); got != tt.want {
			t.Errorf("Greet(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
@@ solution
package solution

import (
	"fmt"
	"strings"
)

func Greet(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Привет, незнакомец!"
	}
	return fmt.Sprintf("Привет, %s!", name)
}
@@ hint
Сначала обрежь пробелы по краям, потом проверь на пустоту.
