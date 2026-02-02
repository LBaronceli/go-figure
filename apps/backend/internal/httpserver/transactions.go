package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/LBaronceli/go-figure/internal/db/sqlc"
)

// maxStringLength exists so we can prevent someone from uploading a massive string.
const (
	maxLedgerEntries = 100
	maxStringLength  = 500
)

type ledgerEntryRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"` // Minor units
}

type createTransactionRequest struct {
	IdempotencyKey string               `json:"idempotency_key"`
	Description    string               `json:"description"`
	Source         string               `json:"source"`
	PostedAt       string               `json:"posted_at"` // ISO8601
	Entries        []ledgerEntryRequest `json:"entries"`
}

type ledgerEntryResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type transactionResponse struct {
	ID             string                `json:"id"`
	IdempotencyKey string                `json:"idempotency_key"`
	Description    string                `json:"description"`
	Source         string                `json:"source"`
	PostedAt       string                `json:"posted_at"`
	CreatedAt      string                `json:"created_at"`
	Entries        []ledgerEntryResponse `json:"entries,omitempty"`
}

// POST /transactions
func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	req.Description = strings.TrimSpace(req.Description)
	req.Source = strings.TrimSpace(strings.ToLower(req.Source))

	if req.IdempotencyKey == "" {
		http.Error(w, "missing idempotency_key", http.StatusBadRequest)
		return
	}
	if len(req.IdempotencyKey) > maxStringLength {
		http.Error(w, "idempotency_key too long", http.StatusBadRequest)
		return
	}
	if len(req.Description) > maxStringLength {
		http.Error(w, "description too long", http.StatusBadRequest)
		return
	}
	if req.Source != "manual" && req.Source != "csv" && req.Source != "api" {
		http.Error(w, "invalid source (must be manual, csv, or api)", http.StatusBadRequest)
		return
	}
	if len(req.Entries) < 2 {
		http.Error(w, "transaction must have at least 2 entries", http.StatusBadRequest)
		return
	}
	if len(req.Entries) > maxLedgerEntries {
		http.Error(w, fmt.Sprintf("too many entries (max %d)", maxLedgerEntries), http.StatusBadRequest)
		return
	}

	var postedAt pgtype.Timestamptz
	if req.PostedAt != "" {
		t, err := time.Parse(time.RFC3339, req.PostedAt)
		if err != nil {
			http.Error(w, "invalid posted_at format (use RFC3339)", http.StatusBadRequest)
			return
		}
		postedAt = pgtype.Timestamptz{Time: t, Valid: true}
	} else {
		postedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	// resolve accounts and validate logic
	accountIDs := make([]pgtype.UUID, 0, len(req.Entries))

	for _, entry := range req.Entries {
		id, err := parseUUID(entry.AccountID)
		if err != nil {
			http.Error(w, "invalid account_id uuid", http.StatusBadRequest)
			return
		}
		accountIDs = append(accountIDs, id)

	}

	accounts, err := s.q.GetAccountsByIDs(r.Context(), accountIDs)
	if err != nil {
		http.Error(w, "failed to fetch accounts", http.StatusInternalServerError)
		return
	}

	// Map accounts for easy lookup
	accMap := make(map[string]db.Account)
	for _, acc := range accounts {
		if acc.ID.Valid {
			idStr := uuid.UUID(acc.ID.Bytes).String()
			accMap[idStr] = acc
		}
	}

	// Validate Existence, Currencies, and Balance
	var sum int64
	var commonCurrency string

	for i, entry := range req.Entries {
		// Normalize ID string from request (assuming parseUUID validated format implicitly)
		uid, _ := uuid.Parse(entry.AccountID) // safe because parseUUID passed
		accID := uid.String()

		acc, found := accMap[accID]
		if !found {
			http.Error(w, fmt.Sprintf("account not found: %s", entry.AccountID), http.StatusBadRequest)
			return
		}

		if i == 0 {
			commonCurrency = acc.Currency
		} else {
			if acc.Currency != commonCurrency {
				http.Error(w, "multi-currency transactions not supported yet (all accounts must have same currency)", http.StatusBadRequest)
				return
			}
		}

		// Check overflow
		if (entry.Amount > 0 && sum > (1<<63-1)-entry.Amount) || (entry.Amount < 0 && sum < -(1<<63-1)-entry.Amount) {
			http.Error(w, "transaction amount overflow", http.StatusBadRequest)
			return
		}
		sum += entry.Amount
	}

	if sum != 0 {
		http.Error(w, "transaction is not balanced (sum must be 0)", http.StatusBadRequest)
		return
	}

	// Execute DB Transaction
	tx, err := s.db.Begin(r.Context())
	if err != nil {
		http.Error(w, "failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := s.q.WithTx(tx)

	// Create Header
	t, err := qtx.CreateTransaction(r.Context(), db.CreateTransactionParams{
		IdempotencyKey: req.IdempotencyKey,
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		Source:         req.Source,
		PostedAt:       postedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			// Idempotency: fetch existing
			existing, getErr := s.q.GetTransactionByIdempotencyKey(r.Context(), req.IdempotencyKey)
			if getErr != nil {
				http.Error(w, "idempotency conflict handling failed", http.StatusInternalServerError)
				return
			}
			// In a real strict implementation, we would verify the payload matches.
			// For now, we return the existing transaction as per plan.
			// We should fetch entries too.
			entries, _ := s.q.ListLedgerEntries(r.Context(), existing.ID)
			writeJSON(w, http.StatusOK, toFullTransactionResponse(existing, entries))
			return
		}
		http.Error(w, "failed to create transaction", http.StatusInternalServerError)
		return
	}

	// Create Entries
	createdEntries := make([]db.LedgerEntry, 0, len(req.Entries))
	for _, entry := range req.Entries {
		accID, _ := parseUUID(entry.AccountID)
		le, err := qtx.CreateLedgerEntry(r.Context(), db.CreateLedgerEntryParams{
			TransactionID: t.ID,
			AccountID:     accID,
			AmountMinor:   entry.Amount,
			Currency:      commonCurrency,
		})
		if err != nil {
			http.Error(w, "failed to create ledger entry", http.StatusInternalServerError)
			return
		}
		createdEntries = append(createdEntries, le)
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toFullTransactionResponse(t, createdEntries))
}

// GET /transactions
func (s *Server) listTransactions(w http.ResponseWriter, r *http.Request) {
	// Simple pagination
	limit := 50
	offset := 0

	// Filters
	var accountID pgtype.UUID
	if v := r.URL.Query().Get("account_id"); v != "" {
		id, err := parseUUID(v)
		if err != nil {
			http.Error(w, "invalid account_id", http.StatusBadRequest)
			return
		}
		accountID = id
	}

	var startDate pgtype.Timestamptz
	if v := r.URL.Query().Get("start_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			http.Error(w, "invalid start_date (use RFC3339)", http.StatusBadRequest)
			return
		}
		startDate = pgtype.Timestamptz{Time: t, Valid: true}
	}

	var endDate pgtype.Timestamptz
	if v := r.URL.Query().Get("end_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			http.Error(w, "invalid end_date (use RFC3339)", http.StatusBadRequest)
			return
		}
		endDate = pgtype.Timestamptz{Time: t, Valid: true}
	}

	txs, err := s.q.ListTransactions(r.Context(), db.ListTransactionsParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		AccountID: accountID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		http.Error(w, "failed to list transactions", http.StatusInternalServerError)
		return
	}

	resp := make([]transactionResponse, 0, len(txs))
	for _, t := range txs {
		// Optimization: For list view, we do not fetch entries to avoid N+1 queries.
		resp = append(resp, toTransactionResponse(t))
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /transactions/{id}
func (s *Server) getTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := parseUUID(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	t, err := s.q.GetTransaction(r.Context(), id)
	if err != nil {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	entries, err := s.q.ListLedgerEntries(r.Context(), t.ID)
	if err != nil {
		http.Error(w, "failed to fetch ledger entries", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toFullTransactionResponse(t, entries))
}

// Helpers

func toTransactionResponse(t db.Transaction) transactionResponse {
	idStr := uuid.UUID(t.ID.Bytes).String()

	posted := ""
	if t.PostedAt.Valid {
		posted = t.PostedAt.Time.Format(time.RFC3339)
	}

	created := t.CreatedAt.Time.Format(time.RFC3339Nano)

	return transactionResponse{
		ID:             idStr,
		IdempotencyKey: t.IdempotencyKey,
		Description:    t.Description.String,
		Source:         t.Source,
		PostedAt:       posted,
		CreatedAt:      created,
	}
}

func toFullTransactionResponse(t db.Transaction, entries []db.LedgerEntry) transactionResponse {
	res := toTransactionResponse(t)
	res.Entries = make([]ledgerEntryResponse, 0, len(entries))

	for _, e := range entries {
		eID := uuid.UUID(e.ID.Bytes).String()
		aID := uuid.UUID(e.AccountID.Bytes).String()
		res.Entries = append(res.Entries, ledgerEntryResponse{
			ID:        eID,
			AccountID: aID,
			Amount:    e.AmountMinor,
			Currency:  e.Currency,
		})
	}
	return res
}
