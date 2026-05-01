// Package domain contains the core business models for the InCharacter platform.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// MessageRole defines the set of allowed senders in a practice session.
type MessageRole string

const (
	// RoleUser represents the human learner.
	RoleUser MessageRole = "user"
	// RoleAssistant represents the AI persona (partner or judge).
	RoleAssistant MessageRole = "assistant"
	// RoleSystem represents hidden instructions or context updates.
	RoleSystem MessageRole = "system"
)

// Message represents a single turn in a roleplay dialog.
// It uses pointers for nullable fields to distinguish between empty strings and missing data.
type Message struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	SessionID   uuid.UUID   `json:"session_id" db:"session_id"`
	SenderRole  MessageRole `json:"sender_role" db:"sender_role"`
	TextContent *string     `json:"text_content,omitempty" db:"text_content"`
	AudioURL    *string     `json:"audio_url,omitempty" db:"audio_url"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

// IsValid checks if the role is one of the predefined constants.
func (r MessageRole) IsValid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	default:
		return false
	}
}

// IsValid checks if the message contains at least some form of content.
func (m *Message) IsValid() bool {
	return m.TextContent != nil || m.AudioURL != nil
}
