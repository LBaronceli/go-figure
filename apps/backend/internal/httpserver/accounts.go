package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/LBaronceli/go-figure/internal/db/sqlc"
)

type createAccountRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
}

type accountResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// POST /accounts
func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	req.Currency = strings.TrimSpace(strings.ToUpper(req.Currency))

	if req.Name == "" || req.Type == "" || req.Currency == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	acc, err := s.q.CreateAccount(r.Context(), db.CreateAccountParams{
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
	})
	if err != nil {
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toAccountResponse(acc))
}

// GET /accounts
func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.q.ListAccounts(r.Context())
	if err != nil {
		http.Error(w, "failed to list accounts", http.StatusInternalServerError)
		return
	}

	resp := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		resp = append(resp, toAccountResponse(a))
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /accounts/{id}
func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	acc, err := s.q.GetAccount(r.Context(), id)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, toAccountResponse(acc))
}

// DELETE /accounts/{id}
func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if _, err := s.q.GetAccount(r.Context(), id); err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	if err := s.q.DeleteAccount(r.Context(), id); err != nil {
		http.Error(w, "failed to delete account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helpers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	if !id.Valid {
		return pgtype.UUID{}, errors.New("invalid uuid")
	}
	return id, nil
}

func toAccountResponse(a db.Account) accountResponse {
	idStr := ""
	if a.ID.Valid {
		idStr = uuid.UUID(a.ID.Bytes).String()
	}

	created := ""
	if a.CreatedAt.Valid {
		created = a.CreatedAt.Time.Format(time.RFC3339Nano)
	}

	updated := ""
	if a.UpdatedAt.Valid {
		updated = a.UpdatedAt.Time.Format(time.RFC3339Nano)
	}

	return accountResponse{
		ID:        idStr,
		Name:      a.Name,
		Type:      a.Type,
		Currency:  a.Currency,
		CreatedAt: created,
		UpdatedAt: updated,
	}
}

