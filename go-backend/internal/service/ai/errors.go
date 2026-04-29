// Package ai defines business errors for AI processing orchestration.
package ai

import "errors"

var (
	// ErrAIProcessingFailed is returned when the Python AI service fails to process audio.
	ErrAIProcessingFailed = errors.New("ai service processing failed")

	// ErrStorageUploadFailed indicates a failure when saving audio to MinIO/S3.
	ErrStorageUploadFailed = errors.New("failed to upload audio to storage")

	// ErrInvalidAudioFormat is returned if the provided audio file is corrupted or unsupported.
	ErrInvalidAudioFormat = errors.New("invalid audio format provided")
)
