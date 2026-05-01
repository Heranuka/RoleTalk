package message

import "errors"

var (
	// ErrInvalidSessionID is returned when the session ID in the URL is malformed.
	ErrInvalidSessionID = errors.New("invalid session id provided")
	// ErrAudioFileMissing is returned when the multipart form does not contain an audio file.
	ErrAudioFileMissing = errors.New("audio file is required in multipart form")
	// ErrInternalServer is a generic error for unexpected failures.
	ErrInternalServer = errors.New("internal server error during voice processing")
)
