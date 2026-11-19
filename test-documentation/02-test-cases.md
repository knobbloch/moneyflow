# Test Cases

ID Format: `TC-<module>-<number>`.  
Priority: P1 (Critical), P2 (Medium), P3 (Low).

## Accounts

| ID | Title | Steps | Expected Result | Priority |
|---|---|---|---|---|
| TC-ACC-001 | Create an account | POST /accounts {Ivan, RUB, 100000} | HTTP 201, status=ACTIVE, balance=100000, Location header is present | P1 |
| TC-ACC-002 | Create an account without initial_balance | POST /accounts {Ivan, RUB} | HTTP 201, balance=0 | P2 |
| TC-ACC-003 | Empty account owner name | POST /accounts {"", RUB} | HTTP 400 VALIDATION_ERROR | P2 |
| TC-ACC-004 | Unsupported currency | POST /accounts {Ivan, XXX} | HTTP 400 VALIDATION_ERROR | P2 |
| TC-ACC-005 | Negative initial balance | initial_balance=-1 | HTTP 400 VALIDATION_ERROR | P1 |
| TC-ACC-006 | Boundary: balance = 1,000,000,000,000 | initial_balance=1000000000000 | HTTP 201 | P3 |
| TC-ACC-007 | Boundary: balance = 1,000,000,000,001 | initial_balance=1000000000001 | HTTP 400 | P3 |
| TC-ACC-008 | Retrieve an account | GET /accounts/{id} | HTTP 200 | P1 |
| TC-ACC-009 | Invalid UUID | GET /accounts/not-a-uuid | HTTP 400 INVALID_ID | P2 |
| TC-ACC-010 | Non-existent account | GET /accounts/{random uuid} | HTTP 404 ACCOUNT_NOT_FOUND | P2 |
| TC-ACC-011 | Block an account | PATCH /accounts/{id}/block | HTTP 200, status=BLOCKED | P1 |
| TC-ACC-012 | Block an already blocked account | PATCH /accounts/{id}/block twice | HTTP 409 ACCOUNT_ALREADY_BLOCKED | P2 |

## Transfers

| ID | Title | Steps | Expected Result | Priority |
|---|---|---|---|---|
| TC-TRF-001 | Successful money transfer | A(RUB,100000) -> B(RUB,0), amount=5000 | HTTP 201; A=95000, B=5000 | P1 |
| TC-TRF-002 | Transfer between accounts with different currencies | A(RUB) -> B(USD) | HTTP 422 CURRENCY_MISMATCH, balances remain unchanged | P1 |
| TC-TRF-003 | Insufficient funds | amount > balance | HTTP 422 INSUFFICIENT_FUNDS, balances remain unchanged | P1 |
| TC-TRF-004 | Negative transfer amount | amount=-1 | HTTP 400 VALIDATION_ERROR | P1 |
| TC-TRF-005 | Zero transfer amount | amount=0 | HTTP 400 VALIDATION_ERROR | P1 |
| TC-TRF-006 | Minimum valid transfer amount | amount=1 | HTTP 201 | P2 |
| TC-TRF-007 | Maximum allowed transfer amount | amount=1000000000 (sufficient funds available) | HTTP 201 | P2 |
| TC-TRF-008 | Transfer amount exceeds the maximum | amount=1000000001 | HTTP 400 | P2 |
| TC-TRF-009 | Transfer to the same account | from == to | HTTP 400 VALIDATION_ERROR | P1 |
| TC-TRF-010 | Transfer to a non-existent recipient | to = random uuid | HTTP 404 ACCOUNT_NOT_FOUND | P1 |
| TC-TRF-011 | Transfer from a blocked account | from account status = BLOCKED | HTTP 422 ACCOUNT_BLOCKED | P1 |
| TC-TRF-012 | Transfer to a blocked account | to account status = BLOCKED | HTTP 422 ACCOUNT_BLOCKED | P1 |
| TC-TRF-013 | Missing Idempotency-Key header | Request without the header | HTTP 400 MISSING_IDEMPOTENCY_KEY | P1 |
| TC-TRF-014 | Retry with the same key and request body | Send two identical requests | Second request: HTTP 200, Idempotent-Replay: true, amount debited only once | P1 |
| TC-TRF-015 | Same key with a different transfer amount | First request: k1 with amount=5000; second request: k1 with amount=1 | HTTP 409 IDEMPOTENCY_KEY_CONFLICT | P1 |
| TC-TRF-016 | Retrieve transfer history | GET /accounts/{id}/transfers | HTTP 200, transfers sorted from newest to oldest | P2 |
| TC-TRF-017 | Limit parameter out of range | limit=0 and limit=101 | HTTP 400 VALIDATION_ERROR | P3 |
| TC-TRF-018 | Concurrent money transfers | 50 simultaneous transfers of 1000 from an account with a balance of 10000 | Exactly 10 successful transfers, balance >= 0, total balance across accounts remains unchanged | P1 |
| TC-TRF-019 | Concurrent requests with the same idempotency key | Send 20 requests using the same key | Exactly one transfer is created | P1 |