ore Ledger & Transaction Model — Checklist

Tick each item top to bottom. Each item is scoped to ~1 hour of work.

---

## 0. Repository & Local Setup

* [x] Create repo and initial commit
* [x] Create folders: `apps/api`, `apps/worker`, `db/migrations`, `deploy`, `scripts`
* [x] Add `.gitignore`, `.editorconfig`, `README.md`
* [ ] Add `docker-compose.yml` with Postgres only
* [ ] Verify Postgres starts and is reachable locally

---

## 1. Database Migrations Setup

* [ ] Choose migration tool (`goose`, `migrate`, or `atlas`)
* [ ] Wire migration tool into repo
* [ ] Add Makefile targets for DB up / migrate / status
* [ ] Run an empty test migration successfully

---

## 2. Core Database Schema

### 2.1 Accounts

* [ ] Create migration: enable UUID extension
* [ ] Create `accounts` table

  * id (UUID)
  * name
  * type (asset/liability/expense/income/equity)
  * currency (ISO code)
  * timestamps
* [ ] Verify constraints work

### 2.2 Transactions

* [ ] Create migration: `transactions` table

  * id (UUID)
  * idempotency_key (unique)
  * description
  * source (manual/csv/api)
  * posted_at
  * created_at
* [ ] Verify idempotency constraint

### 2.3 Ledger Entries

* [ ] Create migration: `ledger_entries` table

  * id (UUID)
  * transaction_id (FK)
  * account_id (FK)
  * amount_minor (BIGINT)
  * currency
  * created_at
* [ ] Add indexes on account_id and transaction_id
* [ ] Verify inserts and joins work

### 2.4 Balances (derived)

* [ ] Create SQL view to sum balances per account
* [ ] Verify balances match ledger entries

---

## 3. API Skeleton

* [ ] Initialize Go module in `apps/api`
* [ ] Add HTTP router
* [ ] Add `/healthz` endpoint
* [ ] Add DB connection package
* [ ] Add `/readyz` endpoint that pings DB
* [ ] Run API via Docker

---

## 4. Accounts API

* [ ] Implement `POST /accounts`
* [ ] Validate account type and currency
* [ ] Implement `GET /accounts`
* [ ] Include computed balance in response
* [ ] Write integration tests for accounts

---

## 5. Transactions & Ledger Posting

### 5.1 Posting Model

* [ ] Define transaction posting request
* [ ] Support idempotency key
* [ ] Support multiple ledger entries per transaction

### 5.2 Core Posting Logic

* [ ] Implement DB transaction wrapper
* [ ] Validate accounts exist
* [ ] Validate currencies match accounts
* [ ] Validate sum(entries.amount_minor) == 0
* [ ] Insert transaction row
* [ ] Insert ledger entries

### 5.3 Idempotency

* [ ] Detect duplicate idempotency key
* [ ] Return existing transaction without duplication

### 5.4 Queries

* [ ] Implement `GET /transactions`
* [ ] Implement `GET /transactions/:id`
* [ ] Filter by account and date range

### 5.5 Tests

* [ ] Reject unbalanced transactions
* [ ] Reject wrong currency
* [ ] Verify idempotent retries

---

## 6. Corrections

* [ ] Implement transaction reversal
* [ ] Create compensating ledger entries
* [ ] Preserve immutability of ledger

---

## 7. Done Criteria

* [ ] Ledger is append-only
* [ ] Balances are derived, not stored
* [ ] Posting is atomic and idempotent
* [ ] Core invariants enforced by tests

---

This completes the **core financial engine** of the project.
:wq

