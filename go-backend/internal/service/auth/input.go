package auth

// RegisterInput holds the user-provided data for standard email registration.
// It maps directly to the initial profile setup requirements.
type RegisterInput struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName string  `json:"display_name"`
	Username    *string `json:"username,omitempty"`
	PhotoURL    *string `json:"photo_url,omitempty"`
}

// LoginInput holds the credentials for standard email login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// OAuthProfile represents unified user data received from any OAuth provider.
// This serves as a standardized DTO to bridge various providers and the internal service.
type OAuthProfile struct {
	ProviderID      string  `json:"provider_id"`
	Email           string  `json:"email"`
	DisplayName     string  `json:"display_name"`
	PhotoURL        *string `json:"photo_url"`
	IsEmailVerified bool    `json:"is_email_verified"`
}
