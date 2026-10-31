package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

// LC — задача на LeetCode, привязанная к уроку.
type LC struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Level string `json:"level"`
	Hint  string `json:"hint"`
}

// Task — практическое задание с автопроверкой через go test.
type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Lesson    string `json:"lesson"`
	Level     string `json:"level"`
	Statement string `json:"statement"`
	Starter   string `json:"starter"`
	Tests     string `json:"tests"`
	Solution  string `json:"-"`
	Hint      string `json:"hint"`
}

type Lesson struct {
	ID      string   `json:"id"`
	Week    int      `json:"week"`
	Day     int      `json:"day"`
	Minutes int      `json:"minutes"`
	Title   string   `json:"title"`
	Goals   string   `json:"goals"`
	Body    string   `json:"body,omitempty"`
	Tasks   []string `json:"tasks"`
	LC      []LC     `json:"lc"`
}

type Week struct {
	Week  int    `json:"week"`
	Title string `json:"title"`
	Goal  string `json:"goal"`
}

type Card struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
	Q   string `json:"q"`
	A   string `json:"a"`
}

type Course struct {
	Weeks   []Week
	Lessons []*Lesson
	Tasks   map[string]*Task
	Cards   []Card
	Pages   map[string]string
	byID    map[string]*Lesson
}

func parseFront(src string) (map[string][]string, string) {
	meta := map[string][]string{}
	src = strings.ReplaceAll(src, "\r\n", "\n")
	if !strings.HasPrefix(src, "---\n") {
		return meta, src
	}
	rest := src[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return meta, src
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		meta[k] = append(meta[k], strings.TrimSpace(v))
	}
	return meta, rest[end+5:]
}

func first(m map[string][]string, k string) string {
	if v := m[k]; len(v) > 0 {
		return v[0]
	}
	return ""
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

func loadCourse(fsys fs.FS) (*Course, error) {
	c := &Course{Tasks: map[string]*Task{}, Pages: map[string]string{}, byID: map[string]*Lesson{}}

	raw, err := fs.ReadFile(fsys, "content/weeks.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &c.Weeks); err != nil {
		return nil, fmt.Errorf("weeks.json: %w", err)
	}

	files, _ := fs.Glob(fsys, "content/lessons/*.md")
	sort.Strings(files)
	for _, f := range files {
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		meta, body := parseFront(string(b))
		l := &Lesson{
			ID: first(meta, "id"), Week: atoi(first(meta, "week")), Day: atoi(first(meta, "day")),
			Minutes: atoi(first(meta, "minutes")), Title: first(meta, "title"), Goals: first(meta, "goals"),
			Body: strings.TrimSpace(body), Tasks: []string{}, LC: []LC{},
		}
		if l.ID == "" {
			return nil, fmt.Errorf("%s: нет id", f)
		}
		for _, line := range meta["lc"] {
			p := strings.Split(line, "|")
			if len(p) < 3 {
				return nil, fmt.Errorf("%s: плохая строка lc: %q", f, line)
			}
			item := LC{Slug: strings.TrimSpace(p[0]), Title: strings.TrimSpace(p[1]), Level: strings.TrimSpace(p[2])}
			if len(p) > 3 {
				item.Hint = strings.TrimSpace(p[3])
			}
			l.LC = append(l.LC, item)
		}
		c.Lessons = append(c.Lessons, l)
		c.byID[l.ID] = l
	}

	tfiles, _ := fs.Glob(fsys, "content/tasks/*.md")
	sort.Strings(tfiles)
	for _, f := range tfiles {
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		id := strings.TrimSuffix(path.Base(f), ".md")
		t, err := parseTask(id, string(b))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		l := c.byID[t.Lesson]
		if l == nil {
			return nil, fmt.Errorf("%s: нет урока %q", f, t.Lesson)
		}
		l.Tasks = append(l.Tasks, t.ID)
		c.Tasks[t.ID] = t
	}

	if b, err := fs.ReadFile(fsys, "content/cards.md"); err == nil {
		c.Cards = parseCards(string(b))
	}
	pages, _ := fs.Glob(fsys, "content/pages/*.md")
	for _, f := range pages {
		b, _ := fs.ReadFile(fsys, f)
		c.Pages[strings.TrimSuffix(path.Base(f), ".md")] = string(b)
	}
	return c, nil
}

func parseTask(id, src string) (*Task, error) {
	meta, body := parseFront(src)
	t := &Task{ID: id, Title: first(meta, "title"), Level: first(meta, "level")}
	t.Lesson, _, _ = strings.Cut(id, "-")
	sections := map[string]*strings.Builder{}
	var cur *strings.Builder
	for _, line := range strings.Split(body, "\n") {
		if name, ok := strings.CutPrefix(line, "@@ "); ok {
			cur = &strings.Builder{}
			sections[strings.TrimSpace(name)] = cur
			continue
		}
		if cur != nil {
			cur.WriteString(line)
			cur.WriteByte('\n')
		}
	}
	get := func(k string) string {
		if b := sections[k]; b != nil {
			return strings.TrimSpace(b.String()) + "\n"
		}
		return ""
	}
	t.Statement = strings.TrimSpace(get("statement"))
	t.Starter, t.Tests, t.Solution = get("starter"), get("tests"), get("solution")
	t.Hint = strings.TrimSpace(get("hint"))
	if t.Title == "" || t.Statement == "" || t.Starter == "\n" || t.Tests == "\n" || t.Solution == "\n" {
		return nil, fmt.Errorf("неполное задание (нужны title, statement, starter, tests, solution)")
	}
	return t, nil
}

func parseCards(src string) []Card {
	var cards []Card
	var cur *Card
	var ans []string
	flush := func() {
		if cur != nil {
			cur.A = strings.TrimSpace(strings.Join(ans, "\n"))
			cards = append(cards, *cur)
		}
		cur, ans = nil, nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n") {
		if h, ok := strings.CutPrefix(line, "## "); ok {
			flush()
			tag := ""
			if strings.HasPrefix(h, "[") {
				if i := strings.Index(h, "]"); i > 0 {
					tag, h = h[1:i], strings.TrimSpace(h[i+1:])
				}
			}
			hs := fnv.New32a()
			hs.Write([]byte(h))
			cur = &Card{ID: fmt.Sprintf("%08x", hs.Sum32()), Tag: tag, Q: h}
			continue
		}
		if cur != nil {
			ans = append(ans, line)
		}
	}
	flush()
	return cards
}
