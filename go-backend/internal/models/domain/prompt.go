package domain

// RoleplayParams defines the variables needed to render a partner persona prompt.
type RoleplayParams struct {
	PartnerRole  string
	Description  string
	SecretMotive string
	UserRole     string
	Goal         string
	Language     string
}

// EvaluationParams defines the variables needed to render a session analysis prompt.
type EvaluationParams struct {
	Goal       string
	Transcript string
}
