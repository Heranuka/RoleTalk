package verificationtoken

import "errors"

var (
	// ErrTokenNotFound is returned when a verification token does not exist,
	// has already been used (and deleted), or has expired.
	ErrTokenNotFound = errors.New("verification token is invalid or has expired")
)
