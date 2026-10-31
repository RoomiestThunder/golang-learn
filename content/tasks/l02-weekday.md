---
title: Перечисление через iota
level: easy
---
@@ statement
Сделай перечисление дней недели `Weekday` (тип на базе `int`) через `iota`: `Monday = 1`, …, `Sunday = 7`.

Реализуй:

- `func (d Weekday) String() string` — русское название ("Понедельник", …, "Воскресенье"); для невалидного значения — `"Неизвестно"`;
- `func (d Weekday) IsWeekend() bool`;
- `func (d Weekday) Next() Weekday` — следующий день по кругу (после воскресенья — понедельник).
@@ starter
package solution

type Weekday int

// TODO: константы Monday..Sunday (Monday = 1) через iota

func (d Weekday) String() string { return "" }

func (d Weekday) IsWeekend() bool { return false }

func (d Weekday) Next() Weekday { return 0 }
@@ tests
package solution

import "testing"

func TestWeekday(t *testing.T) {
	if Monday != 1 || Sunday != 7 || Wednesday != 3 {
		t.Fatalf("значения iota: Monday=%d Wednesday=%d Sunday=%d", Monday, Wednesday, Sunday)
	}
	names := map[Weekday]string{
		Monday: "Понедельник", Tuesday: "Вторник", Wednesday: "Среда", Thursday: "Четверг",
		Friday: "Пятница", Saturday: "Суббота", Sunday: "Воскресенье", Weekday(0): "Неизвестно", Weekday(9): "Неизвестно",
	}
	for d, want := range names {
		if got := d.String(); got != want {
			t.Errorf("Weekday(%d).String() = %q, want %q", int(d), got, want)
		}
	}
}

func TestWeekend(t *testing.T) {
	for d := Monday; d <= Sunday; d++ {
		want := d == Saturday || d == Sunday
		if d.IsWeekend() != want {
			t.Errorf("%v.IsWeekend() = %v", d, !want)
		}
	}
}

func TestNext(t *testing.T) {
	if Monday.Next() != Tuesday || Saturday.Next() != Sunday || Sunday.Next() != Monday {
		t.Error("Next работает неверно")
	}
}
@@ solution
package solution

type Weekday int

const (
	Monday Weekday = iota + 1
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

var names = [...]string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

func (d Weekday) String() string {
	if d < Monday || d > Sunday {
		return "Неизвестно"
	}
	return names[d-1]
}

func (d Weekday) IsWeekend() bool { return d == Saturday || d == Sunday }

func (d Weekday) Next() Weekday {
	if d >= Sunday {
		return Monday
	}
	return d + 1
}
@@ hint
`Monday Weekday = iota + 1` — тогда следующие константы продолжат счёт 2, 3, ... Для названий удобен массив или `switch`.
