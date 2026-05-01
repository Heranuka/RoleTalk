package transactor

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:generate mockgen -package=transactor -destination=mocks.go go-backend/pkg/transactor Pool,Tx

// DBTx defines methods for executing database operations.
type DBTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Tx represents a database transaction.
type Tx interface {
	DBTx
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Pool represents a database connection pool that can start transactions.
type Pool interface {
	Begin(ctx context.Context) (Tx, error)
}
