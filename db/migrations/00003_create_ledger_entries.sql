-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE ledger_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  transaction_id UUID NOT NULL,
  account_id UUID NOT NULL,

  amount_minor BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT ledger_entries_transaction_fk
    FOREIGN KEY (transaction_id) REFERENCES transactions(id),

  CONSTRAINT ledger_entries_account_fk
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries (transaction_id);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries (account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_ledger_entries_account_id;
DROP INDEX IF EXISTS idx_ledger_entries_transaction_id;
DROP TABLE IF EXISTS ledger_entries;


