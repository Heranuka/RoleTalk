// Package postgres provides PostgreSQL infrastructure and adapter setup.
//
// It contains initialization and configuration of the pgx connection pool,
// as well as adapters (PoolAdapter and TxAdapter) that allow the pool and
// transactions to be used with the transactor package for safe transaction
// management in repositories and services.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-backend/pkg/transactor"
)

// PoolAdapter adapts a pgxpool.Pool to implement the transactor.Pool interface.
//
// This allows the Transactor to manage transactions uniformly over any underlying
// PostgreSQL connection pool.
type PoolAdapter struct {
	pool *pgxpool.Pool
}

// NewPoolAdapter creates a new PoolAdapter wrapping the provided pgxpool.Pool.
//
// Parameters:
//   - pool: the PostgreSQL connection pool to wrap.
//
// Returns:
//   - *PoolAdapter: the adapter implementing transactor.Pool.
func NewPoolAdapter(pool *pgxpool.Pool) *PoolAdapter {
	return &PoolAdapter{pool: pool}
}

// Begin starts a new transaction on the underlying pool.
//
// It returns a TxAdapter that implements the transactor.Tx interface.
//
// Parameters:
//   - ctx: the context for controlling transaction start timeout.
//
// Returns:
//   - transactor.Tx: the transaction adapter for the new transaction.
//   - error: any error encountered while starting the transaction.
func (p *PoolAdapter) Begin(ctx context.Context) (transactor.Tx, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &TxAdapter{tx: tx}, nil
}

// TxAdapter adapts a pgx.Tx to implement the transactor.Tx interface.
//
// This allows repositories and services to interact with either a transaction
// or the pool transparently through the DBTx interface.
type TxAdapter struct {
	tx pgx.Tx
}

// Exec executes a SQL command within the transaction.
//
// Parameters:
//   - ctx: the context for query execution.
//   - sql: the SQL statement to execute.
//   - args: optional arguments for the SQL statement.
//
// Returns:
//   - pgconn.CommandTag: metadata about the executed command.
//   - error: any error encountered during execution.
func (t *TxAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return tag, fmt.Errorf("failed to execute query: %w", err)
	}
	return tag, nil
}

// Query executes a SQL query that returns multiple rows.
//
// Parameters:
//   - ctx: the context for query execution.
//   - sql: the SQL query to execute.
//   - args: optional arguments for the SQL query.
//
// Returns:
//   - pgx.Rows: the result set rows.
//   - error: any error encountered during execution.
func (t *TxAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return rows, nil
}

// QueryRow executes a SQL query expected to return a single row.
//
// Parameters:
//   - ctx: the context for query execution.
//   - sql: the SQL query to execute.
//   - args: optional arguments for the SQL query.
//
// Returns:
//   - pgx.Row: the result row for scanning.
func (t *TxAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

// Commit commits the transaction.
//
// Parameters:
//   - ctx: the context for commit operation.
//
// Returns:
//   - error: any error encountered during commit.
func (t *TxAdapter) Commit(ctx context.Context) error {
	if err := t.tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Rollback rolls back the transaction.
//
// Parameters:
//   - ctx: the context for rollback operation.
//
// Returns:
//   - error: any error encountered during rollback.
func (t *TxAdapter) Rollback(ctx context.Context) error {
	if err := t.tx.Rollback(ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}
