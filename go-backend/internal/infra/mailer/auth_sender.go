// Package mailer implements domain-specific email sending logic for the application.
//
// It acts as an adapter between the generic pkg/mailer and the specific
// interfaces required by domain services (like auth.EmailSender).
// It utilizes html/template and go:embed to safely render and inject dynamic data
// into pre-compiled HTML email templates.
package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"

	"go-backend/pkg/mailer"
)

//go:embed templates/*.html
var templateFS embed.FS

// AuthSender implements the auth.EmailSender interface.
type AuthSender struct {
	mailer    *mailer.Mailer
	clientURL string
	templates *template.Template
}

// templateData represents the dynamic payload injected into HTML email templates.
type templateData struct {
	ActionURL string
}

// NewAuthSender creates a new AuthSender instance.
// It parses the embedded HTML templates once upon initialization.
// Returns an error if the templates fail to parse, ensuring fail-fast behavior on startup.
func NewAuthSender(m *mailer.Mailer, clientURL string) (*AuthSender, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return &AuthSender{
		mailer:    m,
		clientURL: clientURL,
		templates: tmpl,
	}, nil
}

// SendVerificationEmail dispatches an email containing the account activation link.
func (s *AuthSender) SendVerificationEmail(ctx context.Context, toEmail string, rawToken string) error {
	subject := "Подтверждение аккаунта Backforge"
	link := fmt.Sprintf("%s/verify-email?token=%s", s.clientURL, rawToken)

	data := templateData{
		ActionURL: link,
	}

	return s.executeAndSend(ctx, "verify_email.html", toEmail, subject, data)
}

// SendPasswordResetEmail dispatches an email containing the password reset link.
func (s *AuthSender) SendPasswordResetEmail(ctx context.Context, toEmail string, rawToken string) error {
	subject := "Сброс пароля Backforge"
	link := fmt.Sprintf("%s/reset-password?token=%s", s.clientURL, rawToken)

	data := templateData{
		ActionURL: link,
	}

	return s.executeAndSend(ctx, "reset_password.html", toEmail, subject, data)
}

// executeAndSend is an internal helper that executes the specified HTML template
// with the provided data and triggers the underlying generic mailer.
func (s *AuthSender) executeAndSend(ctx context.Context, templateName, toEmail, subject string, data templateData) error {
	var body bytes.Buffer

	if err := s.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	if err := s.mailer.SendHTML(ctx, toEmail, subject, body.String()); err != nil {
		return fmt.Errorf("failed to send html email (%s): %w", templateName, err)
	}

	return nil
}
