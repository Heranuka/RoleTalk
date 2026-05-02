// Package auth implements HTTP handlers for user authentication, registration, and identity management.
package auth

import (
	"context"
	"errors"
	"fmt"
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

// Logout handles POST /api/v1/auth/logout.
// It accepts a refresh token and revokes it to prevent further access.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.Logout")
	defer span.End()

	log := logger.FromContext(ctx, h.log)

	var req logoutRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "logout") {
		return
	}

	if err := h.service.Logout(ctx, req.RefreshToken); err != nil {
		// Even if logout fails internally, we usually don't want to leak
		// details to the client, but for tracing we record it.
		h.handleError(ctx, w, err, "logout")
		return
	}

	log.Info("logout request processed")

	// 204 No Content is the standard response for a successful side-effect only request
	w.WriteHeader(http.StatusNoContent)
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

// VerifyEmailWeb extracted from your previous snippet, placed in the correct handler package.
func (h *Handler) VerifyEmailWeb(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.VerifyEmailWeb")
	defer span.End()

	token := r.URL.Query().Get("token")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var title, message, icon, iconColor string

	if token == "" {
		title = "Invalid Link"
		message = "The verification token is missing. Please use the link provided in your email."
		icon = "❓"
		iconColor = "#EF4444"
	} else {
		err := h.service.VerifyEmail(ctx, token)
		if err != nil {
			title = "Verification Failed"
			message = "The link has expired or the token is invalid. Please request a new verification email in the app."
			icon = "⚠️"
			iconColor = "#F59E0B"
		} else {
			title = "Email Verified!"
			message = "Your email has been successfully confirmed. You can now access all features of Role Talk."
			icon = "✅"
			iconColor = "#10B981"
		}
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Role Talk | Verification</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background-color: #F9FAFB; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
			.card { background: white; padding: 48px 32px; border-radius: 32px; box-shadow: 0 20px 50px rgba(0,0,0,0.04); text-align: center; max-width: 420px; width: 90%%; border: 1px solid #F3F4F6; }
			.brand { color: #00A67E; font-weight: 900; font-size: 22px; letter-spacing: -1px; margin-bottom: 32px; }
			.icon-circle { width: 80px; height: 80px; background-color: %s15; border-radius: 50%%; display: flex; justify-content: center; align-items: center; margin: 0 auto 24px; font-size: 40px; }
			h1 { color: #111827; font-size: 26px; font-weight: 800; margin: 0 0 12px; letter-spacing: -0.5px; }
			p { color: #6B7280; font-size: 16px; line-height: 1.6; margin: 0 0 32px; }
			.btn { background-color: #00A67E; color: white; padding: 16px 40px; border-radius: 16px; text-decoration: none; font-weight: 700; font-size: 15px; transition: all 0.2s ease; display: inline-block; box-shadow: 0 4px 12px rgba(0,166,126,0.2); }
			.btn:hover { transform: translateY(-2px); box-shadow: 0 6px 15px rgba(0,166,126,0.3); }
		</style>
	</head>
	<body>
		<div class="card">
			<div class="brand">ROLE TALK</div>
			<div class="icon-circle" style="background-color: %s15; color: %s;">%s</div>
			<h1>%s</h1>
			<p>%s</p>
			<a href="roletalk://verify" class="btn">Return to App</a>
		</div>
	</body>
	</html>`, iconColor, iconColor, iconColor, icon, title, message)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
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
