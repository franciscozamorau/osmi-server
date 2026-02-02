package dto

import "time"

type UserResponse struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone,omitempty"`
	Username          string    `json:"username,omitempty"`
	FirstName         string    `json:"first_name,omitempty"`
	LastName          string    `json:"last_name,omitempty"`
	FullName          string    `json:"full_name,omitempty"`
	AvatarURL         string    `json:"avatar_url,omitempty"`
	DateOfBirth       string    `json:"date_of_birth,omitempty"`
	EmailVerified     bool      `json:"email_verified"`
	PhoneVerified     bool      `json:"phone_verified"`
	PreferredLanguage string    `json:"preferred_language"`
	PreferredCurrency string    `json:"preferred_currency"`
	Timezone          string    `json:"timezone"`
	MFAEnabled        bool      `json:"mfa_enabled"`
	LastLoginAt       string    `json:"last_login_at,omitempty"`
	IsActive          bool      `json:"is_active"`
	IsStaff           bool      `json:"is_staff"`
	IsSuperuser       bool      `json:"is_superuser"`
	LastActiveAt      string    `json:"last_active_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UserStatsResponse struct {
	TotalUsers            int64 `json:"total_users"`
	ActiveUsers           int64 `json:"active_users"`
	EmailVerifiedUsers    int64 `json:"email_verified_users"`
	PhoneVerifiedUsers    int64 `json:"phone_verified_users"`
	MFAEnabledUsers       int64 `json:"mfa_enabled_users"`
	NewUsersLast7Days     int64 `json:"new_users_last_7_days"`
	ActiveUsersLast30Days int64 `json:"active_users_last_30_days"`
}

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
