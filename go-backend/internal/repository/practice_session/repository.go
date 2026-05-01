// Package session implements the database repository for managing practice sessions.
package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	"go-backend/pkg/transactor"
)

var tracer = otel.Tracer("internal/repository/session")

// Repository handles PostgreSQL operations for roleplay sessions.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates a new practice session repository.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create persists a new practice session.
func (r *Repository) Create(ctx context.Context, s *domain.PracticeSession) error {
	ctx, span := tracer.Start(ctx, "Repository.Session.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO practice_sessions (id, user_id, topic_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := db.Exec(ctx, q, s.ID, s.UserID, s.TopicID, s.Status, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		r.handleError(ctx, err, "failed to create practice session")
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

// UpdateStatus changes the status of a practice session (e.g., to 'completed').
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	ctx, span := tracer.Start(ctx, "Repository.Session.UpdateStatus")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `UPDATE practice_sessions SET status = $2, updated_at = NOW() WHERE id = $1`

	res, err := db.Exec(ctx, q, id, status)
	if err != nil {
		r.handleError(ctx, err, "failed to update status")
		return fmt.Errorf("db.Exec: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// GetByID fetches a single session by its UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSession, error) {
	ctx, span := tracer.Start(ctx, "Repository.Session.GetByID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT id, user_id, topic_id, status, created_at, updated_at 
		FROM practice_sessions 
		WHERE id = $1
	`

	var s domain.PracticeSession
	err := db.QueryRow(ctx, q, id).Scan(&s.ID, &s.UserID, &s.TopicID, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		r.handleError(ctx, err, "failed to get session")
		return nil, fmt.Errorf("db.Scan: %w", err)
	}

	return &s, nil
}

// GetActiveByUserID finds the current ongoing practice session for a specific user.
// Returns nil, nil if no active session is found.
func (r *Repository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.PracticeSession, error) {
	ctx, span := tracer.Start(ctx, "Repository.PracticeSession.GetActiveByUserID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT id, user_id, topic_id, status, created_at, updated_at
		FROM practice_sessions
		WHERE user_id = $1 AND status = 'active'
		LIMIT 1
	`

	var s domain.PracticeSession
	err := db.QueryRow(ctx, q, userID).Scan(
		&s.ID, &s.UserID, &s.TopicID, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil //nolint:nilnil
		}
		r.handleError(ctx, err, "failed to query active session")
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return &s, nil
}

// handleError maps and logs database errors for observability.
func (r *Repository) handleError(ctx context.Context, err error, op string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	log.Errorw("practice session database failure", "op", op, "error", err)
}
