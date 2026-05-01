package user

import (
	"github.com/google/uuid"

	"go-backend/internal/models/domain"
)

// userResponse represents the JSON payload for user profile details.
// It includes language preferences and UI-state booleans.
type userResponse struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	HasPassword     bool      `json:"has_password"`
	DisplayName     string    `json:"display_name"`
	Username        *string   `json:"username,omitempty"`
	PhotoURL        *string   `json:"photo_url,omitempty"`
	InterfaceLang   string    `json:"interface_lang"`
	PracticeLang    string    `json:"practice_lang"`
	Role            string    `json:"role"`
}

// updateProfileRequest represents the JSON payload for updating a user's profile.
// It allows partial updates for profile information and language settings.
type updateProfileRequest struct {
	DisplayName   *string `json:"display_name" validate:"omitempty,min=2,max=64"`
	Username      *string `json:"username" validate:"omitempty,min=3,max=32"`
	PhotoURL      *string `json:"photo_url" validate:"omitempty,url"`
	InterfaceLang *string `json:"interface_lang" validate:"omitempty,oneof=en ru"`
	PracticeLang  *string `json:"practice_lang" validate:"omitempty,oneof=en ru es fr de"`
}

// toUserResponse converts a domain.User entity to a userResponse DTO.
// It maps domain state to a structure suitable for the frontend.
func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:              u.ID,
		Email:           u.Email,
		IsEmailVerified: u.IsEmailVerified,
		HasPassword:     u.HasPassword(),
		DisplayName:     u.DisplayName,
		Username:        u.Username,
		PhotoURL:        u.PhotoURL, // Domain now uses PhotoURL consistently
		InterfaceLang:   u.InterfaceLang,
		PracticeLang:    u.PracticeLang,
		Role:            string(u.Role),
	}
}
