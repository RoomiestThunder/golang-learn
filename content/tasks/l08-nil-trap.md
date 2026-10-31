---
title: Найди баг: nil-интерфейс
level: medium
---
@@ statement
В коде ниже классическая ловушка: `Find` возвращает **nil-указатель типа `*NotFoundError`, завёрнутый в `error`**, поэтому проверка `err != nil` у вызывающего срабатывает, хотя ошибки нет.

Исправь `Find` и `Lookup`, чтобы:

- при успехе возвращался настоящий `nil`;
- при отсутствии — `*NotFoundError` с полем `ID`, находимый через `errors.As`;
- `Lookup` по-прежнему использовал `mustFind` (**не меняй** сигнатуру `mustFind`, она эмулирует чужой код: возвращает `*NotFoundError`).

Так же поправь `Lookup`, чтобы он не терял типизацию: ошибка должна оставаться `*NotFoundError` при не найдено и быть `nil` при найдено.
@@ starter
package solution

import "fmt"

type NotFoundError struct{ ID int }

func (e *NotFoundError) Error() string { return fmt.Sprintf("запись %d не найдена", e.ID) }

var db = map[int]string{1: "один", 2: "два"}

// mustFind — «чужой код»: возвращает конкретный тип указателя на ошибку (nil при успехе).
func mustFind(id int) (string, *NotFoundError) {
	if v, ok := db[id]; ok {
		return v, nil
	}
	return "", &NotFoundError{ID: id}
}

// Find возвращает значение или ошибку.
func Find(id int) (string, error) {
	v, nfe := mustFind(id)
	return v, nfe // БАГ: при успехе вернёт непустой интерфейс с nil-указателем
}

// Lookup возвращает значение или "?" и ошибку.
func Lookup(id int) (string, error) {
	var err *NotFoundError
	v, err := mustFind(id)
	if err != nil {
		return "?", err
	}
	return v, err // БАГ того же рода
}
@@ tests
package solution

import (
	"errors"
	"testing"
)

func TestFind(t *testing.T) {
	v, err := Find(1)
	if err != nil {
		t.Fatalf("Find(1): err = %v (тип %T) — ожидался настоящий nil", err, err)
	}
	if v != "один" {
		t.Errorf("v = %q", v)
	}
	_, err = Find(9)
	var nfe *NotFoundError
	if !errors.As(err, &nfe) || nfe.ID != 9 {
		t.Errorf("Find(9): err = %v", err)
	}
}

func TestLookup(t *testing.T) {
	if v, err := Lookup(2); err != nil || v != "два" {
		t.Fatalf("Lookup(2) = %q, %v (%T)", v, err, err)
	}
	v, err := Lookup(7)
	var nfe *NotFoundError
	if v != "?" || !errors.As(err, &nfe) || nfe.ID != 7 {
		t.Errorf("Lookup(7) = %q, %v", v, err)
	}
}
@@ solution
package solution

import "fmt"

type NotFoundError struct{ ID int }

func (e *NotFoundError) Error() string { return fmt.Sprintf("запись %d не найдена", e.ID) }

var db = map[int]string{1: "один", 2: "два"}

func mustFind(id int) (string, *NotFoundError) {
	if v, ok := db[id]; ok {
		return v, nil
	}
	return "", &NotFoundError{ID: id}
}

func Find(id int) (string, error) {
	v, nfe := mustFind(id)
	if nfe != nil {
		return v, nfe
	}
	return v, nil
}

func Lookup(id int) (string, error) {
	v, nfe := mustFind(id)
	if nfe != nil {
		return "?", nfe
	}
	return v, nil
}
@@ hint
Проверяй конкретный указатель на `nil` **до** присваивания в `error`, а при успехе возвращай литерал `nil`.
