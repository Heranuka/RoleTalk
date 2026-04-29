package domain

import "github.com/google/uuid"

type TopicLike struct {
	ID      uuid.UUID `db:"id"`
	UserID  uuid.UUID
	TopicID uuid.UUID
}
