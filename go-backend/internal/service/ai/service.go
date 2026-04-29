// Package ai implements the orchestration logic for voice-to-voice interaction.
package ai

import (
	"bytes"
	"context"
	"fmt"
	"go-backend/internal/models/domain"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go-backend/internal/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("internal/service/ai")

// Service coordinates audio storage, AI processing, and message persistence.
type Service struct {
	storage    StorageProvider
	messages   MessageService
	httpClient *http.Client
	pythonURL  string
	log        *zap.SugaredLogger
}

// NewService creates a new AI orchestration service.
func NewService(
	storage StorageProvider,
	messages MessageService,
	pythonURL string,
	log *zap.SugaredLogger,
) *Service {
	return &Service{
		storage:  storage,
		messages: messages,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   90 * time.Second,
		},
		pythonURL: pythonURL,
		log:       log,
	}
}

// ProcessVoiceTurn orchestrates the full AI interaction loop:
// 1. Buffers audio for multiple reads.
// 2. Uploads user's original voice to S3 for history.
// 3. Calls Python AI service for STT, LLM reasoning, and TTS.
// 4. Saves both User and AI messages to the database.
// 5. Returns the resulting texts and audio link to the handler.
func (s *Service) ProcessVoiceTurn(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
	practiceLang string,
	audioData io.Reader,
) (string, string, string, error) {
	ctx, span := tracer.Start(ctx, "Service.AI.ProcessVoiceTurn")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	requestID := span.SpanContext().TraceID().String()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("session.id", sessionID.String()),
		attribute.String("practice.lang", practiceLang),
	)

	// STEP 1: Buffer the audio data to memory.
	// This is necessary because we need to read the audio twice:
	// once for S3 archival and once for the Python AI service call.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, audioData); err != nil {
		return "", "", "", fmt.Errorf("failed to buffer audio stream: %w", err)
	}
	audioBytes := buf.Bytes()

	// STEP 2: Upload User Audio to S3 (Local MinIO) for archival.
	userFilename := fmt.Sprintf("%s/%s.m4a", userID, uuid.New())
	userAudioURL, err := s.storage.Upload(ctx, "voice-history", userFilename, bytes.NewReader(audioBytes))
	if err != nil {
		span.RecordError(err)
		log.Errorw("failed to archive user audio in s3", "error", err)
		return "", "", "", fmt.Errorf("%w: %v", ErrStorageUploadFailed, err)
	}

	// STEP 3: Invoke Python AI Service.
	// This performs STT (Whisper), LLM (Ollama), and TTS (Piper) in one step.
	aiText, aiAudioBytes, userText, err := s.callPythonAI(ctx, audioBytes, practiceLang, requestID)
	if err != nil {
		span.RecordError(err)
		return "", "", "", fmt.Errorf("%w: %v", ErrAIProcessingFailed, err)
	}

	// STEP 4: Persist the User's message (Transcribed text).
	if err := s.messages.SaveMessage(ctx, sessionID, "user", userText, userAudioURL); err != nil {
		log.Errorw("failed to persist user message in database", "error", err)
		// We continue even if DB fails to ensure the user gets a voice response.
	}

	// STEP 5: Upload the AI's generated response audio to S3.
	aiFilename := fmt.Sprintf("%s/%s.wav", sessionID, uuid.New())
	aiAudioURL, err := s.storage.Upload(ctx, "ai-responses", aiFilename, bytes.NewReader(aiAudioBytes))
	if err != nil {
		log.Errorw("failed to upload generated ai audio to s3", "error", err)
	}

	// STEP 6: Persist the AI's message (Generated text).
	if err := s.messages.SaveMessage(ctx, sessionID, "ai", aiText, aiAudioURL); err != nil {
		log.Errorw("failed to persist ai message in database", "error", err)
	}

	log.Infow("voice dialog turn completed successfully",
		"user_id", userID,
		"session_id", sessionID,
		"trace_id", requestID,
	)

	// Return the results for the Flutter client.
	return userText, aiText, aiAudioURL, nil
}

// GetSessionHistory retrieves dialog history.
// This method implements the interface required by the message.Handler.
func (s *Service) GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*domain.Message, error) {
	ctx, span := tracer.Start(ctx, "Service.AI.GetSessionHistory")
	defer span.End()

	// We delegate the database call to the dedicated message service
	return s.messages.GetSessionHistory(ctx, sessionID)
}

// callPythonAI performs a multipart POST request to the Python AI microservice.
func (s *Service) callPythonAI(
	ctx context.Context,
	audio []byte,
	lang string,
	requestID string,
) (string, []byte, string, error) {
	ctx, span := tracer.Start(ctx, "Service.AI.callPythonAI")
	defer span.End()

	// FIX 2: Construct proper Multipart request as expected by FastAPI UploadFile
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "input.m4a")
	if err != nil {
		return "", nil, "", fmt.Errorf("multipart creator: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return "", nil, "", fmt.Errorf("multipart writer: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", s.pythonURL, body)
	if err != nil {
		return "", nil, "", fmt.Errorf("http request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("X-Practice-Language", lang)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", nil, "", fmt.Errorf("http execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, "", fmt.Errorf("ai server error: status %d", resp.StatusCode)
	}

	// Extract metadata from headers returned by Python script
	userText := resp.Header.Get("X-STT-Transcription")
	aiText := resp.Header.Get("X-LLM-Response")

	aiAudio, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to read response audio: %w", err)
	}

	return aiText, aiAudio, userText, nil
}
