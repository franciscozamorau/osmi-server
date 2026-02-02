package entities

import "time"

type Session struct {
	ID               int64      `json:"id" db:"id"`
	SessionID        string     `json:"session_id" db:"session_uuid"`
	UserID           int64      `json:"user_id" db:"user_id"`
	RefreshTokenHash string     `json:"-" db:"refresh_token_hash"`
	UserAgent        *string    `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress        *string    `json:"ip_address,omitempty" db:"ip_address"`
	DeviceInfo       *string    `json:"device_info,omitempty" db:"device_info"`
	IsValid          bool       `json:"is_valid" db:"is_valid"`
	InvalidatedAt    *time.Time `json:"invalidated_at,omitempty" db:"invalidated_at"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}
