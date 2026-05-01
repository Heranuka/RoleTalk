// Package domain defines core business entities and types used across the application.
package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrTokenGenerationFailed is returned when the crypto/rand reader fails.
	ErrTokenGenerationFailed = errors.New("failed to generate secure random token")
)

// TokenPurpose defines the authorization context for a verification token.
type TokenPurpose string

const (
	//nolint:gosec // This is a domain constant, not a hardcoded credential
	// TokenPurposeEmailVerification is used to confirm a user's email address upon registration.
	TokenPurposeEmailVerification TokenPurpose = "email_verification"

	// TokenPurposePasswordReset is used to grant access to reset a forgotten password.
	TokenPurposePasswordReset TokenPurpose = "password_reset"
)

// VerificationToken represents a secure, time-limited token for out-of-band verification flows.
type VerificationToken struct {
	TokenHash string
	UserID    uuid.UUID
	Purpose   TokenPurpose
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NewVerificationToken generates a secure random token, hashes it for storage,
// and returns both the raw token string (to be emailed) and the domain entity (to be stored).
func NewVerificationToken(userID uuid.UUID, purpose TokenPurpose, ttl time.Duration) (string, *VerificationToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, ErrTokenGenerationFailed
	}

	// Use RawURLEncoding so the token can be safely included in URL query parameters.
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := HashVerificationToken(rawToken)

	entity := &VerificationToken{
		TokenHash: tokenHash,
		UserID:    userID,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
	}

	return rawToken, entity, nil
}

// HashVerificationToken computes a SHA-256 hash of the raw token string.
// This ensures that even if the database is compromised, the tokens cannot be used.
func HashVerificationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// IsExpired checks if the token has passed its expiration time.
func (t *VerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
