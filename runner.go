package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Result struct {
	OK       bool   `json:"ok"`
	Output   string `json:"output"`
	TimedOut bool   `json:"timedOut"`
	Ms       int64  `json:"ms"`
}

// Runner выполняет код пользователя во временных модулях.
type Runner struct {
	goDirective string
	sem         chan struct{}
}

func NewRunner() (*Runner, error) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return nil, fmt.Errorf("не найден компилятор Go в PATH: %w", err)
	}
	m := regexp.MustCompile(`go(\d+)\.(\d+)`).FindStringSubmatch(string(out))
	if m == nil {
		return nil, fmt.Errorf("не удалось разобрать версию: %s", out)
	}
	if atoi(m[1]) == 1 && atoi(m[2]) < 23 {
		return nil, fmt.Errorf("нужен Go 1.23+, найден %s", strings.TrimSpace(string(out)))
	}
	return &Runner{goDirective: m[1] + "." + m[2], sem: make(chan struct{}, 3)}, nil
}

type limitedBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
	max int
}

func (l *limitedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if room := l.max - l.buf.Len(); room > 0 {
		if len(p) > room {
			l.buf.Write(p[:room])
			l.buf.WriteString("\n... вывод обрезан ...\n")
		} else {
			l.buf.Write(p)
		}
	}
	return len(p), nil
}

func runCmd(dir string, timeout time.Duration, name string, args ...string) (string, int, bool) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOFLAGS=", "CGO_ENABLED=0", "GO111MODULE=on")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	out := &limitedBuf{max: 64 << 10}
	cmd.Stdout, cmd.Stderr = out, out
	if err := cmd.Start(); err != nil {
		return err.Error(), -1, false
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		code := 0
		if err != nil {
			code = 1
			if ee, ok := err.(*exec.ExitError); ok {
				code = ee.ExitCode()
			}
		}
		return out.buf.String(), code, false
	case <-time.After(timeout):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
		return out.buf.String(), -1, true
	}
}

func (r *Runner) workdir(files map[string]string) (string, error) {
	dir, err := os.MkdirTemp("", "golearn-*")
	if err != nil {
		return "", err
	}
	files["go.mod"] = "module play\n\ngo " + r.goDirective + "\n"
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}

func clean(out, dir string) string {
	out = strings.ReplaceAll(out, dir, ".")
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		out = strings.ReplaceAll(out, real, ".")
	}
	return strings.TrimSpace(out)
}

// Play собирает и запускает одиночную программу.
func (r *Runner) Play(code string) Result {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()
	start := time.Now()
	dir, err := r.workdir(map[string]string{"main.go": code})
	if err != nil {
		return Result{Output: err.Error()}
	}
	defer os.RemoveAll(dir)

	out, exit, _ := runCmd(dir, 120*time.Second, "go", "build", "-o", "prog", ".")
	if exit != 0 {
		return Result{Output: clean(out, dir), Ms: time.Since(start).Milliseconds()}
	}
	out, exit, to := runCmd(dir, 10*time.Second, filepath.Join(dir, "prog"))
	res := Result{OK: exit == 0, Output: clean(out, dir), TimedOut: to, Ms: time.Since(start).Milliseconds()}
	if to {
		res.Output += "\n\n⏱ Превышен лимит 10 секунд — программа остановлена (возможно, бесконечный цикл или взаимная блокировка)."
	}
	return res
}

// Check запускает тесты задания против кода пользователя.
func (r *Runner) Check(t *Task, code string) Result {
	return r.test(code, t.Tests)
}

func (r *Runner) test(code, tests string) Result {
	r.sem <- struct{}{}
	defer func() { <-r.sem }()
	start := time.Now()
	dir, err := r.workdir(map[string]string{"solution.go": code, "solution_test.go": tests})
	if err != nil {
		return Result{Output: err.Error()}
	}
	defer os.RemoveAll(dir)
	out, exit, to := runCmd(dir, 150*time.Second, "go", "test", "-count=1", "-timeout=20s", "-v", ".")
	res := Result{OK: exit == 0, Output: clean(out, dir), TimedOut: to, Ms: time.Since(start).Milliseconds()}
	if to {
		res.Output += "\n\n⏱ Проверка не уложилась в лимит времени."
	}
	return res
}
