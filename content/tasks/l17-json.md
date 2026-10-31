---
title: JSON: статусы, строгий разбор
level: medium
---
@@ statement
Работа с JSON-API пользователей.

1. Тип `Status` (`StatusActive`, `StatusBlocked`) должен сериализоваться в JSON **строкой** (`"active"`, `"blocked"`). Реализуй `MarshalText`/`UnmarshalText` (разбор без учёта регистра; неизвестное значение → ошибка; сериализация недопустимого значения → ошибка).
2. `EncodeUsers(w io.Writer, users []User) error` — записывает JSON-массив. Поля помечены тегами (см. структуру). Особенность: `nil`-слайс `Tags` должен выводиться как `[]`, а не `null` (входные данные при этом **не менять**).
3. `DecodeUsers(r io.Reader) ([]User, error)` — читает JSON-массив **строго**: неизвестные поля → ошибка; после массива не должно быть лишних данных (проверь, что следующий токен — `io.EOF`).

Поле `Password` никогда не попадает в JSON; `Email` пропускается, если пуст; `Age` — `*int` (null, если не задан).
@@ starter
package solution

import "io"

type Status int

const (
	StatusActive Status = iota
	StatusBlocked
)

type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email,omitempty"`
	Password string   `json:"-"`
	Status   Status   `json:"status"`
	Tags     []string `json:"tags"`
	Age      *int     `json:"age"`
}

// TODO: MarshalText / UnmarshalText для Status

func EncodeUsers(w io.Writer, users []User) error {
	return nil
}

func DecodeUsers(r io.Reader) ([]User, error) {
	return nil, nil
}
@@ tests
package solution

import (
	"bytes"
	"strings"
	"testing"
)

func intp(n int) *int { return &n }

func TestEncode(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Ann", Password: "secret", Status: StatusActive},
		{ID: 2, Name: "Bob", Email: "b@x.io", Status: StatusBlocked, Tags: []string{"go"}, Age: intp(30)},
	}
	var buf bytes.Buffer
	if err := EncodeUsers(&buf, users); err != nil {
		t.Fatal(err)
	}
	want := `[{"id":1,"name":"Ann","status":"active","tags":[],"age":null},` +
		`{"id":2,"name":"Bob","email":"b@x.io","status":"blocked","tags":["go"],"age":30}]`
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Errorf("\n got: %s\nwant: %s", got, want)
	}
	if users[0].Tags != nil {
		t.Error("EncodeUsers не должна менять входные данные")
	}
}

func TestEncodeInvalidStatus(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeUsers(&buf, []User{{ID: 1, Status: Status(42)}}); err == nil {
		t.Error("ожидалась ошибка для недопустимого статуса")
	}
}

func TestDecode(t *testing.T) {
	in := `[{"id":1,"name":"Ann","status":"ACTIVE","tags":["a"],"age":null},{"id":2,"name":"Bob","status":"blocked","tags":[],"age":41}]`
	users, err := DecodeUsers(strings.NewReader(in))
	if err != nil || len(users) != 2 {
		t.Fatalf("%v %v", users, err)
	}
	if users[0].Status != StatusActive || users[0].Age != nil || users[1].Status != StatusBlocked || *users[1].Age != 41 {
		t.Errorf("неверный разбор: %+v %+v", users[0], users[1])
	}
}

func TestDecodeStrict(t *testing.T) {
	bad := []string{
		`[{"id":1,"name":"A","status":"active","tags":[],"age":null,"extra":true}]`, // неизвестное поле
		`[{"id":1,"name":"A","status":"weird","tags":[],"age":null}]`,               // плохой статус
		`[] []`, // лишние данные
		`{"id":1}`,
	}
	for _, in := range bad {
		if _, err := DecodeUsers(strings.NewReader(in)); err == nil {
			t.Errorf("ожидалась ошибка для %s", in)
		}
	}
}
@@ solution
package solution

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Status int

const (
	StatusActive Status = iota
	StatusBlocked
)

var statusNames = []string{"active", "blocked"}

func (s Status) MarshalText() ([]byte, error) {
	if s < 0 || int(s) >= len(statusNames) {
		return nil, fmt.Errorf("недопустимый статус %d", int(s))
	}
	return []byte(statusNames[s]), nil
}

func (s *Status) UnmarshalText(b []byte) error {
	for i, n := range statusNames {
		if strings.EqualFold(n, string(b)) {
			*s = Status(i)
			return nil
		}
	}
	return fmt.Errorf("неизвестный статус %q", b)
}

type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email,omitempty"`
	Password string   `json:"-"`
	Status   Status   `json:"status"`
	Tags     []string `json:"tags"`
	Age      *int     `json:"age"`
}

func EncodeUsers(w io.Writer, users []User) error {
	out := make([]User, len(users))
	copy(out, users)
	for i := range out {
		if out[i].Tags == nil {
			out[i].Tags = []string{}
		}
	}
	return json.NewEncoder(w).Encode(out)
}

func DecodeUsers(r io.Reader) ([]User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var users []User
	if err := dec.Decode(&users); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("лишние данные после JSON-массива")
	}
	return users, nil
}
@@ hint
Для лишних данных: после `dec.Decode(&v)` вызови `dec.Token()` — ожидаем `io.EOF`. Для `[]` вместо `null` подмени `nil` на пустой слайс в **копии** входных данных.
