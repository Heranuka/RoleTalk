package domain

import (
	"time"

	"github.com/google/uuid"
)

type Topic struct {
	ID              uuid.UUID
	AuthorID        *uuid.UUID
	Title           string
	Description     *string
	Emoji           *string
	DifficultyLevel *string
	IsOfficial      bool
	LikesCount      int
	CreatedAt       time.Time
}
