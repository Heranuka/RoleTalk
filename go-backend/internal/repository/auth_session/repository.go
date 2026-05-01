// Package auth_session implements the database repository for managing user sessions.
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

var tracer = otel.Tracer("internal/repository/auth_session")

const (
	createSessionQuery = `
		INSERT INTO auth_sessions (user_id, token_hash, expires_at, revoked)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	getSessionQuery = `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at, updated_at
		FROM auth_sessions
		WHERE token_hash = $1`

	revokeSessionQuery = `
		UPDATE auth_sessions
		SET revoked = TRUE, updated_at = now()
		WHERE token_hash = $1`
)

// Repository handles PostgreSQL operations for Session entities.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates and returns a new Session repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new auth session and populates the domain model with DB-generated fields.
func (r *Repository) Create(ctx context.Context, s *domain.AuthSession) error {
	ctx, span := tracer.Start(ctx, "Repository.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	err := db.QueryRow(ctx, createSessionQuery,
		s.UserID,
		s.TokenHash,
		s.ExpiresAt,
		s.Revoked,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrSessionAlreadyExists
		}

		r.recordError(ctx, err, "create session")
		return fmt.Errorf("execute create session: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a session using the refresh token hash.
func (r *Repository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error) {
	ctx, span := tracer.Start(ctx, "Repository.GetByTokenHash")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	var s domain.AuthSession
	err := db.QueryRow(ctx, getSessionQuery, tokenHash).Scan(
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
		r.recordError(ctx, err, "get session by hash")
		return nil, fmt.Errorf("scan session: %w", err)
	}

	return &s, nil
}

// Revoke invalidates the authentication session by setting the revoked flag.
func (r *Repository) Revoke(ctx context.Context, tokenHash string) error {
	ctx, span := tracer.Start(ctx, "Repository.Revoke")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	cmd, err := db.Exec(ctx, revokeSessionQuery, tokenHash)
	if err != nil {
		r.recordError(ctx, err, "revoke session")
		return fmt.Errorf("execute revoke: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// recordError logs only unexpected infrastructure errors and updates the trace status.
func (r *Repository) recordError(ctx context.Context, err error, op string) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	// We log it as Error only if it's not a business/expected error
	logger.FromContext(ctx, r.log).Errorw("database error",
		"operation", op,
		"error", err,
	)
}
