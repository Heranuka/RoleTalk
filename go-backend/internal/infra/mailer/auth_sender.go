// Package mailer implements domain-specific email sending logic for the application.
package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs" // Добавь этот импорт

	"go-backend/pkg/mailer"
)

//go:embed templates/*.html
var templateFS embed.FS

// AuthSender handles authentication-related emails.
type AuthSender struct {
	mailer    *mailer.Mailer
	clientURL string
	templates *template.Template
}

type templateData struct {
	ActionURL string
}

// NewAuthSender initializes the mailer and parses embedded templates using fs.Sub.
func NewAuthSender(m *mailer.Mailer, clientURL string) (*AuthSender, error) {
	// Rooting the filesystem at "templates" folder to avoid name prefix issues.
	// This makes template names identical to their filenames (e.g., "reset_password.html").
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
		templates: tmpl,
	}, nil
}

// SendVerificationEmail dispatches an activation link.
func (s *AuthSender) SendVerificationEmail(ctx context.Context, toEmail string, rawToken string) error {
	subject := "Confirm your RoleTalk account"
	link := fmt.Sprintf("%s/verify-email?token=%s", s.clientURL, rawToken)

	data := templateData{ActionURL: link}

	// Now we use only the filename, without "templates/" prefix.
	return s.executeAndSend(ctx, "verify_email.html", toEmail, subject, data)
}

// SendPasswordResetEmail dispatches a recovery link.
func (s *AuthSender) SendPasswordResetEmail(ctx context.Context, toEmail string, rawToken string) error {
	subject := "Reset your RoleTalk password"
	link := fmt.Sprintf("%s/reset-password?token=%s", s.clientURL, rawToken)

	data := templateData{ActionURL: link}

	// Use only the filename.
	return s.executeAndSend(ctx, "reset_password.html", toEmail, subject, data)
}

// executeAndSend renders the template and passes it to the generic mailer.
func (s *AuthSender) executeAndSend(ctx context.Context, templateName, toEmail, subject string, data templateData) error {
	var body bytes.Buffer

	// templateName is now "reset_password.html" or "verify_email.html"
	if err := s.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	if err := s.mailer.SendHTML(ctx, toEmail, subject, body.String()); err != nil {
		return fmt.Errorf("failed to dispatch email [%s]: %w", templateName, err)
	}

	return nil
}
