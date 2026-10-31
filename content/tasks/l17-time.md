---
title: Время: календарные дни и рабочие дни
level: medium
---
@@ statement
Три функции для работы с датами.

1. `DaysBetween(a, b time.Time) int` — сколько **календарных дней** от даты `a` до даты `b` (время суток и часовой пояс каждого значения учитываются только для определения его календарной даты в собственной локации; результат отрицателен, если `b` раньше). `DaysBetween(25 янв 23:59, 26 янв 00:01) == 1`.
2. `NextBusinessDay(t time.Time) time.Time` — полночь (00:00 в той же локации, что у `t`) следующего рабочего дня (пн–пт) **после** даты `t`. Из пятницы, субботы и воскресенья — следующий понедельник.
3. `StartOfWeek(t time.Time) time.Time` — полночь понедельника той недели, в которую входит `t` (неделя с понедельника), в локации `t`.

Не прибавляй «24 часа» для перехода на следующий день (из-за перехода на летнее время сутки бывают 23 или 25 часов) — используй `AddDate`/`time.Date`.
@@ starter
package solution

import "time"

func DaysBetween(a, b time.Time) int { return 0 }

func NextBusinessDay(t time.Time) time.Time { return t }

func StartOfWeek(t time.Time) time.Time { return t }
@@ tests
package solution

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day, h, min int) time.Time {
	return time.Date(y, m, day, h, min, 0, 0, time.UTC)
}

func TestDaysBetween(t *testing.T) {
	tests := []struct {
		a, b time.Time
		want int
	}{
		{d(2025, 1, 25, 23, 59), d(2025, 1, 26, 0, 1), 1},
		{d(2025, 1, 26, 0, 1), d(2025, 1, 25, 23, 59), -1},
		{d(2025, 3, 1, 10, 0), d(2025, 3, 1, 22, 0), 0},
		{d(2024, 2, 28, 12, 0), d(2024, 3, 1, 12, 0), 2}, // високосный год
		{d(2023, 12, 31, 0, 0), d(2025, 1, 1, 0, 0), 367},
	}
	for _, tt := range tests {
		if got := DaysBetween(tt.a, tt.b); got != tt.want {
			t.Errorf("DaysBetween(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestDaysBetweenDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("нет базы часовых поясов")
	}
	a := time.Date(2025, 3, 8, 12, 0, 0, 0, loc)  // накануне перехода на летнее время
	b := time.Date(2025, 3, 10, 12, 0, 0, 0, loc) // сутки 9 марта длились 23 часа
	if got := DaysBetween(a, b); got != 2 {
		t.Errorf("DST: got %d, want 2", got)
	}
}

func TestNextBusinessDay(t *testing.T) {
	tests := []struct{ in, want time.Time }{
		{d(2025, 3, 10, 15, 30), d(2025, 3, 11, 0, 0)}, // пн → вт
		{d(2025, 3, 14, 9, 0), d(2025, 3, 17, 0, 0)},   // пт → пн
		{d(2025, 3, 15, 9, 0), d(2025, 3, 17, 0, 0)},   // сб → пн
		{d(2025, 3, 16, 23, 59), d(2025, 3, 17, 0, 0)}, // вс → пн
		{d(2025, 12, 31, 8, 0), d(2026, 1, 1, 0, 0)},   // через границу года
	}
	for _, tt := range tests {
		if got := NextBusinessDay(tt.in); !got.Equal(tt.want) {
			t.Errorf("NextBusinessDay(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestStartOfWeek(t *testing.T) {
	tests := []struct{ in, want time.Time }{
		{d(2025, 3, 10, 15, 30), d(2025, 3, 10, 0, 0)}, // пн
		{d(2025, 3, 13, 1, 0), d(2025, 3, 10, 0, 0)},   // чт
		{d(2025, 3, 16, 23, 59), d(2025, 3, 10, 0, 0)}, // вс относится к неделе, начавшейся 10-го
		{d(2025, 1, 1, 12, 0), d(2024, 12, 30, 0, 0)},  // через границу года
	}
	for _, tt := range tests {
		if got := StartOfWeek(tt.in); !got.Equal(tt.want) {
			t.Errorf("StartOfWeek(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
@@ solution
package solution

import "time"

func midnightUTC(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func DaysBetween(a, b time.Time) int {
	return int(midnightUTC(b).Sub(midnightUTC(a)).Hours() / 24)
}

func NextBusinessDay(t time.Time) time.Time {
	y, m, d := t.Date()
	next := time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func StartOfWeek(t time.Time) time.Time {
	y, m, d := t.Date()
	offset := (int(t.Weekday()) + 6) % 7 // пн=0 ... вс=6
	return time.Date(y, m, d-offset, 0, 0, 0, 0, t.Location())
}
@@ hint
`time.Date(y, m, d+1, 0, 0, 0, 0, loc)` нормализует переполнение дня. Для `DaysBetween` приведи обе даты к «полуночи в UTC» по их Y/M/D и поделить разницу на 24 часа.
