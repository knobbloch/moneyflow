# MoneyFlow: Specification v1

## General Rules

- Money is stored as an integer in the smallest currency unit (kopecks/cents). For example, 1050 = 10.50.
- Supported currencies: RUB, USD, EUR. Transfers are allowed only between accounts with the same currency.
- IDs use the UUID format.
- Error response format:

```json
{
  "error": {
    "code": "SOME_CODE",
    "message": "text"
  }
}
```

## HTTP Response Codes

| Code | Description |
|---|---|
| 200 | Successful read or update operation |
| 201 | Resource successfully created |
| 400 | Invalid JSON, malformed ID, or field validation failure |
| 404 | Resource not found |
| 409 | State conflict (e.g., account is already blocked) |
| 422 | Business rule violation (money transfer) |

## Account

Fields: `id`, `owner_name`, `currency`, `balance`, `status` (`ACTIVE` \| `BLOCKED`), `created_at`.

### POST /api/v1/accounts

Request body: `owner_name`, `currency`, `initial_balance` (optional, defaults to 0).

Validation rules:
- `owner_name`: 1–100 characters after trimming leading and trailing whitespace.
- `currency`: `RUB` | `USD` | `EUR`.
- `initial_balance`: 0–1,000,000,000,000.
- Unknown JSON fields result in HTTP 400.

Successful response:
- HTTP 201.
- Account status: `ACTIVE`.
- `Location` response header is included.

### GET /api/v1/accounts/{id}

- HTTP 200 + account object.
- HTTP 404: `ACCOUNT_NOT_FOUND`.
- HTTP 400: `INVALID_ID`.

### GET /api/v1/accounts/{id}/balance

HTTP 200 response:

```json
{
  "account_id": "...",
  "balance": 0,
  "currency": "RUB"
}
```

### PATCH /api/v1/accounts/{id}/block

- `ACTIVE` → `BLOCKED`: HTTP 200 + account object.
- Already `BLOCKED` → HTTP 409: `ACCOUNT_ALREADY_BLOCKED`.
- Account does not exist → HTTP 404.

## Transfer

### POST /api/v1/transfers

The `Idempotency-Key` request header is required (1–64 characters).

Request body:
- `from_account_id`
- `to_account_id`
- `amount` (integer, 1–1,000,000,000).

| Scenario | Response |
|---|---|
| Successful transfer | HTTP 201 + transfer object, `Location` header |
| Retry with the same key and identical request body | HTTP 200 + original transfer, `Idempotent-Replay: true` response header |
| Same key with a different request body | HTTP 409: `IDEMPOTENCY_KEY_CONFLICT` |
| Missing key or key longer than 64 characters | HTTP 400: `MISSING_IDEMPOTENCY_KEY` |
| Invalid JSON or unknown fields | HTTP 400: `INVALID_JSON` |
| Invalid UUID, amount <= 0 or above the maximum, or `from_account_id == to_account_id` | HTTP 400: `VALIDATION_ERROR` |
| One of the accounts does not exist | HTTP 404: `ACCOUNT_NOT_FOUND` |
| Either account is `BLOCKED` | HTTP 422: `ACCOUNT_BLOCKED` |
| Account currencies do not match | HTTP 422: `CURRENCY_MISMATCH` |
| Sender's balance < amount | HTTP 422: `INSUFFICIENT_FUNDS` |

**Transaction Atomicity:**

A transfer must be atomic: both the debit and credit operations must either succeed together or both be rolled back. Partial transfers are not allowed.

### GET /api/v1/transfers/{id}

- HTTP 200: Transfer successfully retrieved.
- HTTP 400: `INVALID_ID`.
- HTTP 404: `TRANSFER_NOT_FOUND`.

### GET /api/v1/accounts/{id}/transfers?limit=&offset=

Returns transfers where the specified account is either the sender or the recipient, ordered from newest to oldest.

**Query Parameters:**
- `limit`: 1–100 (default: 20).
- `offset`: >= 0 (default: 0).
- Invalid values result in HTTP 400: `VALIDATION_ERROR`.

Response format:

```json
{
  "items": [],
  "limit": 20,
  "offset": 0
}
```

If the account does not exist, the API returns HTTP 404: `ACCOUNT_NOT_FOUND`.