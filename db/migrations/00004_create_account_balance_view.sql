-- +goose Up
CREATE OR REPLACE VIEW account_balances AS
SELECT
  a.id         AS account_id,
  a.name       AS account_name,
  a.type       AS account_type,
  a.currency   AS account_currency,
  COALESCE(SUM(le.amount_minor), 0) AS balance_minor
FROM accounts a
LEFT JOIN ledger_entries le
  ON le.account_id = a.id
GROUP BY
  a.id, a.name, a.type, a.currency;

-- +goose Down
DROP VIEW IF EXISTS account_balances;

