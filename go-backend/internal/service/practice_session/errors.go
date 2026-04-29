package practice

import "errors"

var (
	// ErrSessionNotFound is returned when the practice instance does not exist.
	ErrSessionNotFound = errors.New("practice session not found")

	// ErrTopicNotFound indicates the user tried to start a session with a non-existent topic.
	ErrTopicNotFound = errors.New("the selected topic does not exist")

	// ErrActiveSessionExists prevents a user from having multiple concurrent roleplays.
	ErrActiveSessionExists = errors.New("user already has an active practice session")
)
