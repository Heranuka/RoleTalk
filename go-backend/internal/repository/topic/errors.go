package topic

import "errors"

var (
	// ErrTopicNotFound is returned when a specific scenario ID does not exist in the database.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrLikeAlreadyExists is returned when a user tries to like a topic they have already liked.
	// Maps to a unique constraint violation in topic_likes table.
	ErrLikeAlreadyExists = errors.New("topic already liked by this user")

	// ErrLikeNotFound is returned when trying to unlike a topic that wasn't liked by the user.
	ErrLikeNotFound = errors.New("like record not found")

	// ErrAuthorNotFound is returned when the specified author_id does not exist in users table.
	ErrAuthorNotFound = errors.New("topic author not found")

	// ErrDuplicateTitle is returned when a topic with the same title already exists.
	ErrDuplicateTitle = errors.New("topic with this title already exists")
)
