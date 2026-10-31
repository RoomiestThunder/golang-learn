package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// Проверяет весь контент курса: эталонные решения проходят тесты, шаблоны (starter) — нет,
// все запускаемые примеры из уроков (```go run) компилируются и завершаются без ошибки.
// Выборочно:  VERIFY=l05 go test -run TestContent -v
func TestContent(t *testing.T) {
	c, err := loadCourse(os.DirFS("."))
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRunner()
	if err != nil {
		t.Fatal(err)
	}
	only := os.Getenv("VERIFY")
	match := func(id string) bool { return only == "" || strings.HasPrefix(id, only) }

	for _, task := range c.Tasks {
		if !match(task.ID) {
			continue
		}
		t.Run("task/"+task.ID, func(t *testing.T) {
			t.Parallel()
			if res := r.Check(task, task.Solution); !res.OK {
				t.Errorf("эталонное решение не проходит:\n%s", res.Output)
			}
			if res := r.Check(task, task.Starter); res.OK {
				t.Errorf("шаблон уже проходит тесты — задание ничего не проверяет")
			}
		})
	}

	re := regexp.MustCompile("(?s)```go run\n(.*?)\n```")
	for _, l := range c.Lessons {
		if !match(l.ID) {
			continue
		}
		for i, m := range re.FindAllStringSubmatch(l.Body, -1) {
			code := m[1]
			t.Run("example/"+l.ID+"/"+string(rune('a'+i)), func(t *testing.T) {
				t.Parallel()
				if res := r.Play(code); !res.OK {
					t.Errorf("пример не запускается:\n%s\n--- код ---\n%s", res.Output, code)
				}
			})
		}
		if l.Body == "" || len(l.LC) == 0 {
			t.Errorf("%s: пустой текст или нет задач LeetCode", l.ID)
		}
	}

	if only == "" {
		for name, body := range c.Pages {
			for i, m := range re.FindAllStringSubmatch(body, -1) {
				code := m[1]
				t.Run("page/"+name+"/"+string(rune('a'+i)), func(t *testing.T) {
					t.Parallel()
					if res := r.Play(code); !res.OK {
						t.Errorf("пример не запускается:\n%s", res.Output)
					}
				})
			}
		}
		if len(c.Cards) < 70 {
			t.Errorf("карточек мало: %d", len(c.Cards))
		}
	}
}
