// Package message implements HTTP handlers for voice interaction and chat history.
package message

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	serviceai "go-backend/internal/service/ai"
	"go-backend/internal/transport/http/middleware"
	"go-backend/internal/transport/http/render"
)

var tracer = otel.Tracer("internal/transport/http/handler/message")

// Handler manages message and AI voice interactions with integrated observability.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates and returns a new message Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// ProcessVoiceTurn handles POST /sessions/{id}/voice.
// It processes a multipart audio upload and returns the transcribed text and AI vocal response.
func (h *Handler) ProcessVoiceTurn(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.ProcessVoiceTurn")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	// 1. Extract and validate parameters
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		_ = render.Fail(w, http.StatusBadRequest, ErrInvalidSessionID)
		return
	}

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.Warn("unauthorized access attempt to voice processing")
		_ = render.Fail(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	practiceLang := r.Header.Get("X-Practice-Language")
	if practiceLang == "" {
		practiceLang = "English"
	}

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("session.id", sessionID.String()),
		attribute.String("practice.language", practiceLang),
	)

	// 2. Parse audio file from multipart form safely
	r.Body = http.MaxBytesReader(w, r.Body, 11<<20)        // 11MB to allow some overhead for 10MB file
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		log.Warnw("failed to parse multipart form", "error", err)
		_ = render.Fail(w, http.StatusBadRequest, errors.New("invalid multipart form"))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		_ = render.Fail(w, http.StatusBadRequest, ErrAudioFileMissing)
		return
	}
	defer func() { _ = file.Close() }()

	// 3. Execute AI orchestration service
	userText, aiText, aiAudioURL, err := h.service.ProcessVoiceTurn(ctx, userID, sessionID, practiceLang, file)
	if err != nil {
		h.handleError(ctx, w, err, "process_voice_turn")
		return
	}

	// 4. Construct and send response
	// Note: In production, IDs should be returned from the service layer persisted in DB.
	resp := voiceTurnResponse{
		UserText: userText,
		AIResponse: messageResponse{
			ID:          uuid.New(),
			SenderRole:  "ai",
			TextContent: &aiText,
			AudioURL:    &aiAudioURL,
		},
	}

	log.Infow("voice interaction turn completed", "user_id", userID, "session_id", sessionID)
	_ = render.OK(w, resp)
}

// GetHistory handles GET /sessions/{id}/history.
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.GetHistory")
	defer span.End()

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		_ = render.Fail(w, http.StatusBadRequest, ErrInvalidSessionID)
		return
	}

	messages, err := h.service.GetSessionHistory(ctx, sessionID)
	if err != nil {
		h.handleError(ctx, w, err, "get_session_history")
		return
	}

	_ = render.OK(w, toMessageListResponse(messages))
}

// handleError maps business errors to HTTP responses and records system failures in tracing.
func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error, action string) {
	log := logger.FromContext(ctx, h.log)
	span := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, serviceai.ErrInvalidAudioFormat):
		_ = render.Fail(w, http.StatusBadRequest, err)

	case errors.Is(err, serviceai.ErrAIProcessingFailed):
		span.RecordError(err)
		span.SetStatus(codes.Error, "ai_backend_processing_failed")
		log.Errorw("AI service error", "action", action, "error", err)
		_ = render.Fail(w, http.StatusBadGateway, ErrInternalServer)

	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, "internal_handler_failure")
		log.Errorw("unexpected handler error", "action", action, "error", err)
		_ = render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error())
	}
}
