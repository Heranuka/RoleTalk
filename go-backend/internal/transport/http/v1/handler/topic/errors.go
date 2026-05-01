// Package topic manages the errors returned by the topic HTTP handlers.
package topic

import "errors"

var (
	// ErrTopicNotFound indicates that the requested roleplay scenario does not exist.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrAlreadyLiked is returned when a user attempts to like a topic they have already interacted with.
	ErrAlreadyLiked = errors.New("you have already liked this topic")

	// ErrUnauthorizedAction is returned when a user lacks permission to modify or delete a specific topic.
	ErrUnauthorizedAction = errors.New("you are not authorized to perform this action")

	// ErrInvalidData indicates that the request payload (title, emoji, etc.) failed validation.
	ErrInvalidData = errors.New("invalid topic data provided")

	// ErrInternalServer is a generic error for unexpected system failures.
	ErrInternalServer = errors.New("internal server error")
)
