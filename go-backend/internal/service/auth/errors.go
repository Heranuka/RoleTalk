package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when the email or password does not match.
	// For security reasons, we do not specify which one is incorrect.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrEmailNotVerified is returned when a user attempts to log in before verifying their email.
	ErrEmailNotVerified = errors.New("email address is not verified")

	// ErrInvalidVerificationToken is returned when a token is invalid, expired, or used for the wrong purpose.
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")

	// ErrOAuthExchangeFailed is returned when the service fails to exchange the OAuth code for a user profile.
	ErrOAuthExchangeFailed = errors.New("failed to exchange oauth code")

	// ErrRefreshTokenInvalid is returned when the refresh token is malformed, expired, or unrecognized.
	ErrRefreshTokenInvalid = errors.New("invalid or expired refresh token")

	// ErrRefreshTokenRevoked is returned when attempting to use a refresh token that has been explicitly revoked.
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")

	// ErrRefreshTokenAlreadyExists is returned on a highly improbable token collision.
	ErrRefreshTokenAlreadyExists = errors.New("refresh token already exists")

	// ErrEmailAlreadyVerified is returned when a user requests a verification link but is already verified.
	ErrEmailAlreadyVerified = errors.New("user is already verified")
)
