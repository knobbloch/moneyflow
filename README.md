# MoneyFlow

**Backend API & QA Automation Portfolio Project**

MoneyFlow is a demo banking application designed to demonstrate backend testing and QA automation skills using Go, REST API, and PostgreSQL.

The project simulates basic banking operations, including account management and money transfers. Its main purpose is to demonstrate how to analyze requirements, design test scenarios, identify defects, and build automated API tests.

## Project Goals

The main goal is to demonstrate practical skills required for a Backend QA/Golang Engineer position.

Key areas:

- Requirements analysis and test design
- Manual REST API testing
- Backend test automation using Go
- Database testing with PostgreSQL
- Positive, negative, and boundary testing
- Integration testing
- Concurrency and idempotency testing
- Defect reporting and regression testing
- Git and CI/CD practices

## Application Architecture

The project consists of three main components:

### 1. Banking API

A REST API written in Go.

Main functional modules:

- **Accounts** — account creation, balance retrieval, and account management.
- **Transfers** — money transfers between accounts and transfer history.

### 2. QA Automation

A collection of automated tests written in Go.

The tests interact with the Banking API through HTTP requests and verify application behavior and database state.

### 3. Test Documentation

QA documentation containing:

- Test Plan
- Test Cases
- Bug Reports
- Test Execution Reports

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Backend development and test automation |
| REST API | Communication between application components |
| PostgreSQL | Database |
| Docker | Containerization |
| Docker Compose | Local environment management |
| Postman | Manual API testing |
| Git | Version control |
| GitHub Actions | CI/CD and automated test execution |

## Planned API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/accounts` | Create an account |
| GET | `/api/v1/accounts/{id}` | Get account details |
| GET | `/api/v1/accounts/{id}/balance` | Get account balance |
| PATCH | `/api/v1/accounts/{id}/block` | Block an account |
| POST | `/api/v1/transfers` | Create a money transfer |
| GET | `/api/v1/transfers/{id}` | Get transfer details |
| GET | `/api/v1/accounts/{id}/transfers` | Get account transfer history |

## Testing Scope

The project will cover the following scenarios:

### Functional Testing
- Account creation
- Successful money transfers
- Balance verification
- Account blocking

### Negative Testing
- Transfer with insufficient funds
- Transfer with a negative amount
- Transfer to a non-existent account
- Transfer from a blocked account
- Transfer to the same account

### Boundary Testing
- Zero transfer amount
- Minimum valid transfer amount
- Maximum allowed transfer amount
- Transfer exceeding available balance

### Advanced Testing
- Idempotency
- Concurrent money transfers
- Database consistency
- Transaction atomicity

## Project Structure

```text
moneyflow/
├── banking-api/
│   ├── cmd/
│   ├── internal/
│   │   ├── accounts/
│   │   ├── transfers/
│   │   └── database/
│   └── migrations/
├── qa-automation/
│   ├── tests/
│   │   ├── api/
│   │   ├── integration/
│   │   └── database/
│   ├── helpers/
│   ├── fixtures/
│   └── testdata/
├── postman/
├── test-documentation/
├── .github/
│   └── workflows/
├── docker-compose.yml
└── README.md
```

## Project Status

**Status: In Progress**

The project is being developed incrementally.

Planned milestones:

- [ ] Initialize the project structure
- [ ] Implement the Banking API
- [ ] Integrate PostgreSQL
- [ ] Prepare test documentation
- [ ] Perform manual API testing
- [ ] Implement Go API automation tests
- [ ] Add database and integration tests
- [ ] Test concurrency and idempotency
- [ ] Prepare bug reports and regression results
- [ ] Configure GitHub Actions