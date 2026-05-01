//go:generate mockgen -package=auth -destination=mocks.go go-backend/internal/transport/http/handler/auth Service
package auth

import (
	"context"

	serviceauth "go-backend/internal/service/auth"
)

// Service defines the interface that HTTP handlers use to perform authentication
// and identity management operations.
type Service interface {
	// Register creates a new user account with the provided email and password,
	// and triggers the dispatch of an email verification link.
	Register(ctx context.Context, input serviceauth.RegisterInput) error

	// Login authenticates a user by verifying their email and password.
	// It returns a short-lived access token and a long-lived refresh token on success.
	Login(ctx context.Context, input serviceauth.LoginInput) (string, string, error)

	// VerifyEmail validates the provided secure token and updates the user's
	// status to indicate that their email address has been successfully confirmed.
	VerifyEmail(ctx context.Context, rawToken string) error

	// RequestPasswordReset initiates the password recovery flow. It securely generates
	// a reset token and sends it to the specified email address if the account exists.
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword validates the provided reset token and updates the user's password.
	// It ensures the token is valid, unexpired, and strictly intended for password resets.
	ResetPassword(ctx context.Context, rawToken, newPassword string) error

	// LoginWithGoogle handles the OAuth flow, linking existing accounts by email
	// or creating new ones, then issuing tokens.
	LoginWithGoogle(ctx context.Context, code string) (string, string, error)

	// Refresh consumes a valid, unrevoked refresh token to issue a fresh pair
	// of access and refresh tokens, securely rotating the user's session.
	Refresh(ctx context.Context, oldRawToken string) (string, string, error)

	// ResendVerificationEmail invalidates old verification tokens and sends a fresh one.
	ResendVerificationEmail(ctx context.Context, email string) error
}
