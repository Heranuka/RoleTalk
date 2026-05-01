package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserSkill represents the soft skills profile of a user, typically displayed as a radar chart.
// All values are clamped between 0 and 100 representing percentage-based proficiency.
type UserSkill struct {
	UserID uuid.UUID `json:"user_id"`

	Empathy          int `json:"empathy"`
	Persuasion       int `json:"persuasion"`
	Structure        int `json:"structure"`
	StressResistance int `json:"stress_resistance"`

	UpdatedAt time.Time `json:"updated_at"`
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

// ApplyProgress updates skill values based on AI-calculated increments.
// It ensures that the resulting values are strictly within the [0, 100] range.
func (s *UserSkill) ApplyProgress(emp, pers, struc, stress int) {
	s.Empathy = s.clamp(s.Empathy + emp)
	s.Persuasion = s.clamp(s.Persuasion + pers)
	s.Structure = s.clamp(s.Structure + struc)
	s.StressResistance = s.clamp(s.StressResistance + stress)
	s.UpdatedAt = time.Now()
}

func (s *UserSkill) clamp(val int) int {
	if val > 100 {
		return 100
	}
	if val < 0 {
		return 0
	}
	return val
}
