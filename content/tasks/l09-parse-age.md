---
title: Ошибки: Is, As и оборачивание
level: medium
---
@@ statement
Реализуй разбор возраста с «богатыми» ошибками.

- `ErrEmpty` — sentinel: вход пустой (после `TrimSpace`).
- `*RangeError{Got int}` — тип ошибки (метод `Error()` на указателе): возраст вне диапазона `0..150`.
- `ParseAge(s string) (int, error)`:
  - пустая строка → `ErrEmpty` (можно обернуть через `%w`);
  - не число → ошибка, оборачивающая исходную ошибку `strconv` (должна находиться `errors.As` как `*strconv.NumError`);
  - вне диапазона → `*RangeError`.
- `Classify(err error) string` — возвращает `"ok"` для nil, `"empty"` для `ErrEmpty`, `"range"` для `*RangeError`, `"syntax"` если в цепочке `strconv.ErrSyntax`, иначе `"unknown"`. Пользуйся `errors.Is`/`errors.As`, а не `==`.
@@ starter
package solution

import "errors"

var ErrEmpty = errors.New("пустая строка")

type RangeError struct{ Got int }

func (e *RangeError) Error() string { return "" } // TODO: осмысленный текст

func ParseAge(s string) (int, error) {
	return 0, nil
}

func Classify(err error) string {
	return ""
}
@@ tests
package solution

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func TestParseAge(t *testing.T) {
	if n, err := ParseAge(" 42 "); err != nil || n != 42 {
		t.Fatalf("ParseAge(42) = %d, %v", n, err)
	}
	if _, err := ParseAge("  "); !errors.Is(err, ErrEmpty) {
		t.Errorf("пустая: %v", err)
	}
	_, err := ParseAge("abc")
	var ne *strconv.NumError
	if !errors.As(err, &ne) {
		t.Errorf("не число: %v — ошибка strconv потерялась (нужен %%w)", err)
	}
	_, err = ParseAge("200")
	var re *RangeError
	if !errors.As(err, &re) || re.Got != 200 {
		t.Errorf("вне диапазона: %v", err)
	}
	if _, err = ParseAge("-1"); !errors.As(err, &re) || re.Got != -1 {
		t.Errorf("отрицательный: %v", err)
	}
	if s := re.Error(); s == "" {
		t.Error("у RangeError пустой текст")
	}
}

func TestClassify(t *testing.T) {
	_, e1 := ParseAge("")
	_, e2 := ParseAge("x1")
	_, e3 := ParseAge("999")
	wrapped := fmt.Errorf("контекст: %w", e3)
	tests := []struct {
		err  error
		want string
	}{
		{nil, "ok"}, {e1, "empty"}, {e2, "syntax"}, {e3, "range"}, {wrapped, "range"}, {errors.New("x"), "unknown"},
	}
	for _, tt := range tests {
		if got := Classify(tt.err); got != tt.want {
			t.Errorf("Classify(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}
@@ solution
package solution

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrEmpty = errors.New("пустая строка")

type RangeError struct{ Got int }

func (e *RangeError) Error() string { return fmt.Sprintf("возраст %d вне диапазона 0..150", e.Got) }

func ParseAge(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrEmpty
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("разбор возраста: %w", err)
	}
	if n < 0 || n > 150 {
		return 0, &RangeError{Got: n}
	}
	return n, nil
}

func Classify(err error) string {
	var re *RangeError
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, ErrEmpty):
		return "empty"
	case errors.As(err, &re):
		return "range"
	case errors.Is(err, strconv.ErrSyntax):
		return "syntax"
	default:
		return "unknown"
	}
}
@@ hint
`*strconv.NumError` реализует `Unwrap`, поэтому `errors.Is(err, strconv.ErrSyntax)` работает, если ты обернул ошибку через `%w`.
