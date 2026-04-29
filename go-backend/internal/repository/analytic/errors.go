// Package analytic defines repository-level errors for user skill management.
package analytic

import "errors"

var (
	// ErrAnalyticNotFound is returned when a skill profile for a specific user does not exist in the database.
	ErrAnalyticNotFound = errors.New("analytic record not found")

	// ErrAnalyticAlreadyExists is returned if an attempt is made to create a duplicate skill record for a user.
	ErrAnalyticAlreadyExists = errors.New("analytic record already exists for this user")

	// ErrUpdateFailed is returned when the database operation to modify skills completes without affecting any rows.
	ErrUpdateFailed = errors.New("failed to update analytic record")
)
