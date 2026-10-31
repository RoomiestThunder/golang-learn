---
title: Фигуры и интерфейс Shape
level: easy
---
@@ statement
Дан интерфейс `Shape` с методами `Area() float64` и `Perimeter() float64`. Реализуй типы:

- `Rect{W, H float64}`;
- `Circle{R float64}`;
- `Triangle{A, B, C float64}` (по трём сторонам; площадь по формуле Герона).

И функции, работающие с интерфейсом:

- `TotalArea(shapes []Shape) float64`;
- `Largest(shapes []Shape) Shape` — фигура с наибольшей площадью (`nil` для пустого слайса);
- `SortByArea(shapes []Shape)` — отсортировать **на месте** по возрастанию площади (используй `sort.Slice`).

Также убедись на этапе компиляции, что все три типа реализуют `Shape`.
@@ starter
package solution

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }
type Triangle struct{ A, B, C float64 }

// TODO: методы Area и Perimeter для Rect, Circle, Triangle

func TotalArea(shapes []Shape) float64 { return 0 }
func Largest(shapes []Shape) Shape      { return nil }
func SortByArea(shapes []Shape)         {}
@@ tests
package solution

import (
	"math"
	"testing"
)

var _ = []Shape{Rect{}, Circle{}, Triangle{}} // типы должны реализовывать Shape

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestShapes(t *testing.T) {
	r := Rect{3, 4}
	if !near(r.Area(), 12) || !near(r.Perimeter(), 14) {
		t.Errorf("Rect: %v %v", r.Area(), r.Perimeter())
	}
	c := Circle{2}
	if !near(c.Area(), 4*math.Pi) || !near(c.Perimeter(), 4*math.Pi) {
		t.Errorf("Circle: %v %v", c.Area(), c.Perimeter())
	}
	tr := Triangle{3, 4, 5}
	if !near(tr.Area(), 6) || !near(tr.Perimeter(), 12) {
		t.Errorf("Triangle: %v %v", tr.Area(), tr.Perimeter())
	}
}

func TestAggregates(t *testing.T) {
	shapes := []Shape{Rect{3, 4}, Circle{1}, Triangle{3, 4, 5}}
	if got := TotalArea(shapes); !near(got, 12+math.Pi+6) {
		t.Errorf("TotalArea = %v", got)
	}
	if got := Largest(shapes); got != (Rect{3, 4}) {
		t.Errorf("Largest = %v", got)
	}
	if Largest(nil) != nil {
		t.Error("Largest(nil) должен быть nil")
	}
	SortByArea(shapes)
	if shapes[0] != (Circle{1}) || shapes[1] != (Triangle{3, 4, 5}) || shapes[2] != (Rect{3, 4}) {
		t.Errorf("после сортировки: %v", shapes)
	}
}
@@ solution
package solution

import (
	"math"
	"sort"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }
type Triangle struct{ A, B, C float64 }

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) Area() float64 {
	p := t.Perimeter() / 2
	return math.Sqrt(p * (p - t.A) * (p - t.B) * (p - t.C))
}

func TotalArea(shapes []Shape) float64 {
	var sum float64
	for _, s := range shapes {
		sum += s.Area()
	}
	return sum
}

func Largest(shapes []Shape) Shape {
	var best Shape
	for _, s := range shapes {
		if best == nil || s.Area() > best.Area() {
			best = s
		}
	}
	return best
}

func SortByArea(shapes []Shape) {
	sort.Slice(shapes, func(i, j int) bool { return shapes[i].Area() < shapes[j].Area() })
}
@@ hint
Формула Герона: `p = (a+b+c)/2`, площадь `sqrt(p(p-a)(p-b)(p-c))`.
