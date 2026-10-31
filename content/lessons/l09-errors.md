---
id: l09
week: 2
day: 4
minutes: 120
title: Обработка ошибок
goals: Идиомы error, оборачивание %w, errors.Is/As/Join, sentinel и пользовательские ошибки, чего избегать
lc: valid-parentheses | Valid Parentheses | Easy | Стек; при закрывающей скобке сверяем с вершиной
lc: evaluate-reverse-polish-notation | Evaluate Reverse Polish Notation | Medium | Стек чисел; операнды извлекаются в обратном порядке!
lc: basic-calculator-ii | Basic Calculator II | Medium | Стек с отложенным применением * и /
---

## Философия

В Go ошибка — **обычное значение**, а не исключение. Функция, которая может не получиться, возвращает `error` последним результатом, вызывающий **обязан** решить, что делать.

```go
type error interface {
	Error() string
}
```

```go
f, err := os.Open(name)
if err != nil {
	return fmt.Errorf("открытие конфига: %w", err) // добавили контекст, сохранили причину
}
defer f.Close()
```

Многословно? Да, зато поток управления виден прямо в коде, нет «невидимых» выходов.

## Создание ошибок

```go run
package main

import (
	"errors"
	"fmt"
)

// sentinel-ошибка: пакетная переменная, сравнивается по значению
var ErrNotFound = errors.New("не найдено")

// пользовательский тип несёт данные
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

func find(id int) error {
	if id < 0 {
		return &ValidationError{Field: "id", Msg: "должен быть неотрицательным"}
	}
	if id > 10 {
		return fmt.Errorf("поиск id=%d: %w", id, ErrNotFound) // %w оборачивает
	}
	return nil
}

func main() {
	err := find(42)
	fmt.Println(err)
	fmt.Println(err == ErrNotFound)          // false: ошибка обёрнута
	fmt.Println(errors.Is(err, ErrNotFound)) // true: Is разворачивает цепочку

	err = find(-1)
	var ve *ValidationError
	if errors.As(err, &ve) { // As ищет в цепочке ошибку нужного ТИПА и записывает в ve
		fmt.Println("поле:", ve.Field, "|", ve.Msg)
	}
}
```

### Is, As, Unwrap, Join

- `errors.Is(err, target)` — есть ли в цепочке ошибка, равная `target` (или имеющая метод `Is`).
- `errors.As(err, &target)` — есть ли в цепочке ошибка типа `target`; **target — указатель** на переменную нужного типа.
- `errors.Unwrap(err)` — на один уровень вглубь.
- `errors.Join(errs...)` (Go 1.20) — объединяет несколько ошибок; `Is/As` работают по всем.
- Один `fmt.Errorf` может содержать несколько `%w` (Go 1.20).

```go run
package main

import (
	"errors"
	"fmt"
)

var (
	ErrEmpty = errors.New("пустое имя")
	ErrShort = errors.New("слишком короткий пароль")
)

func validate(name, pass string) error {
	var errs []error
	if name == "" {
		errs = append(errs, ErrEmpty)
	}
	if len(pass) < 8 {
		errs = append(errs, ErrShort)
	}
	return errors.Join(errs...) // nil, если errs пуст
}

func main() {
	err := validate("", "123")
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrEmpty), errors.Is(err, ErrShort))
	fmt.Println(validate("ann", "12345678") == nil)
}
```

## Как правильно работать с ошибками

1. **Добавляй контекст при передаче наверх**: `fmt.Errorf("загрузка профиля %d: %w", id, err)`. Тексты — со строчной буквы, без точки на конце и без «failed to» повторов (иначе будет «failed to X: failed to Y: ...»).
2. **Обрабатывай ошибку один раз**: либо логируй, либо возвращай. «Залогировал и вернул» приводит к дублям в логах. Логирование — на верхнем уровне (handler, main).
3. **Не игнорируй ошибки** (`_ = f.Close()` допустимо осознанно, но `defer f.Close()` для записи — ошибку Close терять нельзя).
4. **Sentinel-ошибки** — часть публичного API: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`. Сравнивай через `errors.Is`, а не `==`.
5. **Типизированные ошибки** — когда вызывающему нужны данные (код статуса, поле). Проверяй через `errors.As`.
6. Не используй `panic` для обычных ошибок. Паника — для нарушенных инвариантов и ошибок программиста; в библиотеках не выпускай панику за границы пакета.
7. Не оборачивай ошибки, если не хочешь делать причину частью API (используй `%v` вместо `%w` — скрываешь детали реализации).

## Паттерны

### Defer + именованная ошибка

```go
func writeFile(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil { // не затираем первую ошибку
			err = cerr
		}
	}()
	_, err = f.Write(data)
	return err
}
```

### Накопление ошибок

Валидация — `errors.Join`; параллельные задачи — `errgroup` (неделя 3); «ошибка в итераторе» — отдельный метод `Err()` как у `bufio.Scanner` и `sql.Rows`.

### Ошибки, которые умеют «временность»

```go
type temporary interface{ Temporary() bool }
```

Так пакет `net` сообщает, имеет ли смысл повторить. Идея: проверяй **поведение**, а не конкретный тип.

### Ошибка и HTTP

На границе API преобразуй доменные ошибки в статусы: `errors.Is(err, ErrNotFound)` → 404, `errors.As(err, &ve)` → 400, остальное → 500 с общим текстом (детали — в лог). Это тема недели 5.

## Типичные ошибки (и вопросы собеседований)

- Сравнение обёрнутой ошибки через `==`.
- Возврат nil-указателя типа `*MyErr` как `error` (см. урок про интерфейсы).
- `if err != nil { return err }` без контекста — потом невозможно понять, где упало.
- `errors.As` с не-указателем во втором аргументе → паника.
- Ошибки в горутинах теряются, если их некуда отправить: собирай через канал или `errgroup`.

## Проверь себя

<details><summary>В чём разница <code>%w</code> и <code>%v</code> в <code>fmt.Errorf</code>?</summary><code>%w</code> сохраняет исходную ошибку в цепочке (её найдут <code>Is/As/Unwrap</code>), <code>%v</code> лишь вставляет текст.</details>

<details><summary>Когда использовать <code>errors.As</code>, а когда <code>errors.Is</code>?</summary>Is — сравнить с конкретным значением (sentinel), As — достать ошибку определённого типа вместе с её данными.</details>

<details><summary>Почему не стоит и логировать, и возвращать ошибку?</summary>Одна ошибка попадёт в логи несколько раз на разных уровнях; решение о логировании принимает верхний уровень.</details>

<details><summary>Когда допустима panic?</summary>При нарушении инвариантов/ошибке программиста при старте (<code>regexp.MustCompile</code>), недостижимом коде; не как обычный механизм ошибок.</details>
