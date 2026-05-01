// Package render provides helpers for decoding HTTP requests,
// validating input payloads, and rendering JSON responses in HTTP handlers.
//
// The package centralizes common HTTP transport logic such as:
//
//   - JSON request decoding
//   - request validation
//   - consistent error handling
//
// It is intended to be used inside the HTTP transport layer of the application.
package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrEmptyBody indicates that the HTTP request body is empty.
	ErrEmptyBody = errors.New("request body is empty")

	// ErrMultipleJSON indicates that the request body contains
	// more than one JSON object.
	ErrMultipleJSON = errors.New("only one JSON object allowed")
)

// Decode reads the JSON body from the provided HTTP request,
// decodes it into dst, and validates the resulting structure.
//
// The function also enforces the following constraints:
//
//   - the request body must not be empty
//   - unknown JSON fields are rejected
//   - only a single JSON object is allowed
//   - the decoded structure must pass validation
//
// The dst argument must be a pointer to the destination struct.
//
// Example:
//
//	var input CreateUserRequest
//	if err := render.Decode(r, &input); err != nil {
//	    // handle error
//	}
func Decode(r *http.Request, dst any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return ErrEmptyBody
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	if dec.More() {
		return ErrMultipleJSON
	}

	if err := Validate(dst); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
