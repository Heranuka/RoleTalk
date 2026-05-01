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

	"github.com/jackc/pgx/v5/pgxpool"

	"go-backend/internal/config"
)

// New creates a new PostgreSQL connection pool using the given DSN and pool configuration.
//
// The pool is configured with maximum and minimum connections, as well as maximum connection lifetime.
//
// Parameters:
//   - ctx: the context for controlling pool creation timeout.
//   - dsn: the PostgreSQL connection string.
//   - cfg: the pool configuration (max/min connections, max lifetime).
//
// Returns:
//   - *pgxpool.Pool: the initialized connection pool.
//   - error: any error encountered during pool creation.
func New(ctx context.Context, dsn string, cfg *config.PoolConfig) (*pgxpool.Pool, error) {
	// Parse pool configuration for PostgreSQL connection.
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	// Initialize connection pool for PostgreSQL.
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	return pool, nil
}
