package api

import (
	"testing"

	. "moneyflow/qa-automation/helpers"
)

func TestTC_ACC_001_CreateAccount(t *testing.T) {
	r := Do(t, "POST", "/api/v1/accounts", nil, map[string]any{
		"owner_name": "Ivan", "currency": "RUB", "initial_balance": 100000,
	})
	AssertStatus(t, r, 201)
	var a Account
	r.Decode(t, &a)
	if a.ID == "" || a.Status != "ACTIVE" || a.Balance != 100000 || a.Currency != "RUB" || a.OwnerName != "Ivan" {
		t.Fatalf("unexpected account: %+v", a)
	}
	if loc := r.Header.Get("Location"); loc != "/api/v1/accounts/"+a.ID {
		t.Fatalf("Location: got %q", loc)
	}
}

func TestTC_ACC_002_CreateWithoutInitialBalance(t *testing.T) {
	r := Do(t, "POST", "/api/v1/accounts", nil, map[string]any{"owner_name": "Ivan", "currency": "RUB"})
	AssertStatus(t, r, 201)
	var a Account
	r.Decode(t, &a)
	if a.Balance != 0 {
		t.Fatalf("balance: got %d, want 0", a.Balance)
	}
}

func createInvalid(t *testing.T, body map[string]any) {
	t.Helper()
	AssertError(t, Do(t, "POST", "/api/v1/accounts", nil, body), 400, "VALIDATION_ERROR")
}

func TestTC_ACC_003_EmptyName(t *testing.T) {
	createInvalid(t, map[string]any{"owner_name": "", "currency": "RUB"})
}

func TestTC_ACC_004_UnsupportedCurrency(t *testing.T) {
	createInvalid(t, map[string]any{"owner_name": "Ivan", "currency": "XXX"})
}

func TestTC_ACC_005_NegativeInitialBalance(t *testing.T) {
	createInvalid(t, map[string]any{"owner_name": "Ivan", "currency": "RUB", "initial_balance": -1})
}

func TestTC_ACC_006_MaxInitialBalance(t *testing.T) {
	r := Do(t, "POST", "/api/v1/accounts", nil, map[string]any{
		"owner_name": "Ivan", "currency": "RUB", "initial_balance": 1_000_000_000_000,
	})
	AssertStatus(t, r, 201)
}

func TestTC_ACC_007_AboveMaxInitialBalance(t *testing.T) {
	createInvalid(t, map[string]any{"owner_name": "Ivan", "currency": "RUB", "initial_balance": 1_000_000_000_001})
}

func TestTC_ACC_008_GetAccount(t *testing.T) {
	a := CreateAccount(t, "USD", 500)
	r := Do(t, "GET", "/api/v1/accounts/"+a.ID, nil, nil)
	AssertStatus(t, r, 200)
	var got Account
	r.Decode(t, &got)
	if got.ID != a.ID || got.Currency != "USD" || got.Balance != 500 {
		t.Fatalf("unexpected account: %+v", got)
	}
}

func TestTC_ACC_009_InvalidUUID(t *testing.T) {
	AssertError(t, Do(t, "GET", "/api/v1/accounts/not-a-uuid", nil, nil), 400, "INVALID_ID")
}

func TestTC_ACC_010_AccountNotFound(t *testing.T) {
	AssertError(t, Do(t, "GET", "/api/v1/accounts/"+NewUUID(), nil, nil), 404, "ACCOUNT_NOT_FOUND")
}

func TestTC_ACC_011_BlockAccount(t *testing.T) {
	a := CreateAccount(t, "RUB", 0)
	r := Do(t, "PATCH", "/api/v1/accounts/"+a.ID+"/block", nil, nil)
	AssertStatus(t, r, 200)
	var got Account
	r.Decode(t, &got)
	if got.Status != "BLOCKED" {
		t.Fatalf("status: got %s, want BLOCKED", got.Status)
	}
}

func TestTC_ACC_012_BlockTwice(t *testing.T) {
	a := CreateAccount(t, "RUB", 0)
	BlockAccount(t, a.ID)
	AssertError(t, Do(t, "PATCH", "/api/v1/accounts/"+a.ID+"/block", nil, nil), 409, "ACCOUNT_ALREADY_BLOCKED")
}