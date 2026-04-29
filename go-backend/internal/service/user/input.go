// Package user provides data structures and logic for user management.
package user

import (
	"go-backend/internal/models/domain"

	"github.com/google/uuid"
)

// CreateWithPasswordInput holds the data required to register a user via email and password.
type CreateWithPasswordInput struct {
	Email       string
	Password    string
	DisplayName string
	Username    *string
	PhotoURL    *string
}

// CreateOAuthInput holds the data required to register a user via a third-party provider.
type CreateOAuthInput struct {
	Email           string
	DisplayName     string
	Username        *string
	PhotoURL        *string
	IsEmailVerified bool
}

// UpdateInput holds data for modifying an existing user's profile.
// Only non-nil fields will be applied.
type UpdateInput struct {
	ID            uuid.UUID
	DisplayName   *string
	Username      *string
	PhotoURL      *string
	InterfaceLang *string
	PracticeLang  *string
	Role          *domain.UserRole
}
