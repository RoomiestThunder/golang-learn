---
title: Semver: разбор и сравнение версий
level: hard
---
@@ statement
Модули Go версионируются по [semver](https://semver.org). Реализуй разбор и сравнение версий.

```go
type Version struct {
	Major, Minor, Patch int
	Pre                 []string // идентификаторы пререлиза: "rc.1" → ["rc", "1"]
}
```

- `ParseVersion(s string) (Version, error)` — принимает `1.2.3`, `v1.2.3`, `1.2.3-rc.1`, `1.2.3+build.5` (метаданные сборки после `+` **игнорируются**). Ошибка для: `1.2`, `1.2.3.4`, `a.b.c`, ведущих нулей (`01.2.3`), пустых идентификаторов пререлиза (`1.2.3-`, `1.2.3-a..b`).
- `Compare(a, b Version) int` (`-1`, `0`, `1`) по правилам semver: сначала Major, Minor, Patch; версия **с** пререлизом младше такой же без него; пререлизы сравниваются идентификатор за идентификатором: числовые — как числа, буквенные — лексикографически (ASCII), числовой идентификатор младше буквенного; если все общие идентификаторы равны, меньше тот, у кого их меньше. (Так `1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0`.)
- `(v Version) String() string` — каноническая запись без `v` и без метаданных: `1.2.3` или `1.2.3-rc.1`.
- `(v Version) Bump(part string) (Version, error)` — `"major"`, `"minor"`, `"patch"`: увеличивает часть, обнуляет младшие, **сбрасывает пререлиз**. Другое значение → ошибка.
- `Max(versions []string) (string, error)` — наибольшая версия из списка (вернуть исходную строку). Пустой список или неверная версия → ошибка.
@@ starter
package solution

type Version struct {
	Major, Minor, Patch int
	Pre                 []string
}

func ParseVersion(s string) (Version, error) { return Version{}, nil }

func Compare(a, b Version) int { return 0 }

func (v Version) String() string { return "" }

func (v Version) Bump(part string) (Version, error) { return v, nil }

func Max(versions []string) (string, error) { return "", nil }
@@ tests
package solution

import (
	"slices"
	"testing"
)

func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := ParseVersion(s)
	if err != nil {
		t.Fatalf("ParseVersion(%q): %v", s, err)
	}
	return v
}

func TestParse(t *testing.T) {
	v := mustParse(t, "v1.22.3-rc.1+build.7")
	if v.Major != 1 || v.Minor != 22 || v.Patch != 3 || !slices.Equal(v.Pre, []string{"rc", "1"}) {
		t.Errorf("got %+v", v)
	}
	if v := mustParse(t, "0.0.0"); v.Major != 0 || len(v.Pre) != 0 {
		t.Errorf("got %+v", v)
	}
	for _, bad := range []string{"", "1.2", "1.2.3.4", "a.b.c", "01.2.3", "1.2.3-", "1.2.3-a..b", "v", "1.2.x", "1.2.3-@"} {
		if _, err := ParseVersion(bad); err == nil {
			t.Errorf("ParseVersion(%q) должна вернуть ошибку", bad)
		}
	}
}

func TestString(t *testing.T) {
	for in, want := range map[string]string{"v1.2.3": "1.2.3", "1.2.3-rc.1+meta": "1.2.3-rc.1", "2.0.0+x": "2.0.0"} {
		if got := mustParse(t, in).String(); got != want {
			t.Errorf("String(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompareOrdering(t *testing.T) {
	// список в порядке возрастания (пример из спецификации semver)
	order := []string{
		"0.9.9", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta",
		"1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0",
	}
	for i := range order {
		for j := range order {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := Compare(mustParse(t, order[i]), mustParse(t, order[j])); got != want {
				t.Errorf("Compare(%s, %s) = %d, want %d", order[i], order[j], got, want)
			}
		}
	}
	if Compare(mustParse(t, "1.0.0+a"), mustParse(t, "v1.0.0+b")) != 0 {
		t.Error("метаданные сборки не влияют на сравнение")
	}
	if Compare(mustParse(t, "1.2.10"), mustParse(t, "1.2.9")) != 1 {
		t.Error("числа сравниваются как числа, а не строки")
	}
}

func TestBump(t *testing.T) {
	tests := []struct{ in, part, want string }{
		{"1.2.3", "patch", "1.2.4"}, {"1.2.3", "minor", "1.3.0"}, {"1.2.3", "major", "2.0.0"},
		{"1.2.3-rc.1", "patch", "1.2.4"}, {"0.9.9", "minor", "0.10.0"},
	}
	for _, tt := range tests {
		got, err := mustParse(t, tt.in).Bump(tt.part)
		if err != nil || got.String() != tt.want {
			t.Errorf("Bump(%s, %s) = %v, %v; want %s", tt.in, tt.part, got, err, tt.want)
		}
	}
	if _, err := mustParse(t, "1.0.0").Bump("build"); err == nil {
		t.Error("неизвестная часть должна давать ошибку")
	}
}

func TestMax(t *testing.T) {
	got, err := Max([]string{"v1.9.0", "v1.10.0", "v1.10.0-rc.1", "v0.99.0"})
	if err != nil || got != "v1.10.0" {
		t.Errorf("Max = %q, %v", got, err)
	}
	if _, err := Max(nil); err == nil {
		t.Error("пустой список")
	}
	if _, err := Max([]string{"1.0.0", "oops"}); err == nil {
		t.Error("плохая версия")
	}
}
@@ solution
package solution

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch int
	Pre                 []string
}

var semverRe = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

func ParseVersion(s string) (Version, error) {
	m := semverRe.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("некорректная версия %q", s)
	}
	var v Version
	v.Major, _ = strconv.Atoi(m[1])
	v.Minor, _ = strconv.Atoi(m[2])
	v.Patch, _ = strconv.Atoi(m[3])
	if m[4] != "" {
		v.Pre = strings.Split(m[4], ".")
	}
	return v, nil
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func isNum(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

func cmpIdent(a, b string) int {
	an, bn := isNum(a), isNum(b)
	switch {
	case an && bn:
		x, _ := strconv.ParseUint(a, 10, 64)
		y, _ := strconv.ParseUint(b, 10, 64)
		if x < y {
			return -1
		} else if x > y {
			return 1
		}
		return 0
	case an:
		return -1
	case bn:
		return 1
	}
	return strings.Compare(a, b)
}

func Compare(a, b Version) int {
	if c := cmpInt(a.Major, b.Major); c != 0 {
		return c
	}
	if c := cmpInt(a.Minor, b.Minor); c != 0 {
		return c
	}
	if c := cmpInt(a.Patch, b.Patch); c != 0 {
		return c
	}
	switch {
	case len(a.Pre) == 0 && len(b.Pre) == 0:
		return 0
	case len(a.Pre) == 0:
		return 1
	case len(b.Pre) == 0:
		return -1
	}
	for i := 0; i < len(a.Pre) && i < len(b.Pre); i++ {
		if c := cmpIdent(a.Pre[i], b.Pre[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(a.Pre), len(b.Pre))
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Pre) > 0 {
		s += "-" + strings.Join(v.Pre, ".")
	}
	return s
}

func (v Version) Bump(part string) (Version, error) {
	switch part {
	case "major":
		return Version{Major: v.Major + 1}, nil
	case "minor":
		return Version{Major: v.Major, Minor: v.Minor + 1}, nil
	case "patch":
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}, nil
	}
	return v, fmt.Errorf("неизвестная часть версии %q", part)
}

func Max(versions []string) (string, error) {
	if len(versions) == 0 {
		return "", errors.New("пустой список версий")
	}
	best, bestV := "", Version{}
	for i, s := range versions {
		v, err := ParseVersion(s)
		if err != nil {
			return "", err
		}
		if i == 0 || Compare(v, bestV) > 0 {
			best, bestV = s, v
		}
	}
	return best, nil
}
@@ hint
Разбор удобнее всего делать регуляркой с группами: `^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-(...))?(?:\+...)?$`. При сравнении пререлизов идентификаторы сравнивай по правилам: число < строка.
