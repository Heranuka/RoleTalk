// Package topic defines business errors for scenario management.
package topic

import "errors"

var (
	// ErrTopicNotFound is returned when the requested scenario does not exist.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrInvalidTopicData is returned when the provided title, emoji or description is malformed.
	ErrInvalidTopicData = errors.New("invalid topic data: title and emoji are required")

	// ErrAlreadyLiked is returned when a user attempts to like a topic they have already liked.
	ErrAlreadyLiked = errors.New("you have already liked this topic")

	// ErrLikeNotFound is returned when a user attempts to unlike a topic they haven't liked yet.
	ErrLikeNotFound = errors.New("like record not found")

	// ErrUnauthorizedAction is returned when a user attempts to modify a topic that belongs to someone else
	// or attempts to modify a system-level official topic.
	ErrUnauthorizedAction = errors.New("user is not authorized to modify this topic")
)
