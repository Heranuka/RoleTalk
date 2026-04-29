// Package verificationtoken provides the repository layer for managing secure,
// time-limited tokens used for flows such as email confirmation and password resets.
package verificationtoken

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	"go-backend/pkg/transactor"
)

var tracer = otel.Tracer("internal/repository/verificationtoken")

// Repository handles database operations related to verification tokens with integrated observability.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates a new verification token repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new verification token hash into the database.
func (r *Repository) Create(ctx context.Context, token *domain.VerificationToken) error {
	ctx, span := tracer.Start(ctx, "Repository.VerificationToken.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("token.user_id", token.UserID.String()),
		attribute.String("token.purpose", string(token.Purpose)),
	)

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO verification_tokens (
			token_hash,
			user_id,
			purpose,
			expires_at
		) VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := db.QueryRow(ctx, q,
		token.TokenHash,
		token.UserID,
		token.Purpose,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		r.handleError(ctx, err, "failed to create verification token")
		return fmt.Errorf("create verification token: %w", err)
	}

	return nil
}

// GetByHash retrieves a valid token by its hash and purpose.
func (r *Repository) GetByHash(
	ctx context.Context,
	tokenHash string,
	purpose domain.TokenPurpose,
) (*domain.VerificationToken, error) {
	ctx, span := tracer.Start(ctx, "Repository.VerificationToken.GetByHash")
	defer span.End()

	span.SetAttributes(attribute.String("token.purpose", string(purpose)))

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			token_hash, user_id, purpose, expires_at, created_at
		FROM verification_tokens
		WHERE token_hash = $1 
		  AND purpose = $2 
		  AND expires_at > now()
	`

	var token domain.VerificationToken
	err := db.QueryRow(ctx, q, tokenHash, purpose).Scan(
		&token.TokenHash,
		&token.UserID,
		&token.Purpose,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		r.handleError(ctx, err, "failed to get verification token by hash")
		return nil, fmt.Errorf("get verification token: %w", err)
	}

	return &token, nil
}

// Delete removes a verification token by its hash to prevent reuse.
func (r *Repository) Delete(ctx context.Context, tokenHash string) error {
	ctx, span := tracer.Start(ctx, "Repository.VerificationToken.Delete")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `DELETE FROM verification_tokens WHERE token_hash = $1`

	res, err := db.Exec(ctx, q, tokenHash)
	if err != nil {
		r.handleError(ctx, err, "failed to delete verification token")
		return fmt.Errorf("delete verification token: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// DeleteAllForUser removes all existing tokens of a specific purpose for a given user.
func (r *Repository) DeleteAllForUser(
	ctx context.Context,
	userID uuid.UUID,
	purpose domain.TokenPurpose,
) error {
	ctx, span := tracer.Start(ctx, "Repository.VerificationToken.DeleteAllForUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("token.user_id", userID.String()),
		attribute.String("token.purpose", string(purpose)),
	)

	db := transactor.GetDB(ctx, r.db)

	const q = `DELETE FROM verification_tokens WHERE user_id = $1 AND purpose = $2`

	_, err := db.Exec(ctx, q, userID, purpose)
	if err != nil {
		r.handleError(ctx, err, "failed to delete tokens for user")
		return fmt.Errorf("delete user tokens: %w", err)
	}

	return nil
}

// handleError maps and reports internal database failures to both logs and traces.
func (r *Repository) handleError(ctx context.Context, err error, op string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	log.Errorw("verification token database failure",
		"operation", op,
		"error", err,
	)
}
