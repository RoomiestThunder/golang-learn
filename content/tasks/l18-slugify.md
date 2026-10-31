---
title: Slugify: пиши по таблице тестов
level: easy
---
@@ statement
Реализуй `Slugify(s string) string` — превращает заголовок в ЧПУ-фрагмент URL.

Правила (они закодированы в таблице тестов):
1. Приводим к нижнему регистру.
2. Буквы (латиница **и кириллица**) и цифры остаются; любая последовательность остальных символов (пробелы, знаки, эмодзи) → один дефис `-`.
3. Дефисы в начале и конце убираются.
4. Пустой результат допустим (`""`).

Это упражнение на привычку «сначала таблица тестов, потом код»: открой «🧪 Тесты», добавь свои случаи в голове — и реализуй.
@@ starter
package solution

func Slugify(s string) string {
	return s
}
@@ tests
package solution

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"empty", "", ""},
		{"simple", "Hello World", "hello-world"},
		{"punctuation", "Go: Tips & Tricks!", "go-tips-tricks"},
		{"multiple spaces", "  many   spaces  ", "many-spaces"},
		{"digits", "Top 10 Go libs", "top-10-go-libs"},
		{"cyrillic", "Привет, Мир!", "привет-мир"},
		{"emoji", "Go 🙂 rocks", "go-rocks"},
		{"only symbols", "!!!", ""},
		{"dashes inside", "a--b - c", "a-b-c"},
		{"already slug", "already-a-slug", "already-a-slug"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.in); got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
@@ solution
package solution

import (
	"strings"
	"unicode"
)

func Slugify(s string) string {
	var sb strings.Builder
	pendingDash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingDash && sb.Len() > 0 {
				sb.WriteByte('-')
			}
			pendingDash = false
			sb.WriteRune(r)
		} else {
			pendingDash = true
		}
	}
	return sb.String()
}
@@ hint
Идём по рунам; для «разделителей» только запоминаем `pendingDash = true`, а дефис пишем перед следующим символом (и только если результат уже не пуст) — тогда лишних дефисов по краям не будет.
