// Package analytic implements HTTP handlers for user skill tracking and performance data.
package analytic

import (
	"context"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-backend/internal/logger"
	serviceanalytic "go-backend/internal/service/analytic"
	"go-backend/internal/transport/http/middleware"
	"go-backend/internal/transport/http/render"
)

var tracer = otel.Tracer("internal/transport/http/handler/analytic")

// Handler manages analytic-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates and returns a new analytic Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// GetMySkills handles GET /users/me/skills requests.
// It returns the performance metrics for the currently authenticated user.
func (h *Handler) GetMySkills(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.GetMySkills")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	// 1. Identify user from auth context
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.Warn("unauthorized access attempt to analytics")
		_ = render.Fail(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	span.SetAttributes(attribute.String("user.id", userID.String()))

	// 2. Fetch data from service
	skills, err := h.service.GetUserSkills(ctx, userID)
	if err != nil {
		h.handleError(ctx, w, err, "get_user_skills")
		return
	}

	// 3. Respond with DTO
	log.Debugw("user skills retrieved", "user_id", userID)
	_ = render.OK(w, toSkillResponse(skills))
}

// handleError maps business errors to HTTP responses and instruments spans for observability.
func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error, action string) {
	log := logger.FromContext(ctx, h.log)
	span := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, serviceanalytic.ErrUserNotFound):
		_ = render.Fail(w, http.StatusNotFound, ErrProfileNotFound)

	default:
		// System failures are recorded in Tempo and Loki
		span.RecordError(err)
		span.SetStatus(codes.Error, "internal_analytic_failure")

		log.Errorw("analytic handler failure",
			"action", action,
			"error", err,
		)

		_ = render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error())
	}
}
