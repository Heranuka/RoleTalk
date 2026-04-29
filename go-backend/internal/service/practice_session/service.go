// Package practice implements business logic for managing roleplay practice sessions.
package practice

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
	repopractice "go-backend/internal/repository/practice_session"
	repotopic "go-backend/internal/repository/topic"
)

var tracer = otel.Tracer("internal/service/practice")

// Service coordinates practice session lifecycles and interactions.
type Service struct {
	repo      Repository
	topicRepo TopicRepository
	log       *zap.SugaredLogger
}

// NewService creates a new practice session Service instance.
func NewService(repo Repository, tRepo TopicRepository, log *zap.SugaredLogger) *Service {
	return &Service{
		repo:      repo,
		topicRepo: tRepo,
		log:       log,
	}
}

// StartSession initializes a new practice instance for a user.
// It verifies the topic existence and ensures no other active session exists for the user.
func (s *Service) StartSession(ctx context.Context, userID, topicID uuid.UUID) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Service.Practice.StartSession")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("topic.id", topicID.String()),
	)

	// 1. Verify that the selected topic exists
	_, err := s.topicRepo.GetByID(ctx, topicID)
	if err != nil {
		if errors.Is(err, repotopic.ErrTopicNotFound) {
			return uuid.Nil, ErrTopicNotFound
		}
		return uuid.Nil, fmt.Errorf("verify topic: %w", err)
	}

	// 2. Check if user already has an active session (Optional Business Rule)
	active, err := s.repo.GetActiveByUserID(ctx, userID)
	if err == nil && active != nil {
		log.Warnw("user attempted to start multiple sessions", "user_id", userID)
		return uuid.Nil, ErrActiveSessionExists
	}

	// 3. Create domain entity
	session := domain.NewPracticeSession(userID, topicID)

	// 4. Persist in database
	if err := s.repo.Create(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create session")
		return uuid.Nil, fmt.Errorf("create session: %w", err)
	}

	log.Infow("new practice session started", "session_id", session.ID, "user_id", userID)
	return session.ID, nil
}

// CompleteSession marks an ongoing session as successfully finished.
func (s *Service) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Practice.CompleteSession")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	err := s.repo.UpdateStatus(ctx, sessionID, "completed")
	if err != nil {
		if errors.Is(err, repopractice.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("update session status: %w", err)
	}

	log.Infow("practice session completed", "session_id", sessionID)
	return nil
}

// GetSession retrieves details of a specific session.
func (s *Service) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.PracticeSession, error) {
	ctx, span := tracer.Start(ctx, "Service.Practice.GetSession")
	defer span.End()

	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repopractice.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	return session, nil
}
