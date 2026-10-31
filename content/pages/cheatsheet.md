# Шпаргалка по Go

Всё, что нужно держать «под рукой». Примеры можно править и запускать.

## Команды

| Команда | Что делает |
|---|---|
| `go run .` | собрать и запустить |
| `go build -o app ./cmd/app` | собрать бинарь |
| `go test ./... -race -cover` | тесты с детектором гонок и покрытием |
| `go test -run TestX/sub -v -count=1` | один тест без кэша |
| `go test -bench=. -benchmem -run=^$` | бенчмарки |
| `go vet ./...` / `gofmt -l .` | статанализ / проверка форматирования |
| `go mod tidy` | синхронизировать зависимости |
| `go get pkg@v1.2.3` | добавить/обновить зависимость |
| `go doc fmt.Printf` | документация |
| `go build -gcflags=-m` | решения escape analysis |
| `go tool pprof -http=:9090 cpu.out` | профиль |

## Синтаксис за минуту

```go run
package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Shape interface{ Area() float64 }

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

var ErrEmpty = errors.New("пусто")

func first[T any](xs []T) (T, error) { // дженерик
	var zero T
	if len(xs) == 0 {
		return zero, ErrEmpty
	}
	return xs[0], nil
}

func main() {
	x := 42                            // :=, вывод типа
	var s []int                        // nil-слайс
	s = append(s, 1, 2, 3)
	m := map[string]int{"a": 1}
	if v, ok := m["b"]; !ok {          // comma ok
		fmt.Println("нет b", v)
	}
	for i, v := range s {              // range
		fmt.Print(i, ":", v, " ")
	}
	for i := range 3 {                 // range по числу (1.22)
		fmt.Print(i)
	}
	fmt.Println()

	var sh Shape = Rect{2, 3}
	switch v := sh.(type) {            // type switch
	case Rect:
		fmt.Println("rect", v.W, v.Area())
	}

	keys := []string{"b", "a"}
	sort.Strings(keys)
	fmt.Println(strings.Join(keys, ","), x)

	if _, err := first([]int{}); errors.Is(err, ErrEmpty) {
		fmt.Println("ошибка:", err)
	}

	defer fmt.Println("defer: выполнится при выходе")
}
```

## Типы и значения

| Тип | Zero value | Заметки |
|---|---|---|
| `int, float64` | `0` | `int` 64-бит на 64-бит. платформе |
| `string` | `""` | неизменяема, UTF-8, `len` — в байтах |
| `bool` | `false` | |
| `[]T` | `nil` | дескриптор (ptr, len, cap) |
| `map[K]V` | `nil` | запись в nil-map — panic |
| `*T`, `chan`, `func`, `interface` | `nil` | |
| `struct` | поля в zero value | сравнима, если поля сравнимы |

## Форматирование (fmt)

| Глагол | Значение |
|---|---|
| `%v` `%+v` `%#v` | значение / с полями / как в коде Go |
| `%T` | тип |
| `%d` `%b` `%x` | целое / двоичное / hex |
| `%s` `%q` | строка / в кавычках |
| `%f` `%.2f` `%e` | вещественное |
| `%t` `%p` `%c` `%U` | bool / указатель / символ / Unicode |
| `%5d` `%-5s` `%05d` | ширина / выравнивание / нули |
| `%w` | обернуть ошибку (только `Errorf`) |

## Слайсы и map: быстрые рецепты

```go
s = append(s, x)                    // добавить
s = append(s[:i], s[i+1:]...)       // удалить i (с порядком); slices.Delete(s, i, i+1)
copy(dst, src)                      // копирование (min длин)
c := slices.Clone(s)                // независимая копия
slices.Sort(s); slices.Contains(s, x); slices.Index(s, x); slices.Reverse(s)
slices.SortFunc(s, func(a, b T) int { return cmp.Compare(a.X, b.X) })
keys := slices.Sorted(maps.Keys(m)) // ключи по порядку
delete(m, k); clear(m)              // удаление / очистка
set := map[T]struct{}{}             // множество
```

## Ошибки

```go
return fmt.Errorf("загрузка %q: %w", name, err) // обернуть
errors.Is(err, ErrNotFound)                      // по значению в цепочке
var ve *ValidationError; errors.As(err, &ve)     // по типу в цепочке
errors.Join(err1, err2)                          // несколько ошибок
```

## Конкурентность: шаблоны

```go
// ждать группу
var wg sync.WaitGroup
for _, x := range xs { wg.Add(1); go func() { defer wg.Done(); work(x) }() }
wg.Wait()

// защита состояния
mu.Lock(); defer mu.Unlock()

// отмена
ctx, cancel := context.WithTimeout(ctx, 2*time.Second); defer cancel()
select { case <-ctx.Done(): return ctx.Err(); case v := <-ch: use(v) }

// worker pool: jobs/results каналы, N воркеров, close(jobs), wg.Wait(), close(results)
// семафор: sem := make(chan struct{}, N); sem <- struct{}{}; defer func() { <-sem }()
// один раз: var once sync.Once; once.Do(init)
```

Правила: канал закрывает отправитель; у каждой горутины есть условие остановки; `ctx` — первый параметр; не копируй `sync.Mutex`.

## HTTP: минимум

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
})
srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
// graceful: srv.Shutdown(ctx)

client := &http.Client{Timeout: 10 * time.Second}
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req); defer resp.Body.Close()
```

## База данных: минимум

```go
db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id).Scan(&name) // sql.ErrNoRows
rows, _ := db.QueryContext(ctx, q, args...); defer rows.Close(); for rows.Next() { rows.Scan(...) }; rows.Err()
tx, _ := db.BeginTx(ctx, nil); defer tx.Rollback(); ...; tx.Commit()
```

## Тесты

```go
func TestX(t *testing.T) {
	tests := []struct{ name string; in, want int }{ {"zero", 0, 0} }
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := F(tt.in); got != tt.want { t.Errorf("F(%d) = %d, want %d", tt.in, got, tt.want) }
		})
	}
}
func BenchmarkX(b *testing.B) { for b.Loop() { F(1) } }
```

## Big-O и подсказки по ограничениям

| n | Допустимая сложность |
|---|---|
| ≤ 20 | O(2ⁿ), O(n!) — перебор |
| ≤ 500 | O(n³) |
| ≤ 5 000 | O(n²) |
| ≤ 10⁶ | O(n log n) |
| ≥ 10⁷ | O(n), O(log n) |

## Шаблоны алгоритмов

| Признак | Шаблон |
|---|---|
| «пара/тройка с суммой», «видел ли?» | хэш-таблица, два указателя |
| непрерывный подмассив/подстрока | скользящее окно, префиксные суммы |
| отсортировано / монотонный предикат | бинарный поиск |
| следующий больший/меньший | монотонный стек |
| k наибольших/наименьших, слияние | куча |
| дерево | рекурсия DFS, BFS по уровням |
| кратчайший путь без весов | BFS |
| порядок с зависимостями | топосортировка |
| все комбинации/перестановки | backtracking |
| оптимум с перекрывающимися подзадачами | динамическое программирование |

## Частые ловушки

- `append` к срезу с запасом `cap` пишет в общий массив — клонируй.
- Запись в nil-map → panic; чтение безопасно.
- Typed nil в интерфейсе: `var e *MyErr; return e` ≠ `nil`.
- `for _, v := range items { v.X = 1 }` меняет копию.
- `defer` в цикле накапливается до выхода из функции.
- `err :=` во внутреннем блоке затеняет внешнюю `err`.
- `resp.Body` нужно закрывать; `http.Get` без таймаута может висеть вечно.
- `time.Now()` в глубине кода → недетерминированные тесты.
- `os.Exit` не выполняет `defer`.
- Горутина без условия остановки → утечка.
