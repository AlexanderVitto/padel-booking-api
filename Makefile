SHELL := /bin/bash

COMPOSE ?= docker compose
MIGRATE_BIN ?= migrate
MIGRATIONS_DIR ?= migrations
ENV_FILE ?= .env

define load_env
	set -a; source $(ENV_FILE); set +a;
endef

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make run                - Run API server (requires .env to be set up)"
	@echo "  make up                 - Start containers (detached)"
	@echo "  make down               - Stop containers (keeps volumes/data)"
	@echo "  make reset              - Stop containers AND remove volumes (DESTROYS DATA)"
	@echo "  make logs               - Tail logs"
	@echo "  make psql               - Open psql inside db container"
	@echo "  make env-check          - Print DATABASE_URL (sanity check)"
	@echo "  make migrate-up         - Run all up migrations"
	@echo "  make migrate-down       - Rollback 1 migration"
	@echo "  make migrate-force V=N  - Force migration version"
	@echo "  make migrate-create NAME=... - Create new migration files"
	@echo "  make db-init            - up + migrate-up"
	@echo "  make seed-dev           - Seed a dev user (UUID aaaaaaaa-...)"

.PHONY: run
run:
	@$(load_env) \
	go run ./cmd/api/main.go

.PHONY: up
up:
	$(COMPOSE) up -d

.PHONY: down
down:
	$(COMPOSE) down

.PHONY: reset
reset:
	@echo "WARNING: This will remove volumes and DESTROY DB data."
	$(COMPOSE) down -v

.PHONY: logs
logs:
	$(COMPOSE) logs -f --tail=200

.PHONY: psql
psql:
	docker exec -it padel_booking_db psql -U postgres -d padel_booking

.PHONY: env-check
env-check:
	@$(load_env) echo "DATABASE_URL=$$DATABASE_URL"

# --- migrations ---
.PHONY: migrate-up
migrate-up:
	@$(load_env) \
	$(MIGRATE_BIN) -database "$$DATABASE_URL" -path $(MIGRATIONS_DIR) up

.PHONY: migrate-down
migrate-down:
	@$(load_env) \
	$(MIGRATE_BIN) -database "$$DATABASE_URL" -path $(MIGRATIONS_DIR) down 1

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(V)" ]; then echo "Usage: make migrate-force V=<version>"; exit 1; fi
	@$(load_env) \
	$(MIGRATE_BIN) -database "$$DATABASE_URL" -path $(MIGRATIONS_DIR) force $(V)

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<migration_name>"; exit 1; fi
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

.PHONY: db-init
db-init: up migrate-up

# --- seeding ---
.PHONY: seed-dev
seed-dev:
	docker exec -i padel_booking_db psql -U postgres -d padel_booking -v ON_ERROR_STOP=1 -c "insert into users (id, email, password_hash, display_name) values ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'dev@example.com', 'dev', 'Dev User') on conflict (email) do update set display_name=excluded.display_name;"
	docker exec -i padel_booking_db psql -U postgres -d padel_booking -v ON_ERROR_STOP=1 -c "insert into venues (id, name, timezone) values ('11111111-1111-1111-1111-111111111111', 'Dev Venue', 'Asia/Jakarta') on conflict (id) do update set name=excluded.name, timezone=excluded.timezone;"
	docker exec -i padel_booking_db psql -U postgres -d padel_booking -v ON_ERROR_STOP=1 -c "insert into courts (id, venue_id, name, is_active) values ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Court 1', true) on conflict (id) do update set name=excluded.name, is_active=excluded.is_active;"
	docker exec -i padel_booking_db psql -U postgres -d padel_booking -v ON_ERROR_STOP=1 -c "insert into courts (id, venue_id, name, is_active) values ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Court 2', true) on conflict (id) do update set name=excluded.name, is_active=excluded.is_active;"