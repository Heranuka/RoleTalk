package domain

type RoleplayParams struct {
	PartnerRole  string
	Description  string
	SecretMotive string
	UserRole     string
	Goal         string
	Language     string
}

type EvaluationParams struct {
	Goal       string
	Transcript string
}
