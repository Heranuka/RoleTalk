// Package topic implements HTTP handlers for scenario discovery and social interactions.
package topic

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	servicetopic "go-backend/internal/service/topic"
	"go-backend/internal/transport/http/middleware"
	"go-backend/internal/transport/http/render"
)

var tracer = otel.Tracer("internal/transport/http/handler/topic")

// Handler manages scenario-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates and returns a new topic Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// GetOfficial handles GET /topics/official requests for the AI Solo tab.
func (h *Handler) GetOfficial(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.GetOfficial")
	defer span.End()

	topics, err := h.service.GetAIRecommended(ctx)
	if err != nil {
		h.handleError(ctx, w, err, "fetch official topics")
		return
	}

	_ = render.OK(w, toTopicListResponse(topics))
}

// GetCommunity handles GET /topics/community requests for the Social hub.
func (h *Handler) GetCommunity(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.GetCommunity")
	defer span.End()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	topics, err := h.service.GetCommunityFeed(ctx, limit, offset)
	if err != nil {
		h.handleError(ctx, w, err, "fetch community feed")
		return
	}

	_ = render.OK(w, toTopicListResponse(topics))
}

// Create handles POST /topics requests to publish a user scenario.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.CreateTopic")
	defer span.End()

	userID, _ := middleware.UserIDFromContext(ctx)

	var req createTopicRequest
	if err := render.Decode(r, &req); err != nil {
		_ = render.Fail(w, http.StatusBadRequest, err)
		return
	}

	id, err := h.service.CreateTopic(ctx, userID, req.Title, req.Description, req.Emoji, req.DifficultyLevel)
	if err != nil {
		h.handleError(ctx, w, err, "create community topic")
		return
	}

	_ = render.Created(w, map[string]interface{}{"id": id})
}

// Delete handles the DELETE /api/v1/topics/{id} request.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Delete")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	// 1. Extract Topic ID from path
	topicID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		_ = render.Fail(w, http.StatusBadRequest, ErrInvalidData)
		return
	}

	// 2. Extract Authenticated User ID from context (provided by Auth Middleware)
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.Error("user id missing from authenticated context")
		_ = render.Fail(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	span.SetAttributes(attribute.String("topic.id", topicID.String()))

	// 3. Invoke Service with ownership validation
	if err := h.service.DeleteTopic(ctx, userID, topicID); err != nil {
		h.handleError(ctx, w, err, "delete_topic")
		return
	}

	// 4. Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// AddLike handles POST /topics/{id}/like requests.
func (h *Handler) AddLike(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.AddLike")
	defer span.End()

	topicID, _ := uuid.Parse(chi.URLParam(r, "id"))
	userID, _ := middleware.UserIDFromContext(ctx)

	span.SetAttributes(attribute.String("topic.id", topicID.String()))

	if err := h.service.LikeTopic(ctx, userID, topicID); err != nil {
		h.handleError(ctx, w, err, "like topic")
		return
	}

	_ = render.OK(w, nil)
}

// RemoveLike handles DELETE /topics/{id}/like requests.
func (h *Handler) RemoveLike(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.RemoveLike")
	defer span.End()

	topicID, _ := uuid.Parse(chi.URLParam(r, "id"))
	userID, _ := middleware.UserIDFromContext(ctx)

	if err := h.service.RemoveLike(ctx, userID, topicID); err != nil {
		h.handleError(ctx, w, err, "remove like")
		return
	}

	_ = render.OK(w, nil)
}

// handleError maps domain/service errors to HTTP responses and records failures in tracing.
func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error, action string) {
	log := logger.FromContext(ctx, h.log)
	span := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, servicetopic.ErrTopicNotFound):
		_ = render.Fail(w, http.StatusNotFound, ErrTopicNotFound)

	case errors.Is(err, servicetopic.ErrAlreadyLiked):
		_ = render.Fail(w, http.StatusConflict, ErrAlreadyLiked)

	case errors.Is(err, servicetopic.ErrUnauthorizedAction):
		_ = render.Fail(w, http.StatusForbidden, ErrUnauthorizedAction)

	case errors.Is(err, servicetopic.ErrInvalidTopicData):
		_ = render.Fail(w, http.StatusBadRequest, ErrInvalidData)

	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, "internal_error")
		log.Errorw("topic handler: service failure", "action", action, "error", err)
		_ = render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error())
	}
}
