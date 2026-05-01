// Package auth implements HTTP handlers for user authentication, registration, and identity management.
package auth

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
	serviceauth "go-backend/internal/service/auth"
	serviceuser "go-backend/internal/service/user"
	"go-backend/internal/transport/http/httputil"
	"go-backend/internal/transport/http/render"
)

var tracer = otel.Tracer("internal/transport/http/handler/auth")

// Handler manages authentication-related requests with integrated observability.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates and returns a new authentication Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// Register handles user account creation.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Register")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	var req registerRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "register") {
		return
	}

	span.SetAttributes(attribute.String("auth.email", req.Email))

	input := serviceauth.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Username:    req.Username,
		PhotoURL:    req.PhotoURL,
	}

	if err := h.service.Register(ctx, input); err != nil {
		h.handleError(ctx, w, err, "registration")
		return
	}

	log.Infow("user registration initiated", "email", req.Email)
	_ = render.Created(w, render.Message{
		Message: "Registration successful. Please check your email to verify your account.",
	})
}

// Login authenticates a user and returns a token pair.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Login")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	var req loginRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "login") {
		return
	}

	at, rt, err := h.service.Login(ctx, serviceauth.LoginInput(req))
	if err != nil {
		h.handleError(ctx, w, err, "login")
		return
	}

	log.Infow("user logged in successfully", "email", req.Email)
	_ = render.OK(w, tokenResponse{AccessToken: at, RefreshToken: rt})
}

// VerifyEmail confirms the user's email address using a provided token.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.VerifyEmail")
	defer span.End()

	var req verifyEmailRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "verify_email") {
		return
	}

	if err := h.service.VerifyEmail(ctx, req.Token); err != nil {
		h.handleError(ctx, w, err, "email verification")
		return
	}

	_ = render.Msg(w, "email verified successfully")
}

// RequestPasswordReset initiates the password recovery process.
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.RequestPasswordReset")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	var req requestResetRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "request_reset") {
		return
	}

	if err := h.service.RequestPasswordReset(ctx, req.Email); err != nil {
		h.handleError(ctx, w, err, "password reset request")
		return
	}

	log.Infow("password reset email sent", "email", req.Email)
	// Success message is generic to prevent email enumeration
	_ = render.Msg(w, "if that email is registered, a password reset link has been sent")
}

// ResetPassword sets a new password using a valid reset token.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.ResetPassword")
	defer span.End()

	var req resetPasswordRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "reset_password") {
		return
	}

	if err := h.service.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		h.handleError(ctx, w, err, "reset password")
		return
	}

	_ = render.Msg(w, "password reset successfully")
}

// ResendVerification sends a new email verification link.
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.ResendVerification")
	defer span.End()

	var req resendVerificationRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "resend_verification") {
		return
	}

	if err := h.service.ResendVerificationEmail(ctx, req.Email); err != nil {
		h.handleError(ctx, w, err, "resend verification")
		return
	}

	_ = render.Msg(w, "if the account exists and is unverified, a new link has been sent")
}

// GoogleCallback handles Google OAuth2 authorization code exchange.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.GoogleCallback")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	var req googleCallbackRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "google_callback") {
		return
	}

	at, rt, err := h.service.LoginWithGoogle(ctx, req.Code)
	if err != nil {
		h.handleError(ctx, w, err, "google oauth")
		return
	}

	log.Info("successful google oauth login")
	_ = render.OK(w, tokenResponse{AccessToken: at, RefreshToken: rt})
}

// Refresh issues a new token pair using a valid refresh token.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Refresh")
	defer span.End()

	var req refreshRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "refresh") {
		return
	}

	at, rt, err := h.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		h.handleError(ctx, w, err, "token refresh")
		return
	}

	_ = render.OK(w, tokenResponse{AccessToken: at, RefreshToken: rt})
}

// handleError maps business and service errors to HTTP responses and records failures in tracing/logs.
func (h *Handler) handleError(ctx context.Context, w http.ResponseWriter, err error, action string) {
	log := logger.FromContext(ctx, h.log)
	currentSpan := trace.SpanFromContext(ctx)

	switch {
	case errors.Is(err, serviceauth.ErrInvalidCredentials):
		currentSpan.AddEvent("auth_failure", trace.WithAttributes(attribute.String("reason", "credentials")))
		_ = render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials)

	case errors.Is(err, serviceauth.ErrEmailNotVerified):
		currentSpan.AddEvent("auth_failure", trace.WithAttributes(attribute.String("reason", "unverified_email")))
		_ = render.Fail(w, http.StatusUnauthorized, ErrEmailNotVerified)

	case errors.Is(err, serviceauth.ErrRefreshTokenInvalid), errors.Is(err, serviceauth.ErrRefreshTokenRevoked):
		currentSpan.AddEvent("auth_failure", trace.WithAttributes(attribute.String("reason", "invalid_refresh_token")))
		_ = render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials)

	case errors.Is(err, serviceauth.ErrInvalidVerificationToken):
		_ = render.Fail(w, http.StatusBadRequest, ErrInvalidToken)

	case errors.Is(err, serviceauth.ErrEmailAlreadyVerified):
		_ = render.Fail(w, http.StatusBadRequest, ErrAlreadyVerified)

	case errors.Is(err, serviceauth.ErrOAuthExchangeFailed):
		log.Errorw("external oauth error", "action", action, "error", err)
		_ = render.Fail(w, http.StatusBadGateway, ErrOAuthFailed)

	case errors.Is(err, serviceuser.ErrUserEmailTaken):
		_ = render.Fail(w, http.StatusConflict, ErrEmailTaken)

	case errors.Is(err, serviceuser.ErrUserUsernameTaken):
		_ = render.Fail(w, http.StatusConflict, ErrUsernameTaken)

	default:
		log.Errorw("authentication service failure", "action", action, "error", err)
		currentSpan.RecordError(err)
		currentSpan.SetStatus(codes.Error, "internal_service_error")
		_ = render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error())
	}
}
