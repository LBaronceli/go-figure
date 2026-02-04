# Go Figure

A production-oriented, open-source **finance, cashflow, and tax engine** built to demonstrate backend, frontend, and infrastructure engineering practices.

This project focuses on **ledger correctness, concurrency, asynchronous processing, and deployability**. UI polish will come on a second phase.

The goal of this project is to be sort off a **Xero** light, focussed on a sole-trader and small operations for personal use, dodging the need for the compliance bloat that would come with that type of software.

GST and tax returns calculation will come on Phase 2.

Note: This repo used AI for coding tests and boilerplate code. The structure, features, security choices, architecture, and most of the core code was "artisanally" crafted by me.

---

## Goals

This project exists to demonstrate:

- Backend engineering in **Go**
  - Concurrency and worker pools
  - Idempotent APIs
  - Background job processing
  - Clear domain modeling

- Frontend development in **React**
  - Data-heavy UI
  - Real-world async workflows

- Database design
  - Transactional integrity
  - Double-entry ledger modeling
  - Migrations and constraints

- Infrastructure & operations
  - Docker-first development
  - Kubernetes / Cloud Run-friendly architecture
  - Separation of API and background workers
  - Observability hooks (metrics, logs)

This is **not** intended to be a full accounting product.

---

## High-Level Architecture

```
┌──────────┐        HTTP        ┌────────────┐
│  Web UI  │ ─────────────────▶ │   API      │
│ (React)  │                    │   (Go)     │
└──────────┘                    └─────┬──────┘
                                      │
                                      │ enqueue jobs
                                      ▼
                                ┌────────────┐
                                │  Worker    │
                                │  (Go)      │
                                └─────┬──────┘
                                      │
                               ┌──────▼──────┐
                               │  Postgres   │
                               │ (Ledger +   │
                               │  Job Queue) │
                               └─────────────┘
```

- **API** handles request/response workflows and validation
- **Workers** handle long-running or CPU-intensive tasks
- **Postgres** is the source of truth for both domain data and job coordination

---

## Core Features

### Accounts & Ledger

- Multiple accounts (cash, credit, savings)
- Double-entry ledger model
- Strong consistency guarantees on balances

### Transactions

- Manual entry
- CSV import (bank export style)
- Deduplication and idempotency
- Pending vs cleared transactions

### Categorisation & Rules

- Rule-based categorisation engine
- Reprocessing on rule changes

### Budgets & Cash-Flow

- Monthly category budgets
- Rolling cash-flow projections
- Trend analysis

### Background Jobs

Handled asynchronously by workers:

- CSV imports
- Transaction reconciliation
- Categorisation reprocessing
- Cash-flow projections
- Scheduled rollups

---

## Tech Stack

### Backend

- **Go**
- HTTP API (chi)
- PostgreSQL
- Postgres-backed job queue (`FOR UPDATE SKIP LOCKED`)
- OpenAPI (planned)

### Frontend

- **React**
- Vite
- TypeScript
- Charts and data-heavy views

### Infrastructure

- Docker & Docker Compose
- Kubernetes (Kustomize or Helm)
- Cloud Run compatible (stateless API + worker)
- Postgres as the only required external dependency

---

## Repository Structure

```
.
├── apps/
│  ├── backend
│     ├── cmd/        # Go HTTP API and background workers
│     └── internal/
│  └── web/           # React frontend
├── db/
│   ├── migrations/
│   └── seed/
├── deploy/
│   ├── k8s/
│   └── helm/
├── docker-compose.yml
└── README.md
```

---

## Job Queue Design

Background jobs are coordinated using **Postgres**, not a separate message broker.

Why:

- Fewer moving parts
- Strong durability guarantees
- Easy local development
- Demonstrates locking, leasing, retries, and backoff

Key concepts:

- Jobs are leased using `SELECT ... FOR UPDATE SKIP LOCKED`
- Workers renew leases while processing
- Automatic retries with exponential backoff
- Dead-lettering after max attempts

---

## Running Locally

### Prerequisites

- Docker
- Docker Compose

### Start everything

```bash
docker compose up --build
```

This will start:

- API
- Worker
- Web UI
- Postgres

---

## Development Philosophy

- **Correctness over cleverness**
- **Explicit over magical**
- **Operational realism**
- **Small, composable services**
- **Clear failure modes**

If a design decision trades simplicity for realism, realism usually wins.

---

## Non-Goals

- Real bank integrations
- Tax/VAT/GST filing (on v1)
- Payroll
- Regulatory compliance
- Payment processing

Those are intentionally out of scope.

---

## Roadmap (Indicative)

- [ ] Core ledger and transaction model
- [ ] CSV import + reconciliation
- [ ] Background worker framework
- [ ] Budgeting and projections
- [ ] Web UI dashboards
- [ ] OpenAPI spec + client generation
- [ ] Metrics and tracing
- [ ] Cloud Run deployment example

---

## License

MIT
