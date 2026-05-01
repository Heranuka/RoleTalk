// Package user implements HTTP handlers for user management.
package user

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
	serviceuser "go-backend/internal/service/user"
	"go-backend/internal/transport/http/middleware"
	"go-backend/internal/transport/http/render"
)

var tracer = otel.Tracer("internal/transport/http/handler/user")

// Handler handles user-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates and returns a new user Handler with a logger.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// GetProfile handles GET /users/me requests.
// It retrieves the profile of the currently authenticated user.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.GetProfile")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.Warn("unauthorized access attempt")
		_ = render.Fail(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	span.SetAttributes(attribute.String("user.id", userID.String()))

	u, err := h.service.GetByID(ctx, userID)
	if err != nil {
		h.handleError(ctx, w, err, "get profile")
		return
	}

	log.Debugw("profile retrieved", "user_id", userID)
	_ = render.OK(w, toUserResponse(u))
}

// UpdateProfile handles PATCH /users/me requests.
// It allows users to update their display name, username, photo, and language settings.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.UpdateProfile")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.Warn("unauthorized update attempt")
		_ = render.Fail(w, http.StatusUnauthorized, ErrUnauthorized)
		return
	}

	span.SetAttributes(attribute.String("user.id", userID.String()))

	var req updateProfileRequest
	if err := render.Decode(r, &req); err != nil {
		validationErrs := render.ValidationErrors(err)
		if len(validationErrs) > 0 {
			_ = render.FailWithDetails(w, http.StatusBadRequest, "validation failed", validationErrs)
			return
		}

		log.Warnw("failed to decode request", "error", err)
		_ = render.FailMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := serviceuser.UpdateInput{
		ID:            userID,
		DisplayName:   req.DisplayName,
		Username:      req.Username,
		PhotoURL:      req.PhotoURL, // Consistent with your request
		InterfaceLang: req.InterfaceLang,
		PracticeLang:  req.PracticeLang,
	}

	if err := h.service.Update(ctx, input); err != nil {
		h.handleError(ctx, w, err, "update profile")
		return
	}

	log.Infow("profile updated successfully", "user_id", userID)
	_ = render.Msg(w, "profile updated successfully")
}

// handleError maps service-level errors to HTTP responses and records failures in the trace and logs.
func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error, action string) {
	log := logger.FromContext(ctx, h.log)

	// Get the existing span from the context (it was started in GetProfile/UpdateProfile)
	span := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, serviceuser.ErrUserNotFound):
		// For expected business errors, we usually don't set span status to Error
		_ = render.Fail(w, http.StatusNotFound, ErrUserNotFound)

	case errors.Is(err, serviceuser.ErrUserUsernameTaken):
		log.Warnw("username conflict", "action", action)
		_ = render.Fail(w, http.StatusConflict, ErrUsernameTaken)

	case errors.Is(err, serviceuser.ErrUserInvalidData):
		_ = render.Fail(w, http.StatusBadRequest, ErrInvalidData)

	default:
		// For unexpected internal errors, we record details for Tempo and Loki
		span.RecordError(err)
		span.SetStatus(codes.Error, "internal server error during "+action)

		log.Errorw("service call failed",
			"action", action,
			"error", err,
		)

		_ = render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error())
	}
}
