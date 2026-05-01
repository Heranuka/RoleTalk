// Package analytic manages user skill progression and performance metrics in PostgreSQL.
package analytic

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

var tracer = otel.Tracer("internal/repository/analytic")

// Repository handles database operations for user skill sets.
type Repository struct {
	db  transactor.DBTx
	log *zap.SugaredLogger
}

// NewRepository creates a new Analytic repository instance.
func NewRepository(db transactor.DBTx, log *zap.SugaredLogger) *Repository {
	return &Repository{db: db, log: log}
}

// Create initializes a new skill profile for a user with zero values.
// This is typically called during user registration.
func (r *Repository) Create(ctx context.Context, userID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Repository.Analytic.Create")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO user_skills (user_id, empathy, persuasion, structure, stress_resistance, updated_at)
		VALUES ($1, 0, 0, 0, 0, NOW())
	`

	_, err := db.Exec(ctx, q, userID)
	if err != nil {
		r.handleError(ctx, err, "failed to initialize user skills")
		return fmt.Errorf("analytic.Create: %w", err)
	}

	return nil
}

// GetByUserID retrieves the current skill profile for a user.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserSkill, error) {
	ctx, span := tracer.Start(ctx, "Repository.Analytic.GetByUserID")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT user_id, empathy, persuasion, structure, stress_resistance, updated_at
		FROM user_skills
		WHERE user_id = $1
	`

	var s domain.UserSkill
	err := db.QueryRow(ctx, q, userID).Scan(
		&s.UserID, &s.Empathy, &s.Persuasion, &s.Structure, &s.StressResistance, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAnalyticNotFound
		}
		r.handleError(ctx, err, "failed to get user skills")
		return nil, fmt.Errorf("analytic.GetByUserID: %w", err)
	}

	return &s, nil
}

// UpdateSkills modifies the user's skill levels and updates the timestamp.
func (r *Repository) UpdateSkills(ctx context.Context, s *domain.UserSkill) error {
	ctx, span := tracer.Start(ctx, "Repository.Analytic.UpdateSkills")
	defer span.End()

	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE user_skills
		SET empathy = $2, persuasion = $3, structure = $4, stress_resistance = $5, updated_at = NOW()
		WHERE user_id = $1
	`

	res, err := db.Exec(ctx, q, s.UserID, s.Empathy, s.Persuasion, s.Structure, s.StressResistance)
	if err != nil {
		r.handleError(ctx, err, "failed to update user skills")
		return fmt.Errorf("analytic.UpdateSkills: %w", err)
	}

	// Check if the record actually existed
	if res.RowsAffected() == 0 {
		return ErrAnalyticNotFound
	}

	return nil
}

// handleError records the technical failure in the trace span and logs it with request context.
func (r *Repository) handleError(ctx context.Context, err error, op string) {
	log := logger.FromContext(ctx, r.log)
	span := trace.SpanFromContext(ctx)

	span.RecordError(err)
	span.SetStatus(codes.Error, op)

	log.Errorw(op, "error", err)
}
