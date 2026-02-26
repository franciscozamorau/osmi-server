// internal/api/dto/response/notification_response.go
package response

import "time"

// NotificationResponse representa la respuesta de una notificación
type NotificationResponse struct {
	ID                string                 `json:"id"`
	TemplateID        *string                `json:"template_id,omitempty"`
	TemplateName      *string                `json:"template_name,omitempty"`
	RecipientEmail    *string                `json:"recipient_email,omitempty"`
	RecipientPhone    *string                `json:"recipient_phone,omitempty"`
	RecipientName     *string                `json:"recipient_name,omitempty"`
	RecipientUserID   *string                `json:"recipient_user_id,omitempty"`
	RecipientLanguage string                 `json:"recipient_language"`
	Subject           string                 `json:"subject"`
	Body              string                 `json:"body"`
	Channel           string                 `json:"channel"`
	Status            string                 `json:"status"`
	Attempts          int                    `json:"attempts"`
	MaxAttempts       int                    `json:"max_attempts"`
	NextRetryAt       *time.Time             `json:"next_retry_at,omitempty"`
	LastError         *string                `json:"last_error,omitempty"`
	ErrorCode         *string                `json:"error_code,omitempty"`
	ErrorHistory      []NotificationError    `json:"error_history,omitempty"`
	ProviderMessageID *string                `json:"provider_message_id,omitempty"`
	ProviderResponse  map[string]interface{} `json:"provider_response,omitempty"`
	ContextData       map[string]interface{} `json:"context_data,omitempty"`
	ScheduledFor      *time.Time             `json:"scheduled_for,omitempty"`
	SentAt            *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt       *time.Time             `json:"delivered_at,omitempty"`
	OpenCount         int                    `json:"open_count"`
	ClickCount        int                    `json:"click_count"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// NotificationError representa un error en el envío de notificaciones
type NotificationError struct {
	Attempt     int       `json:"attempt"`
	Error       string    `json:"error"`
	Code        string    `json:"code"`
	Timestamp   time.Time `json:"timestamp"`
	ProviderRaw *string   `json:"provider_raw,omitempty"`
}

// NotificationTemplateResponse representa una plantilla de notificación
type NotificationTemplateResponse struct {
	ID                  string            `json:"id"`
	Code                string            `json:"code"`
	Name                string            `json:"name"`
	SubjectTranslations map[string]string `json:"subject_translations"`
	BodyTranslations    map[string]string `json:"body_translations"`
	AvailableVariables  []string          `json:"available_variables"`
	Channel             string            `json:"channel"`
	IsActive            bool              `json:"is_active"`
	Priority            int               `json:"priority"`
	Category            string            `json:"category"`
	Tags                []string          `json:"tags,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// NotificationListResponse representa una lista paginada de notificaciones
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	TotalPages    int                    `json:"total_pages"`
}

// NotificationStatsResponse representa estadísticas de notificaciones
type NotificationStatsResponse struct {
	TotalNotifications  int64   `json:"total_notifications"`
	SentNotifications   int64   `json:"sent_notifications"`
	FailedNotifications int64   `json:"failed_notifications"`
	DeliveryRate        float64 `json:"delivery_rate"`
	OpenRate            float64 `json:"open_rate"`
	ClickRate           float64 `json:"click_rate"`
	AvgDeliveryTime     float64 `json:"avg_delivery_time_ms"`
}

// NotificationChannelStats representa estadísticas por canal
type NotificationChannelStats struct {
	Channel      string  `json:"channel"`
	Count        int64   `json:"count"`
	DeliveryRate float64 `json:"delivery_rate"`
	OpenRate     float64 `json:"open_rate"`
}
