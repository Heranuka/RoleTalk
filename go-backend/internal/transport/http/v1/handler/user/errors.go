package user

import "errors"

var (
	// ErrUserNotFound indicates the user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUnauthorized indicates the user is not authenticated.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrUsernameTaken indicates the requested username is already in use by another account.
	ErrUsernameTaken = errors.New("username is already taken")

	// ErrInternalServer indicates an unexpected internal server error.
	ErrInternalServer = errors.New("internal server error")

	// ErrInvalidData indicates invalid data
	ErrInvalidData = errors.New("invalid data")
)
