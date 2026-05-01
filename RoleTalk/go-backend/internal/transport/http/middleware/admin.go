// Package middleware provides HTTP middleware components for the Backforge API.
// It includes utilities for logging, authentication, authorization, and rate limiting.
package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-backend/internal/transport/http/render"
)

var (
	// ErrForbidden indicates the user does not have sufficient privileges.
	ErrForbidden = errors.New("forbidden: admin access required")
)

// UserRoleChecker defines minimal interface for checking user roles.
// This allows middleware to depend on an interface, not concrete service.
type UserRoleChecker interface {
	// IsAdmin checks if a given user ID corresponds to an admin user.
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

// AdminOnly returns a middleware that ensures the user has admin role.
// Depends on UserRoleChecker interface, not concrete service.
func AdminOnly(log *zap.SugaredLogger, userSvc UserRoleChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				log.Warn("user ID not found in context")
				if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
					log.With(zap.Error(err)).Warn("failed to send unauthorized response")
				}
				return
			}

			isAdmin, err := userSvc.IsAdmin(r.Context(), userID)
			if err != nil {
				log.With(zap.Error(err), "user_id", userID).Error("failed to check admin role")
				if sendErr := render.FailMessage(w, http.StatusInternalServerError, "failed to check user role"); sendErr != nil {
					log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
				}
				return
			}

			if !isAdmin {
				log.With("user_id", userID).Warn("user attempted to access admin route without permission")
				if sendErr := render.Fail(w, http.StatusForbidden, ErrForbidden); sendErr != nil {
					log.With(zap.Error(sendErr)).Warn("failed to send forbidden response")
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
