package entities

import "time"

type User struct {
	ID           int64   `json:"id" db:"id"`
	PublicID     string  `json:"public_id" db:"public_uuid"`
	Email        string  `json:"email" db:"email"`
	Phone        *string `json:"phone,omitempty" db:"phone"`
	Username     *string `json:"username,omitempty" db:"username"`
	PasswordHash string  `json:"-" db:"password_hash"`

	FirstName   *string    `json:"first_name,omitempty" db:"first_name"`
	LastName    *string    `json:"last_name,omitempty" db:"last_name"`
	FullName    *string    `json:"full_name,omitempty" db:"full_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`

	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified bool       `json:"phone_verified" db:"phone_verified"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty" db:"verified_at"`

	PreferredLanguage string `json:"preferred_language" db:"preferred_language"`
	PreferredCurrency string `json:"preferred_currency" db:"preferred_currency"`
	Timezone          string `json:"timezone" db:"timezone"`

	MFAEnabled          bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret           *string    `json:"-" db:"mfa_secret"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP         *string    `json:"last_login_ip,omitempty" db:"last_login_ip"`
	FailedLoginAttempts int32      `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty" db:"locked_until"`

	IsActive    bool `json:"is_active" db:"is_active"`
	IsStaff     bool `json:"is_staff" db:"is_staff"`
	IsSuperuser bool `json:"is_superuser" db:"is_superuser"`

	LastActiveAt time.Time `json:"last_active_at" db:"last_active_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
