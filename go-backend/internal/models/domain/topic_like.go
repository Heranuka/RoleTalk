package domain

import "github.com/google/uuid"

// TopicLike represents a user's like on a specific roleplay topic.
type TopicLike struct {
	ID      uuid.UUID `db:"id"`
	UserID  uuid.UUID
	TopicID uuid.UUID
}
