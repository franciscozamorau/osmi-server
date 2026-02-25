package response

import "time"

type UserResponse struct {
	ID                string     `json:"id"`
	PublicID          string     `json:"public_id"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	Phone             *string    `json:"phone,omitempty"`
	FirstName         *string    `json:"first_name,omitempty"`
	LastName          *string    `json:"last_name,omitempty"`
	FullName          *string    `json:"full_name,omitempty"`
	AvatarURL         *string    `json:"avatar_url,omitempty"`
	DateOfBirth       *string    `json:"date_of_birth,omitempty"`
	PreferredLanguage string     `json:"preferred_language"`
	PreferredCurrency string     `json:"preferred_currency"`
	Timezone          string     `json:"timezone"`
	Role              string     `json:"role"`
	IsActive          bool       `json:"is_active"`
	EmailVerified     bool       `json:"email_verified"`
	PhoneVerified     bool       `json:"phone_verified"`
	MFAEnabled        bool       `json:"mfa_enabled"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	LastActiveAt      time.Time  `json:"last_active_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
}

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
