---
title: Валидация с errors.Join
level: easy
---
@@ statement
Реализуй `ValidateUser(u User) error`, которая проверяет **все** правила и возвращает объединённую ошибку (`errors.Join`), а не только первую:

- пустое имя (после `TrimSpace`) → `ErrNameEmpty`;
- в `Email` нет `@` или он первый/последний символ → `ErrEmailInvalid`;
- `Age < 18` → `ErrUnderage`.

Для валидного пользователя — `nil`. Тесты проверяют, что через `errors.Is` находится **каждая** из ошибок.
@@ starter
package solution

import "errors"

var (
	ErrNameEmpty    = errors.New("имя не заполнено")
	ErrEmailInvalid = errors.New("некорректный email")
	ErrUnderage     = errors.New("нет 18 лет")
)

type User struct {
	Name  string
	Email string
	Age   int
}

func ValidateUser(u User) error {
	return nil
}
@@ tests
package solution

import (
	"errors"
	"testing"
)

func TestValidateUser(t *testing.T) {
	if err := ValidateUser(User{"Ann", "ann@example.com", 30}); err != nil {
		t.Fatalf("валидный пользователь: %v", err)
	}
	err := ValidateUser(User{"  ", "bad", 10})
	for _, target := range []error{ErrNameEmpty, ErrEmailInvalid, ErrUnderage} {
		if !errors.Is(err, target) {
			t.Errorf("в ошибке нет %v: %v", target, err)
		}
	}
	err = ValidateUser(User{"Bob", "@x", 20})
	if !errors.Is(err, ErrEmailInvalid) || errors.Is(err, ErrNameEmpty) || errors.Is(err, ErrUnderage) {
		t.Errorf("ожидалась только ErrEmailInvalid: %v", err)
	}
}
@@ solution
package solution

import (
	"errors"
	"strings"
)

var (
	ErrNameEmpty    = errors.New("имя не заполнено")
	ErrEmailInvalid = errors.New("некорректный email")
	ErrUnderage     = errors.New("нет 18 лет")
)

type User struct {
	Name  string
	Email string
	Age   int
}

func ValidateUser(u User) error {
	var errs []error
	if strings.TrimSpace(u.Name) == "" {
		errs = append(errs, ErrNameEmpty)
	}
	if i := strings.Index(u.Email, "@"); i <= 0 || i == len(u.Email)-1 {
		errs = append(errs, ErrEmailInvalid)
	}
	if u.Age < 18 {
		errs = append(errs, ErrUnderage)
	}
	return errors.Join(errs...)
}
@@ hint
`errors.Join()` без аргументов (или только с nil) возвращает `nil`.
