// Package message manages the flow of dialog data between users and AI/Partners.
package message

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"
)

var tracer = otel.Tracer("internal/service/message")

// Service handles persisting and retrieving dialog messages.
type Service struct {
	repo Repository
	log  *zap.SugaredLogger
}

// NewService creates a new Message service instance.
func NewService(repo Repository, log *zap.SugaredLogger) *Service {
	return &Service{
		repo: repo,
		log:  log,
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

// GetSessionHistory retrieves all messages for a specific session to display in UI.
func (s *Service) GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error) {
	ctx, span := tracer.Start(ctx, "Service.Message.GetSessionHistory")
	defer span.End()

	return s.repo.GetBySessionID(ctx, sessionID)
}

func (s *Service) logger(ctx context.Context) *zap.SugaredLogger {
	return logger.FromContext(ctx, s.log)
}
