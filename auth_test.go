package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func authServer(t *testing.T) (*auth, http.Handler) {
	t.Helper()
	a := newAuth("correct-horse-battery")
	a.failDelay = 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("secret")) })
	return a, guard(a.wrap(next), true)
}

func do(h http.Handler, method, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = "localhost:8088"
	r.Header.Set("X-Requested-With", "golearn")
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestAuthBlocksAnonymous(t *testing.T) {
	_, h := authServer(t)
	if w := do(h, "GET", "/api/progress", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("/api без входа: %d, ждали 401", w.Code)
	}
	w := do(h, "GET", "/", "")
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Errorf("/ без входа: %d %q, ждали редирект на /login", w.Code, w.Header().Get("Location"))
	}
	if w := do(h, "GET", "/login", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Пароль") {
		t.Errorf("страница входа недоступна: %d", w.Code)
	}
}

func TestAuthLogin(t *testing.T) {
	_, h := authServer(t)
	if w := do(h, "POST", "/auth/login", `{"password":"wrong"}`); w.Code != http.StatusUnauthorized || len(w.Result().Cookies()) != 0 {
		t.Errorf("неверный пароль: %d, cookies=%v", w.Code, w.Result().Cookies())
	}
	w := do(h, "POST", "/auth/login", `{"password":"correct-horse-battery"}`)
	cookies := w.Result().Cookies()
	if w.Code != http.StatusOK || len(cookies) != 1 {
		t.Fatalf("верный пароль: %d, cookies=%v", w.Code, cookies)
	}
	if c := cookies[0]; !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie должна быть HttpOnly и SameSite=Lax: %+v", c)
	}
	if got := do(h, "GET", "/api/anything", "", cookies[0]); got.Code != http.StatusOK || got.Body.String() != "secret" {
		t.Errorf("с cookie доступ закрыт: %d %q", got.Code, got.Body.String())
	}
}

func TestAuthLoginNeedsHeader(t *testing.T) {
	_, h := authServer(t)
	r := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"password":"correct-horse-battery"}`))
	r.Host = "localhost:8088"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("вход без служебного заголовка: %d, ждали 403", w.Code)
	}
}

func TestAuthSecureCookieBehindProxy(t *testing.T) {
	_, h := authServer(t)
	r := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"password":"correct-horse-battery"}`))
	r.Host = "localhost:8088"
	r.Header.Set("X-Requested-With", "golearn")
	r.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if cs := w.Result().Cookies(); len(cs) != 1 || !cs[0].Secure {
		t.Errorf("за HTTPS-прокси cookie должна быть Secure: %v", cs)
	}
}

func TestSessionToken(t *testing.T) {
	a := newAuth("correct-horse-battery")
	now := time.Now()
	tok := a.token(now)
	if !a.valid(tok, now) {
		t.Error("свежий токен не прошёл")
	}
	if a.valid(tok, now.Add(sessionTTL+time.Minute)) {
		t.Error("просроченный токен прошёл")
	}
	exp, sig, _ := strings.Cut(tok, ".")
	for name, bad := range map[string]string{
		"подпись":       exp + "." + sig[:len(sig)-2] + "AA",
		"срок":          "9999999999." + sig,
		"без подписи":   exp,
		"пустой":        "",
		"чужой пароль":  newAuth("another-password-1").token(now),
		"мусор в сроке": "abc." + a.sign("abc"),
	} {
		if a.valid(bad, now) {
			t.Errorf("подделанный токен (%s) прошёл проверку", name)
		}
	}
}

func TestLoginThrottle(t *testing.T) {
	a, h := authServer(t)
	a.pending.Store(8) // очередь из попыток уже заполнена
	if w := do(h, "POST", "/auth/login", `{"password":"correct-horse-battery"}`); w.Code != http.StatusTooManyRequests {
		t.Errorf("переполненная очередь: %d, ждали 429", w.Code)
	}
	if n := a.pending.Load(); n != 8 {
		t.Errorf("счётчик очереди сбился: %d", n)
	}
}
