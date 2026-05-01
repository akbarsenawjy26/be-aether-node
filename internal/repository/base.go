package repository

import (
	"context"

	"aether-node/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transaction wraps a pgxpool.Pool and provides methods for managing database transactions.
type Transaction struct {
	pool *pgxpool.Pool
}

// NewTransaction creates a new Transaction helper from a pgx connection pool.
func NewTransaction(pool *pgxpool.Pool) *Transaction {
	return &Transaction{pool: pool}
}

// BeginTx starts a new database transaction and returns a transaction-wrapped *db.Queries.
// The returned cleanup function commits if commit=true, otherwise rolls back.
// Usage:
//   tx := NewTransaction(pool)
//   txQueries, cleanup, err := tx.BeginTx(ctx)
//   if err != nil {
//       return err
//   }
//   defer cleanup(false) // rollback by default
//   // use txQueries for operations
//   cleanup(true) // commit when done
func (t *Transaction) BeginTx(ctx context.Context) (*db.Queries, func(commit bool) error, error) {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	queries := db.New(tx) // Use db.New directly on tx (which is pgx.Tx, a DBTX)

	cleanup := func(commit bool) error {
		if commit {
			return tx.Commit(ctx)
		}
		return tx.Rollback(ctx)
	}

	return queries, cleanup, nil
}

// Begin starts a new database transaction.
// Returns the pgx.Tx directly for advanced use cases.
// The caller is responsible for Commit/Rollback.
func (t *Transaction) Begin(ctx context.Context) (pgx.Tx, error) {
	return t.pool.Begin(ctx)
}
