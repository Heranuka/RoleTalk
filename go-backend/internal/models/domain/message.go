package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	SenderRole  string
	TextContent *string
	AudioURL    *string
	CreatedAt   time.Time
}
