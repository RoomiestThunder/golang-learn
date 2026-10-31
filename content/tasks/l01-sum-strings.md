---
title: Сумма чисел из строк
level: easy
---
@@ statement
Напиши `SumStrings(args []string) (int, error)`: преобразуй каждую строку в `int` и верни сумму.

Если какая-то строка не число — верни `0` и ошибку, в тексте которой есть сама плохая строка (используй `fmt.Errorf` с `%w`, чтобы сохранить исходную ошибку `strconv`). Пустой список → `0, nil`.

Идиома Go: функция возвращает `(результат, error)`, вызывающий проверяет `if err != nil`.
@@ starter
package solution

// SumStrings суммирует числа, записанные строками.
func SumStrings(args []string) (int, error) {
	return 0, nil
}
@@ tests
package solution

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestSumStrings(t *testing.T) {
	got, err := SumStrings([]string{"1", "2", "-3", "10"})
	if err != nil || got != 10 {
		t.Fatalf("got %d, %v; want 10, nil", got, err)
	}
	got, err = SumStrings(nil)
	if err != nil || got != 0 {
		t.Fatalf("nil: got %d, %v", got, err)
	}
}

func TestSumStringsError(t *testing.T) {
	got, err := SumStrings([]string{"1", "abc", "3"})
	if err == nil {
		t.Fatal("ожидалась ошибка")
	}
	if got != 0 {
		t.Errorf("при ошибке сумма должна быть 0, got %d", got)
	}
	if !strings.Contains(err.Error(), "abc") {
		t.Errorf("в ошибке нет плохой строки: %v", err)
	}
	var ne *strconv.NumError
	if !errors.As(err, &ne) {
		t.Errorf("ошибку нужно оборачивать через %%w, чтобы errors.As нашёл *strconv.NumError")
	}
}
@@ solution
package solution

import (
	"fmt"
	"strconv"
)

func SumStrings(args []string) (int, error) {
	sum := 0
	for _, a := range args {
		n, err := strconv.Atoi(a)
		if err != nil {
			return 0, fmt.Errorf("плохое число %q: %w", a, err)
		}
		sum += n
	}
	return sum, nil
}
@@ hint
`strconv.Atoi` возвращает `(int, error)`. Оберни ошибку: `fmt.Errorf("...: %w", err)`.
