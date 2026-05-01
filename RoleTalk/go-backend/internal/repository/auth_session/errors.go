// Package session provides the repository layer for accessing session entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, and manage user sessions.
package auth_session

import "errors"

var (
	// ErrSessionNotFound is returned when the provided session does not exist
	// in the database or has been deleted.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionAlreadyExists is returned when attempting to create a session
	// that violates the unique constraint on the token field.
	ErrSessionAlreadyExists = errors.New("session already exists")
)
