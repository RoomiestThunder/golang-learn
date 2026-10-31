---
id: l26
week: 6
day: 1
minutes: 120
title: Аутентификация и безопасность
goals: Пароли, сессии vs JWT, подпись HS256 своими руками, авторизация, типичные уязвимости веб-сервисов
lc: encode-and-decode-tinyurl | Encode and Decode TinyURL | Medium | Прообраз итогового проекта: счётчик + map в обе стороны (или base62)
lc: validate-ip-address | Validate IP Address | Medium | Разбор строки с аккуратной проверкой всех граничных случаев
---

## Аутентификация и авторизация

- **Аутентификация** — «кто ты?» (логин/пароль, токен, сертификат).
- **Авторизация** — «что тебе можно?» (роли, права, владение ресурсом).

Базовое правило: **никогда не доверяй клиенту**. Идентификатор пользователя берётся из проверенного токена/сессии, а не из тела запроса (`{"user_id": 5}` — классическая уязвимость IDOR).

## Пароли

- Пароли хранят **только в виде медленного хэша с солью**: bcrypt, scrypt, argon2id (`golang.org/x/crypto/bcrypt`, `argon2`). Не `md5`/`sha256` — они слишком быстрые, перебираются на GPU.
- Стоимость (cost) подбирают под текущее железо (bcrypt cost 10–12).
- Сравнение — функциями библиотеки (`bcrypt.CompareHashAndPassword`) — они устойчивы к timing-атакам.
- Одинаковые сообщения для «нет пользователя» и «неверный пароль»; лимит попыток/блокировка; не логируй пароли.

```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
...
err = bcrypt.CompareHashAndPassword(hash, []byte(password)) // nil — совпало
```

## Сессии или токены

| | Серверные сессии (cookie) | JWT (токены) |
|---|---|---|
| Состояние | на сервере (БД/Redis), в cookie лишь ID | в самом токене, сервер без состояния |
| Отзыв | легко (удалить сессию) | сложно (до истечения нужен blacklist/короткий TTL) |
| Масштабирование | общий стор сессий | любой инстанс проверяет подпись |
| Подходит | браузерные приложения | API между сервисами, мобильные клиенты |

Практика: **короткоживущий access-токен (5–15 мин) + refresh-токен** (длинный, хранится в БД, ротируется, может быть отозван).

Cookie для сессий: `HttpOnly` (недоступна JS), `Secure` (только HTTPS), `SameSite=Lax/Strict` (защита от CSRF).

## JWT изнутри

JWT = три части в base64url через точку: `header.payload.signature`.

- header: `{"alg":"HS256","typ":"JWT"}`
- payload (claims): `{"sub":"42","role":"admin","exp":1735689600}` — **не шифруется**, любой может прочитать: секретов внутри нет!
- signature: `HMAC-SHA256(secret, header + "." + payload)`

```go run
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var enc = base64.RawURLEncoding // без «=» в конце

type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func sign(secret []byte, signingInput string) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(signingInput))
	return m.Sum(nil)
}

func Issue(secret []byte, c Claims) string {
	h := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p, _ := json.Marshal(c)
	in := h + "." + enc.EncodeToString(p)
	return in + "." + enc.EncodeToString(sign(secret, in))
}

func Verify(secret []byte, token string, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("формат токена")
	}
	got, err := enc.DecodeString(parts[2])
	if err != nil {
		return Claims{}, errors.New("подпись")
	}
	// hmac.Equal — сравнение за константное время (нельзя через == или bytes.Equal)
	if !hmac.Equal(got, sign(secret, parts[0]+"."+parts[1])) {
		return Claims{}, errors.New("подпись не сходится")
	}
	raw, _ := enc.DecodeString(parts[1])
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return Claims{}, err
	}
	if now.Unix() >= c.Exp {
		return Claims{}, errors.New("токен истёк")
	}
	return c, nil
}

func main() {
	secret := []byte("super-secret-key")
	now := time.Now()
	tok := Issue(secret, Claims{"42", "admin", now.Add(15 * time.Minute).Unix()})
	fmt.Println(strings.Count(tok, "."), len(tok) > 50)

	c, err := Verify(secret, tok, now)
	fmt.Println(c, err)

	// подмена роли: меняем payload, подпись остаётся старой
	parts := strings.Split(tok, ".")
	evil, _ := json.Marshal(Claims{"42", "root", now.Add(time.Hour).Unix()})
	forged := parts[0] + "." + enc.EncodeToString(evil) + "." + parts[2]
	_, err = Verify(secret, forged, now)
	fmt.Println("подделка:", err)

	_, err = Verify(secret, tok, now.Add(time.Hour))
	fmt.Println("просрочен:", err)
}
```

В продакшене используют проверенные библиотеки (`golang-jwt/jwt`), но **знать механику обязательно**. Важные правила проверки:

1. **Жёстко проверяй алгоритм** (`alg`): атака `alg: none` и подмена RS256→HS256 исторически ломали системы.
2. Проверяй `exp`, `nbf`, `iss`, `aud`.
3. Секрет HS256 — длинный случайный (≥ 32 байта) из окружения; для распределённых систем лучше асимметричные ключи (RS256/EdDSA: приватным подписывает один сервис, публичным проверяют все).
4. Не клади чувствительные данные в payload.
5. Токен передаётся в заголовке `Authorization: Bearer <token>`.

## Middleware аутентификации

```go
func Auth(secret []byte) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			claims, err := Verify(secret, tok, time.Now())
			if !ok || err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(withClaims(r.Context(), claims)))
		})
	}
}
```

Авторизация поверх: `RequireRole("admin")`, проверка владения (`order.UserID == claims.Sub`) — на уровне сервиса.

## OAuth2 / OIDC в двух словах

«Войти через Google/GitHub»: пользователь аутентифицируется у провайдера (Authorization Code Flow + PKCE), ваш сервис получает `code`, меняет его на токены, проверяет `id_token` (OIDC). В Go: `golang.org/x/oauth2`. Не изобретай свои схемы: используй готовых провайдеров/Keycloak/Auth0.

## Чек-лист безопасности веб-сервиса (OWASP)

| Угроза | Защита |
|---|---|
| SQL-инъекции | параметризованные запросы |
| XSS | `html/template` (автоэкранирование), заголовок `Content-Security-Policy` |
| CSRF | SameSite cookie, CSRF-токены для cookie-аутентификации |
| IDOR / нарушение авторизации | проверка владения на сервере для каждого ресурса |
| Брутфорс | rate limit, блокировка, капча |
| Утечка секретов | переменные окружения, Vault; не в git и не в логах |
| SSRF | белый список хостов при запросах по URL от пользователя |
| Небезопасная десериализация / большие тела | `MaxBytesReader`, `DisallowUnknownFields` |
| Уязвимые зависимости | `govulncheck`, Dependabot |
| Транспорт | всегда TLS; HSTS |
| Заголовки | `X-Content-Type-Options`, `X-Frame-Options`, CORS по белому списку |

> [!INTERVIEW] «Почему нельзя сравнивать хэши/подписи через `==`?» — ответ занимает время, зависящее от позиции первого несовпавшего байта → timing-атака позволяет подобрать подпись. Используй `hmac.Equal`/`subtle.ConstantTimeCompare`.

## Проверь себя

<details><summary>Чем отличаются 401 и 403?</summary>401 — не аутентифицирован; 403 — аутентифицирован, но нет прав.</details>

<details><summary>Почему в payload JWT нельзя класть пароль?</summary>Payload только закодирован (base64), а не зашифрован — любой читатель токена видит содержимое; подпись защищает от изменения, но не от чтения.</details>

<details><summary>Почему sha256 не годится для хранения паролей?</summary>Слишком быстрый — позволяет перебирать миллиарды вариантов в секунду; нужны адаптивные медленные функции (bcrypt/scrypt/argon2) с солью.</details>

<details><summary>Как отозвать JWT до истечения?</summary>Хранить список отозванных (jti) / версию токена пользователя, либо использовать короткий access-токен и отзывать refresh-токены.</details>
