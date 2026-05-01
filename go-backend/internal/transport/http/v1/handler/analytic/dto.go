package analytic

import (
	"go-backend/internal/models/domain"
	"time"
)

// skillResponse represents the user's performance metrics for the UI radar chart.
// Values are percentages ranging from 0 to 100.
type skillResponse struct {
	Empathy          int       `json:"empathy"`
	Persuasion       int       `json:"persuasion"`
	Structure        int       `json:"structure"`
	StressResistance int       `json:"stress_resistance"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// toSkillResponse maps the domain UserSkill entity to a transport DTO.
func toSkillResponse(s *domain.UserSkill) skillResponse {
	return skillResponse{
		Empathy:          s.Empathy,
		Persuasion:       s.Persuasion,
		Structure:        s.Structure,
		StressResistance: s.StressResistance,
		UpdatedAt:        s.UpdatedAt,
	}
}
