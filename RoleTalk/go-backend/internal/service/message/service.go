// Package message manages the flow of dialog data between users and AI/Partners.
package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
)

var tracer = otel.Tracer("internal/service/message")

// Service handles persisting and retrieving dialog messages.
type Service struct {
	repo    Repository
	storage StorageProvider
	log     *zap.SugaredLogger
}

// NewService creates a new Message service instance.
func NewService(repo Repository, storage StorageProvider, log *zap.SugaredLogger) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		log:     log,
	}
}

// SaveMessage persists a dialog turn (either from user or AI).
func (s *Service) SaveMessage(ctx context.Context, sessionID uuid.UUID, role, content, audioURL string) error {
	ctx, span := tracer.Start(ctx, "Service.Message.SaveMessage")
	defer span.End()

	msg := &domain.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		SenderRole:  role,
		TextContent: &content,
		AudioURL:    &audioURL,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		s.logger(ctx).Errorw("failed to save message", "session_id", sessionID, "error", err)
		return fmt.Errorf("persist message: %w", err)
	}

	return nil
}

// GetSessionHistory retrieves messages and converts S3 keys into temporary public URLs for the UI.
func (s *Service) GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error) {
	ctx, span := tracer.Start(ctx, "Service.Message.GetSessionHistory")
	defer span.End()

	log := s.logger(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	// 1. Get raw messages from Repository
	messages, err := s.repo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("fetch messages: %w", err)
	}

	// 2. Convert internal S3 paths to public Presigned URLs
	// We set expiration to 15 minutes - enough for a practice session.
	for _, m := range messages {
		if m.AudioURL != nil && *m.AudioURL != "" {
			presignedURL, err := s.storage.GetPresignedURL(ctx, *m.AudioURL, 15*time.Minute)
			if err != nil {
				log.Warnw("failed to sign audio url", "key", *m.AudioURL, "error", err)
				continue // Continue so we don't break the whole list if one icon is missing
			}
			m.AudioURL = &presignedURL
		}
	}

	return messages, nil
}

func (s *Service) logger(ctx context.Context) *zap.SugaredLogger {
	return logger.FromContext(ctx, s.log)
}
