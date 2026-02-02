-- name: CreateAccount :one
INSERT INTO accounts (
    name,
    type,
    currency
) VALUES (
    $1, $2, $3
)
RETURNING
    id,
    name,
    type,
    currency,
    created_at,
    updated_at;

-- name: GetAccount :one
SELECT 
  id,
  name,
  type,
  currency,
  created_at,
  updated_at
FROM accounts
WHERE id = $1;

-- name: ListAccounts :many
SELECT 
  id,
  name,
  type,
  currency,
  created_at,
  updated_at
FROM accounts
ORDER BY created_at DESC;

-- name: UpdateAccount :one
UPDATE accounts
SET
    name = COALESCE(sqlc.narg('name'), name),
    type = COALESCE(sqlc.narg('type'), type)
WHERE id = $1
RETURNING
    id,
    name,
    type,
    currency,
    created_at,
    updated_at;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;

-- name: GetAccountsByIDs :many
SELECT * FROM accounts
WHERE id = ANY($1::uuid[]);

