// Package oauthconnection provides the repository layer for managing third-party
// identity provider connections linked to local user accounts.
package oauthconnection

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	"go-backend/pkg/transactor"
)

var tracer = otel.Tracer("internal/repository/oauthconnection")

// Repository handles database operations for OAuth connections with observability support.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates and returns a new OAuth connection repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new OAuth connection linking a local user to a third-party provider.
func (r *Repository) Create(ctx context.Context, conn *domain.OAuthConnection) error {
	ctx, span := tracer.Start(ctx, "Repository.OAuthConnection.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("oauth.provider", string(conn.Provider)),
		attribute.String("oauth.provider_user_id", conn.ProviderUserID),
	)

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO oauth_connections (
			id, user_id, provider, provider_user_id, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := db.Exec(ctx, q,
		conn.ID,
		conn.UserID,
		conn.Provider,
		conn.ProviderUserID,
		conn.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateConnection
		}
		r.handleError(ctx, err, "failed to create oauth connection")
		return fmt.Errorf("oauth_connections.Create: %w", err)
	}

	return nil
}

// GetByProviderUserID retrieves a connection using the provider name and the external ID.
func (r *Repository) GetByProviderUserID(
	ctx context.Context,
	provider domain.OAuthProvider,
	providerUserID string,
) (*domain.OAuthConnection, error) {
	ctx, span := tracer.Start(ctx, "Repository.OAuthConnection.GetByProviderUserID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT id, user_id, provider, provider_user_id, created_at
		FROM oauth_connections
		WHERE provider = $1 AND provider_user_id = $2
	`

	var conn domain.OAuthConnection
	err := db.QueryRow(ctx, q, provider, providerUserID).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.Provider,
		&conn.ProviderUserID,
		&conn.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}
		r.handleError(ctx, err, "failed to fetch oauth connection by provider id")
		return nil, fmt.Errorf("oauth_connections.GetByProviderUserID: %w", err)
	}

	return &conn, nil
}

// GetByUserID retrieves all OAuth connections linked to a specific local user.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OAuthConnection, error) {
	ctx, span := tracer.Start(ctx, "Repository.OAuthConnection.GetByUserID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT id, user_id, provider, provider_user_id, created_at
		FROM oauth_connections
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(ctx, q, userID)
	if err != nil {
		r.handleError(ctx, err, "failed to query oauth connections by user id")
		return nil, fmt.Errorf("oauth_connections.GetByUserID: %w", err)
	}
	defer rows.Close()

	var connections []*domain.OAuthConnection
	for rows.Next() {
		conn := &domain.OAuthConnection{}
		if err := rows.Scan(
			&conn.ID,
			&conn.UserID,
			&conn.Provider,
			&conn.ProviderUserID,
			&conn.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oauth connection: %w", err)
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// Delete removes a specific OAuth connection by unlinking the external account.
func (r *Repository) Delete(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	ctx, span := tracer.Start(ctx, "Repository.OAuthConnection.Delete")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM oauth_connections
		WHERE user_id = $1 AND provider = $2
	`

	cmd, err := db.Exec(ctx, q, userID, provider)
	if err != nil {
		r.handleError(ctx, err, "failed to delete oauth connection")
		return fmt.Errorf("oauth_connections.Delete: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return ErrConnectionNotFound
	}

	return nil
}

// handleError records database failures in the trace span and logs them via Zap.
func (r *Repository) handleError(ctx context.Context, err error, op string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	log.Errorw("oauth connection database error",
		"operation", op,
		"error", err,
	)
}
