#!/bin/bash
# Create a new migration file pair

set -e

NAME=${1:-new_migration}
TIMESTAMP=$(date +%Y%m%d%H%M%S)
MIGRATIONS_DIR="./migrations"

UP_FILE="${MIGRATIONS_DIR}/${TIMESTAMP}_${NAME}.up.sql"
DOWN_FILE="${MIGRATIONS_DIR}/${TIMESTAMP}_${NAME}.down.sql"

if [ -f "$UP_FILE" ]; then
    echo "Migration already exists: $UP_FILE"
    exit 1
fi

cat > "$UP_FILE" << 'EOF'
-- Migration: TIMESTAMP_NAME.up.sql
-- Description:

EOF

cat > "$DOWN_FILE" << 'EOF'
-- Migration: TIMESTAMP_NAME.down.sql
-- Rollback:

EOF

echo "Created: $UP_FILE"
echo "Created: $DOWN_FILE"
