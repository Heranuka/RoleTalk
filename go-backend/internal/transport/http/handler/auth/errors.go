package auth

import "errors"

var (
	// ErrInvalidCredentials indicates the provided email or password is incorrect.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrEmailNotVerified indicates the user must confirm their email before logging in.
	ErrEmailNotVerified = errors.New("email address is not verified")

	// ErrInvalidToken indicates a verification or reset token is invalid, expired, or used.
	ErrInvalidToken = errors.New("invalid or expired token")

	// ErrOAuthFailed indicates an error occurred while communicating with the OAuth provider.
	ErrOAuthFailed = errors.New("failed to authenticate via third-party provider")

	// ErrEmailTaken indicates the requested email is already registered.
	ErrEmailTaken = errors.New("email is already registered")

	// ErrUsernameTaken indicates the requested username is already in use.
	ErrUsernameTaken = errors.New("username is already taken")

	// ErrInternalServer indicates an unexpected internal server error.
	ErrInternalServer = errors.New("internal server error")

	// ErrAlreadyVerified indicates the user's email is already verified.
	ErrAlreadyVerified = errors.New("email is already verified")
)
