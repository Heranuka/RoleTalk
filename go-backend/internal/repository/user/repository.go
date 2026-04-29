// Package user provides the repository layer for managing user data in PostgreSQL.
package user

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

var tracer = otel.Tracer("internal/repository/user")

// Repository handles database operations for the User entity with observability support.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates and returns a new User repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create inserts a new User into the database and returns the generated UUID.
func (r *Repository) Create(ctx context.Context, u *domain.User) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "User.Repository.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO users (
			email,
			password_hash,
			is_email_verified,
			display_name,
			photo_url,
			interface_lang,
			practice_lang,
			username,
			role
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id uuid.UUID
	err := db.QueryRow(ctx, q,
		u.Email,
		u.PasswordHash,
		u.IsEmailVerified,
		u.DisplayName,
		u.PhotoURL,
		u.InterfaceLang,
		u.PracticeLang,
		u.Username,
		u.Role,
	).Scan(&id)

	if err != nil {
		r.handleError(ctx, err, "failed to create user")
		return uuid.Nil, handlePostgresError(err, "create user")
	}

	return id, nil
}

// GetByEmail retrieves a User from the database using their unique email address.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "User.Repository.GetByEmail")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, email, password_hash, is_email_verified,
			display_name, photo_url, interface_lang, practice_lang,
			username, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	u, err := r.scanUser(db.QueryRow(ctx, q, email))
	if err != nil {
		// Log internal error but return a clean error for the service layer
		if !errors.Is(err, ErrUserNotFound) {
			r.handleError(ctx, err, "database error: get user by email")
		}
		return nil, err
	}

	return u, nil
}

// GetByID retrieves a User from the database using their unique UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "User.Repository.GetByID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, email, password_hash, is_email_verified,
			display_name, photo_url, interface_lang, practice_lang,
			username, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	u, err := r.scanUser(db.QueryRow(ctx, q, id))
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			r.handleError(ctx, err, "database error: get user by id")
		}
		return nil, err
	}

	return u, nil
}

// Update modifies an existing User record and updates the updated_at timestamp.
func (r *Repository) Update(ctx context.Context, u *domain.User) error {
	ctx, span := tracer.Start(ctx, "User.Repository.Update")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE users
		SET
			email             = $2,
			password_hash     = $3,
			is_email_verified = $4,
			display_name      = $5,
			photo_url         = $6,
			interface_lang    = $7,
			practice_lang     = $8,
			username          = $9,
			role              = $10,
			updated_at        = now()
		WHERE id = $1
	`

	cmdTag, err := db.Exec(ctx, q,
		u.ID,
		u.Email,
		u.PasswordHash,
		u.IsEmailVerified,
		u.DisplayName,
		u.PhotoURL,
		u.InterfaceLang,
		u.PracticeLang,
		u.Username,
		u.Role,
	)

	if err != nil {
		r.handleError(ctx, err, "failed to update user")
		return handlePostgresError(err, "update user")
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// IsAdmin checks if the user associated with the given ID has administrative privileges.
func (r *Repository) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	ctx, span := tracer.Start(ctx, "User.Repository.IsAdmin")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `SELECT role FROM users WHERE id = $1`

	var role domain.UserRole
	err := db.QueryRow(ctx, q, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrUserNotFound
		}
		r.handleError(ctx, err, "failed to check user role")
		return false, fmt.Errorf("query user role: %w", err)
	}

	return role == domain.UserRoleAdmin, nil
}

// scanUser is a helper to map a single database row to a domain.User entity.
func (r *Repository) scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsEmailVerified,
		&u.DisplayName,
		&u.PhotoURL,
		&u.InterfaceLang,
		&u.PracticeLang,
		&u.Username,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user row failed: %w", err)
	}

	return &u, nil
}

// handleError reports an error to the current span and logs it with request context.
func (r *Repository) handleError(ctx context.Context, err error, op string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	log.Errorw(op, "error", err)
}

// handlePostgresError maps PostgreSQL-specific constraint violations to domain errors.
func handlePostgresError(err error, operation string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "users_email_key" {
				return ErrUserEmailTaken
			}
			if pgErr.ConstraintName == "users_username_key" {
				return ErrUserUsernameTaken
			}
		case pgerrcode.InvalidTextRepresentation, pgerrcode.InvalidParameterValue:
			return ErrUserInvalidRole
		}
	}
	return fmt.Errorf("db %s failed: %w", operation, err)
}
