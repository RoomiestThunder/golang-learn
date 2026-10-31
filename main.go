// golearn — локальная платформа для изучения Go: теория, практика с автопроверкой,
// задачи LeetCode с синхронизацией прогресса, карточки для собеседования.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

//go:embed web content
var assets embed.FS

type Server struct {
	mu       sync.Mutex
	prog     *Progress
	dataPath string
	runner   *Runner
	dev      bool
	course   *Course
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8088", "адрес сервера (только локальный)")
	data := flag.String("data", "progress.json", "файл с прогрессом")
	dev := flag.Bool("dev", false, "читать content/ и web/ с диска (правка материалов без пересборки)")
	flag.Parse()

	runner, err := NewRunner()
	if err != nil {
		log.Fatal(err)
	}
	s := &Server{dataPath: *data, runner: runner, dev: *dev, prog: loadProgress(*data)}
	if _, err := s.getCourse(); err != nil {
		log.Fatalf("контент: %v", err)
	}
	go s.backgroundSync()

	host, _, _ := net.SplitHostPort(*addr)
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		log.Fatal("сервер выполняет произвольный код, слушать можно только localhost")
	}
	fmt.Printf("\n  Go Learn запущен:  http://%s\n  Прогресс хранится в: %s\n\n", *addr, *data)
	log.Fatal(http.ListenAndServe(*addr, s.routes()))
}

func (s *Server) fsys() fs.FS {
	if s.dev {
		return os.DirFS(".")
	}
	return assets
}

func (s *Server) getCourse() (*Course, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.course == nil || s.dev {
		c, err := loadCourse(s.fsys())
		if err != nil {
			return nil, err
		}
		s.course = c
	}
	return s.course, nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	web, _ := fs.Sub(s.fsys(), "web")
	mux.Handle("/", http.FileServerFS(web))
	mux.HandleFunc("GET /api/course", s.handleCourse)
	mux.HandleFunc("GET /api/lesson/{id}", s.handleLesson)
	mux.HandleFunc("GET /api/progress", func(w http.ResponseWriter, r *http.Request) { writeRaw(w, s.snapshot()) })
	mux.HandleFunc("GET /api/cards", s.handleCards)
	mux.HandleFunc("GET /api/page/{name}", s.handlePage)
	mux.HandleFunc("GET /api/solution/{task}", s.handleSolution)
	mux.HandleFunc("POST /api/run", s.handleRun)
	mux.HandleFunc("POST /api/check/{task}", s.handleCheck)
	mux.HandleFunc("POST /api/draft/{task}", s.handleDraft)
	mux.HandleFunc("POST /api/lesson/{id}/done", s.handleLessonDone)
	mux.HandleFunc("POST /api/lc/toggle", s.handleLCToggle)
	mux.HandleFunc("POST /api/lc/sync", s.handleLCSync)
	mux.HandleFunc("POST /api/card", s.handleCard)
	mux.HandleFunc("POST /api/check-item", s.handleCheckItem)
	return guard(mux)
}

// extraHostAllowed разрешает Host, который приходит через проброс портов GitHub Codespaces
// (<codespace>-<port>.<домен>; порт в Codespaces приватный — доступен только владельцу)
// и явно перечисленные в GOLEARN_ALLOWED_HOSTS хосты (через запятую).
func extraHostAllowed(h string) bool {
	for _, a := range strings.Split(os.Getenv("GOLEARN_ALLOWED_HOSTS"), ",") {
		if a = strings.TrimSpace(a); a != "" && strings.EqualFold(h, a) {
			return true
		}
	}
	name, domain := os.Getenv("CODESPACE_NAME"), os.Getenv("GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN")
	return name != "" && domain != "" && strings.HasPrefix(h, name+"-") && strings.HasSuffix(h, "."+domain)
}

// guard защищает от запросов с чужих сайтов (CSRF / DNS rebinding):
// принимаем только localhost-хосты и требуем свой заголовок на изменяющих запросах.
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			h = r.Host
		}
		if h != "127.0.0.1" && h != "localhost" && h != "[::1]" && h != "::1" && !extraHostAllowed(h) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Header.Get("X-Requested-With") != "golearn" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func writeRaw(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(b)
}

func fail(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func decode(r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(v) == nil
}

func (s *Server) snapshot() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.Marshal(s.prog)
	return b
}

func (s *Server) update(fn func(p *Progress)) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.prog)
	s.prog.touch()
	if err := saveProgress(s.dataPath, s.prog); err != nil {
		log.Printf("не удалось сохранить прогресс: %v", err)
	}
	b, _ := json.Marshal(s.prog)
	return b
}

// ---- обработчики ----

func (s *Server) handleCourse(w http.ResponseWriter, r *http.Request) {
	c, err := s.getCourse()
	if err != nil {
		fail(w, 500, err.Error())
		return
	}
	type lessonSum struct {
		ID      string   `json:"id"`
		Week    int      `json:"week"`
		Day     int      `json:"day"`
		Minutes int      `json:"minutes"`
		Title   string   `json:"title"`
		Goals   string   `json:"goals"`
		Tasks   []string `json:"tasks"`
		LC      []string `json:"lc"`
	}
	sums := make([]lessonSum, 0, len(c.Lessons))
	for _, l := range c.Lessons {
		ls := lessonSum{l.ID, l.Week, l.Day, l.Minutes, l.Title, l.Goals, l.Tasks, nil}
		for _, p := range l.LC {
			ls.LC = append(ls.LC, p.Slug)
		}
		sums = append(sums, ls)
	}
	tasks := map[string]map[string]string{}
	for id, t := range c.Tasks {
		tasks[id] = map[string]string{"title": t.Title, "level": t.Level, "lesson": t.Lesson}
	}
	writeJSON(w, map[string]any{"weeks": c.Weeks, "lessons": sums, "tasks": tasks, "cards": len(c.Cards)})
}

func (s *Server) handleLesson(w http.ResponseWriter, r *http.Request) {
	c, err := s.getCourse()
	if err != nil {
		fail(w, 500, err.Error())
		return
	}
	l := c.byID[r.PathValue("id")]
	if l == nil {
		fail(w, 404, "урок не найден")
		return
	}
	tasks := make([]*Task, 0, len(l.Tasks))
	for _, id := range l.Tasks {
		tasks = append(tasks, c.Tasks[id])
	}
	writeJSON(w, map[string]any{"lesson": l, "tasks": tasks})
}

func (s *Server) handleCards(w http.ResponseWriter, r *http.Request) {
	c, err := s.getCourse()
	if err != nil {
		fail(w, 500, err.Error())
		return
	}
	writeJSON(w, c.Cards)
}

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	c, err := s.getCourse()
	if err != nil {
		fail(w, 500, err.Error())
		return
	}
	p, ok := c.Pages[r.PathValue("name")]
	if !ok {
		fail(w, 404, "страница не найдена")
		return
	}
	writeJSON(w, map[string]string{"body": p})
}

func (s *Server) handleSolution(w http.ResponseWriter, r *http.Request) {
	c, _ := s.getCourse()
	t := c.Tasks[r.PathValue("task")]
	if t == nil {
		fail(w, 404, "нет такого задания")
		return
	}
	s.update(func(p *Progress) { p.task(t.ID).Peeked = true })
	writeJSON(w, map[string]string{"solution": t.Solution})
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	var in struct{ Code string }
	if !decode(r, &in) || len(in.Code) > 200_000 {
		fail(w, 400, "плохой запрос")
		return
	}
	writeJSON(w, s.runner.Play(in.Code))
}

func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	c, _ := s.getCourse()
	t := c.Tasks[r.PathValue("task")]
	var in struct{ Code string }
	if t == nil || !decode(r, &in) || len(in.Code) > 200_000 {
		fail(w, 400, "плохой запрос")
		return
	}
	res := s.runner.Check(t, in.Code)
	var snap []byte
	s.update(func(p *Progress) {
		st := p.task(t.ID)
		st.Attempts++
		st.Code = in.Code
		if res.OK && !st.Solved {
			st.Solved, st.SolvedAt = true, time.Now().Format(time.RFC3339)
		}
	})
	snap = s.snapshot()
	writeJSON(w, map[string]any{"result": res, "progress": json.RawMessage(snap)})
}

func (s *Server) handleDraft(w http.ResponseWriter, r *http.Request) {
	var in struct{ Code string }
	if !decode(r, &in) || len(in.Code) > 200_000 {
		fail(w, 400, "плохой запрос")
		return
	}
	id := r.PathValue("task")
	s.mu.Lock()
	s.prog.task(id).Code = in.Code
	_ = saveProgress(s.dataPath, s.prog)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLessonDone(w http.ResponseWriter, r *http.Request) {
	var in struct{ Done bool }
	if !decode(r, &in) {
		fail(w, 400, "плохой запрос")
		return
	}
	id := r.PathValue("id")
	writeRaw(w, s.update(func(p *Progress) {
		if in.Done {
			p.Lessons[id] = time.Now().Format(time.RFC3339)
		} else {
			delete(p.Lessons, id)
		}
	}))
}

func (s *Server) handleLCToggle(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Slug string
		Done bool
	}
	if !decode(r, &in) || in.Slug == "" {
		fail(w, 400, "плохой запрос")
		return
	}
	writeRaw(w, s.update(func(p *Progress) {
		if in.Done {
			p.LC[in.Slug] = &LCState{Done: true, Source: "manual", At: time.Now().Format(time.RFC3339)}
		} else {
			delete(p.LC, in.Slug)
		}
	}))
}

func (s *Server) syncLC(ctx context.Context, user string) (added int, err error) {
	subs, err := fetchAccepted(ctx, user)
	if err != nil {
		return 0, err
	}
	s.update(func(p *Progress) {
		p.LCUser = user
		p.LCSynced = time.Now().Format(time.RFC3339)
		for _, sub := range subs {
			if st := p.LC[sub.TitleSlug]; st == nil || !st.Done {
				p.LC[sub.TitleSlug] = &LCState{Done: true, Source: "auto", At: time.Now().Format(time.RFC3339)}
				added++
			}
		}
	})
	return added, nil
}

func (s *Server) handleLCSync(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username string }
	if !decode(r, &in) {
		fail(w, 400, "плохой запрос")
		return
	}
	user := strings.TrimSpace(in.Username)
	if user == "" {
		s.mu.Lock()
		user = s.prog.LCUser
		s.mu.Unlock()
	}
	if user == "" || len(user) > 64 {
		fail(w, 400, "укажи ник на LeetCode")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	added, err := s.syncLC(ctx, user)
	if err != nil {
		fail(w, 502, err.Error())
		return
	}
	writeJSON(w, map[string]any{"added": added, "progress": json.RawMessage(s.snapshot())})
}

// backgroundSync раз в несколько минут подтягивает решения с LeetCode, пока открыта платформа.
func (s *Server) backgroundSync() {
	for range time.Tick(3 * time.Minute) {
		s.mu.Lock()
		user := s.prog.LCUser
		s.mu.Unlock()
		if user == "" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		if n, err := s.syncLC(ctx, user); err != nil {
			log.Printf("leetcode sync: %v", err)
		} else if n > 0 {
			log.Printf("leetcode sync: +%d решённых задач", n)
		}
		cancel()
	}
}

func (s *Server) handleCard(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID  string
		Box int
	}
	if !decode(r, &in) || in.ID == "" {
		fail(w, 400, "плохой запрос")
		return
	}
	writeRaw(w, s.update(func(p *Progress) { p.Cards[in.ID] = max(0, min(in.Box, 4)) }))
}

func (s *Server) handleCheckItem(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Key string
		On  bool
	}
	if !decode(r, &in) || in.Key == "" {
		fail(w, 400, "плохой запрос")
		return
	}
	writeRaw(w, s.update(func(p *Progress) {
		if in.On {
			p.Checks[in.Key] = true
		} else {
			delete(p.Checks, in.Key)
		}
	}))
}
