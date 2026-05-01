# Migration Makefile
# Usage: make migrate-up | make migrate-down | make migrate-status

DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= aether_node
MIGRATIONS_DIR ?= ./migrations

# PostgreSQL connection string
export PGPASSWORD=$(DB_PASSWORD)
PSQL = psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

# Run all pending migrations (up)
migrate-up:
	@echo "Running migrations..."
	@for f in $(MIGRATIONS_DIR)/*.up.sql; do \
		echo "Applying: $$f"; \
		$(PSQL) -f $$f; \
	done
	@echo "Migrations complete."

# Rollback all migrations (down)
migrate-down:
	@echo "Rolling back migrations..."
	@for f in $(sort $(wildcard $(MIGRATIONS_DIR)/*.down.sql)); do \
		echo "Rolling back: $$f"; \
		$(PSQL) -f $$f; \
	done
	@echo "Rollback complete."

# Create a new migration file
new-migration name=$(name):
	@./scripts/create_migration.sh $(name)

# Show migration status
status:
	@echo "Checking migration status..."
	$(PSQL) -c "SELECT table_name FROM information_table WHERE table_schema = 'public' ORDER BY table_name;"
