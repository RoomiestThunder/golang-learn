---
id: l05
week: 1
day: 5
minutes: 120
title: Массивы, слайсы и map
goals: Понять устройство слайса (ptr/len/cap), подводные камни append, работу с map
lc: two-sum | Two Sum | Easy | map значение→индекс; ищем complement = target - x за один проход
lc: contains-duplicate | Contains Duplicate | Easy | map[int]struct{} как множество
lc: move-zeroes | Move Zeroes | Easy | Два указателя: write и read; всё на месте
lc: remove-duplicates-from-sorted-array | Remove Duplicates from Sorted Array | Easy | Два указателя по отсортированному массиву
---

## Массивы: значения фиксированной длины

Длина — часть типа: `[3]int` и `[4]int` — разные типы. Массив при присваивании и передаче в функцию **копируется целиком**. Напрямую используются редко (хэши `[32]byte`, ключи map, матрицы фиксированного размера).

```go run
package main

import "fmt"

func main() {
	a := [3]int{1, 2, 3}
	b := a // копия
	b[0] = 99
	fmt.Println(a, b, len(a))

	c := [...]string{"x", "y"} // длину посчитает компилятор
	d := [5]int{1: 10, 3: 30}  // инициализация по индексам
	fmt.Println(c, d, a == [3]int{1, 2, 3}) // массивы сравнимы через ==
}
```

## Слайсы

Слайс — это **дескриптор из трёх слов**: указатель на массив, `len` (длина) и `cap` (ёмкость). Сам слайс копируется дёшево, но копия **разделяет те же данные**.

```
slice {ptr ─────┐, len: 3, cap: 5}
				▼
		[ 1 | 2 | 3 | _ | _ ]   ← базовый массив
```

```go run
package main

import "fmt"

func main() {
	s := make([]int, 3, 5) // тип, len, cap
	fmt.Println(s, len(s), cap(s))

	s = append(s, 1, 2) // влезло в cap: тот же массив
	s = append(s, 3)    // не влезло: выделен новый массив (обычно x2), данные скопированы
	fmt.Println(s, len(s), cap(s))

	lit := []string{"a", "b", "c", "d"}
	part := lit[1:3] // [low:high) — половинный интервал; len=2, cap=3 (до конца базового массива)
	fmt.Println(part, len(part), cap(part))

	part[0] = "ИЗМЕНЕНО" // меняет и lit — общая память!
	fmt.Println(lit)

	var nilS []int           // nil-слайс: len=0, cap=0, безопасен
	empty := []int{}         // пустой, но не nil
	fmt.Println(nilS == nil, empty == nil, len(nilS))
}
```

### Главный подводный камень: append и разделяемая память

```go run
package main

import "fmt"

func main() {
	a := make([]int, 3, 10)
	b := append(a, 100) // в cap помещается → пишем в ОБЩИЙ массив
	c := append(a, 200) // перезаписали тот же элемент
	fmt.Println(b[3], c[3]) // 200 200 — сюрприз!

	x := []int{1, 2, 3} // cap == len
	y := append(x, 4)   // не влезает → новый массив
	y[0] = 99
	fmt.Println(x[0], y[0]) // 1 99 — теперь независимы

	// 3-индексный срез ограничивает cap и защищает от затирания соседей
	safe := a[0:3:3]
	z := append(safe, 5) // всегда копирование
	fmt.Println(len(z), cap(safe))
}
```

Правила, которые спасают от багов:

1. **Всегда присваивай результат**: `s = append(s, x)`.
2. Не полагайся на то, разделяют ли два слайса память, — если нужна независимая копия, делай её явно: `dst := slices.Clone(src)` или `copy`.
3. Слайс, вырезанный из большого массива, держит **весь** массив в памяти (утечка) — копируй нужный кусок.
4. Если размер известен заранее, выдели ёмкость: `make([]T, 0, n)` — меньше аллокаций.

### Копирование, удаление, вставка

```go run
package main

import (
	"fmt"
	"slices"
)

func main() {
	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, 3)
	n := copy(dst, src) // копирует min(len(dst), len(src)) элементов
	fmt.Println(n, dst)

	// удаление i-го с сохранением порядка
	i := 1
	s := slices.Clone(src)
	s = append(s[:i], s[i+1:]...)
	fmt.Println(s)

	// то же стандартной функцией (Go 1.21+) и вставка
	s = slices.Delete(slices.Clone(src), 1, 2)
	s = slices.Insert(s, 1, 100, 200)
	fmt.Println(s)

	// быстрый вариант без сохранения порядка
	t := slices.Clone(src)
	t[i] = t[len(t)-1]
	t = t[:len(t)-1]
	fmt.Println(t)

	// пакет slices: Contains, Index, Sort, Reverse, Max, Min, BinarySearch, Equal...
	u := []int{3, 1, 2}
	slices.Sort(u)
	idx, found := slices.BinarySearch(u, 2)
	fmt.Println(u, slices.Contains(u, 3), idx, found, slices.Max(u))
}
```

### Двумерные слайсы

```go
grid := make([][]int, rows)
for i := range grid {
	grid[i] = make([]int, cols) // каждую строку выделяем отдельно
}
```

## map

Хэш-таблица `map[K]V`. Ключ — сравнимый тип (числа, строки, bool, массивы, структуры из сравнимых полей, указатели, интерфейсы; **не** слайсы, map, функции).

```go run
package main

import (
	"fmt"
	"sort"
)

func main() {
	ages := map[string]int{"alice": 30, "bob": 25}
	ages["carol"] = 41

	v, ok := ages["dave"] // «comma ok»: отличаем отсутствие от нулевого значения
	fmt.Println(v, ok)
	fmt.Println(ages["нет такого"]) // отсутствующий ключ → zero value, не ошибка

	delete(ages, "bob") // удаление; удалять отсутствующий ключ безопасно
	fmt.Println(len(ages))

	// порядок обхода НЕ определён — для стабильного вывода сортируем ключи
	keys := make([]string, 0, len(ages))
	for k := range ages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k, ages[k])
	}

	// множество: map[T]struct{} — struct{} занимает 0 байт
	seen := map[int]struct{}{}
	seen[1] = struct{}{}
	_, has := seen[1]
	fmt.Println(has)

	// map значений-слайсов: append работает на zero value (nil-слайс)
	groups := map[int][]string{}
	groups[3] = append(groups[3], "кот", "пёс")
	fmt.Println(groups)

	// map структур: нельзя менять поле «на месте» → m[k].f = ... не компилируется
	type P struct{ N int }
	pm := map[string]P{"a": {1}}
	p := pm["a"]
	p.N++
	pm["a"] = p
	fmt.Println(pm)
}
```

Что важно знать:

- **nil-map**: чтение работает, **запись паникует**. Создавай через `make(map[K]V)` или литерал.
- map **небезопасна для конкурентного доступа** (fatal error: concurrent map writes) — нужен `sync.Mutex`/`sync.Map` (неделя 3).
- Элементы map не адресуемы (`&m[k]` нельзя); при росте таблицы элементы переезжают.
- Значения в map сохраняют «горячие» данные; для редких ключей и больших `V` храни указатели.
- Новые builtin: `clear(m)` очищает map/слайс, `min`/`max` для упорядоченных типов; пакет `maps` (`maps.Keys` возвращает итератор, `maps.Clone`, `maps.DeleteFunc`).

> [!INTERVIEW] «Как устроена map в Go?» — хэш-таблица с бакетами (в Go 1.24 — Swiss Tables); итерация рандомизируется, чтобы не полагаться на порядок; рост — постепенная эвакуация (в старой реализации). Детали — на неделе 6.

## Проверь себя

<details><summary>Чем отличается массив от слайса?</summary>Массив — значение фиксированной длины (копируется), слайс — ссылочный дескриптор (ptr, len, cap) на массив.</details>

<details><summary>Что напечатает <code>a := []int{1,2,3}; b := a[:2]; b = append(b, 99); fmt.Println(a)</code>?</summary><code>[1 2 99]</code> — append поместился в cap и перезаписал общий массив.</details>

<details><summary>Что будет при записи в nil-map? При чтении?</summary>Запись — panic; чтение возвращает zero value.</details>

<details><summary>Как безопасно итерироваться по map в детерминированном порядке?</summary>Собрать ключи в слайс, отсортировать, обходить слайс.</details>

<details><summary>Чем <code>len</code> отличается от <code>cap</code> у слайса?</summary>len — число доступных элементов, cap — сколько поместится до реаллокации (от начала слайса до конца базового массива).</details>
