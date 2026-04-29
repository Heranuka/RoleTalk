// Package render provides helpers for decoding HTTP requests,
// validating input payloads, and rendering JSON responses in HTTP handlers.
//
// The package centralizes common HTTP transport logic such as:
//
//   - JSON response encoding
//   - consistent API response formats
//   - HTTP status helpers
//
// It is intended to be used within the HTTP transport layer of the application.
package render

import (
	"errors"
	"net/http"
)

// ErrUnknown represents a generic fallback error used when
// no specific error message is available.
var ErrUnknown = errors.New("unknown error")

// Fail writes an error response using the provided HTTP status code
// and error value.
//
// If err is nil, the response will contain a generic "unknown error" message.
// The error message is serialized into the standard API error envelope.
func Fail(w http.ResponseWriter, status int, err error) error {
	if err == nil {
		err = ErrUnknown
	}

	return JSON(w, status, Error{
		Error: err.Error(),
	})
}

// FailMessage writes an error response using the provided HTTP status code
// and a plain message string.
//
// If the message is empty, a generic "unknown error" message is used.
func FailMessage(w http.ResponseWriter, status int, message string) error {
	if message == "" {
		message = ErrUnknown.Error()
	}

	return JSON(w, status, Error{
		Error: message,
	})
}

// FailWithDetails writes an error response with additional structured details.
// The details field can contain any extra information useful to the client,
// such as validation errors.
//
// If the message is empty, a generic "unknown error" message is used.
func FailWithDetails(w http.ResponseWriter, status int, message string, details any) error {
	if message == "" {
		message = ErrUnknown.Error()
	}

	return JSON(w, status, Error{
		Error:   message,
		Details: details,
	})
}
