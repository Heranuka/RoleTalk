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
	"encoding/json"
	"net/http"
)

// Message represents a standard success response containing a message.
type Message struct {
	Message string `json:"message"`
}

// Error represents a standard error response envelope returned by the API.
// The Details field may contain additional structured information about the error.
type Error struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// JSON writes the provided data as a JSON response with the given HTTP status code.
// It sets the appropriate Content-Type header and encodes the response using json.Encoder.
//
// The function returns any encoding error that occurs during JSON serialization.
func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	return enc.Encode(data)
}

// OK writes a JSON response with HTTP status 200 (OK).
func OK(w http.ResponseWriter, data any) error {
	return JSON(w, http.StatusOK, data)
}

// Created writes a JSON response with HTTP status 201 (Created).
func Created(w http.ResponseWriter, data any) error {
	return JSON(w, http.StatusCreated, data)
}

// NoContent writes an HTTP 204 (No Content) response without a body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Msg writes a standard message response with HTTP status 200.
//
// Example response:
//
//	{
//	  "message": "operation completed successfully"
//	}
func Msg(w http.ResponseWriter, msg string) error {
	return OK(w, Message{Message: msg})
}
