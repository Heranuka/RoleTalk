// Package mailer provides a generic, reusable, and context-aware SMTP email sender.
//
// It handles the construction of MIME headers, HTML bodies, and secure SMTP
// authentication. It wraps standard net/smtp and adds context cancellation
// support to prevent hanging requests during network timeouts.
package mailer

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"
)

// ErrCRLFInjection is returned when newline characters are detected in email headers.
var ErrCRLFInjection = errors.New("crlf injection detected in email headers")

// ErrInvalidEmailAddress is returned when sender or recipient email address is invalid.
var ErrInvalidEmailAddress = errors.New("invalid email address")

// Config holds the SMTP configuration required to send emails.
type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
}

// Mailer is a generic SMTP email sender.
type Mailer struct {
	cfg Config
}

// New creates a new Mailer instance with the provided configuration.
func New(cfg Config) *Mailer {
	return &Mailer{
		cfg: cfg,
	}
}

// SendHTML sends an HTML-formatted email to the specified recipient.
//
// It respects the provided context, allowing for timeout or cancellation
// during the SMTP transmission.
func (m *Mailer) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	// 1. Sanitize inputs to prevent Email Header Injection (CRLF Injection).
	if strings.ContainsAny(to, "\r\n") || strings.ContainsAny(m.cfg.FromAddress, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return ErrCRLFInjection
	}

	// 2. Validate and canonicalize email address syntax for recipient and sender.
	parsedTo, err := mail.ParseAddress(to)
	if err != nil {
		return ErrInvalidEmailAddress
	}
	parsedFrom, err := mail.ParseAddress(m.cfg.FromAddress)
	if err != nil {
		return ErrInvalidEmailAddress
	}

	// 3. Construct standard MIME headers for HTML email using canonical mailbox addresses.
	var msgBuilder strings.Builder
	fmt.Fprintf(&msgBuilder, "From: %s\r\n", parsedFrom.Address)
	fmt.Fprintf(&msgBuilder, "To: %s\r\n", parsedTo.Address)
	fmt.Fprintf(&msgBuilder, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")

	// A blank line separates headers from the body in standard SMTP.
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(htmlBody)

	msg := []byte(msgBuilder.String())
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	// 4. Execute sending in a goroutine to support context cancellation.
	ch := make(chan error, 1)
	go func() {
		ch <- smtp.SendMail(addr, auth, parsedFrom.Address, []string{parsedTo.Address}, msg)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending canceled or timed out: %w", ctx.Err())
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("smtp send failed: %w", err)
		}
		return nil
	}
}
