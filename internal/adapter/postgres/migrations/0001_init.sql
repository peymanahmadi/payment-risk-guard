CREATE TABLE IF NOT EXISTS transactions (
    id            UUID PRIMARY KEY,
    account_id    TEXT NOT NULL,
    amount        NUMERIC(18, 2) NOT NULL,
    currency      TEXT NOT NULL,
    country       TEXT NOT NULL DEFAULT '',
    ip_address    TEXT NOT NULL DEFAULT '',
    payment_type  TEXT NOT NULL DEFAULT '',
    occurred_at   TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_occurred
    ON transactions (account_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS alerts (
    id             UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions (id),
    account_id     TEXT NOT NULL,
    score          NUMERIC(5, 2) NOT NULL,
    severity       TEXT NOT NULL,
    reasons        TEXT[] NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_alerts_account_created
    ON alerts (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_alerts_severity
    ON alerts (severity);