package httpserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(r chi.Router, db *pgxpool.Pool) {
	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(db))
}
