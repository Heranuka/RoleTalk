// Package render provides helpers for decoding HTTP requests,
// validating input payloads, and rendering JSON responses in HTTP handlers.
//
// The package centralizes common HTTP transport logic such as:
//
//   - JSON response encoding
//   - consistent API response formats
//   - HTTP status helpers
//   - request validation utilities
//
// It is intended to be used within the HTTP transport layer of the application.
package render

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is the shared validator instance used across the render package.
var validate = func() *validator.Validate {
	v := validator.New()

	// RegisterTagNameFunc configures the validator to use JSON tag names
	// instead of struct field names when reporting validation errors.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "" {
			return fld.Name
		}

		name = strings.Split(name, ",")[0]

		if name == "-" || name == "" {
			return fld.Name
		}

		return name
	})

	return v
}()

// Validate validates the provided struct using the configured validator.
//
// The function expects v to be a struct or a pointer to a struct containing
// validation tags compatible with github.com/go-playground/validator.
//
// Example:
//
//	type CreateUserRequest struct {
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	err := render.Validate(req)
func Validate(v any) error {
	if err := validate.Struct(v); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// ValidationErrors converts validator validation errors into a
// client-friendly map of field names to error messages.
//
// The returned map uses JSON field names instead of Go struct field names.
//
// Example response:
//
//	{
//	  "email": "must be a valid email",
//	  "password": "must be at least 8 characters"
//	}
//
// If the provided error is not a validator.ValidationErrors instance,
// an empty map is returned.
func ValidationErrors(err error) map[string]string {
	errs := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errs
	}

	for _, e := range validationErrors {
		field := e.Field()

		switch e.Tag() {
		case "required":
			errs[field] = "field is required"

		case "email":
			errs[field] = "must be a valid email"

		case "min":
			if e.Kind() == reflect.String {
				errs[field] = fmt.Sprintf("must be at least %s characters", e.Param())
			} else {
				errs[field] = fmt.Sprintf("must be at least %s", e.Param())
			}

		case "max":
			if e.Kind() == reflect.String {
				errs[field] = fmt.Sprintf("must be at most %s characters", e.Param())
			} else {
				errs[field] = fmt.Sprintf("must be at most %s", e.Param())
			}

		case "oneof":
			errs[field] = fmt.Sprintf("must be one of: %s", e.Param())

		case "uuid":
			errs[field] = "must be a valid UUID"

		case "url":
			errs[field] = "must be a valid URL"

		default:
			errs[field] = "invalid value"
		}
	}

	return errs
}
