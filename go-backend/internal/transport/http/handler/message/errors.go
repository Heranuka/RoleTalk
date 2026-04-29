package message

import "errors"

var (
	ErrInvalidSessionID = errors.New("invalid session id provided")
	ErrAudioFileMissing = errors.New("audio file is required in multipart form")
	ErrInternalServer   = errors.New("internal server error during voice processing")
)
