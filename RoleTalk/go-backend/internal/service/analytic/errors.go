package analytic

import "errors"

var (
	// ErrUserNotFound indicates the user does not have a skill profile.
	ErrUserNotFound = errors.New("user skill profile not found")
	// ErrInvalidScore indicates the provided progress data is malformed.
	ErrInvalidScore = errors.New("invalid skill score provided")
)
