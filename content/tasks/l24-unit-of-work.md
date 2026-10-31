---
title: Транзакция: WithTx и Transfer
level: medium
---
@@ statement
Смоделируй работу с транзакциями поверх минимальных интерфейсов (как `database/sql`):

```go
type DB interface{ Begin(ctx context.Context) (Tx, error) }
type Tx interface {
	Exec(ctx context.Context, query string, args ...any) error
	Commit() error
	Rollback() error
}
```

1. `WithTx(ctx, db, fn func(Tx) error) error`:
   - `Begin`; ошибка Begin → вернуть её (без вызова `fn`);
   - `fn` вернула ошибку → `Rollback`, вернуть **ошибку fn** (если и откат упал — объединить обе через `errors.Join`);
   - `fn` запаниковала → `Rollback` и **повторная паника** с тем же значением;
   - успех → `Commit`; его ошибку вернуть (с контекстом через `%w`).
2. `Transfer(ctx, db, from, to int, amount int64) error` — внутри `WithTx` выполняет два `Exec`:
   - `UPDATE accounts SET balance = balance - $1 WHERE id = $2` (аргументы `amount, from`);
   - `UPDATE accounts SET balance = balance + $1 WHERE id = $2` (аргументы `amount, to`).

   Валидация до начала транзакции: `amount <= 0` или `from == to` → `ErrInvalidTransfer`, транзакция не открывается. Если упал любой `Exec`, откат и ошибка наружу.
@@ starter
package solution

import (
	"context"
	"errors"
)

var ErrInvalidTransfer = errors.New("некорректный перевод")

type DB interface{ Begin(ctx context.Context) (Tx, error) }

type Tx interface {
	Exec(ctx context.Context, query string, args ...any) error
	Commit() error
	Rollback() error
}

func WithTx(ctx context.Context, db DB, fn func(Tx) error) error {
	return nil
}

func Transfer(ctx context.Context, db DB, from, to int, amount int64) error {
	return nil
}
@@ tests
package solution

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeTx struct {
	log       *[]string
	failExec  int // номер Exec (с 1), на котором вернуть ошибку; 0 — не падать
	execCount int
	commitErr error
	rbErr     error
	args      [][]any
}

func (f *fakeTx) Exec(_ context.Context, q string, args ...any) error {
	f.execCount++
	*f.log = append(*f.log, "exec:"+q)
	f.args = append(f.args, args)
	if f.failExec == f.execCount {
		return errors.New("exec failed")
	}
	return nil
}
func (f *fakeTx) Commit() error   { *f.log = append(*f.log, "commit"); return f.commitErr }
func (f *fakeTx) Rollback() error { *f.log = append(*f.log, "rollback"); return f.rbErr }

type fakeDB struct {
	log      []string
	tx       *fakeTx
	beginErr error
}

func newDB() *fakeDB {
	db := &fakeDB{}
	db.tx = &fakeTx{log: &db.log}
	return db
}

func (d *fakeDB) Begin(context.Context) (Tx, error) {
	if d.beginErr != nil {
		return nil, d.beginErr
	}
	d.log = append(d.log, "begin")
	return d.tx, nil
}

func TestWithTxSuccess(t *testing.T) {
	db := newDB()
	err := WithTx(context.Background(), db, func(tx Tx) error { return tx.Exec(context.Background(), "q") })
	if err != nil || strings.Join(db.log, ",") != "begin,exec:q,commit" {
		t.Errorf("err=%v log=%v", err, db.log)
	}
}

func TestWithTxRollbackOnError(t *testing.T) {
	db := newDB()
	boom := errors.New("boom")
	err := WithTx(context.Background(), db, func(Tx) error { return boom })
	if !errors.Is(err, boom) || strings.Join(db.log, ",") != "begin,rollback" {
		t.Errorf("err=%v log=%v", err, db.log)
	}

	db = newDB()
	db.tx.rbErr = errors.New("rollback failed")
	err = WithTx(context.Background(), db, func(Tx) error { return boom })
	if !errors.Is(err, boom) || !errors.Is(err, db.tx.rbErr) {
		t.Errorf("обе ошибки должны быть доступны: %v", err)
	}
}

func TestWithTxPanic(t *testing.T) {
	db := newDB()
	defer func() {
		if r := recover(); r != "oops" {
			t.Errorf("паника должна пробрасываться, recover=%v", r)
		}
		if strings.Join(db.log, ",") != "begin,rollback" {
			t.Errorf("log=%v", db.log)
		}
	}()
	WithTx(context.Background(), db, func(Tx) error { panic("oops") })
}

func TestWithTxBeginAndCommitErrors(t *testing.T) {
	db := newDB()
	db.beginErr = errors.New("no conn")
	called := false
	err := WithTx(context.Background(), db, func(Tx) error { called = true; return nil })
	if !errors.Is(err, db.beginErr) || called {
		t.Errorf("begin: err=%v called=%v", err, called)
	}

	db = newDB()
	db.tx.commitErr = errors.New("commit failed")
	err = WithTx(context.Background(), db, func(Tx) error { return nil })
	if !errors.Is(err, db.tx.commitErr) {
		t.Errorf("commit: %v", err)
	}
}

func TestTransfer(t *testing.T) {
	db := newDB()
	if err := Transfer(context.Background(), db, 1, 2, 500); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"begin",
		"exec:UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		"exec:UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		"commit",
	}
	if !reflect.DeepEqual(db.log, want) {
		t.Errorf("log:\n%v\nwant:\n%v", db.log, want)
	}
	if !reflect.DeepEqual(db.tx.args, [][]any{{int64(500), 1}, {int64(500), 2}}) {
		t.Errorf("args = %v", db.tx.args)
	}
}

func TestTransferFailures(t *testing.T) {
	db := newDB()
	db.tx.failExec = 2
	if err := Transfer(context.Background(), db, 1, 2, 10); err == nil || db.log[len(db.log)-1] != "rollback" {
		t.Errorf("падение второго Exec: err=%v log=%v", err, db.log)
	}
	for _, c := range []struct {
		from, to int
		amt      int64
	}{{1, 1, 5}, {1, 2, 0}, {1, 2, -5}} {
		db := newDB()
		if err := Transfer(context.Background(), db, c.from, c.to, c.amt); !errors.Is(err, ErrInvalidTransfer) || len(db.log) != 0 {
			t.Errorf("%+v: err=%v log=%v (транзакция не должна открываться)", c, err, db.log)
		}
	}
}
@@ solution
package solution

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidTransfer = errors.New("некорректный перевод")

type DB interface{ Begin(ctx context.Context) (Tx, error) }

type Tx interface {
	Exec(ctx context.Context, query string, args ...any) error
	Commit() error
	Rollback() error
}

func WithTx(ctx context.Context, db DB, fn func(Tx) error) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func Transfer(ctx context.Context, db DB, from, to int, amount int64) error {
	if amount <= 0 || from == to {
		return ErrInvalidTransfer
	}
	return WithTx(ctx, db, func(tx Tx) error {
		if err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from); err != nil {
			return fmt.Errorf("списание: %w", err)
		}
		if err := tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to); err != nil {
			return fmt.Errorf("зачисление: %w", err)
		}
		return nil
	})
}
@@ hint
Панику ловит `defer` с `recover()`: откатить и сделать `panic(p)` снова. Ошибку `fn` сравни с ошибкой отката через `errors.Join`.
