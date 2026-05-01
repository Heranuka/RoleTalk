// Package transactor provides a simple and safe way to manage PostgreSQL transactions.
//
// It wraps pgx connection pool and provides transaction boundary control
// with automatic rollback on panic or error, while allowing nested transaction-aware
// repository calls via context.
package transactor

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type txKey struct{}

// Transactor manages PostgreSQL transactions using a connection pool.
type Transactor struct {
	pool Pool
	lg   *zap.SugaredLogger
}

// NewTransactor creates a new Transactor instance.
//
// It accepts a pgx connection pool and a sugared logger for error reporting.
func NewTransactor(pool Pool, lg *zap.SugaredLogger) *Transactor {
	return &Transactor{
		pool: pool,
		lg:   lg,
	}
}

// WithinTx executes the provided function inside a database transaction.
//
// If the function returns an error or panics, the transaction is rolled back.
// If the function completes successfully, the transaction is committed.
//
// The transaction is made available to nested calls via context.
// Use GetDB() or repositories that respect context to obtain the transactional connection.
//
// Example:
//
//	err := transactor.WithinTx(ctx, func(ctx context.Context) error {
//	    // all repository calls inside this closure will use the same transaction
//	    return nil
//	})
func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}

	// Inject transaction into context so nested calls can use it.
	ctx = contextWithTx(ctx, tx)

	// Ensure rollback is called if commit fails or panic occurs.
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck // rollback in panic path: safe to ignore (tx may be committed/closed)
			_ = tx.Rollback(ctx)
			panic(r) // re-panic after rollback
		}

		// Normal error path — rollback if commit wasn't called.
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.lg.Errorw("tx rollback error", "error", err)
		}
	}()

	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// contextWithTx injects the active transaction into the context.
func contextWithTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetDB returns either the active transaction (if present in context) or the original pool.
//
// This function allows repositories to transparently use either a transaction
// (when called inside WithinTx) or the pool directly (when called outside a transaction).
//
// Typical usage in repositories:
//
//	db := transactor.GetDB(ctx, r.pool)
//	// then use db.Exec / db.QueryRow / db.Query
func GetDB(ctx context.Context, pool DBTx) DBTx {
	if tx, ok := ctx.Value(txKey{}).(Tx); ok {
		return tx
	}

	return pool
}
