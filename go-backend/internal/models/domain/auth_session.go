// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuthSession represents a refresh token session for maintaining user authentication.
type AuthSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewAuthSession initializes a session for a new login.
func NewAuthSession(userID uuid.UUID, tokenHash string, expiresAt time.Time) *AuthSession {
	return &AuthSession{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
