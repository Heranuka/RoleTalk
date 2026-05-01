//go:generate mockgen -package=auth -destination=mocks.go go-backend/internal/service/auth UserService,SessionRepository,OAuthConnectionRepository,VerificationTokenRepository,Transactor,EmailSender,OAuthClient
package auth

import (
	"context"

	"github.com/google/uuid"

	"go-backend/internal/models/domain"
	"go-backend/internal/service/user"
)

// UserService defines the required operations from the user domain service.
// It acts as an internal boundary between the auth logic and user management.
type UserService interface {
	// CreateWithPassword creates a new local user with an email and hashed password.
	CreateWithPassword(ctx context.Context, input user.CreateWithPasswordInput) (uuid.UUID, error)

	// CreateOAuthUser creates a new user originating from a third-party identity provider.
	CreateOAuthUser(ctx context.Context, input user.CreateOAuthInput) (uuid.UUID, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByID retrieves a user by their unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// MarkEmailVerified updates the user's status to indicate their email is confirmed.
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error

	// SetNewPassword securely hashes and updates the user's password.
	SetNewPassword(ctx context.Context, id uuid.UUID, newPassword string) error
}

// AuthSessionRepository defines persistence operations for user sessions (refresh tokens).
type AuthSessionRepository interface {
	// Create stores a new session record containing the refresh token hash.
	Create(ctx context.Context, s *domain.AuthSession) error

	// GetByTokenHash retrieves an active session by the SHA-256 hash of its refresh token.
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error)

	// Revoke explicitly invalidates a session, preventing future token refreshes.
	Revoke(ctx context.Context, tokenHash string) error
}

// OAuthConnectionRepository manages links between local accounts and third-party providers.
type OAuthConnectionRepository interface {
	// Create persists a new link between a local user and an external OAuth profile.
	Create(ctx context.Context, conn *domain.OAuthConnection) error

	// GetByProviderUserID retrieves an existing connection using the external provider's user ID.
	GetByProviderUserID(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.OAuthConnection, error)
}

// VerificationTokenRepository manages secure, time-limited tokens used for out-of-band flows.
type VerificationTokenRepository interface {
	// Create stores the hash of a newly generated verification token.
	Create(ctx context.Context, token *domain.VerificationToken) error

	// GetByHash retrieves a token by its hash, strictly enforcing the requested purpose.
	GetByHash(ctx context.Context, tokenHash string, purpose domain.TokenPurpose) (*domain.VerificationToken, error)

	// Delete removes a token from the database, typically after it has been successfully used.
	Delete(ctx context.Context, tokenHash string) error

	// DeleteAllForUser invalidates all existing tokens of a specific purpose for a given user,
	// ensuring only the most recently requested token is active.
	DeleteAllForUser(ctx context.Context, userID uuid.UUID, purpose domain.TokenPurpose) error
}

// EmailSender defines the contract for dispatching transactional emails to users.
type EmailSender interface {
	// SendVerificationEmail dispatches an email containing the account activation link.
	SendVerificationEmail(ctx context.Context, toEmail string, rawToken string) error

	// SendPasswordResetEmail dispatches an email containing the password reset link.
	SendPasswordResetEmail(ctx context.Context, toEmail string, rawToken string) error
}

// OAuthClient defines the contract for communicating with the OAuth API (Yandex, VK, etc.).
type OAuthClient interface {
	// ExchangeCode swaps the temporary authorization code for an access token
	// and fetches the user's primary profile data and email.
	ExchangeCode(ctx context.Context, code string) (*OAuthProfile, error)
}

// Transactor provides transactional execution scope for database operations.
type Transactor interface {
	// WithinTx executes the given function inside a database transaction.
	// The transaction is committed on success or rolled back on error or panic.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
