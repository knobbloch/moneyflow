# Execution Log

## Run 1: Ad-hoc Smoke Testing (curl)

An unplanned quick check performed after the first API startup.

Test cases were not selected based on priority. The main focus was to verify that core transfer and idempotency scenarios were working correctly.

The planned priority-based test execution was performed in Run 2.

- **Date:** 2026-11-18
- **Tester:** kbob
- **Environment:** Go API (local, :8080), PostgreSQL 16 (Docker, :5433)
- **Version:** commit `<git rev-parse --short HEAD>`

| ID | Result | Actual Result | Bug |
|---|---|---|---|
| TC-TRF-001 | PASS | HTTP 201, Location header present; A=95000, B=5000 | |
| TC-TRF-003 | PASS | HTTP 422 INSUFFICIENT_FUNDS; balances remained unchanged | |
| TC-TRF-013 | PASS | HTTP 400 MISSING_IDEMPOTENCY_KEY | |
| TC-TRF-014 | PASS | HTTP 200, Idempotent-Replay: true, same transfer ID; amount debited only once | |
| TC-TRF-015 | PASS | HTTP 409 IDEMPOTENCY_KEY_CONFLICT | |
| TC-TRF-002 | PASS | HTTP 422 CURRENCY_MISMATCH | |

### Summary

- **Executed:** 6
- **PASS:** 6
- **FAIL:** 0
- **NOT RUN:** All remaining test cases from `02-test-cases.md`

## Run 2: Automated Testing (go test), Priority-Based Execution

- **Date:** 2026-11-18
- **Environment:** Go API (:8080), PostgreSQL 16 (Docker, :5432)
- **Version:** commit `<git rev-parse --short HEAD>`
- **Results:** `test-documentation/run2-output.txt`

| Group | Test Cases | Result |
|---|---|---|
| Accounts | TC-ACC-001..012 | 12 PASS |
| Transfers | TC-TRF-001..017 | 17 PASS |
| Concurrency | TC-TRF-018..020 | 3 PASS, stable with `-count=10` |
| Database | TC-DB-001..004 | 4 PASS |

### Summary

- **Executed:** 36
- **PASS:** 36
- **FAIL:** 0
- All P1 test cases passed.
- No open defects.

**Observation:**

The `created_at` field is returned in the database server's local time zone rather than UTC.

Priority: P3.

See `BUG-001` (if documented).