package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

type Account struct {
	ID        string    `json:"id"`
	OwnerName string    `json:"owner_name"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Transfer struct {
	ID             string    `json:"id"`
	FromAccountID  string    `json:"from_account_id"`
	ToAccountID    string    `json:"to_account_id"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewKey() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "qa-" + hex.EncodeToString(b)
}

func NewUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func CreateAccount(t testing.TB, currency string, balance int64) Account {
	t.Helper()
	r := Do(t, "POST", "/api/v1/accounts", nil, map[string]any{
		"owner_name":      "QA",
		"currency":        currency,
		"initial_balance": balance,
	})
	AssertStatus(t, r, 201)
	var a Account
	r.Decode(t, &a)
	return a
}

func BlockAccount(t testing.TB, id string) {
	t.Helper()
	AssertStatus(t, Do(t, "PATCH", "/api/v1/accounts/"+id+"/block", nil, nil), 200)
}

func Balance(t testing.TB, id string) int64 {
	t.Helper()
	r := Do(t, "GET", "/api/v1/accounts/"+id+"/balance", nil, nil)
	AssertStatus(t, r, 200)
	var b struct {
		Balance int64 `json:"balance"`
	}
	r.Decode(t, &b)
	return b.Balance
}

func AssertBalance(t testing.TB, id string, want int64) {
	t.Helper()
	if got := Balance(t, id); got != want {
		t.Fatalf("balance of %s: got %d, want %d", id, got, want)
	}
}

func transferBody(from, to string, amount int64) map[string]any {
	return map[string]any{"from_account_id": from, "to_account_id": to, "amount": amount}
}

// SendTransfer безопасен для горутин (без testing.T).
func SendTransfer(key, from, to string, amount int64) (Response, error) {
	headers := map[string]string{}
	if key != "" {
		headers["Idempotency-Key"] = key
	}
	return Send("POST", "/api/v1/transfers", headers, transferBody(from, to, amount))
}

func PostTransfer(t testing.TB, key, from, to string, amount int64) Response {
	t.Helper()
	r, err := SendTransfer(key, from, to, amount)
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	return r
}