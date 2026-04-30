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

// Service coordinates practice session lifecycles and triggers post-session analytics.
type Service struct {
	repo        Repository
	topicRepo   TopicRepository
	analyticSvc AnalyticService // New dependency for AI evaluation
	log         *zap.SugaredLogger
}

// NewService creates a new practice session Service instance.
func NewService(
	repo Repository,
	tRepo TopicRepository,
	aSvc AnalyticService,
	log *zap.SugaredLogger,
) *Service {
	return &Service{
		repo:        repo,
		topicRepo:   tRepo,
		analyticSvc: aSvc,
		log:         log,
	}
}

// StartSession initializes a new practice instance for a user.
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

	// 2. Check if user already has an active session
	active, err := s.repo.GetActiveByUserID(ctx, userID)
	if err == nil && active != nil {
		log.Warnw("user attempted to start multiple sessions", "user_id", userID)
		return uuid.Nil, ErrActiveSessionExists
	}

	// 3. Create domain entity (PracticeSession instead of Session)
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

// CompleteSession marks an ongoing session as finished and triggers AI skill evaluation.
func (s *Service) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Practice.CompleteSession")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	// 1. Fetch the session to get the UserID for the analytics service
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repopractice.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("fetch session: %w", err)
	}

	// 2. Update status to 'completed' in DB
	err = s.repo.UpdateStatus(ctx, sessionID, "completed")
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// 3. Trigger AI Evaluation
	// This is the core "Feedback Loop". We pass the UserID and SessionID.
	// The Analytic service will fetch the transcript and update user_skills.
	err = s.analyticSvc.EvaluateSession(ctx, session.UserID, sessionID)
	if err != nil {
		// We log this as an error but don't strictly fail the "completion"
		// because the session is already marked as done in the DB.
		log.Errorw("post-session evaluation failed",
			"session_id", sessionID,
			"error", err,
		)
		return fmt.Errorf("evaluate session: %w", err)
	}

	log.Infow("practice session completed and evaluated", "session_id", sessionID)
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
