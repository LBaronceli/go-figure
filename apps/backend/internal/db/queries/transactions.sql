-- name: CreateTransaction :one
INSERT INTO transactions (
  idempotency_key,
  description,
  source,
  posted_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (
  transaction_id,
  account_id,
  amount_minor,
  currency
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1 LIMIT 1;

-- name: GetTransactionByIdempotencyKey :one
SELECT * FROM transactions
WHERE idempotency_key = $1 LIMIT 1;

-- name: ListTransactions :many
SELECT * FROM transactions
ORDER BY posted_at DESC, created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListLedgerEntries :many
SELECT * FROM ledger_entries
WHERE transaction_id = $1
ORDER BY amount_minor DESC;
