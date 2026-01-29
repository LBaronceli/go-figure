package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	// middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// health
	r.Get("/healthz", s.healthz)
	r.Get("/readyz", s.readyz)

	// accounts
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", s.createAccount)
		r.Get("/", s.listAccounts)
		r.Get("/{id}", s.getAccount)
		r.Delete("/{id}", s.deleteAccount)
	})

	return r
}

