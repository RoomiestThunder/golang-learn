---
title: Банковский счёт: методы и embedding
level: medium
---
@@ statement
Смоделируй счета.

`Account`:
- неэкспортируемые поля `owner string`, `balance int` (в копейках);
- `NewAccount(owner string, initial int) (*Account, error)` — ошибка `ErrInvalidAmount`, если `owner` пуст или `initial < 0`;
- `Deposit(amount int) error` — `amount > 0`, иначе `ErrInvalidAmount`;
- `Withdraw(amount int) error` — `ErrInvalidAmount` для `amount <= 0`, `ErrInsufficientFunds`, если не хватает денег (баланс не меняется);
- `Balance() int`, `Owner() string`.

`SavingsAccount` **встраивает** `Account` и добавляет `rate int` (процент в год):
- `NewSavings(owner string, initial, rate int) (*SavingsAccount, error)`;
- `AddInterest()` — начислить `balance * rate / 100` (целочисленно, вниз).

`Transfer(from, to *Account, amount int) error` — списать у `from` и зачислить `to`. При ошибке списания ничего не меняется. `ErrInsufficientFunds` и `ErrInvalidAmount` должны находиться через `errors.Is` даже если обёрнуты.
@@ starter
package solution

import "errors"

var (
	ErrInvalidAmount     = errors.New("некорректная сумма")
	ErrInsufficientFunds = errors.New("недостаточно средств")
)

type Account struct {
	// TODO
}

func NewAccount(owner string, initial int) (*Account, error) { return nil, nil }
func (a *Account) Deposit(amount int) error                  { return nil }
func (a *Account) Withdraw(amount int) error                 { return nil }
func (a *Account) Balance() int                              { return 0 }
func (a *Account) Owner() string                             { return "" }

type SavingsAccount struct {
	// TODO: встроить Account
}

func NewSavings(owner string, initial, rate int) (*SavingsAccount, error) { return nil, nil }
func (s *SavingsAccount) AddInterest()                                    {}

func Transfer(from, to *Account, amount int) error { return nil }
@@ tests
package solution

import (
	"errors"
	"testing"
)

func TestAccount(t *testing.T) {
	a, err := NewAccount("Ann", 1000)
	if err != nil || a.Balance() != 1000 || a.Owner() != "Ann" {
		t.Fatalf("NewAccount: %v %v", a, err)
	}
	if _, err := NewAccount("", 5); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("пустое имя: %v", err)
	}
	if _, err := NewAccount("x", -1); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("отрицательный баланс: %v", err)
	}
	if err := a.Deposit(500); err != nil || a.Balance() != 1500 {
		t.Errorf("Deposit: %v %d", err, a.Balance())
	}
	if err := a.Deposit(0); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("Deposit(0): %v", err)
	}
	if err := a.Withdraw(2000); !errors.Is(err, ErrInsufficientFunds) || a.Balance() != 1500 {
		t.Errorf("Withdraw(2000): %v, баланс %d", err, a.Balance())
	}
	if err := a.Withdraw(-5); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("Withdraw(-5): %v", err)
	}
	if err := a.Withdraw(1500); err != nil || a.Balance() != 0 {
		t.Errorf("Withdraw(1500): %v %d", err, a.Balance())
	}
}

func TestSavings(t *testing.T) {
	s, err := NewSavings("Bob", 10_000, 7)
	if err != nil {
		t.Fatal(err)
	}
	s.AddInterest()
	if s.Balance() != 10_700 { // методы Account продвинуты через embedding
		t.Errorf("баланс после процентов = %d", s.Balance())
	}
	s.Deposit(1)
	if s.Owner() != "Bob" || s.Balance() != 10_701 {
		t.Errorf("owner=%q balance=%d", s.Owner(), s.Balance())
	}
}

func TestTransfer(t *testing.T) {
	a, _ := NewAccount("a", 100)
	b, _ := NewAccount("b", 0)
	if err := Transfer(a, b, 70); err != nil || a.Balance() != 30 || b.Balance() != 70 {
		t.Fatalf("перевод: %v %d %d", err, a.Balance(), b.Balance())
	}
	err := Transfer(a, b, 50)
	if !errors.Is(err, ErrInsufficientFunds) || a.Balance() != 30 || b.Balance() != 70 {
		t.Errorf("неудачный перевод изменил балансы: %v %d %d", err, a.Balance(), b.Balance())
	}
	if err := Transfer(a, b, 0); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("нулевой перевод: %v", err)
	}
}
@@ solution
package solution

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAmount     = errors.New("некорректная сумма")
	ErrInsufficientFunds = errors.New("недостаточно средств")
)

type Account struct {
	owner   string
	balance int
}

func NewAccount(owner string, initial int) (*Account, error) {
	if owner == "" || initial < 0 {
		return nil, ErrInvalidAmount
	}
	return &Account{owner: owner, balance: initial}, nil
}

func (a *Account) Deposit(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.balance += amount
	return nil
}

func (a *Account) Withdraw(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.balance {
		return ErrInsufficientFunds
	}
	a.balance -= amount
	return nil
}

func (a *Account) Balance() int  { return a.balance }
func (a *Account) Owner() string { return a.owner }

type SavingsAccount struct {
	Account
	rate int
}

func NewSavings(owner string, initial, rate int) (*SavingsAccount, error) {
	acc, err := NewAccount(owner, initial)
	if err != nil {
		return nil, err
	}
	return &SavingsAccount{Account: *acc, rate: rate}, nil
}

func (s *SavingsAccount) AddInterest() { s.balance += s.balance * s.rate / 100 }

func Transfer(from, to *Account, amount int) error {
	if err := from.Withdraw(amount); err != nil {
		return fmt.Errorf("перевод: %w", err)
	}
	if err := to.Deposit(amount); err != nil {
		from.balance += amount // откат
		return fmt.Errorf("перевод: %w", err)
	}
	return nil
}
@@ hint
`SavingsAccount` объявляется как `struct { Account; rate int }`. Поля встроенного `Account` доступны внутри пакета напрямую (`s.balance`).
