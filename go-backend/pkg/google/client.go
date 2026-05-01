// Package google provides a context-aware HTTP client for interacting with Google OAuth2 APIs.
// It handles authorization code exchange and user profile retrieval with built-in observability.
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"go-backend/internal/logger"
)

var tracer = otel.Tracer("internal/infra/oauth/google")

const (
	tokenURL = "https://oauth2.googleapis.com/token" //nolint:gosec
	infoURL  = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// UserProfile represents the user data returned by Google's UserInfo endpoint.
type UserProfile struct {
	// ID is the unique identifier for the user (Google's 'sub' claim).
	ID string `json:"sub"`
	// Email is the user's primary email address.
	Email string `json:"email"`
	// Name is the user's full display name.
	Name string `json:"name"`
	// PhotoURL is the URL to the user's profile picture.
	PhotoURL string `json:"picture"`
	// EmailVerified indicates whether Google has verified the email address.
	EmailVerified bool `json:"email_verified"`
}

// Config holds the credentials required for the Google OAuth2 flow.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Client is a Google API client that handles HTTP communication with observability support.
type Client struct {
	cfg        Config
	httpClient *http.Client
	log        *zap.SugaredLogger
}

// New creates a new Google API client with a logger and default 10-second timeout.
func New(cfg Config, log *zap.SugaredLogger) *Client {
	return &Client{
		cfg: cfg,
		log: log,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ExchangeCode executes the OAuth2 cycle: exchanges the authorization code for an access token
// and then retrieves the user's profile information.
func (c *Client) ExchangeCode(ctx context.Context, code string) (*UserProfile, error) {
	ctx, span := tracer.Start(ctx, "Google.ExchangeCode")
	defer span.End()

	log := logger.FromContext(ctx, c.log)

	token, err := c.getAccessToken(ctx, code)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("exchange code failure: %w", err)
	}

	profile, err := c.getUserProfile(ctx, token)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("fetch profile failure: %w", err)
	}

	log.Infow("google profile retrieved", "email", profile.Email, "google_id", profile.ID)
	return profile, nil
}

// getAccessToken exchanges the authorization code for a Google access token.
func (c *Client) getAccessToken(ctx context.Context, code string) (string, error) {
	ctx, span := tracer.Start(ctx, "Google.getAccessToken")
	defer span.End()

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", c.cfg.ClientID)
	data.Set("client_secret", c.cfg.ClientSecret)
	data.Set("redirect_uri", c.cfg.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := c.doJSONRequest(req, &result); err != nil {
		return "", err
	}

	if result.Error != "" {
		err := fmt.Errorf("google oauth error: %s - %s", result.Error, result.ErrorDesc)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	return result.AccessToken, nil
}

// getUserProfile retrieves user information using the provided Bearer token.
func (c *Client) getUserProfile(ctx context.Context, token string) (*UserProfile, error) {
	ctx, span := tracer.Start(ctx, "Google.getUserProfile")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	var profile UserProfile
	if err := c.doJSONRequest(req, &profile); err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.String("google.user_id", profile.ID))
	return &profile, nil
}

// doJSONRequest executes an HTTP request and decodes the JSON response.
func (c *Client) doJSONRequest(req *http.Request, target any) error {
	log := logger.FromContext(req.Context(), c.log)

	resp, err := c.httpClient.Do(req) //nolint:gosec
	if err != nil {
		return fmt.Errorf("http execute: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorw("google api error response",
			"status", resp.StatusCode,
			"body", string(body),
		)
		return fmt.Errorf("google api returned status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}

	return nil
}
