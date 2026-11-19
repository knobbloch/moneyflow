package api

import (
	"testing"

	. "moneyflow/qa-automation/helpers"
)

// pair создаёт два счёта в одной валюте.
func pair(t *testing.T, fromBalance int64) (Account, Account) {
	t.Helper()
	return CreateAccount(t, "RUB", fromBalance), CreateAccount(t, "RUB", 0)
}

func TestTC_TRF_001_SuccessfulTransfer(t *testing.T) {
	a, b := pair(t, 100000)
	r := PostTransfer(t, NewKey(), a.ID, b.ID, 5000)
	AssertStatus(t, r, 201)
	var tr Transfer
	r.Decode(t, &tr)
	if tr.Amount != 5000 || tr.Currency != "RUB" || tr.Status != "COMPLETED" ||
		tr.FromAccountID != a.ID || tr.ToAccountID != b.ID {
		t.Fatalf("unexpected transfer: %+v", tr)
	}
	if r.Header.Get("Location") != "/api/v1/transfers/"+tr.ID {
		t.Fatalf("Location: got %q", r.Header.Get("Location"))
	}
	AssertBalance(t, a.ID, 95000)
	AssertBalance(t, b.ID, 5000)
}

func TestTC_TRF_002_CurrencyMismatch(t *testing.T) {
	a := CreateAccount(t, "RUB", 1000)
	b := CreateAccount(t, "USD", 0)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 100), 422, "CURRENCY_MISMATCH")
	AssertBalance(t, a.ID, 1000)
	AssertBalance(t, b.ID, 0)
}

func TestTC_TRF_003_InsufficientFunds(t *testing.T) {
	a, b := pair(t, 1000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 1001), 422, "INSUFFICIENT_FUNDS")
	AssertBalance(t, a.ID, 1000)
	AssertBalance(t, b.ID, 0)
}

func TestTC_TRF_004_NegativeAmount(t *testing.T) {
	a, b := pair(t, 1000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, -1), 400, "VALIDATION_ERROR")
	AssertBalance(t, a.ID, 1000)
}

func TestTC_TRF_005_ZeroAmount(t *testing.T) {
	a, b := pair(t, 1000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 0), 400, "VALIDATION_ERROR")
}

func TestTC_TRF_006_MinAmount(t *testing.T) {
	a, b := pair(t, 1000)
	AssertStatus(t, PostTransfer(t, NewKey(), a.ID, b.ID, 1), 201)
	AssertBalance(t, a.ID, 999)
	AssertBalance(t, b.ID, 1)
}

func TestTC_TRF_007_MaxAmount(t *testing.T) {
	a, b := pair(t, 1_000_000_000)
	AssertStatus(t, PostTransfer(t, NewKey(), a.ID, b.ID, 1_000_000_000), 201)
	AssertBalance(t, a.ID, 0)
	AssertBalance(t, b.ID, 1_000_000_000)
}

func TestTC_TRF_008_AboveMaxAmount(t *testing.T) {
	a, b := pair(t, 2_000_000_000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 1_000_000_001), 400, "VALIDATION_ERROR")
	AssertBalance(t, a.ID, 2_000_000_000)
}

func TestTC_TRF_009_SameAccount(t *testing.T) {
	a := CreateAccount(t, "RUB", 1000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, a.ID, 100), 400, "VALIDATION_ERROR")
	AssertBalance(t, a.ID, 1000)
}

func TestTC_TRF_010_RecipientNotFound(t *testing.T) {
	a := CreateAccount(t, "RUB", 1000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, NewUUID(), 100), 404, "ACCOUNT_NOT_FOUND")
	AssertBalance(t, a.ID, 1000)
}

func TestTC_TRF_011_SenderBlocked(t *testing.T) {
	a, b := pair(t, 1000)
	BlockAccount(t, a.ID)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 100), 422, "ACCOUNT_BLOCKED")
	AssertBalance(t, a.ID, 1000)
	AssertBalance(t, b.ID, 0)
}

func TestTC_TRF_012_RecipientBlocked(t *testing.T) {
	a, b := pair(t, 1000)
	BlockAccount(t, b.ID)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, b.ID, 100), 422, "ACCOUNT_BLOCKED")
	AssertBalance(t, a.ID, 1000)
	AssertBalance(t, b.ID, 0)
}

func TestTC_TRF_013_MissingIdempotencyKey(t *testing.T) {
	a, b := pair(t, 1000)
	AssertError(t, PostTransfer(t, "", a.ID, b.ID, 100), 400, "MISSING_IDEMPOTENCY_KEY")
	AssertBalance(t, a.ID, 1000)
}

func TestTC_TRF_014_IdempotentReplay(t *testing.T) {
	a, b := pair(t, 100000)
	key := NewKey()
	r1 := PostTransfer(t, key, a.ID, b.ID, 5000)
	AssertStatus(t, r1, 201)
	r2 := PostTransfer(t, key, a.ID, b.ID, 5000)
	AssertStatus(t, r2, 200)
	if r2.Header.Get("Idempotent-Replay") != "true" {
		t.Fatalf("missing Idempotent-Replay header")
	}
	var t1, t2 Transfer
	r1.Decode(t, &t1)
	r2.Decode(t, &t2)
	if t1.ID != t2.ID {
		t.Fatalf("replay returned different transfer: %s vs %s", t1.ID, t2.ID)
	}
	AssertBalance(t, a.ID, 95000)
	AssertBalance(t, b.ID, 5000)
}

func TestTC_TRF_015_IdempotencyKeyConflict(t *testing.T) {
	a, b := pair(t, 100000)
	key := NewKey()
	AssertStatus(t, PostTransfer(t, key, a.ID, b.ID, 5000), 201)
	AssertError(t, PostTransfer(t, key, a.ID, b.ID, 1), 409, "IDEMPOTENCY_KEY_CONFLICT")
	AssertBalance(t, a.ID, 95000)
}

func TestTC_TRF_016_TransferHistory(t *testing.T) {
	a, b := pair(t, 10000)
	var last Transfer
	for i := 0; i < 3; i++ {
		r := PostTransfer(t, NewKey(), a.ID, b.ID, 100)
		AssertStatus(t, r, 201)
		r.Decode(t, &last)
	}
	r := Do(t, "GET", "/api/v1/accounts/"+a.ID+"/transfers", nil, nil)
	AssertStatus(t, r, 200)
	var page struct {
		Items  []Transfer `json:"items"`
		Limit  int        `json:"limit"`
		Offset int        `json:"offset"`
	}
	r.Decode(t, &page)
	if len(page.Items) != 3 || page.Limit != 20 || page.Offset != 0 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if page.Items[0].ID != last.ID {
		t.Fatalf("newest transfer must be first")
	}

	// получатель видит те же переводы
	r = Do(t, "GET", "/api/v1/accounts/"+b.ID+"/transfers?limit=2", nil, nil)
	AssertStatus(t, r, 200)
	r.Decode(t, &page)
	if len(page.Items) != 2 {
		t.Fatalf("limit=2: got %d items", len(page.Items))
	}
}

func TestTC_TRF_017_HistoryInvalidPaging(t *testing.T) {
	a := CreateAccount(t, "RUB", 0)
	for _, q := range []string{"?limit=0", "?limit=101", "?offset=-1", "?limit=abc"} {
		AssertError(t, Do(t, "GET", "/api/v1/accounts/"+a.ID+"/transfers"+q, nil, nil), 400, "VALIDATION_ERROR")
	}
}