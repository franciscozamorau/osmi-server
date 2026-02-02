package entities

import "time"

type Webhook struct {
	ID        int64  `json:"id" db:"id"`
	WebhookID string `json:"webhook_id" db:"public_uuid"`

	Provider  string `json:"provider" db:"provider"`
	EventType string `json:"event_type" db:"event_type"`
	TargetURL string `json:"target_url" db:"target_url"`

	SecretToken     *string `json:"-" db:"secret_token"`
	SignatureHeader *string `json:"signature_header,omitempty" db:"signature_header"`

	IsActive        bool       `json:"is_active" db:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty" db:"last_triggered_at"`

	Config *string `json:"config,omitempty" db:"config"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
