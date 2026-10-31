---
title: Проект: хранилище и валидация URL
level: medium
---
@@ statement
Ядро бизнес-логики сокращателя. `Encode` (base62) из предыдущего задания уже включён в шаблон — используй его.

**`ValidateURL(raw string) (string, error)`** — проверяет и нормализует ссылку:
- обрезать пробелы по краям;
- схема только `http` или `https` (любой регистр → в нижнем), непустой хост, длина не более 2048 символов;
- нормализация: схема и **хост** — в нижний регистр; остальное без изменений;
- иначе — ошибка, для которой `errors.Is(err, ErrInvalidURL)`. Недопустимы: пустая строка, `example.com` (без схемы), `ftp://x.com`, `http://`, `javascript:alert(1)`.

**`MemoryStore`** — потокобезопасное хранилище в памяти:
- `NewMemoryStore() *MemoryStore` — счётчик ID начинается со `125`;
- `Shorten(rawURL string) (string, error)` — валидирует URL, **дедуплицирует** (один и тот же нормализованный URL → всегда один и тот же код), новый URL получает следующий ID → `Encode(id)`;
- `Resolve(code string) (string, error)` — возвращает URL и **увеличивает счётчик кликов**; неизвестный код → `ErrNotFound`;
- `Stats(code string) (Stats, error)` — `Stats{Code, URL string; Clicks int64}` без увеличения счётчика; `ErrNotFound` для неизвестного кода.
@@ starter
package solution

import "errors"

var (
	ErrNotFound   = errors.New("ссылка не найдена")
	ErrInvalidURL = errors.New("некорректный URL")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Encode — готовая реализация base62 из предыдущего задания.
func Encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [11]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}

type Stats struct {
	Code   string
	URL    string
	Clicks int64
}

func ValidateURL(raw string) (string, error) {
	return raw, nil
}

type MemoryStore struct {
	// TODO
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (s *MemoryStore) Shorten(rawURL string) (string, error) { return "", nil }

func (s *MemoryStore) Resolve(code string) (string, error) { return "", ErrNotFound }

func (s *MemoryStore) Stats(code string) (Stats, error) { return Stats{}, ErrNotFound }
@@ tests
package solution

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestValidateURL(t *testing.T) {
	good := map[string]string{
		"https://example.com":                  "https://example.com",
		"  HTTPS://Example.COM/Path?q=1#frag ": "https://example.com/Path?q=1#frag",
		"http://localhost:8080/a/b":            "http://localhost:8080/a/b",
		"HTTP://Sub.Domain.org":                "http://sub.domain.org",
	}
	for in, want := range good {
		got, err := ValidateURL(in)
		if err != nil || got != want {
			t.Errorf("ValidateURL(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	long := "https://example.com/" + strings.Repeat("a", 2100)
	for _, bad := range []string{"", "   ", "example.com", "ftp://x.com/file", "http://", "https:///path", "javascript:alert(1)", "mailto:a@b.c", long} {
		if _, err := ValidateURL(bad); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("ValidateURL(%q): err = %v, want ErrInvalidURL", bad, err)
		}
	}
}

func TestShortenResolve(t *testing.T) {
	s := NewMemoryStore()
	c1, err := s.Shorten("https://example.com/a")
	if err != nil || c1 != Encode(125) {
		t.Fatalf("первый код %q, %v; ожидался %q", c1, err, Encode(125))
	}
	c2, _ := s.Shorten("https://example.com/b")
	if c2 != Encode(126) {
		t.Errorf("второй код = %q", c2)
	}
	// дедупликация (после нормализации)
	if again, _ := s.Shorten("  HTTPS://EXAMPLE.com/a "); again != c1 {
		t.Errorf("дубликат дал новый код %q вместо %q", again, c1)
	}
	if next, _ := s.Shorten("https://example.com/c"); next != Encode(127) {
		t.Errorf("повтор не должен расходовать ID: %q", next)
	}
	if _, err := s.Shorten("nonsense"); !errors.Is(err, ErrInvalidURL) {
		t.Errorf("невалидный URL: %v", err)
	}

	url, err := s.Resolve(c1)
	if err != nil || url != "https://example.com/a" {
		t.Errorf("Resolve = %q, %v", url, err)
	}
	if _, err := s.Resolve("zzzz"); !errors.Is(err, ErrNotFound) {
		t.Errorf("неизвестный код: %v", err)
	}
}

func TestStatsAndClicks(t *testing.T) {
	s := NewMemoryStore()
	code, _ := s.Shorten("https://go.dev")
	for i := 0; i < 3; i++ {
		s.Resolve(code)
	}
	st, err := s.Stats(code)
	if err != nil || st != (Stats{code, "https://go.dev", 3}) {
		t.Errorf("Stats = %+v, %v", st, err)
	}
	s.Stats(code) // не увеличивает счётчик
	if st, _ := s.Stats(code); st.Clicks != 3 {
		t.Errorf("Stats не должен считать клики: %d", st.Clicks)
	}
	if _, err := s.Stats("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Stats(неизвестный): %v", err)
	}
}

func TestConcurrent(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup
	codes := make([]string, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := s.Shorten(fmt.Sprintf("https://example.com/%d", i%20)) // 20 разных URL × 5 повторов
			if err != nil {
				t.Error(err)
			}
			codes[i] = c
		}()
	}
	wg.Wait()
	uniq := map[string]bool{}
	for _, c := range codes {
		uniq[c] = true
	}
	if len(uniq) != 20 {
		t.Errorf("уникальных кодов %d, ожидалось 20", len(uniq))
	}

	code := codes[0]
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				s.Resolve(code)
			}
		}()
	}
	wg.Wait()
	if st, _ := s.Stats(code); st.Clicks != 5000 {
		t.Errorf("кликов %d, ожидалось 5000 (потерянные обновления)", st.Clicks)
	}
}
@@ solution
package solution

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var (
	ErrNotFound   = errors.New("ссылка не найдена")
	ErrInvalidURL = errors.New("некорректный URL")
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [11]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}

type Stats struct {
	Code   string
	URL    string
	Clicks int64
}

func ValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return "", fmt.Errorf("%w: пустой или слишком длинный", ErrInvalidURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	scheme := strings.ToLower(u.Scheme)
	if (scheme != "http" && scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("%w: нужна ссылка http(s) с хостом", ErrInvalidURL)
	}
	u.Scheme = scheme
	u.Host = strings.ToLower(u.Host)
	return u.String(), nil
}

type entry struct {
	url    string
	clicks int64
}

type MemoryStore struct {
	mu     sync.Mutex
	nextID uint64
	byCode map[string]*entry
	byURL  map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{nextID: 125, byCode: map[string]*entry{}, byURL: map[string]string{}}
}

func (s *MemoryStore) Shorten(rawURL string) (string, error) {
	norm, err := ValidateURL(rawURL)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if code, ok := s.byURL[norm]; ok {
		return code, nil
	}
	code := Encode(s.nextID)
	s.nextID++
	s.byCode[code] = &entry{url: norm}
	s.byURL[norm] = code
	return code, nil
}

func (s *MemoryStore) Resolve(code string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.byCode[code]
	if !ok {
		return "", ErrNotFound
	}
	e.clicks++
	return e.url, nil
}

func (s *MemoryStore) Stats(code string) (Stats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.byCode[code]
	if !ok {
		return Stats{}, ErrNotFound
	}
	return Stats{Code: code, URL: e.url, Clicks: e.clicks}, nil
}
@@ hint
`url.Parse` + проверка `Scheme` и `Host`. Два map (код → запись и URL → код) под одним мьютексом. Нормализованный URL — ключ дедупликации.
