// Package session implements the database repository for managing user sessions,
// encompassing both authentication tokens and active practice states.
package auth_session

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	"go-backend/pkg/transactor"
)

var tracer = otel.Tracer("internal/repository/session")

// Repository handles PostgreSQL operations for Session entities.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates and returns a new Session repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new session record. Initially, practice fields like topic_id are often null.
func (r *Repository) Create(ctx context.Context, s *domain.AuthSession) error {
	ctx, span := tracer.Start(ctx, "Repository.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO sessions (
		    id, user_id,  token_hash, expires_at, revoked
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.Exec(ctx, q,
		s.ID,
		s.UserID,
		s.TokenHash,
		s.ExpiresAt,
		s.Revoked,
	)

	if err != nil {
		r.handleError(ctx, err, "create session")
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrSessionAlreadyExists
		}
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a session using the refresh token hash.
func (r *Repository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error) {
	ctx, span := tracer.Start(ctx, "Repository.GetByTokenHash")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, user_id,  token_hash, 
			expires_at, revoked, created_at, updated_at
		FROM sessions
		WHERE token_hash = $1
	`

	var s domain.AuthSession
	err := db.QueryRow(ctx, q, tokenHash).Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.ExpiresAt,
		&s.Revoked,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		r.handleError(ctx, err, "get session by token hash")
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return &s, nil
}

// Revoke invalidates the authentication session.
func (r *Repository) Revoke(ctx context.Context, tokenHash string) error {
	ctx, span := tracer.Start(ctx, "Repository.Revoke")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE sessions
		SET 
			revoked = TRUE,
			updated_at = now()
		WHERE token_hash = $1
	`

	cmd, err := db.Exec(ctx, q, tokenHash)
	if err != nil {
		r.handleError(ctx, err, "revoke session")
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// handleError logs internal database errors with context and records them in the trace.
func (r *Repository) handleError(ctx context.Context, err error, op string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	log.Errorw("database operation failed",
		"op", op,
		"error", err,
	)
}
