// Package ai implements the orchestration logic for voice-to-voice interaction.
package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("internal/service/ai")

// Service coordinates audio storage, AI processing, and message persistence.
type Service struct {
	storage             StorageProvider
	messages            MessageService
	topicRepo           TopicRepository
	practiceSessionRepo PracticeSessionRepository
	prompt              PromptService
	provider            Provider
	log                 *zap.SugaredLogger
}

// NewService creates a new AI orchestration service.
func NewService(
	storage StorageProvider,
	messages MessageService,
	topicRepo TopicRepository,
	practiceSessionRepo PracticeSessionRepository,
	prompt PromptService,
	provider Provider,
	log *zap.SugaredLogger,
) *Service {
	return &Service{
		storage:             storage,
		messages:            messages,
		topicRepo:           topicRepo,
		practiceSessionRepo: practiceSessionRepo,
		prompt:              prompt,
		provider:            provider,
		log:                 log,
	}
}

// ProcessVoiceTurn orchestrates the full AI interaction loop.
// It returns transcribed user text, AI response text, and the URL to the AI voice message.
func (s *Service) ProcessVoiceTurn(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	practiceLang string,
	audioData io.Reader,
) (string, string, string, error) {
	ctx, span := tracer.Start(ctx, "Service.ProcessVoiceTurn",
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("session.id", sessionID.String()),
		),
	)
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	// 1. Fetch Context
	session, err := s.practiceSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", "", "", fmt.Errorf("session lookup: %w", err)
	}

	topic, err := s.topicRepo.GetByID(ctx, session.TopicID)
	if err != nil {
		return "", "", "", fmt.Errorf("topic lookup: %w", err)
	}

	// 2. Buffer Audio (Required for multiple reads)
	audioBytes, err := io.ReadAll(audioData)
	if err != nil {
		return "", "", "", fmt.Errorf("audio read: %w", err)
	}

	// 3. System Prompt
	params := domain.RoleplayParams{
		PartnerRole:  topic.PartnerRole,
		Description:  s.ptrString(topic.Description),
		SecretMotive: s.ptrString(topic.PartnerSecretMotive),
		UserRole:     topic.MyRole,
		Goal:         topic.Goal,
		Language:     practiceLang,
	}

	systemPrompt, err := s.prompt.RenderRoleplay(params)
	if err != nil {
		return "", "", "", fmt.Errorf("prompt generation: %w", err)
	}

	// 4. Archive User Audio (Concurrent)
	userAudioURLChan := make(chan string, 1)
	go func() {
		filename := fmt.Sprintf("users/%s/sessions/%s/%s.m4a", userID, sessionID, uuid.New())
		url, uploadErr := s.storage.Upload(ctx, "voice-history", filename, bytes.NewReader(audioBytes))
		if uploadErr != nil {
			log.Errorw("s3 upload failed", "error", uploadErr)
		}
		userAudioURLChan <- url
	}()

	// 5. Invoke AI Provider (Strict assignment according to interface)
	// Order: UserText, AIText, AIAudioBytes, err
	userTranscription, aiText, aiAudioBytes, err := s.provider.ProcessVoiceTurn(ctx, audioBytes, practiceLang, systemPrompt)
	if err != nil {
		return "", "", "", fmt.Errorf("ai provider error: %w", err)
	}

	userAudioURL := <-userAudioURLChan

	// 6. Persist Messages
	if err := s.messages.SaveMessage(ctx, sessionID, domain.RoleUser, userTranscription, userAudioURL); err != nil {
		log.Errorw("failed to save user message", "error", err)
	}

	// 7. Store AI Voice Response
	aiResFilename := fmt.Sprintf("ai/%s/%s.wav", sessionID, uuid.New())
	aiAudioURL, err := s.storage.Upload(ctx, "ai-responses", aiResFilename, bytes.NewReader(aiAudioBytes))
	if err != nil {
		return userTranscription, aiText, "", fmt.Errorf("ai s3 upload: %w", err)
	}

	if err := s.messages.SaveMessage(ctx, sessionID, domain.RoleAssistant, aiText, aiAudioURL); err != nil {
		log.Errorw("failed to save ai message", "error", err)
	}

	log.Infow("turn processed successfully", "session_id", sessionID)

	return userTranscription, aiText, aiAudioURL, nil
}

// GetSessionHistory retrieves the dialog history for a specific session.
// It delegates the call to the underlying MessageService.
// This is required to satisfy the MessageHandler interface.
func (s *Service) GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error) {
	ctx, span := tracer.Start(ctx, "Service.AI.GetSessionHistory")
	defer span.End()

	res, err := s.messages.GetSessionHistory(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session history: %w", err)
	}
	return res, nil
}

func (s *Service) ptrString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
