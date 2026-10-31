---
title: Безопасный SQL-билдер
level: medium
---
@@ statement
Напиши небольшой построитель SQL для PostgreSQL, устойчивый к инъекциям: значения — только через плейсхолдеры `$1, $2, ...`, а идентификаторы (таблицы, колонки, `ORDER BY`) проверяются.

Валидный идентификатор: `^[a-z_][a-z0-9_]*$` (длина до 63). Любой другой → ошибка, для которой `errors.Is(err, ErrInvalidIdentifier)` истинно.

- `InsertSQL(table string, cols []string) (string, error)` → `INSERT INTO users (name, email) VALUES ($1, $2)`; пустой список колонок — ошибка.
- `SelectSQL(table string, cols []string, where []Cond, orderBy string, limit int) (string, []any, error)`:
  - `SELECT a, b FROM t`; пустой `cols` — ошибка (никаких `SELECT *`);
  - `where` — условия через `AND`: `Cond{Col, Op string; Val any}`, `Op` из набора `=`, `<>`, `<`, `>`, `<=`, `>=`, `LIKE` (иначе ошибка `ErrInvalidOperator`); значения возвращаются вторым результатом в порядке плейсхолдеров;
  - `orderBy` — `""`, `"col"`, `"col ASC"` или `"col DESC"` (регистр ключевого слова любой, в SQL выводится заглавными); иное — ошибка;
  - `limit > 0` → ` LIMIT <n>` (число вставляется в текст — оно типа `int`, это безопасно).
@@ starter
package solution

import "errors"

var (
	ErrInvalidIdentifier = errors.New("недопустимый идентификатор")
	ErrInvalidOperator   = errors.New("недопустимый оператор")
)

type Cond struct {
	Col string
	Op  string
	Val any
}

func InsertSQL(table string, cols []string) (string, error) {
	return "", nil
}

func SelectSQL(table string, cols []string, where []Cond, orderBy string, limit int) (string, []any, error) {
	return "", nil, nil
}
@@ tests
package solution

import (
	"errors"
	"reflect"
	"testing"
)

func TestInsertSQL(t *testing.T) {
	got, err := InsertSQL("users", []string{"name", "email", "created_at"})
	want := "INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)"
	if err != nil || got != want {
		t.Errorf("got %q, %v; want %q", got, err, want)
	}
	if _, err := InsertSQL("users", nil); err == nil {
		t.Error("пустой список колонок")
	}
	for _, bad := range [][2]string{{"users; DROP TABLE x", "a"}, {"users", "a, b"}, {"Users", "a"}, {"1users", "a"}, {"users", "a--"}, {"", "a"}} {
		_, err := InsertSQL(bad[0], []string{bad[1]})
		if !errors.Is(err, ErrInvalidIdentifier) {
			t.Errorf("InsertSQL(%q, %q): err = %v", bad[0], bad[1], err)
		}
	}
}

func TestSelectSQL(t *testing.T) {
	sql, args, err := SelectSQL("users", []string{"id", "name"},
		[]Cond{{"age", ">=", 18}, {"name", "LIKE", "A%"}}, "created_at desc", 10)
	wantSQL := "SELECT id, name FROM users WHERE age >= $1 AND name LIKE $2 ORDER BY created_at DESC LIMIT 10"
	if err != nil || sql != wantSQL || !reflect.DeepEqual(args, []any{18, "A%"}) {
		t.Errorf("got %q %v %v\nwant %q", sql, args, err, wantSQL)
	}

	sql, args, err = SelectSQL("orders", []string{"id"}, nil, "", 0)
	if err != nil || sql != "SELECT id FROM orders" || len(args) != 0 {
		t.Errorf("минимальный: %q %v %v", sql, args, err)
	}

	sql, _, _ = SelectSQL("t", []string{"a"}, nil, "id", 0)
	if sql != "SELECT a FROM t ORDER BY id" {
		t.Errorf("order без направления: %q", sql)
	}
}

func TestSelectSQLInjection(t *testing.T) {
	cols := []string{"id"}
	if _, _, err := SelectSQL("users", cols, []Cond{{"name", "= 1 OR 1=1 --", "x"}}, "", 0); !errors.Is(err, ErrInvalidOperator) {
		t.Errorf("оператор: %v", err)
	}
	if _, _, err := SelectSQL("users", cols, []Cond{{"name; DROP", "=", "x"}}, "", 0); !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("колонка в where: %v", err)
	}
	for _, ob := range []string{"id; DROP TABLE users", "id DESC, name", "id SIDEWAYS", "(select 1)"} {
		if _, _, err := SelectSQL("users", cols, nil, ob, 0); err == nil {
			t.Errorf("orderBy %q должен быть отклонён", ob)
		}
	}
	if _, _, err := SelectSQL("users", nil, nil, "", 0); err == nil {
		t.Error("пустой список колонок")
	}
	if _, _, err := SelectSQL("users", []string{"id FROM secrets --"}, nil, "", 0); !errors.Is(err, ErrInvalidIdentifier) {
		t.Errorf("колонка: %v", err)
	}
}
@@ solution
package solution

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidIdentifier = errors.New("недопустимый идентификатор")
	ErrInvalidOperator   = errors.New("недопустимый оператор")
)

type Cond struct {
	Col string
	Op  string
	Val any
}

var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

func checkIdent(s string) error {
	if !identRe.MatchString(s) {
		return fmt.Errorf("%w: %q", ErrInvalidIdentifier, s)
	}
	return nil
}

func checkIdents(list []string) error {
	if len(list) == 0 {
		return errors.New("нужна хотя бы одна колонка")
	}
	for _, s := range list {
		if err := checkIdent(s); err != nil {
			return err
		}
	}
	return nil
}

func InsertSQL(table string, cols []string) (string, error) {
	if err := checkIdent(table); err != nil {
		return "", err
	}
	if err := checkIdents(cols); err != nil {
		return "", err
	}
	ph := make([]string, len(cols))
	for i := range cols {
		ph[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(ph, ", ")), nil
}

var allowedOps = map[string]bool{"=": true, "<>": true, "<": true, ">": true, "<=": true, ">=": true, "LIKE": true}

func SelectSQL(table string, cols []string, where []Cond, orderBy string, limit int) (string, []any, error) {
	if err := checkIdent(table); err != nil {
		return "", nil, err
	}
	if err := checkIdents(cols); err != nil {
		return "", nil, err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "SELECT %s FROM %s", strings.Join(cols, ", "), table)

	var args []any
	for i, c := range where {
		if err := checkIdent(c.Col); err != nil {
			return "", nil, err
		}
		op := strings.ToUpper(c.Op)
		if !allowedOps[op] {
			return "", nil, fmt.Errorf("%w: %q", ErrInvalidOperator, c.Op)
		}
		if i == 0 {
			sb.WriteString(" WHERE ")
		} else {
			sb.WriteString(" AND ")
		}
		args = append(args, c.Val)
		fmt.Fprintf(&sb, "%s %s $%d", c.Col, op, len(args))
	}

	if orderBy != "" {
		parts := strings.Fields(orderBy)
		if len(parts) == 0 || len(parts) > 2 {
			return "", nil, fmt.Errorf("недопустимый ORDER BY: %q", orderBy)
		}
		if err := checkIdent(parts[0]); err != nil {
			return "", nil, err
		}
		sb.WriteString(" ORDER BY " + parts[0])
		if len(parts) == 2 {
			dir := strings.ToUpper(parts[1])
			if dir != "ASC" && dir != "DESC" {
				return "", nil, fmt.Errorf("недопустимое направление сортировки: %q", parts[1])
			}
			sb.WriteString(" " + dir)
		}
	}
	if limit > 0 {
		fmt.Fprintf(&sb, " LIMIT %d", limit)
	}
	return sb.String(), args, nil
}
@@ hint
Все «опасные» части (идентификаторы, операторы, направление сортировки) проверяй по белому списку/регулярке, значения — только через `$N` и слайс аргументов.
