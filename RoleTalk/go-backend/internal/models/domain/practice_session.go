package domain

import (
	"time"

	"github.com/google/uuid"
)

// PracticeSession represents a specific roleplay practice instance.
type PracticeSession struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	TopicID uuid.UUID

	// Status can be: "active", "completed", "abandoned"
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPracticeSession starts a new practice instance for a user and a topic.
func NewPracticeSession(userID, topicID uuid.UUID) *PracticeSession {
	return &PracticeSession{
		ID:        uuid.New(),
		UserID:    userID,
		TopicID:   topicID,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Complete marks the session as successfully finished.
func (s *PracticeSession) Complete() {
	s.Status = "completed"
	s.UpdatedAt = time.Now()
}
