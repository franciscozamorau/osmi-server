package request

import "time"

type CreateNotificationRequest struct {
	TemplateCode      string                 `json:"template_code" validate:"required"`
	RecipientEmail    *string                `json:"recipient_email,omitempty" validate:"omitempty,email"`
	RecipientPhone    *string                `json:"recipient_phone,omitempty" validate:"omitempty,phone"`
	RecipientUserID   *string                `json:"recipient_user_id,omitempty" validate:"omitempty,uuid4"`
	RecipientName     *string                `json:"recipient_name,omitempty" validate:"omitempty,max=255"`
	RecipientLanguage string                 `json:"recipient_language" validate:"required,oneof=es en pt"`
	Channel           string                 `json:"channel" validate:"required,oneof=email sms push"`
	ContextData       map[string]interface{} `json:"context_data,omitempty"`
	ScheduledFor      *time.Time             `json:"scheduled_for,omitempty"`
	Priority          int                    `json:"priority,omitempty" validate:"omitempty,min=1,max=10"`
}

type UpdateNotificationRequest struct {
	Status            *string    `json:"status,omitempty" validate:"omitempty,oneof=pending processing sent delivered failed cancelled"`
	NextRetryAt       *time.Time `json:"next_retry_at,omitempty"`
	LastError         *string    `json:"last_error,omitempty" validate:"omitempty,max=1000"`
	ErrorCode         *string    `json:"error_code,omitempty" validate:"omitempty,max=50"`
	ProviderMessageID *string    `json:"provider_message_id,omitempty" validate:"omitempty,max=255"`
	ProviderResponse  *string    `json:"provider_response,omitempty"`
	SentAt            *time.Time `json:"sent_at,omitempty"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
}

type NotificationFilter struct {
	RecipientEmail  *string    `json:"recipient_email,omitempty" validate:"omitempty,email"`
	RecipientUserID *string    `json:"recipient_user_id,omitempty" validate:"omitempty,uuid4"`
	Channel         *string    `json:"channel,omitempty" validate:"omitempty,oneof=email sms push"`
	Status          *string    `json:"status,omitempty" validate:"omitempty,oneof=pending processing sent delivered failed cancelled"`
	TemplateCode    *string    `json:"template_code,omitempty"`
	DateFrom        *time.Time `json:"date_from,omitempty"`
	DateTo          *time.Time `json:"date_to,omitempty"`
	HasError        *bool      `json:"has_error,omitempty"`
	RetryPending    *bool      `json:"retry_pending,omitempty"`
}

type RetryNotificationRequest struct {
	NotificationID string `json:"notification_id" validate:"required,uuid4"`
	ForceRetry     bool   `json:"force_retry"`
	MaxRetries     *int   `json:"max_retries,omitempty" validate:"omitempty,min=1,max=10"`
	RetryDelay     *int   `json:"retry_delay,omitempty" validate:"omitempty,min=30,max=86400"`
}

type BulkNotificationRequest struct {
	TemplateCode string                  `json:"template_code" validate:"required"`
	Recipients   []NotificationRecipient `json:"recipients" validate:"required,min=1,max=1000"`
	Channel      string                  `json:"channel" validate:"required,oneof=email sms push"`
	ContextData  map[string]interface{}  `json:"context_data,omitempty"`
	ScheduledFor *time.Time              `json:"scheduled_for,omitempty"`
	Priority     int                     `json:"priority" validate:"min=1,max=10"`
}

type NotificationRecipient struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,phone"`
	UserID   *string `json:"user_id,omitempty" validate:"omitempty,uuid4"`
	Name     string  `json:"name" validate:"required,max=255"`
	Language string  `json:"language" validate:"required,oneof=es en pt"`
}
