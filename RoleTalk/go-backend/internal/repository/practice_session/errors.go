// Package session defines repository-level errors for practice session management.
package session

import "errors"

var (
	// ErrSessionNotFound is returned when the requested practice session does not exist.
	ErrSessionNotFound = errors.New("practice session not found")

	// ErrActiveSessionExists is returned when a user tries to start a new session while having one "active".
	ErrActiveSessionExists = errors.New("user already has an active practice session")
)
