package api

import (
	"strings"
	"testing"

	. "moneyflow/qa-automation/helpers"
)

func postAccountRaw(t *testing.T, raw string) Response {
	t.Helper()
	return Do(t, "POST", "/api/v1/accounts", nil, raw)
}

func TestTC_ACC_013_MalformedJSON(t *testing.T) {
	AssertError(t, postAccountRaw(t, `{"owner_name": `), 400, "INVALID_JSON")
}

func TestTC_ACC_014_UnknownField(t *testing.T) {
	AssertError(t, postAccountRaw(t, `{"owner_name":"Ivan","currency":"RUB","foo":1}`), 400, "INVALID_JSON")
}

func TestTC_ACC_015_EmptyBody(t *testing.T) {
	AssertError(t, postAccountRaw(t, ``), 400, "INVALID_JSON")
}

func TestTC_ACC_016_NameLength100(t *testing.T) {
	AssertStatus(t, postAccountRaw(t,
		`{"owner_name":"`+strings.Repeat("a", 100)+`","currency":"RUB"}`), 201)
}

func TestTC_ACC_017_NameLength101(t *testing.T) {
	AssertError(t, postAccountRaw(t,
		`{"owner_name":"`+strings.Repeat("a", 101)+`","currency":"RUB"}`), 400, "VALIDATION_ERROR")
}

func TestTC_ACC_018_NameOnlySpaces(t *testing.T) {
	createInvalid(t, map[string]any{"owner_name": "     ", "currency": "RUB"})
}

func TestTC_ACC_019_BalanceOfMissingAccount(t *testing.T) {
	AssertError(t, Do(t, "GET", "/api/v1/accounts/"+NewUUID()+"/balance", nil, nil), 404, "ACCOUNT_NOT_FOUND")
}

func TestTC_ACC_020_BlockMissingAccount(t *testing.T) {
	AssertError(t, Do(t, "PATCH", "/api/v1/accounts/"+NewUUID()+"/block", nil, nil), 404, "ACCOUNT_NOT_FOUND")
}

func TestTC_ACC_021_UppercaseUUID(t *testing.T) {
	a := CreateAccount(t, "RUB", 700)
	r := Do(t, "GET", "/api/v1/accounts/"+strings.ToUpper(a.ID), nil, nil)
	AssertStatus(t, r, 200)
	var got Account
	r.Decode(t, &got)
	if got.ID != a.ID {
		t.Fatalf("id: got %s, want %s", got.ID, a.ID)
	}
}

func TestTC_ACC_022_TrailingDataAfterJSON(t *testing.T) {
	AssertError(t, postAccountRaw(t,
		`{"owner_name":"Ivan","currency":"RUB"} garbage`), 400, "INVALID_JSON")
}

func TestTC_ACC_023_CreatedAtIsUTC(t *testing.T) {
	a := CreateAccount(t, "RUB", 0)
	if _, offset := a.CreatedAt.Zone(); offset != 0 {
		t.Fatalf("created_at is not UTC, offset %d s: %s", offset, a.CreatedAt)
	}
}