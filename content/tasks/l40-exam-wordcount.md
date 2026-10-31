---
title: Экзамен 2: параллельный подсчёт слов
level: hard
---
@@ statement
Реализуй

```go
type WordCount struct {
	Word  string
	Count int
}

func TopWords(ctx context.Context, readers []io.Reader, n, workers int) ([]WordCount, error)
```

Функция подсчитывает слова во **всех** потоках данных и возвращает `n` самых частых.

- **Слово** — максимальная последовательность Unicode-букв и цифр (`unicode.IsLetter`/`IsDigit`); регистр не важен (приводи к нижнему); всё остальное — разделители.
- Потоки обрабатываются **параллельно** пулом не более чем из `workers` горутин (`workers < 1` → 1). Читай поток **потоково** (не загружай целиком в память): подойдёт `bufio.Reader` + `ReadRune`.
- Подсчёт в каждом воркере — в своей локальной map, затем слияние в общую (без гонок).
- Результат: по убыванию `Count`, при равенстве — по возрастанию `Word`; не более `n` элементов (`n <= 0` → пустой результат).
- Если любой поток вернул ошибку (не `io.EOF`) — остальные работы останавливаются (отмена внутреннего контекста), функция возвращает ошибку, оборачивающую исходную (`errors.Is`).
- Если `ctx` отменён — вернуть `ctx.Err()` (проверка `errors.Is`). Воркеры должны проверять отмену регулярно (например, каждые несколько тысяч рун).
- К моменту возврата все горутины функции должны завершиться.
@@ starter
package solution

import (
	"context"
	"io"
)

type WordCount struct {
	Word  string
	Count int
}

func TopWords(ctx context.Context, readers []io.Reader, n, workers int) ([]WordCount, error) {
	return nil, nil
}
@@ tests
package solution

import (
	"context"
	"errors"
	"io"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"
)

func rd(ss ...string) []io.Reader {
	out := make([]io.Reader, len(ss))
	for i, s := range ss {
		out[i] = strings.NewReader(s)
	}
	return out
}

func TestBasic(t *testing.T) {
	got, err := TopWords(context.Background(), rd("The cat. the DOG, the dog!", "cat cat", "Привет, привет! мир"), 3, 2)
	want := []WordCount{{"cat", 3}, {"the", 3}, {"dog", 2}}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("got %v, %v; want %v", got, err, want)
	}
	got, _ = TopWords(context.Background(), rd("Привет, привет! мир"), 10, 1)
	if !slices.Equal(got, []WordCount{{"привет", 2}, {"мир", 1}}) {
		t.Errorf("юникод и n > числа слов: %v", got)
	}
}

func TestEdgeCases(t *testing.T) {
	if got, err := TopWords(context.Background(), nil, 5, 3); err != nil || len(got) != 0 {
		t.Errorf("нет потоков: %v %v", got, err)
	}
	if got, _ := TopWords(context.Background(), rd("a b c"), 0, 3); len(got) != 0 {
		t.Errorf("n=0: %v", got)
	}
	if got, _ := TopWords(context.Background(), rd("...!!!", "   "), 5, 0); len(got) != 0 {
		t.Errorf("нет слов: %v", got)
	}
	got, _ := TopWords(context.Background(), rd("a1 b2 a1 B2 a1"), 2, 1)
	if !slices.Equal(got, []WordCount{{"a1", 3}, {"b2", 2}}) {
		t.Errorf("цифры в словах: %v", got)
	}
	// слово, разорванное границей внутреннего буфера Read
	got, _ = TopWords(context.Background(), []io.Reader{iotest.OneByteReader(strings.NewReader("hello world hello"))}, 1, 1)
	if !slices.Equal(got, []WordCount{{"hello", 2}}) {
		t.Errorf("чтение по одному байту: %v", got)
	}
}

func TestReaderError(t *testing.T) {
	boom := errors.New("boom")
	rs := append(rd("a b c", "d e f"), iotest.ErrReader(boom))
	_, err := TopWords(context.Background(), rs, 3, 2)
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, ожидалась обёрнутая boom", err)
	}
}

func TestCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := TopWords(ctx, rd("a b c"), 3, 2); !errors.Is(err, context.Canceled) {
		t.Errorf("уже отменённый контекст: %v", err)
	}
}

type slowReader struct {
	r        io.Reader
	cur      *atomic.Int32
	peak     *atomic.Int32
	started  bool
}

func (s *slowReader) Read(p []byte) (int, error) {
	if !s.started {
		s.started = true
		n := s.cur.Add(1)
		for {
			pk := s.peak.Load()
			if n <= pk || s.peak.CompareAndSwap(pk, n) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	n, err := s.r.Read(p)
	if err == io.EOF {
		s.cur.Add(-1)
	}
	return n, err
}

func TestWorkerLimit(t *testing.T) {
	var cur, peak atomic.Int32
	rs := make([]io.Reader, 12)
	for i := range rs {
		rs[i] = &slowReader{r: strings.NewReader("x y z"), cur: &cur, peak: &peak}
	}
	got, err := TopWords(context.Background(), rs, 1, 3)
	if err != nil || len(got) != 1 || got[0] != (WordCount{"x", 12}) {
		t.Fatalf("got %v, %v", got, err)
	}
	if p := peak.Load(); p > 3 || p < 2 {
		t.Errorf("пик одновременных потоков %d, лимит 3", p)
	}
}

func TestBigAndNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	rs := make([]io.Reader, 64)
	for i := range rs {
		rs[i] = strings.NewReader(strings.Repeat("alpha beta gamma delta alpha beta alpha ", 20_000))
	}
	start := time.Now()
	got, err := TopWords(context.Background(), rs, 2, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []WordCount{{"alpha", 64 * 60_000}, {"beta", 64 * 40_000}}) {
		t.Errorf("got %v", got)
	}
	if time.Since(start) > 5*time.Second {
		t.Errorf("слишком медленно: %v", time.Since(start))
	}
	time.Sleep(50 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("утечка горутин: было %d, стало %d", before, after)
	}
}

func TestCancelMidway(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	pr, pw := io.Pipe() // читатель, который никогда не закончится сам
	defer pw.Close()
	go func() {
		for ctx.Err() == nil {
			pw.Write([]byte("word "))
			time.Sleep(time.Millisecond)
		}
		pw.Close()
	}()
	done := make(chan error, 1)
	go func() {
		_, err := TopWords(ctx, []io.Reader{pr}, 1, 1)
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("TopWords не реагирует на отмену контекста")
	}
}
@@ solution
package solution

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"unicode"
)

type WordCount struct {
	Word  string
	Count int
}

func countWords(ctx context.Context, r io.Reader) (map[string]int, error) {
	counts := map[string]int{}
	br := bufio.NewReader(r)
	var word strings.Builder
	flush := func() {
		if word.Len() > 0 {
			counts[word.String()]++
			word.Reset()
		}
	}
	for i := 0; ; i++ {
		if i%4096 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		c, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				flush()
				return counts, nil
			}
			return nil, fmt.Errorf("чтение: %w", err)
		}
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			word.WriteRune(unicode.ToLower(c))
		} else {
			flush()
		}
	}
}

func TopWords(ctx context.Context, readers []io.Reader, n, workers int) ([]WordCount, error) {
	if workers < 1 {
		workers = 1
	}
	parent := ctx
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan io.Reader)
	var (
		mu       sync.Mutex
		total    = map[string]int{}
		firstErr error
		wg       sync.WaitGroup
	)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range jobs {
				counts, err := countWords(ctx, r)
				mu.Lock()
				if err != nil {
					if firstErr == nil {
						firstErr = err
					}
					cancel()
				} else {
					for word, c := range counts {
						total[word] += c
					}
				}
				mu.Unlock()
			}
		}()
	}

feed:
	for _, r := range readers {
		select {
		case jobs <- r:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()

	if err := parent.Err(); err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if n <= 0 {
		return nil, nil
	}
	out := make([]WordCount, 0, len(total))
	for w, c := range total {
		out = append(out, WordCount{w, c})
	}
	slices.SortFunc(out, func(a, b WordCount) int {
		if a.Count != b.Count {
			return b.Count - a.Count
		}
		return strings.Compare(a.Word, b.Word)
	})
	if len(out) > n {
		out = out[:n]
	}
	return out, nil
}
@@ hint
Раздай воркерам потоки через канал `jobs`; каждый воркер считает в своей локальной map и после завершения потока сливает её в общую под мьютексом. Ошибка → `cancel()` внутреннего контекста. Сортировка: `slices.SortFunc` с составным компаратором.
