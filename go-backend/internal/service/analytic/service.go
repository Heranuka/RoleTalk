// Package analytic implements business logic for user progress tracking and skill analysis.
package analytic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
)

var tracer = otel.Tracer("internal/service/analytic")

// Service coordinates skill updates and retrieves user performance metrics.
type Service struct {
	repo       Repository
	transactor Transactor
	log        *zap.SugaredLogger
}

// NewService creates a new Analytic service instance.
func NewService(repo Repository, transactor Transactor, log *zap.SugaredLogger) *Service {
	return &Service{
		repo:       repo,
		transactor: transactor,
		log:        log,
	}
}

// GetUserSkills retrieves the current skill profile for a specific user.
func (s *Service) GetUserSkills(ctx context.Context, userID uuid.UUID) (*domain.UserSkill, error) {
	ctx, span := tracer.Start(ctx, "Service.Analytic.GetUserSkills")
	defer span.End()

	skills, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		// Log internal error but let high-level errors propagate
		s.logger(ctx).Errorw("failed to fetch user skills", "user_id", userID, "error", err)
		return nil, fmt.Errorf("fetch skills: %w", err)
	}

	return skills, nil
}

// InitializeProfile creates a starting skill record for a new user.
func (s *Service) InitializeProfile(ctx context.Context, userID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Analytic.InitializeProfile")
	defer span.End()

	if err := s.repo.Create(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "initialization failed")
		return fmt.Errorf("initialize skills: %w", err)
	}

	s.logger(ctx).Infow("skill profile initialized", "user_id", userID)
	return nil
}

// ProcessSessionProgress updates user skills based on a completed roleplay session.
// It retrieves current stats, applies increments provided by the AI, and persists the update.
func (s *Service) ProcessSessionProgress(
	ctx context.Context,
	userID uuid.UUID,
	empInc, persInc, strucInc, stressInc int,
) error {
	ctx, span := tracer.Start(ctx, "Service.Analytic.ProcessSessionProgress")
	defer span.End()

	log := s.logger(ctx)

	// We use a transaction to ensure that we read and update consistently
	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Get current skills
		skills, err := s.repo.GetByUserID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("get current skills: %w", err)
		}

		// 2. Apply business logic (clamping 0-100) inside the domain layer
		skills.ApplyProgress(empInc, persInc, strucInc, stressInc)

		// 3. Persist updated skills
		if err := s.repo.UpdateSkills(txCtx, skills); err != nil {
			return fmt.Errorf("update skills: %w", err)
		}

		log.Infow("user skills updated from session",
			"user_id", userID,
			"empathy_change", empInc,
			"persuasion_change", persInc,
		)

		return nil
	})
}

// logger returns a contextualized logger instance with trace information.
func (s *Service) logger(ctx context.Context) *zap.SugaredLogger {
	return logger.FromContext(ctx, s.log)
}
