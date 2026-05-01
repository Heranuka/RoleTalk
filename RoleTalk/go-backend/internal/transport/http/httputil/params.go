// Package httputil provides HTTP transport helper utilities.
package httputil

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ParseQueryInt extracts an integer query parameter by name.
// Returns the default value if the parameter is missing or invalid.
func ParseQueryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// GetQueryPtrString returns a pointer to a string query parameter, or nil if not present.
func GetQueryPtrString(r *http.Request, key string) *string {
	if v := r.URL.Query().Get(key); v != "" {
		return &v
	}
	return nil
}

// URLParamUUID extracts and parses a UUID URL parameter.
//
// Returns parsed UUID and true if successful.
// If parsing fails, uuid.Nil and false are returned.
func URLParamUUID(r *http.Request, name string) (uuid.UUID, bool) {
	value := chi.URLParam(r, name)

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, false
	}

	return id, true
}
