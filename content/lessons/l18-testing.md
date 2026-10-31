---
id: l18
week: 4
day: 3
minutes: 120
title: Тестирование
goals: Table-driven тесты, подтесты, бенчмарки, fuzzing, покрытие, моки через интерфейсы, httptest
lc: find-the-index-of-the-first-occurrence-in-a-string | Find the Index of the First Occurrence in a String | Easy | Напиши наивно, затем придумай набор граничных тестов
lc: string-to-integer-atoi | String to Integer (atoi) | Medium | Задача про граничные случаи: пробелы, знак, переполнение
lc: add-binary | Add Binary | Easy | Сложение столбиком с переносом; строки разной длины
---

Тестирование встроено в язык и toolchain: никаких фреймворков для начала не нужно.

## Базовые правила

- Файл `xxx_test.go` рядом с кодом; пакет тот же (`package foo`, white-box) или `package foo_test` (black-box, только публичный API).
- Функции `func TestXxx(t *testing.T)`; запуск: `go test ./...`, `-v` подробно, `-run Regexp` по имени, `-count=1` без кэша, `-race`, `-cover`, `-timeout 30s`, `-short`.
- `t.Errorf` — сообщить об ошибке и **продолжить**; `t.Fatalf` — сообщить и **остановить** тест; `t.Skip` — пропустить.
- Формат сообщения: `got ..., want ...`. Не используй `assert`-библиотеки без нужды: стандартного `if got != want` достаточно.

```go
func TestAbs(t *testing.T) {
	got := Abs(-3)
	if got != 3 {
		t.Errorf("Abs(-3) = %d, want 3", got)
	}
}
```

## Table-driven тесты и подтесты

Каноническая форма: таблица входов/ожиданий + `t.Run` для каждого случая (даёт имена в выводе, можно запускать выборочно `-run TestAbs/negative`).

```go
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
		{"min int edge", math.MinInt64 + 1, math.MaxInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // подтесты параллельно (с Go 1.22 переменная tt безопасна)
			if got := Abs(tt.in); got != tt.want {
				t.Errorf("Abs(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
```

Что тестировать: обычные значения, **границы** (0, пусто, nil, max/min), ошибочные входы, повторные вызовы, конкурентность.

## Помощники

```go
func newTestServer(t *testing.T) *Server {
	t.Helper() // ошибки будут указывать на строку вызова помощника, а не внутрь него
	dir := t.TempDir()   // автоудаляемая временная папка
	t.Cleanup(func() { /* закрыть ресурсы после теста */ })
	t.Setenv("APP_ENV", "test") // переменная окружения на время теста
	...
}
```

Фикстуры (файлы с данными) кладут в папку `testdata/` — её игнорирует `go build`. Для сложных результатов используют **golden files** (сравнение с эталонным файлом, обновляемым флагом `-update`).

`TestMain(m *testing.M)` — общая настройка/очистка пакета (поднять БД, `os.Exit(m.Run())`).

## Тестирование через интерфейсы (моки и фейки)

```go
type Notifier interface{ Send(to, msg string) error }

type fakeNotifier struct {
	sent []string
	err  error
}
func (f *fakeNotifier) Send(to, msg string) error { f.sent = append(f.sent, to+":"+msg); return f.err }

func TestSignup(t *testing.T) {
	n := &fakeNotifier{}
	svc := NewService(n)
	if err := svc.Signup("a@x.io"); err != nil { t.Fatal(err) }
	if len(n.sent) != 1 { t.Errorf("sent = %v", n.sent) }
}
```

Предпочитай **фейки** (простые рабочие реализации в памяти) генерируемым моками с «ожиданиями вызовов» — тесты меньше привязаны к реализации. Генераторы: `mockgen`, `moq`; ассерты и моки: `testify` — популярны, но не обязательны.

## Бенчмарки

```go
func BenchmarkConcat(b *testing.B) {
	b.ReportAllocs()           // показывать аллокации
	for b.Loop() {             // Go 1.24+; раньше: for i := 0; i < b.N; i++
		_ = strings.Repeat("x", 100)
	}
}
```

Запуск: `go test -bench=. -benchmem -run=^$`; сравнение версий — `benchstat`. Профили: `-cpuprofile cpu.out -memprofile mem.out` → `go tool pprof` (неделя 6).

## Fuzzing

Автоматический поиск входов, ломающих код (паники, нарушение инвариантов):

```go
func FuzzReverse(f *testing.F) {
	f.Add("hello")                       // seed corpus
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) { return }
		if rev2 := Reverse(Reverse(s)); rev2 != s {
			t.Errorf("двойной реверс %q → %q", s, rev2)
		}
	})
}
// go test -fuzz=FuzzReverse -fuzztime=30s
```

Найденные падения сохраняются в `testdata/fuzz/` и становятся обычными регрессионными тестами.

## Example-тесты

```go
func ExampleGreet() {
	fmt.Println(Greet("Go"))
	// Output: Привет, Go!
}
```

Попадают в документацию и **проверяются** при `go test`.

## HTTP-тесты: httptest

```go
func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewMux().ServeHTTP(rec, req) // вызываем обработчик напрямую, без сети

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "ok" {
		t.Errorf("body = %q", got)
	}
}
```

`httptest.NewServer(handler)` поднимает настоящий сервер на случайном порту — для тестов HTTP-клиентов (в неделе 5 ты этим воспользуешься).

## Покрытие

```bash
go test -cover ./...
go test -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

Покрытие — индикатор, а не цель: 100% строк не гарантируют отсутствие багов. Важнее тестировать поведение и риски.

## Хорошие практики

1. Тесты **быстрые, независимые, детерминированные**. Никаких `time.Sleep` для ожидания — используй каналы/условия или внедряемые часы (в Go 1.25 для этого есть `testing/synctest`).
2. Не тестируй приватные детали — тестируй поведение.
3. Тесты на ошибки: проверяй `errors.Is/As`, а не текст.
4. Запускай `go test -race ./...` в CI.
5. Конкурентный код тестируй многократно: `go test -count=100 -race`.
6. Интеграционные тесты (БД, сеть) отделяй тегом сборки (`//go:build integration`) или `testing.Short()`; БД поднимай в Docker (`testcontainers-go`).

## Практика этого урока

В задании ты получишь функции **с ошибками** и таблицу тестов: читай сообщения о падении, находи баги. Это реальный рабочий цикл: тест красный → понять почему → исправить → зелёный.

## Проверь себя

<details><summary>Разница между <code>t.Error</code> и <code>t.Fatal</code>?</summary>Error помечает тест проваленным и продолжает выполнение, Fatal — и завершает текущий тест немедленно.</details>

<details><summary>Зачем нужен <code>t.Helper()</code>?</summary>Чтобы сообщение об ошибке указывало на строку в тесте, откуда вызван помощник.</details>

<details><summary>Как запустить только один подтест?</summary><code>go test -run 'TestAbs/negative'</code>.</details>

<details><summary>Что такое fuzzing и когда он полезен?</summary>Автоматическая генерация входов для поиска падений/нарушений инвариантов; полезен для парсеров, кодеков, валидаторов.</details>
