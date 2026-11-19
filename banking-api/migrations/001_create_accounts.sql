CREATE TABLE accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_name  VARCHAR(100) NOT NULL,
    currency    CHAR(3)      NOT NULL CHECK (currency IN ('RUB', 'USD', 'EUR')),
    balance     BIGINT       NOT NULL DEFAULT 0 CHECK (balance >= 0),
    status      VARCHAR(10)  NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'BLOCKED')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);