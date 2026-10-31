---
title: Свой errgroup
level: hard
---
@@ statement
Реализуй упрощённый аналог `golang.org/x/sync/errgroup`:

```go
func WithContext(ctx context.Context) (*Group, context.Context)
func (g *Group) SetLimit(n int)          // n > 0: не более n одновременных задач; вызывается до Go
func (g *Group) Go(f func() error)       // запуск (если лимит достигнут — блокируется до свободного слота)
func (g *Group) Wait() error             // ждёт все задачи, возвращает ПЕРВУЮ ошибку
```

- Первая ошибка отменяет контекст, возвращённый `WithContext`.
- `Wait` тоже отменяет контекст (освобождает ресурсы) и возвращает первую ошибку (или `nil`).
- Последующие ошибки отбрасываются.
- `Group` без `WithContext` (нулевое значение) тоже должен работать, просто без отмены.
@@ starter
package solution

import "context"

type Group struct {
	// TODO
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	return &Group{}, ctx
}

func (g *Group) SetLimit(n int) {}

func (g *Group) Go(f func() error) { f() }

func (g *Group) Wait() error { return nil }
@@ tests
package solution

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroupFirstError(t *testing.T) {
	g, ctx := WithContext(context.Background())
	boom := errors.New("boom")
	g.Go(func() error { return nil })
	g.Go(func() error { time.Sleep(10 * time.Millisecond); return boom })
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return errors.New("не отменили контекст")
		}
	})
	if err := g.Wait(); !errors.Is(err, boom) {
		t.Errorf("Wait = %v, ожидалась первая ошибка boom", err)
	}
	if ctx.Err() == nil {
		t.Error("после Wait контекст должен быть отменён")
	}
}

func TestGroupNoError(t *testing.T) {
	g, ctx := WithContext(context.Background())
	var n atomic.Int32
	for i := 0; i < 20; i++ {
		g.Go(func() error { n.Add(1); return nil })
	}
	if err := g.Wait(); err != nil || n.Load() != 20 {
		t.Errorf("err=%v n=%d", err, n.Load())
	}
	_ = ctx
}

func TestGroupLimit(t *testing.T) {
	var g Group // нулевое значение
	g.SetLimit(3)
	var cur, peak atomic.Int32
	for i := 0; i < 12; i++ {
		g.Go(func() error {
			n := cur.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(15 * time.Millisecond)
			cur.Add(-1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	if p := peak.Load(); p > 3 || p < 2 {
		t.Errorf("пик параллелизма %d (лимит 3)", p)
	}
}

func TestGroupBlocksAtLimit(t *testing.T) {
	var g Group
	g.SetLimit(1)
	release := make(chan struct{})
	g.Go(func() error { <-release; return nil })
	returned := make(chan struct{})
	go func() {
		g.Go(func() error { return nil }) // должен заблокироваться, пока первая задача не завершится
		close(returned)
	}()
	select {
	case <-returned:
		t.Fatal("Go не заблокировался при достигнутом лимите")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("Go не разблокировался")
	}
	g.Wait()
}
@@ solution
package solution

import (
	"context"
	"sync"
)

type Group struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
	cancel  context.CancelFunc
	sem     chan struct{}
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) SetLimit(n int) {
	if n > 0 {
		g.sem = make(chan struct{}, n)
	}
}

func (g *Group) Go(f func() error) {
	if g.sem != nil {
		g.sem <- struct{}{} // блокируемся, пока нет свободного слота
	}
	g.wg.Add(1)
	go func() {
		defer func() {
			if g.sem != nil {
				<-g.sem
			}
			g.wg.Done()
		}()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
@@ hint
Лимит — буферизованный канал-семафор: `Go` кладёт в него значение перед запуском, горутина достаёт по завершении. Первую ошибку фиксируй через `sync.Once`.
