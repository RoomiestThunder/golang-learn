---
title: Проект: кодировщик base62
level: easy
---
@@ statement
Первый кирпичик сокращателя ссылок: кодирование числовых ID в короткие строки.

Алфавит (по порядку): `0-9`, затем `a-z`, затем `A-Z` — всего 62 символа (`0` = 0, `a` = 10, `Z` = 61).

- `Encode(n uint64) string` — запись числа в системе счисления с основанием 62 (старшая цифра первой, без ведущих нулей). `Encode(0) == "0"`, `Encode(61) == "Z"`, `Encode(62) == "10"`.
- `Decode(s string) (uint64, error)` — обратная операция. Ошибки:
  - пустая строка или символ не из алфавита → `ErrInvalidCode`;
  - значение не помещается в `uint64` → `ErrOverflow`.
  Обе ошибки проверяются через `errors.Is`.

Для любого `n`: `Decode(Encode(n)) == n` (включая `math.MaxUint64`).
@@ starter
package solution

import "errors"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	ErrInvalidCode = errors.New("недопустимый код")
	ErrOverflow    = errors.New("число не помещается в uint64")
)

func Encode(n uint64) string {
	return ""
}

func Decode(s string) (uint64, error) {
	return 0, nil
}
@@ tests
package solution

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := map[uint64]string{0: "0", 1: "1", 9: "9", 10: "a", 35: "z", 36: "A", 61: "Z", 62: "10", 63: "11", 125: "21", 3843: "ZZ", 3844: "100"}
	for n, want := range tests {
		if got := Encode(n); got != want {
			t.Errorf("Encode(%d) = %q, want %q", n, got, want)
		}
	}
	if got := Encode(math.MaxUint64); got != "lYGhA16ahyf" {
		t.Errorf("Encode(MaxUint64) = %q", got)
	}
}

func TestDecode(t *testing.T) {
	for s, want := range map[string]uint64{"0": 0, "Z": 61, "10": 62, "21": 125, "ZZ": 3843, "lYGhA16ahyf": math.MaxUint64} {
		got, err := Decode(s)
		if err != nil || got != want {
			t.Errorf("Decode(%q) = %d, %v; want %d", s, got, err, want)
		}
	}
	for _, bad := range []string{"", "a b", "-1", "ab!", "é", "a_b"} {
		if _, err := Decode(bad); !errors.Is(err, ErrInvalidCode) {
			t.Errorf("Decode(%q): err = %v, want ErrInvalidCode", bad, err)
		}
	}
	for _, big := range []string{"lYGhA16ahyg", "zzzzzzzzzzzz", "ZZZZZZZZZZZZZZ"} {
		if _, err := Decode(big); !errors.Is(err, ErrOverflow) {
			t.Errorf("Decode(%q): err = %v, want ErrOverflow", big, err)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		n := rand.Uint64() >> uint(rand.IntN(64))
		got, err := Decode(Encode(n))
		if err != nil || got != n {
			t.Fatalf("round trip %d → %q → %d, %v", n, Encode(n), got, err)
		}
	}
}
@@ solution
package solution

import (
	"errors"
	"math"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	ErrInvalidCode = errors.New("недопустимый код")
	ErrOverflow    = errors.New("число не помещается в uint64")
)

func Encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [11]byte // 62^11 > 2^64
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}

func Decode(s string) (uint64, error) {
	if s == "" {
		return 0, ErrInvalidCode
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		d := strings.IndexByte(alphabet, s[i])
		if d < 0 {
			return 0, ErrInvalidCode
		}
		if n > (math.MaxUint64-uint64(d))/62 {
			return 0, ErrOverflow
		}
		n = n*62 + uint64(d)
	}
	return n, nil
}
@@ hint
Encode: пока `n > 0` — остаток `n % 62` даёт символ, `n /= 62`; собирай с конца буфера. Decode: `n = n*62 + digit`; перед умножением проверь переполнение: `n > (MaxUint64 - d) / 62`.
