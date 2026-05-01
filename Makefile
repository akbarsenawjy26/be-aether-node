# Variables
BINARY_NAME=aether-node
MAIN_FILE=./cmd/server/main.go
GO=go

# Load .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default target
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  run             Run the application"
	@echo "  build           Build the binary"
	@echo "  test            Run tests"
	@echo "  tidy            Run go mod tidy"
	@echo "  clean           Clean build artifacts"
	@echo "  migrate-up      Run database migrations"
	@echo "  migrate-down    Rollback database migrations"
	@echo "  docker-up       Start services with docker-compose"
	@echo "  docker-down     Stop services with docker-compose"
	@echo "  docker-logs     Show docker-compose logs"

# Development
.PHONY: run
run:
	$(GO) run $(MAIN_FILE)

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: test
test:
	$(GO) test -v ./...

# Build
.PHONY: build
build:
	$(GO) build -o $(BINARY_NAME) $(MAIN_FILE)

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

# Migrations
.PHONY: migrate-up
migrate-up:
	@echo "Running migrations..."
	@for f in ./migrations/*.up.sql; do \
		echo "Applying: $$f"; \
		PGPASSWORD=$(DATABASE_PASSWORD) psql -h $(DATABASE_HOST) -p $(DATABASE_PORT) -U $(DATABASE_USER) -d $(DATABASE_NAME) -f $$f; \
	done
	@echo "Migrations complete."

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	@# Reverse sort for down migrations
	@ls ./migrations/*.down.sql | sort -r | while read f; do \
		echo "Rolling back: $$f"; \
		PGPASSWORD=$(DATABASE_PASSWORD) psql -h $(DATABASE_HOST) -p $(DATABASE_PORT) -U $(DATABASE_USER) -d $(DATABASE_NAME) -f $$f; \
	done
	@echo "Rollback complete."

# Docker
.PHONY: docker-up
docker-up:
	docker-compose up -d

.PHONY: docker-down
docker-down:
	docker-compose down

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f
