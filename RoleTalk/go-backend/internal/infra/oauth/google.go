// Package oauth provides adapters for third-party OAuth providers,
// mapping their specific responses to our internal application domain.
package oauth

import (
	"context"
	"fmt"

	"go-backend/internal/service/auth"
	"go-backend/pkg/google"
)

// GoogleAdapter implements the auth.OAuthClient interface for Google.
// It translates the Google-specific profile into the application's domain profile.
type GoogleAdapter struct {
	client *google.Client
}

// NewGoogleAdapter creates a new adapter wrapping the generic Google client.
func NewGoogleAdapter(client *google.Client) *GoogleAdapter {
	return &GoogleAdapter{
		client: client,
	}
}

// ExchangeCode calls the Google API and maps the result to auth.OAuthProfile.
// It uses the standard OpenID Connect fields provided by Google.
func (a *GoogleAdapter) ExchangeCode(ctx context.Context, code string) (*auth.OAuthProfile, error) {
	googleProfile, err := a.client.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google exchange failed: %w", err)
	}

	return &auth.OAuthProfile{
		ProviderID:  googleProfile.ID,
		Email:       googleProfile.Email,
		DisplayName: googleProfile.Name,
		PhotoURL:    &googleProfile.PhotoURL, // TODO нужно протом подправить, убрать указатель или оставить
	}, nil
}
