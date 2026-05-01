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

	return s.repo.GetCommunity(ctx, limit, offset)
}

// CreateTopic publishes a new scenario created by a user.
func (s *Service) CreateTopic(ctx context.Context, userID uuid.UUID, title, desc, emoji, level string) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "Service.Topic.CreateTopic")
	defer span.End()

	if title == "" || emoji == "" {
		return uuid.Nil, ErrInvalidTopicData
	}

	topic := &domain.Topic{
		ID:              uuid.New(),
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

// LikeTopic handles the social interaction of liking a scenario.
func (s *Service) LikeTopic(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Topic.LikeTopic")
	defer span.End()

	// Use transaction to ensure both like record and counter are updated
	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		return s.repo.AddLike(txCtx, userID, topicID)
	})
}

func (s *Service) RemoveLike(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Topic.RemoveLike")
	defer span.End()

	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		return s.repo.RemoveLike(txCtx, userID, topicID)
	})
}

// DeleteTopic verifies ownership and removes the specified scenario.
func (s *Service) DeleteTopic(ctx context.Context, userID, topicID uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "Service.Topic.DeleteUserTopic")
	defer span.End()

	log := s.logger(ctx)

	// 1. Fetch the topic to check ownership
	topic, err := s.repo.GetByID(ctx, topicID)
	if err != nil {
		return err // Already mapped to ErrTopicNotFound in repo
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
