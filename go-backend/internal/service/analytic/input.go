package analytic

// sessionScores is an internal DTO used to unmarshal the AI's evaluation.
// The JSON tags must strictly match the keys provided by the LLM in its response.
type sessionScores struct {
	// Empathy represents the emotional intelligence score.
	Empathy int `json:"empathy"`
	// Persuasion represents the ability to influence the partner.
	Persuasion int `json:"persuasion"`
	// Structure represents the logical flow of the conversation.
	Structure int `json:"structure"`
	// Stress represents the user's performance under pressure.
	// Note: We map this to StressResistance in our domain model.
	Stress int `json:"stress_resistance"`
}
