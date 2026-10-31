---
title: Температуры: свои типы
level: easy
---
@@ statement
Объяви два **разных типа** на базе `float64`: `Celsius` и `Fahrenheit`. Реализуй:

- `CToF(c Celsius) Fahrenheit` и `FToC(f Fahrenheit) Celsius` (формулы `F = C*9/5 + 32`);
- метод `String()` для `Celsius`, который печатает значение с одним знаком после запятой и суффиксом `°C` (например `36.6°C`), и такой же для `Fahrenheit` с суффиксом `°F`.

Смысл задания — убедиться, что компилятор не даст смешать эти типы без явного преобразования.
@@ starter
package solution

type Celsius float64
type Fahrenheit float64

func CToF(c Celsius) Fahrenheit {
	return 0
}

func FToC(f Fahrenheit) Celsius {
	return 0
}

// TODO: методы String() для Celsius и Fahrenheit
@@ tests
package solution

import (
	"fmt"
	"math"
	"testing"
)

func TestConvert(t *testing.T) {
	if got := CToF(100); math.Abs(float64(got)-212) > 1e-9 {
		t.Errorf("CToF(100) = %v", got)
	}
	if got := FToC(-40); math.Abs(float64(got)+40) > 1e-9 {
		t.Errorf("FToC(-40) = %v", got)
	}
	if got := FToC(CToF(36.6)); math.Abs(float64(got)-36.6) > 1e-9 {
		t.Errorf("round trip = %v", got)
	}
}

func TestString(t *testing.T) {
	if s := fmt.Sprint(Celsius(36.6)); s != "36.6°C" {
		t.Errorf("Celsius.String() = %q", s)
	}
	if s := fmt.Sprint(Fahrenheit(98.64)); s != "98.6°F" {
		t.Errorf("Fahrenheit.String() = %q", s)
	}
	if s := fmt.Sprintf("%v", Celsius(-5)); s != "-5.0°C" {
		t.Errorf("got %q", s)
	}
}
@@ solution
package solution

import "fmt"

type Celsius float64
type Fahrenheit float64

func CToF(c Celsius) Fahrenheit { return Fahrenheit(c*9/5 + 32) }
func FToC(f Fahrenheit) Celsius { return Celsius((f - 32) * 5 / 9) }

func (c Celsius) String() string    { return fmt.Sprintf("%.1f°C", float64(c)) }
func (f Fahrenheit) String() string { return fmt.Sprintf("%.1f°F", float64(f)) }
@@ hint
Внутри `String()` приводи значение к `float64`, иначе `Sprintf("%.1f", c)` снова вызовет `String()` и получится бесконечная рекурсия.
