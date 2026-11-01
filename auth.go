package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Платформа выполняет произвольный Go-код, поэтому при размещении на сервере вход
// только по паролю (GOLEARN_PASSWORD): без него слушать что-либо кроме localhost нельзя.

const (
	minPasswordLen = 12
	sessionTTL     = 30 * 24 * time.Hour
	cookieName     = "golearn_session"
)

//go:embed login.html
var loginPage []byte

type auth struct {
	pwSum     [sha256.Size]byte
	key       []byte
	mu        sync.Mutex    // попытки входа идут строго по одной
	pending   atomic.Int32  // сколько попыток уже ждёт своей очереди
	failDelay time.Duration // пауза после неверного пароля: ≈1 попытка в секунду на весь сервер
}

func newAuth(password string) *auth {
	key := sha256.Sum256([]byte("golearn/session/v1\x00" + password))
	return &auth{pwSum: sha256.Sum256([]byte(password)), key: key[:], failDelay: time.Second}
}

func (a *auth) sign(exp string) string {
	m := hmac.New(sha256.New, a.key)
	m.Write([]byte(exp))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func (a *auth) token(now time.Time) string {
	exp := strconv.FormatInt(now.Add(sessionTTL).Unix(), 10)
	return exp + "." + a.sign(exp)
}

func (a *auth) valid(token string, now time.Time) bool {
	exp, sig, ok := strings.Cut(token, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(a.sign(exp))) {
		return false
	}
	ts, err := strconv.ParseInt(exp, 10, 64)
	return err == nil && now.Unix() < ts
}

func (a *auth) authed(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	return err == nil && a.valid(c.Value, time.Now())
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func (a *auth) wrap(next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(loginPage)
	})
	mux.HandleFunc("POST /auth/login", a.handleLogin)
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.authed(r) {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			fail(w, http.StatusUnauthorized, "нужен вход")
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}))
	return mux
}

func (a *auth) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct{ Password string }
	if !decode(r, &in) {
		fail(w, http.StatusBadRequest, "плохой запрос")
		return
	}
	if a.pending.Add(1) > 8 {
		a.pending.Add(-1)
		fail(w, http.StatusTooManyRequests, "слишком много попыток, подожди минуту")
		return
	}
	defer a.pending.Add(-1)

	a.mu.Lock()
	defer a.mu.Unlock()
	sum := sha256.Sum256([]byte(in.Password))
	if subtle.ConstantTimeCompare(sum[:], a.pwSum[:]) != 1 {
		time.Sleep(a.failDelay)
		fail(w, http.StatusUnauthorized, "неверный пароль")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: a.token(time.Now()), Path: "/",
		MaxAge: int(sessionTTL.Seconds()), HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: isHTTPS(r),
	})
	writeJSON(w, map[string]bool{"ok": true})
}
