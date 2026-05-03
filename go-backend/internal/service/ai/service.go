// Package ai implements the orchestration logic for voice-to-voice interaction.
package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"go-backend/internal/logger"
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
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
// It waits for the AI response and uploads response audio before returning transcribed text, AI reply text,
// an optional presigned MinIO URL, and the durable object key (for same-origin playback through the REST API).
// User clip archiving and message persistence continue in the background.
func (s *Service) ProcessVoiceTurn(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	practiceLang string,
	audioData io.Reader,
) (userTranscription string, aiText string, aiAudioURL string, aiObjectKey string, err error) {
	ctx, span := tracer.Start(ctx, "Service.ProcessVoiceTurn")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	// 1. Fetch Context
	session, err := s.practiceSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("session lookup: %w", err)
	}
	topic, err := s.topicRepo.GetByID(ctx, session.TopicID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("topic lookup: %w", err)
	}

	// 2. Prepare System Prompt
	systemPrompt, err := s.prompt.RenderRoleplay(domain.RoleplayParams{
		PartnerRole:  topic.PartnerRole,
		Description:  s.ptrString(topic.Description),
		SecretMotive: s.ptrString(topic.PartnerSecretMotive),
		UserRole:     topic.MyRole,
		Goal:         topic.Goal,
		Language:     practiceLang,
	})
	if err != nil {
		return "", "", "", "", fmt.Errorf("prompt generation: %w", err)
	}

	// 3. Buffer Audio (Required for dual-use: gRPC call + background S3 upload)
	audioBytes, err := io.ReadAll(audioData)
	if err != nil {
		return "", "", "", "", fmt.Errorf("audio read: %w", err)
	}

	// 4. Critical Path: Call AI Provider (STT -> LLM -> TTS)
	userTranscription, aiText, aiAudioBytes, err := s.provider.ProcessVoiceTurn(ctx, audioBytes, practiceLang, systemPrompt)
	if err != nil {
		return "", "", "", "", fmt.Errorf("ai provider error: %w", err)
	}

	// 5. Critical Path: Upload AI Response Audio to S3
	aiResFilename := fmt.Sprintf("ai/%s/%s.wav", sessionID, uuid.New())
	aiStorageKey, err := s.storage.Upload(ctx, "ai-responses", aiResFilename, bytes.NewReader(aiAudioBytes))
	if err != nil {
		return userTranscription, aiText, "", "", fmt.Errorf("ai s3 upload: %w", err)
	}

	// 6. Presigned URL for clients that reach MinIO directly; mobile clients typically use REST proxy + audio_object_key instead.
	if aiPresignURL, presignErr := s.storage.GetURL(ctx, aiStorageKey); presignErr != nil {
		log.Warnw("ai audio presign failed; client may still replay via authenticated GET", "error", presignErr)
	} else {
		aiAudioURL = aiPresignURL
	}

	// 7. Background path: non-critical operations (archiving and DB persistence).
	// We use a detached context to ensure these finish even if the user disconnects.
	detachedCtx := context.WithoutCancel(ctx)

	go func(uText, aText, aKey string, uAudio []byte) {
		// Archive user's original voice clip
		userFilename := fmt.Sprintf("users/%s/sessions/%s/%s.m4a", userID, sessionID, uuid.New())
		userStorageKey, uploadErr := s.storage.Upload(detachedCtx, "voice-history", userFilename, bytes.NewReader(uAudio))
		if uploadErr != nil {
			log.Errorw("background user audio upload failed", "error", uploadErr)
		}

		// Persist the conversation turn to the database using STORAGE KEYS (not temporary URLs)
		if err := s.messages.SaveMessage(detachedCtx, sessionID, domain.RoleUser, uText, userStorageKey); err != nil {
			log.Errorw("failed to save user message in background", "error", err)
		}

		if err := s.messages.SaveMessage(detachedCtx, sessionID, domain.RoleAssistant, aText, aKey); err != nil {
			log.Errorw("failed to save ai message in background", "error", err)
		}

		log.Debugw("background archiving completed", "session_id", sessionID)
	}(userTranscription, aiText, aiStorageKey, audioBytes)

	log.Infow("voice turn processed successfully", "session_id", sessionID)

	return userTranscription, aiText, aiAudioURL, aiStorageKey, nil
}

func validAssistantWaveObjectKey(sessionID uuid.UUID, objectKey string) bool {
	objectKey = filepath.ToSlash(filepath.Clean(strings.TrimSpace(objectKey)))
	wantPrefix := fmt.Sprintf("ai-responses/ai/%s/", sessionID.String())
	return strings.HasPrefix(objectKey, wantPrefix) &&
		len(objectKey) > len(wantPrefix) &&
		!strings.Contains(objectKey, "..")
}

// StreamValidatedSessionAIWave streams a WAV artifact for the latest AI reply inside the user's session bucket layout.
func (s *Service) StreamValidatedSessionAIWave(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	objectKey string,
) (_ io.ReadCloser, err error) {
	if strings.TrimSpace(objectKey) == "" {
		return nil, ErrInvalidStoredObjectPath
	}

	session, err := s.practiceSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session lookup: %w", err)
	}
	if session.UserID != userID {
		return nil, ErrAISessionPlaybackForbidden
	}
	if !validAssistantWaveObjectKey(sessionID, objectKey) {
		return nil, ErrInvalidStoredObjectPath
	}

	return s.storage.Load(ctx, objectKey)
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
