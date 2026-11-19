package api

import (
	"fmt"
	"strings"
	"testing"

	. "moneyflow/qa-automation/helpers"
)

func postTransferRaw(t *testing.T, raw string) Response {
	t.Helper()
	return Do(t, "POST", "/api/v1/transfers", map[string]string{"Idempotency-Key": NewKey()}, raw)
}

func TestTC_TRF_021_MalformedJSON(t *testing.T) {
	AssertError(t, postTransferRaw(t, `{"amount": `), 400, "INVALID_JSON")
}

func TestTC_TRF_022_UnknownField(t *testing.T) {
	a, b := pair(t, 1000)
	raw := fmt.Sprintf(`{"from_account_id":"%s","to_account_id":"%s","amount":100,"foo":1}`, a.ID, b.ID)
	AssertError(t, postTransferRaw(t, raw), 400, "INVALID_JSON")
	AssertBalance(t, a.ID, 1000)
}

func TestTC_TRF_023_InvalidUUIDInBody(t *testing.T) {
	_, b := pair(t, 1000)
	AssertError(t, PostTransfer(t, NewKey(), "not-a-uuid", b.ID, 100), 400, "VALIDATION_ERROR")
	a := CreateAccount(t, "RUB", 1000)
	AssertError(t, PostTransfer(t, NewKey(), a.ID, "123", 100), 400, "VALIDATION_ERROR")
}

func TestTC_TRF_024_GetTransfer(t *testing.T) {
	a, b := pair(t, 1000)
	r := PostTransfer(t, NewKey(), a.ID, b.ID, 100)
	AssertStatus(t, r, 201)
	var created Transfer
	r.Decode(t, &created)

	g := Do(t, "GET", "/api/v1/transfers/"+created.ID, nil, nil)
	AssertStatus(t, g, 200)
	var got Transfer
	g.Decode(t, &got)
	if got.ID != created.ID || got.Amount != 100 || got.FromAccountID != a.ID || got.ToAccountID != b.ID {
		t.Fatalf("unexpected transfer: %+v", got)
	}
}

func TestTC_TRF_025_GetTransferNotFound(t *testing.T) {
	AssertError(t, Do(t, "GET", "/api/v1/transfers/"+NewUUID(), nil, nil), 404, "TRANSFER_NOT_FOUND")
}

func TestTC_TRF_026_GetTransferInvalidID(t *testing.T) {
	AssertError(t, Do(t, "GET", "/api/v1/transfers/abc", nil, nil), 400, "INVALID_ID")
}

func TestTC_TRF_027_HistoryOfMissingAccount(t *testing.T) {
	AssertError(t, Do(t, "GET", "/api/v1/accounts/"+NewUUID()+"/transfers", nil, nil), 404, "ACCOUNT_NOT_FOUND")
}

func TestTC_TRF_028_HistoryOffsetBeyondEnd(t *testing.T) {
	a, b := pair(t, 1000)
	AssertStatus(t, PostTransfer(t, NewKey(), a.ID, b.ID, 100), 201)
	r := Do(t, "GET", "/api/v1/accounts/"+a.ID+"/transfers?offset=1000", nil, nil)
	AssertStatus(t, r, 200)
	if !strings.Contains(string(r.Body), `"items":[]`) {
		t.Fatalf("items must be an empty array, got: %s", r.Body)
	}
}

func TestTC_TRF_029_UppercaseUUIDs(t *testing.T) {
	a, b := pair(t, 1000)
	r := PostTransfer(t, NewKey(), strings.ToUpper(a.ID), strings.ToUpper(b.ID), 100)
	AssertStatus(t, r, 201)
	var tr Transfer
	r.Decode(t, &tr)
	if tr.FromAccountID != a.ID || tr.ToAccountID != b.ID {
		t.Fatalf("ids must be returned in lowercase: %+v", tr)
	}
	AssertBalance(t, a.ID, 900)
	AssertBalance(t, b.ID, 100)
}

func TestTC_TRF_030_IdempotencyKeyLength(t *testing.T) {
	a, b := pair(t, 1000)
	k := NewKey()
	k64 := k + strings.Repeat("x", 64-len(k))
	AssertStatus(t, PostTransfer(t, k64, a.ID, b.ID, 100), 201)
	AssertError(t, PostTransfer(t, k64+"x", a.ID, b.ID, 100), 400, "MISSING_IDEMPOTENCY_KEY")
	AssertBalance(t, a.ID, 900)
}

func TestTC_TRF_031_SenderNotFound(t *testing.T) {
	b := CreateAccount(t, "RUB", 0)
	AssertError(t, PostTransfer(t, NewKey(), NewUUID(), b.ID, 100), 404, "ACCOUNT_NOT_FOUND")
	AssertBalance(t, b.ID, 0)
}

func TestTC_TRF_032_AmountAsString(t *testing.T) {
	a, b := pair(t, 1000)
	raw := fmt.Sprintf(`{"from_account_id":"%s","to_account_id":"%s","amount":"100"}`, a.ID, b.ID)
	AssertError(t, postTransferRaw(t, raw), 400, "INVALID_JSON")
	AssertBalance(t, a.ID, 1000)
}