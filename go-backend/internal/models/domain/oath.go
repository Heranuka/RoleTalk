// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthProvider represents supported third-party identity providers.
type OAuthProvider string

const (
	// OAuthProviderGoogle represents Google OAuth login.
	OAuthProviderGoogle OAuthProvider = "google"
)

// OAuthConnection links a local user account to a third-party identity provider.
type OAuthConnection struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       OAuthProvider
	ProviderUserID string // stored as a string to accommodate various provider formats
	CreatedAt      time.Time
}

// NewOAuthConnection creates a new association between a local user and an OAuth provider.
func NewOAuthConnection(userID uuid.UUID, provider OAuthProvider, providerUserID string) *OAuthConnection {
	return &OAuthConnection{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}
}
