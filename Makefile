ifneq (,$(wildcard .env))
	include .env
	export
endif

MIGRATIONS_DIR ?= ./migrations
DATABASE_URL := postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(SSL_MODE)
PG_CONTAINER_NAME := postgres-dev
TEST_DB_URL := postgres://postgres:postgres@localhost:5432/pvz_test?sslmode=disable

.PHONY: go
go:
	@go run cmd/main.go


.PHONY: pg
pg:
	@if [ $$(docker ps -aq -f name=$(PG_CONTAINER_NAME)) ]; then \
		echo "Postgres container already exists"; \
		docker start -a $(PG_CONTAINER_NAME); \
	else \
		echo "Starting new postgres container..."; \
		docker run --name $(PG_CONTAINER_NAME) \
			-e POSTGRES_PASSWORD=$(DATABASE_PASSWORD) \
			-e POSTGRES_USER=$(DATABASE_USER) \
			-e POSTGRES_DB=$(DATABASE_NAME) \
			-p $(DATABASE_PORT):$(DATABASE_PORT) \
			-d postgres; \
	fi

.PHONY: psql
psql:
	docker exec -it $(PG_CONTAINER_NAME) psql -U $(DATABASE_USER) -d $(DATABASE_NAME)

.PHONY: migrate_up
migrate_up:
	migrate -path $(MIGRATIONS_DIR) -database '$(DATABASE_URL)' up

.PHONY: migrate_down
migrate_down:
	migrate -path $(MIGRATIONS_DIR) -database '$(DATABASE_URL)' down

.PHONY: test_migrate_up
test_migrate_up:
	migrate -path $(MIGRATIONS_DIR) -database '$(TEST_DB_URL)' up

.PHONY: test_migrate_down
test_migrate_down:
	migrate -path $(MIGRATIONS_DIR) -database '$(TEST_DB_URL)' down

.PHONY: test
test:
	go test ./internal/... -cover