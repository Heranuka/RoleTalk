package domain

import (
	"time"

	"github.com/google/uuid"
)

// Topic represents a roleplay scenario schema.
// It contains metadata for the UI and specific instructions for the AI orchestration.
type Topic struct {
	ID              uuid.UUID
	AuthorID        *uuid.UUID
	Title           string
	Description     *string
	Emoji           *string
	DifficultyLevel *string
	IsOfficial      bool
	LikesCount      int

	// --- НОВЫЕ ПОЛЯ ДЛЯ РОЛЕВОЙ ИГРЫ ---

	// MyRole is the persona assigned to the user (e.g., "Frustrated Customer").
	MyRole string

	// PartnerRole is the persona assigned to the AI (e.g., "Support Agent").
	PartnerRole string

	// PartnerEmoji is the avatar for the AI in the session screen.
	PartnerEmoji string

	// Goal is the specific objective the user must achieve to finish the session.
	// Used by the Analytics service to check if the user succeeded.
	Goal string

	// PartnerSecretMotive is the hidden instruction for the AI to make the dialog dynamic.
	// (e.g., "Don't give a refund unless the user mentions the legal department").
	PartnerSecretMotive *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
