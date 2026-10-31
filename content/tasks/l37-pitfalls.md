---
title: Пять классических ловушек Go
level: medium
---
@@ statement
Ниже пять функций, каждая содержит типичную для Go ошибку из тех, что любят на собеседованиях. Тесты покажут симптом; найди и исправь причину в каждой.

1. `ApplyDiscount(items []Item, pct float64)` — должна уменьшить цену каждого товара **на месте** на `pct` процентов.
2. `Record(m map[string]Stat, word, ctx string)` — увеличивает счётчик слова и запоминает последний контекст.
3. `ProcessAll(paths, open, use)` — открывает ресурсы по очереди, обрабатывает и закрывает; одновременно должен быть открыт **не более одного** ресурса; все открытые закрываются при любом выходе (в том числе по ошибке).
4. `LoadConfig(read, parse)` — прочитать и разобрать конфигурацию, вернуть конфиг и ошибку (ошибки чтения и разбора должны возвращаться).
5. `WithExtra(base []string, extra string) []string` — вернуть **новый** слайс с добавленным элементом; `base` и результаты предыдущих вызовов не должны меняться.
@@ starter
package solution

type Item struct {
	Name  string
	Price float64
}

func ApplyDiscount(items []Item, pct float64) {
	for _, it := range items {
		it.Price *= 1 - pct/100
	}
}

type Stat struct {
	Count int
	Last  string
}

func Record(m map[string]Stat, word, ctx string) {
	s := m[word]
	s.Count++
	s.Last = ctx
}

type Resource interface{ Close() error }

func ProcessAll(paths []string, open func(string) (Resource, error), use func(Resource) error) error {
	for _, p := range paths {
		r, err := open(p)
		if err != nil {
			return err
		}
		defer r.Close()
		if err := use(r); err != nil {
			return err
		}
	}
	return nil
}

type Config struct {
	Name string
	Port int
}

func LoadConfig(read func() ([]byte, error), parse func([]byte) (Config, error)) (Config, error) {
	var cfg Config
	data, err := read()
	if err == nil {
		cfg, err := parse(data)
		_, _ = cfg, err
	}
	return cfg, err
}

func WithExtra(base []string, extra string) []string {
	return append(base, extra)
}
@@ tests
package solution

import (
	"errors"
	"slices"
	"testing"
)

func TestApplyDiscount(t *testing.T) {
	items := []Item{{"a", 100}, {"b", 50}}
	ApplyDiscount(items, 10)
	if items[0].Price != 90 || items[1].Price != 45 {
		t.Errorf("цены не изменились на месте: %+v", items)
	}
}

func TestRecord(t *testing.T) {
	m := map[string]Stat{}
	Record(m, "go", "first")
	Record(m, "go", "second")
	Record(m, "rust", "x")
	if m["go"] != (Stat{2, "second"}) || m["rust"].Count != 1 {
		t.Errorf("m = %+v", m)
	}
}

type fakeRes struct {
	id     string
	closed bool
	reg    *registry
}

func (r *fakeRes) Close() error {
	if !r.closed {
		r.closed = true
		r.reg.open--
	}
	return nil
}

type registry struct {
	open, maxOpen int
	all           []*fakeRes
}

func (g *registry) opener(fail string) func(string) (Resource, error) {
	return func(p string) (Resource, error) {
		if p == fail {
			return nil, errors.New("open failed: " + p)
		}
		r := &fakeRes{id: p, reg: g}
		g.open++
		g.maxOpen = max(g.maxOpen, g.open)
		g.all = append(g.all, r)
		return r, nil
	}
}

func TestProcessAll(t *testing.T) {
	g := &registry{}
	used := 0
	err := ProcessAll([]string{"a", "b", "c", "d"}, g.opener(""), func(Resource) error { used++; return nil })
	if err != nil || used != 4 {
		t.Fatalf("err=%v used=%d", err, used)
	}
	if g.maxOpen != 1 {
		t.Errorf("одновременно было открыто до %d ресурсов (defer в цикле копится до конца функции)", g.maxOpen)
	}
	for _, r := range g.all {
		if !r.closed {
			t.Errorf("ресурс %s не закрыт", r.id)
		}
	}
}

func TestProcessAllErrors(t *testing.T) {
	g := &registry{}
	boom := errors.New("use failed")
	err := ProcessAll([]string{"a", "b", "c"}, g.opener(""), func(r Resource) error {
		if r.(*fakeRes).id == "b" {
			return boom
		}
		return nil
	})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
	for _, r := range g.all {
		if !r.closed {
			t.Errorf("при ошибке ресурс %s остался открытым", r.id)
		}
	}
	g = &registry{}
	if err := ProcessAll([]string{"a", "x", "c"}, g.opener("x"), func(Resource) error { return nil }); err == nil {
		t.Error("ошибка открытия должна возвращаться")
	}
	if g.open != 0 {
		t.Error("после ошибки открытия ничего не должно оставаться открытым")
	}
}

func TestLoadConfig(t *testing.T) {
	good := func() ([]byte, error) { return []byte("x"), nil }
	parseOK := func([]byte) (Config, error) { return Config{"svc", 8080}, nil }
	if cfg, err := LoadConfig(good, parseOK); err != nil || cfg != (Config{"svc", 8080}) {
		t.Errorf("успех: %+v %v", cfg, err)
	}
	readErr := errors.New("read failed")
	if _, err := LoadConfig(func() ([]byte, error) { return nil, readErr }, parseOK); !errors.Is(err, readErr) {
		t.Errorf("ошибка чтения: %v", err)
	}
	parseErr := errors.New("parse failed")
	if cfg, err := LoadConfig(good, func([]byte) (Config, error) { return Config{}, parseErr }); !errors.Is(err, parseErr) || cfg != (Config{}) {
		t.Errorf("ошибка разбора: %+v %v", cfg, err)
	}
}

func TestWithExtra(t *testing.T) {
	base := make([]string, 2, 10) // есть запас cap — append затирает общий массив
	base[0], base[1] = "x", "y"
	a := WithExtra(base, "a")
	b := WithExtra(base, "b")
	if !slices.Equal(a, []string{"x", "y", "a"}) || !slices.Equal(b, []string{"x", "y", "b"}) {
		t.Errorf("a=%v b=%v (результаты делят один массив)", a, b)
	}
	if len(base) != 2 || base[0] != "x" {
		t.Errorf("base изменён: %v", base)
	}
}
@@ solution
package solution

import (
	"slices"
)

type Item struct {
	Name  string
	Price float64
}

func ApplyDiscount(items []Item, pct float64) {
	for i := range items {
		items[i].Price *= 1 - pct/100
	}
}

type Stat struct {
	Count int
	Last  string
}

func Record(m map[string]Stat, word, ctx string) {
	s := m[word]
	s.Count++
	s.Last = ctx
	m[word] = s
}

type Resource interface{ Close() error }

func ProcessAll(paths []string, open func(string) (Resource, error), use func(Resource) error) error {
	for _, p := range paths {
		if err := processOne(p, open, use); err != nil {
			return err
		}
	}
	return nil
}

func processOne(p string, open func(string) (Resource, error), use func(Resource) error) error {
	r, err := open(p)
	if err != nil {
		return err
	}
	defer r.Close()
	return use(r)
}

type Config struct {
	Name string
	Port int
}

func LoadConfig(read func() ([]byte, error), parse func([]byte) (Config, error)) (Config, error) {
	data, err := read()
	if err != nil {
		return Config{}, err
	}
	cfg, err := parse(data)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func WithExtra(base []string, extra string) []string {
	return append(slices.Clone(base), extra)
}
@@ hint
(1) `for _, it := range` даёт копию — индексируй `items[i]`. (2) значение из map — копия, запиши обратно. (3) `defer` в цикле срабатывает в конце функции — вынеси тело в отдельную функцию. (4) `:=` в блоке создаёт новые переменные (shadowing). (5) `append` к срезу с запасом ёмкости пишет в общий массив — клонируй.
