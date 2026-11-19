CREATE TABLE transfers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account_id  UUID        NOT NULL REFERENCES accounts(id),
    to_account_id    UUID        NOT NULL REFERENCES accounts(id),
    amount           BIGINT      NOT NULL CHECK (amount > 0),
    currency         CHAR(3)     NOT NULL,
    status           VARCHAR(10) NOT NULL DEFAULT 'COMPLETED',
    idempotency_key  VARCHAR(64) NOT NULL UNIQUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_account_id <> to_account_id)
);

CREATE INDEX idx_transfers_from ON transfers(from_account_id, created_at DESC);
CREATE INDEX idx_transfers_to   ON transfers(to_account_id, created_at DESC);