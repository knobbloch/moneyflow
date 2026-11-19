# Test Plan

## 1. Objective

Verify that the Banking API complies with `01-specification.md`, including correct account and transfer operations, financial data integrity, idempotency, and behavior under concurrent requests.

## 2. Scope

**In Scope:**
- All endpoints defined in the specification
- Input validation
- HTTP status codes and error response formats
- Database state verification after operations
- Concurrent money transfers

**Out of Scope:**
- Performance and load testing
- Security testing (the API does not implement authentication)
- UI testing

## 3. Testing Approach

| Level | Method | Location |
|---|---|---|
| Smoke Testing | Manual testing using curl after each API startup | Run 1 |
| API Testing | Go automated tests sending HTTP requests to the running API | qa-automation/tests/api |
| Database Testing | Go tests verifying database state after operations | qa-automation/tests/database |
| Integration/Concurrency Testing | Go tests involving multiple requests and concurrent goroutines | qa-automation/tests/integration |
| Exploratory Testing | Postman-based exploratory testing to identify defects beyond predefined test cases | postman/ |

## 4. Priorities and Test Execution Order

1. Smoke testing (6 test cases, Run 1).
2. All P1 test cases from `02-test-cases.md`.
3. P2 test cases, followed by P3.
4. Concurrent test cases (TC-TRF-018, TC-TRF-019) are executed separately and repeated 10 times to identify potential instability.

## 5. Test Environment

- Go 1.22+
- PostgreSQL 16 (Docker Compose)
- API: `:8080`
- Database: `:5432`

Before a full test run, the database environment is recreated using:

```bash
docker compose down -v && docker compose up -d
```

## 6. Test Execution

- **Manual Smoke Testing:** curl commands (will be added to `postman/`).
- **Automated Tests