package analytic

import "errors"

var (
	// ErrProfileNotFound is returned when a user does not have an initialized skill record.
	ErrProfileNotFound = errors.New("user skill profile not found")
	// ErrInternalServer is a generic error for unexpected system failures.
	ErrInternalServer = errors.New("internal server error")
)
