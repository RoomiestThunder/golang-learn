---
title: JWT HS256 своими руками
level: hard
---
@@ statement
Реализуй подпись и проверку токенов формата JWT (HS256) только на стандартной библиотеке.

```go
type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role,omitempty"`
	Exp  int64  `json:"exp"` // Unix-секунды
}
```

- `Sign(secret []byte, c Claims) string` → `base64url(header).base64url(payload).base64url(signature)`. Header — JSON `{"alg":"HS256","typ":"JWT"}`; кодировка — `base64.RawURLEncoding` (без `=`); подпись — `HMAC-SHA256(secret, header + "." + payload)`.
- `Verify(secret []byte, token string, now time.Time) (Claims, error)` проверяет **строго**:
  1. ровно три части, корректный base64url — иначе `ErrInvalidToken`;
  2. заголовок — JSON с `"alg":"HS256"`; любой другой алгоритм (в том числе `none`) → `ErrInvalidToken`;
  3. подпись верна — сравнение через `hmac.Equal` (константное время) → иначе `ErrInvalidToken`;
  4. payload — корректный JSON → иначе `ErrInvalidToken`;
  5. токен **не истёк**: `now.Unix() >= Exp` → `ErrExpired`.
  
  Подпись проверяй **до** разбора содержимого payload и до проверки срока.
@@ starter
package solution

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("недействительный токен")
	ErrExpired      = errors.New("токен просрочен")
)

type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role,omitempty"`
	Exp  int64  `json:"exp"`
}

func Sign(secret []byte, c Claims) string {
	return ""
}

func Verify(secret []byte, token string, now time.Time) (Claims, error) {
	return Claims{}, nil
}
@@ tests
package solution

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	secret = []byte("0123456789abcdef0123456789abcdef")
	now    = time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	b64    = base64.RawURLEncoding
)

func mac(key []byte, data string) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return b64.EncodeToString(m.Sum(nil))
}

func TestSignFormat(t *testing.T) {
	tok := Sign(secret, Claims{Sub: "42", Role: "admin", Exp: now.Add(time.Hour).Unix()})
	parts := strings.Split(tok, ".")
	if len(parts) != 3 || strings.Contains(tok, "=") {
		t.Fatalf("формат: %q", tok)
	}
	hdr, err := b64.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var h map[string]string
	if json.Unmarshal(hdr, &h) != nil || h["alg"] != "HS256" || h["typ"] != "JWT" {
		t.Errorf("header = %s", hdr)
	}
	if parts[2] != mac(secret, parts[0]+"."+parts[1]) {
		t.Error("подпись не совпадает с эталонным HMAC-SHA256")
	}
}

func TestRoundTrip(t *testing.T) {
	want := Claims{Sub: "42", Role: "admin", Exp: now.Add(15 * time.Minute).Unix()}
	got, err := Verify(secret, Sign(secret, want), now)
	if err != nil || got != want {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestVerifyRejects(t *testing.T) {
	good := Sign(secret, Claims{Sub: "42", Role: "user", Exp: now.Add(time.Hour).Unix()})
	parts := strings.Split(good, ".")

	forgedPayload, _ := json.Marshal(Claims{Sub: "42", Role: "admin", Exp: now.Add(time.Hour).Unix()})
	forgedPayloadB64 := b64.EncodeToString(forgedPayload)

	noneHeader := b64.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	hs512Header := b64.EncodeToString([]byte(`{"alg":"HS512","typ":"JWT"}`))
	// токен с alg=none и ПРАВИЛЬНОЙ HMAC-подписью — всё равно должен отвергаться
	noneSigned := noneHeader + "." + parts[1] + "." + mac(secret, noneHeader+"."+parts[1])
	hs512 := hs512Header + "." + parts[1] + "." + mac(secret, hs512Header+"."+parts[1])

	badPayload := b64.EncodeToString([]byte("not json"))
	badPayloadTok := parts[0] + "." + badPayload + "." + mac(secret, parts[0]+"."+badPayload)

	tests := map[string]string{
		"пусто":             "",
		"две части":         parts[0] + "." + parts[1],
		"четыре части":      good + ".x",
		"мусор base64":      "@@@." + parts[1] + "." + parts[2],
		"подменён payload":  parts[0] + "." + forgedPayloadB64 + "." + parts[2],
		"чужой секрет":      Sign([]byte("another-secret"), Claims{Sub: "42", Exp: now.Add(time.Hour).Unix()}),
		"alg=none без подп": noneHeader + "." + parts[1] + ".",
		"alg=none с подп":   noneSigned,
		"alg=HS512":         hs512,
		"payload не JSON":   badPayloadTok,
		"пустая подпись":    parts[0] + "." + parts[1] + ".",
	}
	for name, tok := range tests {
		if _, err := Verify(secret, tok, now); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("%s: err = %v, ожидался ErrInvalidToken", name, err)
		}
	}
}

func TestExpiry(t *testing.T) {
	tok := Sign(secret, Claims{Sub: "1", Exp: now.Unix()})
	if _, err := Verify(secret, tok, now); !errors.Is(err, ErrExpired) {
		t.Errorf("Exp == now → просрочен, got %v", err)
	}
	if _, err := Verify(secret, tok, now.Add(-time.Second)); err != nil {
		t.Errorf("за секунду до Exp токен валиден: %v", err)
	}
	if _, err := Verify(secret, tok, now.Add(time.Hour)); !errors.Is(err, ErrExpired) {
		t.Errorf("через час: %v", err)
	}
	// подпись проверяется раньше срока: просроченный, но подделанный токен — ErrInvalidToken
	parts := strings.Split(tok, ".")
	forged := parts[0] + "." + parts[1] + "." + b64.EncodeToString([]byte("bad"))
	if _, err := Verify(secret, forged, now.Add(time.Hour)); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("подделка важнее срока: %v", err)
	}
}
@@ solution
package solution

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("недействительный токен")
	ErrExpired      = errors.New("токен просрочен")
)

type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role,omitempty"`
	Exp  int64  `json:"exp"`
}

var enc = base64.RawURLEncoding

func computeMAC(secret []byte, input string) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(input))
	return m.Sum(nil)
}

func Sign(secret []byte, c Claims) string {
	header := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(c)
	input := header + "." + enc.EncodeToString(payload)
	return input + "." + enc.EncodeToString(computeMAC(secret, input))
}

func Verify(secret []byte, token string, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	hdrRaw, err := enc.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var hdr struct {
		Alg string `json:"alg"`
	}
	if json.Unmarshal(hdrRaw, &hdr) != nil || hdr.Alg != "HS256" {
		return Claims{}, ErrInvalidToken
	}
	sig, err := enc.DecodeString(parts[2])
	if err != nil || !hmac.Equal(sig, computeMAC(secret, parts[0]+"."+parts[1])) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := enc.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var c Claims
	if json.Unmarshal(payload, &c) != nil {
		return Claims{}, ErrInvalidToken
	}
	if now.Unix() >= c.Exp {
		return Claims{}, ErrExpired
	}
	return c, nil
}
@@ hint
Подписывается строка `header_b64 + "." + payload_b64` (именно закодированные части, не исходный JSON). Сначала проверь алгоритм и подпись, и только потом доверяй содержимому.
