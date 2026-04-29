package message

import "errors"

var (
	// ErrSessionNotFound is returned when trying to add a message to a non-existent session.
	ErrSessionNotFound = errors.New("parent session not found")
	// ErrMessageCreationFailed is returned when the database fails to persist the message.
	ErrMessageCreationFailed = errors.New("could not create message")
)
