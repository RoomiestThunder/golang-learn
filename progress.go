package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type TaskState struct {
	Solved   bool   `json:"solved"`
	Attempts int    `json:"attempts"`
	SolvedAt string `json:"solvedAt,omitempty"`
	Code     string `json:"code,omitempty"`
	Peeked   bool   `json:"peeked,omitempty"` // открывал решение
}

type LCState struct {
	Done   bool   `json:"done"`
	Source string `json:"source"` // auto | manual
	At     string `json:"at"`
}

type Progress struct {
	Lessons  map[string]string     `json:"lessons"`
	Tasks    map[string]*TaskState `json:"tasks"`
	LC       map[string]*LCState   `json:"lc"`
	LCUser   string                `json:"lcUser"`
	LCSynced string                `json:"lcSynced"`
	Cards    map[string]int        `json:"cards"`
	Checks   map[string]bool       `json:"checks"`
	Days     map[string]int        `json:"days"`
	Started  string                `json:"started"`
}

func newProgress() *Progress {
	return &Progress{
		Lessons: map[string]string{}, Tasks: map[string]*TaskState{}, LC: map[string]*LCState{},
		Cards: map[string]int{}, Checks: map[string]bool{}, Days: map[string]int{},
		Started: time.Now().Format("2006-01-02"),
	}
}

func loadProgress(path string) *Progress {
	p := newProgress()
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, p)
	}
	if p.Lessons == nil {
		p.Lessons = map[string]string{}
	}
	if p.Tasks == nil {
		p.Tasks = map[string]*TaskState{}
	}
	if p.LC == nil {
		p.LC = map[string]*LCState{}
	}
	if p.Cards == nil {
		p.Cards = map[string]int{}
	}
	if p.Checks == nil {
		p.Checks = map[string]bool{}
	}
	if p.Days == nil {
		p.Days = map[string]int{}
	}
	return p
}

func (p *Progress) touch() { p.Days[time.Now().Format("2006-01-02")]++ }

func (p *Progress) task(id string) *TaskState {
	if p.Tasks[id] == nil {
		p.Tasks[id] = &TaskState{}
	}
	return p.Tasks[id]
}

func saveProgress(path string, p *Progress) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(filepath.Dir(path), ".progress.tmp")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		// не все файловые системы (например, Cloud Storage FUSE) умеют rename — пишем напрямую
		_ = os.Remove(tmp)
		return os.WriteFile(path, b, 0o644)
	}
	return nil
}

// ---- LeetCode: публичный GraphQL, читает последние принятые решения профиля ----

type acSub struct {
	Title     string `json:"title"`
	TitleSlug string `json:"titleSlug"`
	Timestamp string `json:"timestamp"`
}

func fetchAccepted(ctx context.Context, user string) ([]acSub, error) {
	q := `query recentAc($username: String!, $limit: Int!) {
  recentAcSubmissionList(username: $username, limit: $limit) { title titleSlug timestamp }
}`
	body, _ := json.Marshal(map[string]any{"query": q, "variables": map[string]any{"username": user, "limit": 50}})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://leetcode.com/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://leetcode.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (golearn local)")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("нет связи с leetcode.com: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("leetcode.com ответил %d", resp.StatusCode)
	}
	var out struct {
		Data struct {
			List []acSub `json:"recentAcSubmissionList"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, errors.New("неожиданный ответ leetcode.com")
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("leetcode: %s", out.Errors[0].Message)
	}
	return out.Data.List, nil
}
