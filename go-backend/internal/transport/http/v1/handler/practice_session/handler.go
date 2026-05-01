package practice_session

import (
	"context"
	"errors"
	"go-backend/internal/logger"
	"go-backend/internal/transport/http/middleware"
	"go-backend/internal/transport/http/render"
	"net/http"

	servicepractice "go-backend/internal/service/practice_session"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("internal/transport/http/handler/practice")

// Handler manages practice-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates and returns a new practice session Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// Start handles POST /api/v1/sessions requests to initialize a new roleplay.
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Practice.Start")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	// 1. Identify user
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		_ = render.Fail(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	// 2. Decode request
	var req startSessionRequest
	if err := render.Decode(r, &req); err != nil {
		_ = render.Fail(w, http.StatusBadRequest, err)
		return
	}

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("topic.id", req.TopicID.String()),
	)

	// 3. Invoke Service
	id, err := h.service.StartSession(ctx, userID, req.TopicID)
	if err != nil {
		h.handleError(ctx, w, err, "start_session")
		return
	}

	log.Infow("practice session initialized", "session_id", id, "user_id", userID)
	_ = render.Created(w, map[string]uuid.UUID{"id": id})
}

// GetByID handles GET /api/v1/sessions/{id} requests.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Practice.GetByID")
	defer span.End()

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		_ = render.Fail(w, http.StatusBadRequest, errors.New("invalid session id"))
		return
	}

	session, err := h.service.GetSession(ctx, sessionID)
	if err != nil {
		h.handleError(ctx, w, err, "get_session")
		return
	}

	_ = render.OK(w, toSessionResponse(session))
}

// Complete handles POST /api/v1/sessions/{id}/complete requests.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Practice.Complete")
	defer span.End()

	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		_ = render.Fail(w, http.StatusBadRequest, errors.New("invalid session id"))
		return
	}

	if err := h.service.CompleteSession(ctx, sessionID); err != nil {
		h.handleError(ctx, w, err, "complete_session")
		return
	}

	_ = render.OK(w, nil)
}

// handleError maps practice service errors to HTTP responses and records failures for observability.
func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error, action string) {
	log := logger.FromContext(ctx, h.log)
	span := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, servicepractice.ErrTopicNotFound):
		_ = render.Fail(w, http.StatusNotFound, err)

	case errors.Is(err, servicepractice.ErrActiveSessionExists):
		_ = render.Fail(w, http.StatusConflict, err)

	case errors.Is(err, servicepractice.ErrSessionNotFound):
		_ = render.Fail(w, http.StatusNotFound, err)

	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, "internal_failure")
		log.Errorw("practice handler failure", "action", action, "error", err)
		_ = render.FailMessage(w, http.StatusInternalServerError, "internal server error")
	}
}
