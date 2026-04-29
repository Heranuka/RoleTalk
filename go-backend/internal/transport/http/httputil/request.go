// Package httputil provides HTTP transport helper utilities.
package httputil

import (
	"net/http"

	"go.uber.org/zap"

	"go-backend/internal/transport/http/render"
)

// DecodeAndValidate decodes JSON request and handles validation errors.
//
// Returns true if the request was invalid and response already sent.
func DecodeAndValidate(
	r *http.Request,
	w http.ResponseWriter,
	log *zap.SugaredLogger,
	dst any,
	action string,
) bool {
	if err := render.Decode(r, dst); err != nil {
		details := render.ValidationErrors(err)

		if len(details) > 0 {
			log.With(zap.Any("details", details)).
				Warnf("%s validation failed", action)

			if sendErr := render.FailWithDetails(w, http.StatusBadRequest, "validation failed", details); sendErr != nil {
				log.With(zap.Error(sendErr)).
					Warn("failed to send validation error response")
			}

			return true
		}

		log.With(zap.Error(err)).
			Warnf("failed to decode %s request", action)

		if sendErr := render.FailMessage(w, http.StatusBadRequest, "invalid request payload"); sendErr != nil {
			log.With(zap.Error(sendErr)).
				Warn("failed to send invalid payload response")
		}

		return true
	}

	return false
}
