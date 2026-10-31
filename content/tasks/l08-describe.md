---
title: Type switch: Describe
level: easy
---
@@ statement
Напиши `Describe(v any) string` через `switch v.(type)`:

| Значение | Результат |
|---|---|
| `nil` | `"nil"` |
| `int` | `"int:<n>"` (например `int:42`) |
| `string` | `"string(<len> байт):<s>"`, длина в **байтах**, например `string(2 байт):hi` |
| `bool` | `"bool:true"` / `"bool:false"` |
| `[]int` | `"[]int len=<n>"` |
| значение, реализующее `error` | `"error:<текст>"` |
| значение, реализующее `fmt.Stringer` | `"stringer:<результат String()>"` |
| всё остальное | `"unknown:<%T>"` |

Порядок веток важен: если тип реализует `error` и `Stringer` одновременно — приоритет у `error`.
@@ starter
package solution

func Describe(v any) string {
	return ""
}
@@ tests
package solution

import (
	"errors"
	"fmt"
	"testing"
)

type temp float64

func (t temp) String() string { return fmt.Sprintf("%.1f градусов", float64(t)) }

type both struct{}

func (both) Error() string  { return "err" }
func (both) String() string { return "str" }

func TestDescribe(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{nil, "nil"},
		{42, "int:42"},
		{"hi", "string(2 байт):hi"},
		{"да", "string(4 байт):да"},
		{true, "bool:true"},
		{[]int{1, 2, 3}, "[]int len=3"},
		{errors.New("boom"), "error:boom"},
		{temp(36.6), "stringer:36.6 градусов"},
		{both{}, "error:err"},
		{3.5, "unknown:float64"},
		{map[string]int{}, "unknown:map[string]int"},
	}
	for _, tt := range tests {
		if got := Describe(tt.in); got != tt.want {
			t.Errorf("Describe(%#v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
@@ solution
package solution

import "fmt"

func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int:%d", x)
	case string:
		return fmt.Sprintf("string(%d байт):%s", len(x), x)
	case bool:
		return fmt.Sprintf("bool:%t", x)
	case []int:
		return fmt.Sprintf("[]int len=%d", len(x))
	case error:
		return "error:" + x.Error()
	case fmt.Stringer:
		return "stringer:" + x.String()
	default:
		return fmt.Sprintf("unknown:%T", x)
	}
}
@@ hint
В `case nil` переменная имеет тип `any`; в ветках с одним типом — этот тип. Ветку `error` ставь перед `fmt.Stringer`.
