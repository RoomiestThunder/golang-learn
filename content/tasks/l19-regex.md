---
title: regexp: почты и логи
level: medium
---
@@ statement
1. `ExtractEmails(text string) []string` — все адреса электронной почты в порядке появления, **без повторов** (регистр сравнивается как есть). Формат: `локальная_часть@домен.зона`, где локальная часть — буквы, цифры и `._+-`; домен — буквы, цифры, дефис и точки, минимум одна точка. Если адресов нет — пустой результат.
2. `ParseLogLine(line string) (LogEntry, error)` — разбирает строки вида

   `2025-03-09T14:30:05Z [ERROR] db: connection lost`

   в `LogEntry{Time time.Time, Level, Component, Message string}`. Формат: время RFC3339, уровень в квадратных скобках (заглавные буквы), компонент (буквы/цифры/`_`/`-`) и двоеточие, затем сообщение до конца строки. Для строки другого формата — ошибка. Используй именованные группы и **скомпилируй регулярку один раз** на уровне пакета.
@@ starter
package solution

import "time"

type LogEntry struct {
	Time      time.Time
	Level     string
	Component string
	Message   string
}

func ExtractEmails(text string) []string {
	return nil
}

func ParseLogLine(line string) (LogEntry, error) {
	return LogEntry{}, nil
}
@@ tests
package solution

import (
	"slices"
	"testing"
	"time"
)

func TestExtractEmails(t *testing.T) {
	text := "Пишите ann@example.com или bob.smith+go@mail.co.uk; повтор: ann@example.com. Не почта: @nobody, a@b, plain.text"
	got := ExtractEmails(text)
	want := []string{"ann@example.com", "bob.smith+go@mail.co.uk"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := ExtractEmails("нет адресов"); len(got) != 0 {
		t.Errorf("got %q", got)
	}
}

func TestParseLogLine(t *testing.T) {
	e, err := ParseLogLine("2025-03-09T14:30:05Z [ERROR] db: connection lost: retry in 5s")
	if err != nil {
		t.Fatal(err)
	}
	want := LogEntry{
		Time:      time.Date(2025, 3, 9, 14, 30, 5, 0, time.UTC),
		Level:     "ERROR",
		Component: "db",
		Message:   "connection lost: retry in 5s",
	}
	if !e.Time.Equal(want.Time) || e.Level != want.Level || e.Component != want.Component || e.Message != want.Message {
		t.Errorf("got %+v, want %+v", e, want)
	}
}

func TestParseLogLineBad(t *testing.T) {
	for _, line := range []string{
		"", "garbage", "2025-03-09 14:30:05 [ERROR] db: x", "2025-03-09T14:30:05Z ERROR db: x",
		"2025-03-09T14:30:05Z [error] db: x", "2025-13-45T14:30:05Z [INFO] db: x",
	} {
		if _, err := ParseLogLine(line); err == nil {
			t.Errorf("ожидалась ошибка для %q", line)
		}
	}
}
@@ solution
package solution

import (
	"fmt"
	"regexp"
	"time"
)

type LogEntry struct {
	Time      time.Time
	Level     string
	Component string
	Message   string
}

var (
	emailRe = regexp.MustCompile(`[A-Za-z0-9._+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+`)
	logRe   = regexp.MustCompile(`^(?P<time>\S+) \[(?P<level>[A-Z]+)\] (?P<comp>[\w-]+): (?P<msg>.*)$`)
)

func ExtractEmails(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range emailRe.FindAllString(text, -1) {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

func ParseLogLine(line string) (LogEntry, error) {
	m := logRe.FindStringSubmatch(line)
	if m == nil {
		return LogEntry{}, fmt.Errorf("строка не соответствует формату: %q", line)
	}
	t, err := time.Parse(time.RFC3339, m[logRe.SubexpIndex("time")])
	if err != nil {
		return LogEntry{}, fmt.Errorf("плохое время: %w", err)
	}
	return LogEntry{
		Time:      t,
		Level:     m[logRe.SubexpIndex("level")],
		Component: m[logRe.SubexpIndex("comp")],
		Message:   m[logRe.SubexpIndex("msg")],
	}, nil
}
@@ hint
`logRe.SubexpIndex("level")` возвращает индекс именованной группы в результате `FindStringSubmatch`. Сообщение — всё после первого `": "`, поэтому группа `(?P<msg>.*)`.
