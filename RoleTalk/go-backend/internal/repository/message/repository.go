// Package message implements the database repository for persisting dialog history.
package message

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

var tracer = otel.Tracer("internal/repository/message")

// Repository handles PostgreSQL operations for Message entities.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates and returns a new Message repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new message into the database.
func (r *Repository) Create(ctx context.Context, m *domain.Message) error {
	ctx, span := tracer.Start(ctx, "Message.Repository.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO messages (
			id, session_id, sender_role, text_content, audio_url, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := db.Exec(ctx, q,
		m.ID,
		m.SessionID,
		m.SenderRole,
		m.TextContent,
		m.AudioURL,
		m.CreatedAt,
	)

	if err != nil {
		r.handleInternalError(ctx, err, "failed to insert message")
		return r.handlePostgresError(err, "create message")
	}

	return nil
}

// GetBySessionID retrieves all messages for a specific session ordered by time.
func (r *Repository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error) {
	ctx, span := tracer.Start(ctx, "Message.Repository.GetBySessionID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT id, session_id, sender_role, text_content, audio_url, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := db.Query(ctx, q, sessionID)
	if err != nil {
		r.handleInternalError(ctx, err, "failed to query messages")
		return nil, r.handlePostgresError(err, "query messages")
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		m := &domain.Message{}
		err := rows.Scan(
			&m.ID, &m.SessionID, &m.SenderRole, &m.TextContent, &m.AudioURL, &m.CreatedAt,
		)
		if err != nil {
			r.handleInternalError(ctx, err, "failed to scan message row")
			return nil, r.handlePostgresError(err, "scan messages")
		}
		messages = append(messages, m)
	}

	return messages, nil
}

// handleInternalError records technical failures in the trace span and logs them.
func (r *Repository) handleInternalError(ctx context.Context, err error, message string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, message)

	log.Errorw(message, "error", err)
}

// handlePostgresError maps PostgreSQL-specific errors to domain errors.
func (r *Repository) handlePostgresError(err error, operation string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.ForeignKeyViolation {
			return ErrSessionNotFound
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil // Returning empty list is fine for GetBySessionID
	}

	return fmt.Errorf("db %s failed: %w", operation, err)
}
