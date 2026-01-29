package httpserver

import (
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/LBaronceli/go-figure/internal/db/sqlc"
)

type Server struct {
	db *pgxpool.Pool
	q  *db.Queries
}

func NewServer(dbpool *pgxpool.Pool) *Server {
	return &Server{
		db: dbpool,
		q:  db.New(dbpool),
	}
}

