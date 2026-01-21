-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  idempotency_key TEXT NOT NULL,

  description TEXT,

  source TEXT NOT NULL,
  posted_at TIMESTAMPTZ,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT transactions_idempotency_key_unique UNIQUE (idempotency_key),
  CONSTRAINT transactions_source_check CHECK (source IN ('manual', 'csv', 'api'))
);

CREATE INDEX idx_transactions_posted_at ON transactions (posted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_posted_at;
DROP TABLE IF EXISTS transactions;

