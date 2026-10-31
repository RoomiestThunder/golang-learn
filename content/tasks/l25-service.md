---
title: Сервис с внедрением зависимостей
level: hard
---
@@ statement
Реализуй `UserService` с регистрацией и входом. Зависимости (`UserRepo`, `Hasher`, `Clock`) приходят через конструктор — в тестах подставляются фейки.

**`Register(ctx, email, password) (User, error)`**
- email: `TrimSpace` + нижний регистр; должен содержать `@` (не первым/последним символом); пароль ≥ 8 символов → иначе `ErrInvalid`;
- если репозиторий нашёл пользователя → ошибка, оборачивающая `ErrEmailTaken` (`%w`);
- хэш через `Hasher.Hash`, `CreatedAt` из `Clock.Now()`; сохранить через `repo.Create`; ошибку репозитория вернуть обёрнутой (`%w`);
- возвращается пользователь с заполненным ID.

**`Login(ctx, email, password) (User, error)`**
- нет такого пользователя (`repo.ByEmail` вернул `ErrNotFound`) → `ErrBadCredentials` (тот же ответ, что и для неверного пароля — не раскрываем, существует ли email);
- если у пользователя уже `Failures >= MaxFailures` (3) → `ErrLocked` (пароль даже не проверяется);
- неверный пароль → `repo.RecordFailure(id)` и `ErrBadCredentials`;
- верный → `repo.ResetFailures(id)` (если были неудачи) и вернуть пользователя.
- Прочие ошибки репозитория вернуть обёрнутыми.
@@ starter
package solution

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid        = errors.New("некорректные данные")
	ErrEmailTaken     = errors.New("email уже занят")
	ErrNotFound       = errors.New("не найдено")
	ErrBadCredentials = errors.New("неверный email или пароль")
	ErrLocked         = errors.New("аккаунт заблокирован")
)

const MaxFailures = 3

type User struct {
	ID        int
	Email     string
	PassHash  string
	CreatedAt time.Time
	Failures  int
}

type UserRepo interface {
	ByEmail(ctx context.Context, email string) (User, error) // ErrNotFound, если нет
	Create(ctx context.Context, u *User) error                // заполняет u.ID
	RecordFailure(ctx context.Context, id int) error
	ResetFailures(ctx context.Context, id int) error
}

type Hasher interface {
	Hash(password string) string
	Verify(hash, password string) bool
}

type Clock interface{ Now() time.Time }

type UserService struct {
	// TODO: зависимости
}

func NewUserService(r UserRepo, h Hasher, c Clock) *UserService { return &UserService{} }

func (s *UserService) Register(ctx context.Context, email, password string) (User, error) {
	return User{}, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (User, error) {
	return User{}, nil
}
@@ tests
package solution

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memRepo struct {
	users     map[string]*User
	nextID    int
	createErr error
	resets    int
}

func newRepo() *memRepo { return &memRepo{users: map[string]*User{}, nextID: 1} }

func (m *memRepo) ByEmail(_ context.Context, e string) (User, error) {
	if u, ok := m.users[e]; ok {
		return *u, nil
	}
	return User{}, ErrNotFound
}
func (m *memRepo) Create(_ context.Context, u *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	u.ID = m.nextID
	m.nextID++
	cp := *u
	m.users[u.Email] = &cp
	return nil
}
func (m *memRepo) RecordFailure(_ context.Context, id int) error {
	for _, u := range m.users {
		if u.ID == id {
			u.Failures++
		}
	}
	return nil
}
func (m *memRepo) ResetFailures(_ context.Context, id int) error {
	for _, u := range m.users {
		if u.ID == id {
			u.Failures = 0
			m.resets++
		}
	}
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(pw string) string        { return "h:" + pw }
func (fakeHasher) Verify(h, pw string) bool     { return h == "h:"+pw }

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

var now = time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC)

func setup() (*UserService, *memRepo) {
	r := newRepo()
	return NewUserService(r, fakeHasher{}, fixedClock{now}), r
}

func TestRegister(t *testing.T) {
	svc, repo := setup()
	ctx := context.Background()
	u, err := svc.Register(ctx, "  Ann@Example.COM ", "password1")
	if err != nil || u.ID != 1 || u.Email != "ann@example.com" || u.PassHash != "h:password1" || !u.CreatedAt.Equal(now) {
		t.Fatalf("Register: %+v, %v", u, err)
	}
	if _, ok := repo.users["ann@example.com"]; !ok {
		t.Error("пользователь не сохранён под нормализованным email")
	}
	if _, err := svc.Register(ctx, "ANN@example.com", "otherpass1"); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("дубликат: %v", err)
	}
	for _, c := range [][2]string{{"no-at", "password1"}, {"@x.io", "password1"}, {"a@", "password1"}, {"ok@x.io", "short"}, {"", ""}} {
		if _, err := svc.Register(ctx, c[0], c[1]); !errors.Is(err, ErrInvalid) {
			t.Errorf("Register(%q,%q): %v", c[0], c[1], err)
		}
	}
	repo.createErr = errors.New("db down")
	if _, err := svc.Register(ctx, "new@x.io", "password1"); !errors.Is(err, repo.createErr) {
		t.Errorf("ошибка репозитория должна быть обёрнута: %v", err)
	}
}

func TestLogin(t *testing.T) {
	svc, repo := setup()
	ctx := context.Background()
	svc.Register(ctx, "ann@x.io", "password1")

	if u, err := svc.Login(ctx, "ANN@x.io ", "password1"); err != nil || u.Email != "ann@x.io" {
		t.Fatalf("успешный вход: %+v %v", u, err)
	}
	if _, err := svc.Login(ctx, "ghost@x.io", "password1"); !errors.Is(err, ErrBadCredentials) {
		t.Errorf("несуществующий: %v", err)
	}

	for i := 0; i < MaxFailures; i++ {
		if _, err := svc.Login(ctx, "ann@x.io", "wrong"); !errors.Is(err, ErrBadCredentials) {
			t.Fatalf("неверный пароль #%d: %v", i, err)
		}
	}
	if repo.users["ann@x.io"].Failures != MaxFailures {
		t.Errorf("Failures = %d", repo.users["ann@x.io"].Failures)
	}
	// теперь заблокирован — даже верный пароль не принимается
	if _, err := svc.Login(ctx, "ann@x.io", "password1"); !errors.Is(err, ErrLocked) {
		t.Errorf("блокировка: %v", err)
	}
}

func TestLoginResetsFailures(t *testing.T) {
	svc, repo := setup()
	ctx := context.Background()
	svc.Register(ctx, "bob@x.io", "password1")
	svc.Login(ctx, "bob@x.io", "nope")
	if _, err := svc.Login(ctx, "bob@x.io", "password1"); err != nil {
		t.Fatal(err)
	}
	if repo.users["bob@x.io"].Failures != 0 || repo.resets != 1 {
		t.Errorf("после успешного входа счётчик сбрасывается: failures=%d resets=%d", repo.users["bob@x.io"].Failures, repo.resets)
	}
}
@@ solution
package solution

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalid        = errors.New("некорректные данные")
	ErrEmailTaken     = errors.New("email уже занят")
	ErrNotFound       = errors.New("не найдено")
	ErrBadCredentials = errors.New("неверный email или пароль")
	ErrLocked         = errors.New("аккаунт заблокирован")
)

const MaxFailures = 3

type User struct {
	ID        int
	Email     string
	PassHash  string
	CreatedAt time.Time
	Failures  int
}

type UserRepo interface {
	ByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, u *User) error
	RecordFailure(ctx context.Context, id int) error
	ResetFailures(ctx context.Context, id int) error
}

type Hasher interface {
	Hash(password string) string
	Verify(hash, password string) bool
}

type Clock interface{ Now() time.Time }

type UserService struct {
	repo   UserRepo
	hasher Hasher
	clock  Clock
}

func NewUserService(r UserRepo, h Hasher, c Clock) *UserService {
	return &UserService{repo: r, hasher: h, clock: c}
}

func normalize(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func (s *UserService) Register(ctx context.Context, email, password string) (User, error) {
	email = normalize(email)
	if i := strings.Index(email, "@"); i <= 0 || i == len(email)-1 || len(password) < 8 {
		return User{}, ErrInvalid
	}
	_, err := s.repo.ByEmail(ctx, email)
	switch {
	case err == nil:
		return User{}, fmt.Errorf("регистрация %s: %w", email, ErrEmailTaken)
	case !errors.Is(err, ErrNotFound):
		return User{}, fmt.Errorf("поиск пользователя: %w", err)
	}
	u := User{Email: email, PassHash: s.hasher.Hash(password), CreatedAt: s.clock.Now()}
	if err := s.repo.Create(ctx, &u); err != nil {
		return User{}, fmt.Errorf("сохранение пользователя: %w", err)
	}
	return u, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (User, error) {
	u, err := s.repo.ByEmail(ctx, normalize(email))
	if errors.Is(err, ErrNotFound) {
		return User{}, ErrBadCredentials
	}
	if err != nil {
		return User{}, fmt.Errorf("поиск пользователя: %w", err)
	}
	if u.Failures >= MaxFailures {
		return User{}, ErrLocked
	}
	if !s.hasher.Verify(u.PassHash, password) {
		if err := s.repo.RecordFailure(ctx, u.ID); err != nil {
			return User{}, fmt.Errorf("учёт неудачи: %w", err)
		}
		return User{}, ErrBadCredentials
	}
	if u.Failures > 0 {
		if err := s.repo.ResetFailures(ctx, u.ID); err != nil {
			return User{}, fmt.Errorf("сброс счётчика: %w", err)
		}
		u.Failures = 0
	}
	return u, nil
}
@@ hint
Конструктор сохраняет три интерфейса в полях. Различай `ErrNotFound` (нормальный случай при регистрации) и прочие ошибки репозитория (их нужно вернуть).
