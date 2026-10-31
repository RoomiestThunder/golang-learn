---
title: Отмена через context
level: medium
---
@@ statement
1. `SleepCtx(ctx context.Context, d time.Duration) error` — ждёт `d`, но просыпается раньше, если `ctx` отменён: тогда возвращает `ctx.Err()`; иначе `nil`. Не используй `time.Sleep`; освобождай таймер (`timer.Stop`).
2. `ProcessAll(ctx context.Context, items []int, f func(context.Context, int) (int, error)) ([]int, error)` — последовательно применяет `f` к элементам. Перед каждым элементом проверь отмену контекста. Если контекст отменён — верни **уже накопленные** результаты и `ctx.Err()`. Если `f` вернула ошибку — верни накопленное и эту ошибку, обёрнутую с индексом (`fmt.Errorf("элемент %d: %w", i, err)`).
3. `WithValueRequestID` — пара функций `WithRequestID(ctx, id string) context.Context` и `RequestID(ctx) (string, bool)` с **собственным типом ключа**.
@@ starter
package solution

import (
	"context"
	"time"
)

func SleepCtx(ctx context.Context, d time.Duration) error {
	time.Sleep(d)
	return nil
}

func ProcessAll(ctx context.Context, items []int, f func(context.Context, int) (int, error)) ([]int, error) {
	return nil, nil
}

func WithRequestID(ctx context.Context, id string) context.Context { return ctx }

func RequestID(ctx context.Context) (string, bool) { return "", false }
@@ tests
package solution

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

func TestSleepCtx(t *testing.T) {
	if err := SleepCtx(context.Background(), 20*time.Millisecond); err != nil {
		t.Errorf("обычное ожидание: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := SleepCtx(ctx, 5*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v", err)
	}
	if time.Since(start) > time.Second {
		t.Error("SleepCtx не проснулась по отмене")
	}
	cctx, ccancel := context.WithCancel(context.Background())
	ccancel()
	if err := SleepCtx(cctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("уже отменённый: %v", err)
	}
}

func TestProcessAll(t *testing.T) {
	double := func(_ context.Context, x int) (int, error) { return x * 2, nil }
	got, err := ProcessAll(context.Background(), []int{1, 2, 3}, double)
	if err != nil || !slices.Equal(got, []int{2, 4, 6}) {
		t.Fatalf("got %v, %v", got, err)
	}

	// отмена на середине
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	f := func(_ context.Context, x int) (int, error) {
		calls++
		if x == 3 {
			cancel()
		}
		return x, nil
	}
	got, err = ProcessAll(ctx, []int{1, 2, 3, 4, 5}, f)
	if !errors.Is(err, context.Canceled) || !slices.Equal(got, []int{1, 2, 3}) || calls != 3 {
		t.Errorf("отмена: got %v, err %v, calls %d", got, err, calls)
	}

	// ошибка f
	boom := errors.New("boom")
	fail := func(_ context.Context, x int) (int, error) {
		if x == 2 {
			return 0, boom
		}
		return x, nil
	}
	got, err = ProcessAll(context.Background(), []int{1, 2, 3}, fail)
	if !errors.Is(err, boom) || !slices.Equal(got, []int{1}) {
		t.Errorf("ошибка: got %v, err %v", got, err)
	}
}

func TestRequestID(t *testing.T) {
	if _, ok := RequestID(context.Background()); ok {
		t.Error("пустой контекст")
	}
	ctx := WithRequestID(context.Background(), "r-1")
	if id, ok := RequestID(ctx); !ok || id != "r-1" {
		t.Errorf("RequestID = %q %v", id, ok)
	}
	if v := ctx.Value("request_id"); v != nil { // string-ключ не должен совпадать
		t.Error("ключ должен быть собственного типа, а не string")
	}
}
@@ solution
package solution

import (
	"context"
	"fmt"
	"time"
)

func SleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func ProcessAll(ctx context.Context, items []int, f func(context.Context, int) (int, error)) ([]int, error) {
	out := make([]int, 0, len(items))
	for i, it := range items {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		v, err := f(ctx, it)
		if err != nil {
			return out, fmt.Errorf("элемент %d: %w", i, err)
		}
		out = append(out, v)
	}
	return out, nil
}

type ctxKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok
}
@@ hint
В `ProcessAll` в начале каждой итерации: `if err := ctx.Err(); err != nil { return out, err }`. Ключ контекста: `type ctxKey struct{}`.
