// Package topic implements business logic for managing roleplay scenarios.
package topic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
)

var tracer = otel.Tracer("internal/service/topic")

// Service coordinates operations for both official and user-generated scenarios.
type Service struct {
	repo       Repository
	transactor Transactor
	log        *zap.SugaredLogger
}

// NewService creates a new Topic service instance.
func NewService(repo Repository, transactor Transactor, log *zap.SugaredLogger) *Service {
	return &Service{
		repo:       repo,
		transactor: transactor,
		log:        log,
	}
}

// GetAIRecommended retrieves curated scenarios for solo practice with AI.
func (s *Service) GetAIRecommended(ctx context.Context) ([]*domain.Topic, error) {
	ctx, span := tracer.Start(ctx, "Service.Topic.GetAIRecommended")
	defer span.End()

	topics, err := s.repo.GetOfficial(ctx)
	if err != nil {
		s.logger(ctx).Errorw("failed to fetch official topics", "error", err)
		return nil, fmt.Errorf("get official: %w", err)
	}

	return topics, nil
}

// GetCommunityFeed retrieves popular user-generated scenarios with pagination.
func (s *Service) GetCommunityFeed(ctx context.Context, limit, offset int) ([]*domain.Topic, error) {
	ctx, span := tracer.Start(ctx, "Service.Topic.GetCommunityFeed")
	defer span.End()

	res, err := s.repo.GetCommunity(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch community feed: %w", err)
	}
	return res, nil
}

// CreateTopic publishes a new scenario created by a user.
func (s *Service) CreateTopic(ctx context.Context, userID uuid.UUID, title, desc, emoji, level string) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Service.Topic.CreateTopic")
	defer span.End()

	if title == "" || emoji == "" {
		return uuid.Nil, ErrInvalidTopicData
	}

	topic := &domain.Topic{
		AuthorID:        &userID,
		Title:           title,
		Description:     &desc,
		Emoji:           &emoji,
		DifficultyLevel: &level,
		IsOfficial:      false,
	}

	id, err := s.repo.Create(ctx, topic)
	if err != nil {
		return uuid.Nil, fmt.Errorf("publish topic: %w", err)
	}

	s.logger(ctx).Infow("community topic published", "topic_id", id, "author_id", userID)
	return id, nil
}

// LikeTopic records a user's appreciation for a specific topic within a database transaction.
// It ensures that the operation is atomic and follows strict error wrapping standards.
func (s *Service) LikeTopic(ctx context.Context, userID, topicID uuid.UUID) error {
	// Execute the operation within a transaction to maintain data integrity.
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.AddLike(txCtx, userID, topicID); err != nil {
			// Wrap with lower-level context (Repository layer).
			return fmt.Errorf("repository: %w", err)
		}
		return nil
	})

	if err != nil {
		// Wrap with high-level business context (Service layer).
		// This satisfies the wrapcheck linter and provides a clear audit trail.
		return fmt.Errorf("topic service: failed to like topic: %w", err)
	}

	return nil
}

// RemoveLike deletes a user's appreciation for a specific topic.
func (s *Service) RemoveLike(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Topic.RemoveLike")
	defer span.End()

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		err := s.repo.RemoveLike(txCtx, userID, topicID)
		if err != nil {
			return fmt.Errorf("remove like: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("remove like: %w", err)
	}
	return nil
}

// DeleteTopic verifies ownership and removes the specified scenario.
func (s *Service) DeleteTopic(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Topic.DeleteUserTopic")
	defer span.End()

	log := s.logger(ctx)

	// 1. Fetch the topic to check ownership
	topic, err := s.repo.GetByID(ctx, topicID)
	if err != nil {
		return fmt.Errorf("get topic: %w", err)
	}

	// 2. Authorization check: only the author can delete their topic.
	// We also prevent users from deleting "Official" topics.
	if topic.IsOfficial || topic.AuthorID == nil || *topic.AuthorID != userID {
		log.Warnw("unauthorized delete attempt", "user_id", userID, "topic_id", topicID)
		return ErrUnauthorizedAction
	}

	// 3. Perform deletion
	if err := s.repo.Delete(ctx, topicID); err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	log.Infow("topic deleted", "topic_id", topicID, "deleted_by", userID)
	return nil
}

// logger returns a contextualized logger instance.
func (s *Service) logger(ctx context.Context) *zap.SugaredLogger {
	return logger.FromContext(ctx, s.log)
}
