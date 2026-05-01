package user

import "errors"

var (
	// ErrUserNotFound is returned when the requested user does not exist in the database.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserEmailTaken is returned when attempting to create or update a user with an email
	// address that is already registered to another account.
	ErrUserEmailTaken = errors.New("email address is already taken")

	// ErrUserUsernameTaken is returned when attempting to claim a username
	// that is already associated with another user account.
	ErrUserUsernameTaken = errors.New("username is already taken")

	// ErrUserInvalidRole is returned when a user role value does not match any of the allowed
	// values defined in the user_role enum type in the database.
	ErrUserInvalidRole = errors.New("invalid user role")
)
