package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	. "moneyflow/qa-automation/helpers"
)

func TestTC_DB_001_TransferKeepsTotalAndWritesRecord(t *testing.T) {
	db := DB(t)
	a := CreateAccount(t, "RUB", 100000)
	b := CreateAccount(t, "RUB", 0)
	key := NewKey()
	AssertStatus(t, PostTransfer(t, key, a.ID, b.ID, 5000), 201)

	var sum int64
	err := db.QueryRow(context.Background(),
		`SELECT SUM(balance) FROM accounts WHERE id IN ($1, $2)`, a.ID, b.ID).Scan(&sum)
	if err != nil {
		t.Fatal(err)
	}
	if sum != 100000 {
		t.Fatalf("sum of balances: got %d, want 100000", sum)
	}

	var amount int64
	var status string
	err = db.QueryRow(context.Background(),
		`SELECT amount, status FROM transfers WHERE idempotency_key = $1`, key).Scan(&amount, &status)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 5000 || status != "COMPLETED" {
		t.Fatalf("transfers row: amount=%d status=%s", amount, status)
	}
}

func TestTC_DB_002_FailedTransferLeavesNoTrace(t *testing.T) {
	db := DB(t)
	a := CreateAccount(t, "RUB", 1000)
	b := CreateAccount(t, "RUB", 0)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 5000), 422, "INSUFFICIENT_FUNDS")

	var n int
	err := db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM transfers WHERE from_account_id = $1`, a.ID).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("transfers rows after failed transfer: got %d, want 0", n)
	}
}

func TestTC_DB_003_NegativeBalanceRejectedByDB(t *testing.T) {
	db := DB(t)
	a := CreateAccount(t, "RUB", 0)
	_, err := db.Exec(context.Background(), `UPDATE accounts SET balance = -1 WHERE id = $1`, a.ID)
	if err == nil {
		t.Fatal("DB accepted a negative balance: CHECK constraint is missing")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("expected CHECK violation (23514), got: %v", err)
	}
}

func TestTC_DB_004_ReplayCreatesSingleRow(t *testing.T) {
	db := DB(t)
	a := CreateAccount(t, "RUB", 10000)
	b := CreateAccount(t, "RUB", 0)
	key := NewKey()
	for i := 0; i < 3; i++ {
		PostTransfer(t, key, a.ID, b.ID, 100)
	}
	var n int
	err := db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM transfers WHERE idempotency_key = $1`, key).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("transfers rows for one key: got %d, want 1", n)
	}
}