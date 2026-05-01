package user

import "errors"

var (
	// ErrUserNotFound is returned when the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserEmailTaken is returned when trying to use an email that is already registered.
	ErrUserEmailTaken = errors.New("email address is already taken")

	// ErrUserUsernameTaken is returned when trying to claim a username that is already in use.
	ErrUserUsernameTaken = errors.New("username is already taken")

	// ErrUserInvalidRole is returned when an invalid user role is provided.
	ErrUserInvalidRole = errors.New("invalid user role")

	// ErrUserInvalidData is returned when essential user data (like email) is missing or malformed.
	ErrUserInvalidData = errors.New("invalid user data provided")

	// ErrPasswordTooLong is returned when the password exceeds the secure hashing limit.
	ErrPasswordTooLong = errors.New("password exceeds maximum allowed length")
)
