// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserSkill represents the soft skills profile of a user, typically displayed as a radar chart.
// Each skill value is expected to be a percentage ranging from 0 to 100.
type UserSkill struct {
	UserID uuid.UUID

	Empathy          int
	Persuasion       int
	Structure        int
	StressResistance int

	UpdatedAt time.Time
}

// NewUserSkill initializes a new skill profile for a user with default values.
func NewUserSkill(userID uuid.UUID) *UserSkill {
	return &UserSkill{
		UserID:           userID,
		Empathy:          0,
		Persuasion:       0,
		Structure:        0,
		StressResistance: 0,
		UpdatedAt:        time.Now(),
	}
}

// Validate checks if all skill values are within the allowed range (0-100).
func (s *UserSkill) Validate() bool {
	return s.isValidRange(s.Empathy) &&
		s.isValidRange(s.Persuasion) &&
		s.isValidRange(s.Structure) &&
		s.isValidRange(s.StressResistance)
}

// isValidRange is a private helper to check percentage boundaries.
func (s *UserSkill) isValidRange(val int) bool {
	return val >= 0 && val <= 100
}

// ApplyProgress updates skill values based on increments.
// It ensures that the resulting values do not exceed 100 or drop below 0.
func (s *UserSkill) ApplyProgress(emp, pers, struc, stress int) {
	s.Empathy = s.clamp(s.Empathy + emp)
	s.Persuasion = s.clamp(s.Persuasion + pers)
	s.Structure = s.clamp(s.Structure + struc)
	s.StressResistance = s.clamp(s.StressResistance + stress)
	s.UpdatedAt = time.Now()
}

// clamp ensures the value stays between 0 and 100.
func (s *UserSkill) clamp(val int) int {
	if val > 100 {
		return 100
	}
	if val < 0 {
		return 0
	}
	return val
}
