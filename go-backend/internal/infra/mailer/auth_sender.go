// Package mailer implements domain-specific email sending logic for the application.
package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"

	"go-backend/pkg/mailer"
)

//go:embed templates/*.html
var templateFS embed.FS

// AuthSender handles authentication-related emails, providing links to both
// the web-based API handlers and the frontend client.
type AuthSender struct {
	mailer    *mailer.Mailer
	clientURL string
	apiURL    string
	templates *template.Template
}

type templateData struct {
	ActionURL string
}

// NewAuthSender initializes the mailer and parses embedded templates.
// It requires both clientURL (frontend) and apiURL (backend) to generate correct links.
func NewAuthSender(m *mailer.Mailer, clientURL string, apiURL string) (*AuthSender, error) {
	// Rooting the filesystem at "templates" folder to avoid name prefix issues.
	strippedFS, err := fs.Sub(templateFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to create template sub-filesystem: %w", err)
	}

	tmpl, err := template.ParseFS(strippedFS, "*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return &AuthSender{
		mailer:    m,
		clientURL: clientURL,
		apiURL:    apiURL,
		templates: tmpl,
	}, nil
}

// SendVerificationEmail dispatches an activation link.
// IMPORTANT: This link points to the API URL (port 8080) to trigger the browser landing page.
func (s *AuthSender) SendVerificationEmail(ctx context.Context, toEmail string, rawToken string) error {
	subject := "Confirm your RoleTalk account"

	// Points to the GET handler: /api/v1/auth/verify-email
	link := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.apiURL, rawToken)

	data := templateData{ActionURL: link}

	return s.executeAndSend(ctx, "verify_email.html", toEmail, subject, data)
}

// SendPasswordResetEmail dispatches a recovery link.
// Usually points to the Frontend Client (port 5173) where the "New Password" form lives.
func (s *AuthSender) SendPasswordResetEmail(ctx context.Context, toEmail string, rawToken string) error {
	subject := "Reset your RoleTalk password"

	// Points to the Frontend: /reset-password
	link := fmt.Sprintf("%s/reset-password?token=%s", s.clientURL, rawToken)

	data := templateData{ActionURL: link}

	return s.executeAndSend(ctx, "reset_password.html", toEmail, subject, data)
}

// executeAndSend renders the template and passes it to the generic mailer.
func (s *AuthSender) executeAndSend(ctx context.Context, templateName, toEmail, subject string, data templateData) error {
	var body bytes.Buffer

	if err := s.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	if err := s.mailer.SendHTML(ctx, toEmail, subject, body.String()); err != nil {
		return fmt.Errorf("failed to dispatch email [%s]: %w", templateName, err)
	}

	return nil
}
