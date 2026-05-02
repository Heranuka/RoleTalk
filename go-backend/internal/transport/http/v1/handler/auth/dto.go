package auth

// registerRequest represents the JSON payload for a new user registration.
// It is aligned with the User domain entity and requires a DisplayName.
type registerRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=8,max=72"`
	DisplayName string  `json:"display_name" validate:"required,min=2,max=64"`
	Username    *string `json:"username,omitempty" validate:"omitempty,min=3,max=32"`
	PhotoURL    *string `json:"photo_url,omitempty" validate:"omitempty,url"`
}

// loginRequest represents the JSON payload for authenticating an existing user.
type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// logoutRequest represents the JSON payload to terminate a session.
type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// verifyEmailRequest represents the payload for confirming a user's email address.
type verifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// requestResetRequest represents the payload to initiate a password reset process.
type requestResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// resetPasswordRequest represents the payload to set a new password using a reset token.
type resetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

// googleCallbackRequest represents the payload sent by the frontend
// after Google OAuth2 redirection with an authorization code.
type googleCallbackRequest struct {
	Code string `json:"code" validate:"required"`
}

// refreshRequest represents the JSON payload to exchange a refresh token for new credentials.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// tokenResponse contains the JWT access token and opaque refresh token.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// resendVerificationRequest represents the payload to request a new verification email link.
type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
