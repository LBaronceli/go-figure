package sqlc_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	db "github.com/LBaronceli/go-figure/internal/db/sqlc"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	connString := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, connString, "DATABASE_URL must be set for tests")

	pool, err := pgxpool.New(context.Background(), connString)
	require.NoError(t, err)

	return pool
}

func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	pool := setupDB(t)
	defer pool.Close()

	q := db.New(pool)

	account, err := q.CreateAccount(ctx, db.CreateAccountParams{
		Name:     "Cash",
		Type:     "asset",
		Currency: "NZD",
	})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, account.ID)
	require.Equal(t, "Cash", account.Name)
	require.NotZero(t, account.CreatedAt)
	require.NotZero(t, account.UpdatedAt)
}

