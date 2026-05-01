package oauthconnection

import "errors"

var (
	// ErrConnectionNotFound is returned when the requested OAuth connection does not exist.
	ErrConnectionNotFound = errors.New("oauth connection not found")

	// ErrDuplicateConnection is returned when attempting to link a third-party account
	// that is already linked to another user in the system.
	ErrDuplicateConnection = errors.New("this third-party account is already linked to another user")
)
