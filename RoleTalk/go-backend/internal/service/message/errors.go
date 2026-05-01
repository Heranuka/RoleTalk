// Package message defines business errors for dialog and message history.
package message

import "errors"

var (
	// ErrSessionNotFound is returned when trying to associate a message with a non-existent practice session.
	ErrSessionNotFound = errors.New("practice session not found")

	// ErrInvalidMessageContent is returned when the message text or metadata is missing.
	ErrInvalidMessageContent = errors.New("message content cannot be empty")

	// ErrStorageUploadFailed indicates a failure when saving the audio file to the S3/MinIO storage.
	ErrStorageUploadFailed = errors.New("failed to upload audio content to storage")
)
