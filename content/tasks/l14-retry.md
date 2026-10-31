---
title: Retry с backoff и отменой
level: medium
---
@@ statement
Реализуй `Retry(ctx context.Context, attempts int, base time.Duration, fn func() error) error`.

- Вызывай `fn`, пока она не вернёт `nil`, но не более `attempts` раз (`attempts < 1` → считай 1).
- Между попытками пауза с экспоненциальным ростом: `base`, `2*base`, `4*base`, ... (после последней попытки пауза не нужна).
- Ждать нужно **с учётом отмены**: если `ctx` отменён во время паузы или до вызова — вернуть `ctx.Err()` (проверяй через `errors.Is`) немедленно.
- Если все попытки исчерпаны — вернуть ошибку `fmt.Errorf("после %d попыток: %w", attempts, lastErr)`.
- Если `fn` вернула ошибку, обёрнутую в `ErrPermanent` (проверка `errors.Is(err, ErrPermanent)`), повторять не нужно — вернуть её сразу.
@@ starter
package solution

import (
	"context"
	"errors"
	"time"
)

var ErrPermanent = errors.New("постоянная ошибка")

func Retry(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
	return fn()
}
@@ tests
package solution

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetrySuccessAfterFailures(t *testing.T) {
	calls := 0
	start := time.Now()
	err := Retry(context.Background(), 5, 10*time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("temp")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	if d := time.Since(start); d < 28*time.Millisecond { // паузы 10 + 20 мс
		t.Errorf("паузы слишком короткие: %v", d)
	}
}

func TestRetryExhausted(t *testing.T) {
	base := errors.New("down")
	calls := 0
	err := Retry(context.Background(), 3, time.Millisecond, func() error { calls++; return base })
	if calls != 3 || !errors.Is(err, base) {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	if err.Error() == base.Error() {
		t.Error("ошибка должна содержать контекст про число попыток")
	}
	calls = 0
	Retry(context.Background(), 0, time.Millisecond, func() error { calls++; return base })
	if calls != 1 {
		t.Errorf("attempts<1 → одна попытка, было %d", calls)
	}
}

func TestRetryPermanent(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), 5, time.Millisecond, func() error {
		calls++
		return fmt.Errorf("bad request: %w", ErrPermanent)
	})
	if calls != 1 || !errors.Is(err, ErrPermanent) {
		t.Errorf("calls=%d err=%v", calls, err)
	}
}

func TestRetryContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := Retry(ctx, 10, 500*time.Millisecond, func() error { return errors.New("fail") })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v", err)
	}
	if time.Since(start) > time.Second {
		t.Error("Retry не прервалась по отмене контекста")
	}
}
@@ solution
package solution

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrPermanent = errors.New("постоянная ошибка")

func Retry(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	delay := base
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = fn()
		if last == nil {
			return nil
		}
		if errors.Is(last, ErrPermanent) {
			return last
		}
		if i == attempts-1 {
			break
		}
		t := time.NewTimer(delay)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
		delay *= 2
	}
	return fmt.Errorf("после %d попыток: %w", attempts, last)
}
@@ hint
Пауза — `select` на `timer.C` и `ctx.Done()`. После каждой паузы `delay *= 2`.
