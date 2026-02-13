# padel-booking-api

A backend API for padel court booking system (MVP), built with Go + Gin + PostgreSQL.

## Tech stack

- Go + Gin
- PostgreSQL (Docker Compose)
- sqlc (starting week 3)
- Migrations (starting week 2)
- JWT access + refresh tokens (starting week 3/4)

## Getting started

### Prerequisites

- Go 1.22+
- (Optional) Docker Desktop (required starting week 2)

### Run (week 1)

1. Copy env file:
   ```bash
   cp .env.example .env
   ```
2. Runn the API:
   ```bash
   go run ./cmd/api
   ```

### Endpoints (week 1)

- `GET /healthz`
- `GET /readyz`
- `GET /v1/courts`
- `GET /v1/bookings`

## Roadmap

- Week 2: PostgreSQL + migrations + booking schema
- Week 3: sqlc queries + CRUD
- Week 4: React web client
- Later: court chat rooms + realtime
