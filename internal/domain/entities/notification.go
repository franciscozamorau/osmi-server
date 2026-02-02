package entities

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Notification struct {
	ID         int64  `json:"id" db:"id"`
	TemplateID *int64 `json:"template_id,omitempty" db:"template_id"`

	RecipientEmail    *string `json:"recipient_email,omitempty" db:"recipient_email"`
	RecipientPhone    *string `json:"recipient_phone,omitempty" db:"recipient_phone"`
	RecipientName     *string `json:"recipient_name,omitempty" db:"recipient_name"`
	RecipientUserID   *int64  `json:"recipient_user_id,omitempty" db:"recipient_user_id"`
	RecipientLanguage string  `json:"recipient_language" db:"recipient_language"`

	Subject string `json:"subject" db:"subject"`
	Body    string `json:"body" db:"body"`

	Channel string `json:"channel" db:"channel"`
	Status  string `json:"status" db:"status"`

	Attempts      int32      `json:"attempts" db:"attempts"`
	MaxAttempts   int32      `json:"max_attempts" db:"max_attempts"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	RetryDelay    int32      `json:"retry_delay" db:"retry_delay"`
	BackoffFactor float64    `json:"backoff_factor" db:"backoff_factor"`
	LastError     *string    `json:"last_error,omitempty" db:"last_error"`
	ErrorCode     *string    `json:"error_code,omitempty" db:"error_code"`
	ErrorHistory  *string    `json:"error_history,omitempty" db:"error_history"`

	ProviderMessageID *string `json:"provider_message_id,omitempty" db:"provider_message_id"`
	ProviderResponse  *string `json:"provider_response,omitempty" db:"provider_response"`

	ContextData *string `json:"context_data,omitempty" db:"context_data"`

	ScheduledFor time.Time  `json:"scheduled_for" db:"scheduled_for"`
	SentAt       *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`

	OpenCount  int32 `json:"open_count" db:"open_count"`
	ClickCount int32 `json:"click_count" db:"click_count"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CanRetry verifica si se puede reintentar el envío
func (n *Notification) CanRetry() bool {
	if n.Status != "failed" && n.Status != "pending" {
		return false
	}
	if n.Attempts >= n.MaxAttempts {
		return false
	}
	if n.NextRetryAt != nil && time.Now().Before(*n.NextRetryAt) {
		return false
	}
	if n.IsExpired() {
		return false
	}
	return true
}

// ScheduleRetry programa un reintento con backoff exponencial
func (n *Notification) ScheduleRetry(errorMsg string, errorCode string) error {
	if !n.CanRetry() {
		return errors.New("notification cannot be retried")
	}

	n.Status = "pending"
	n.Attempts++

	// Cálculo de backoff exponencial
	delay := time.Duration(n.RetryDelay) * time.Second
	for i := 0; i < int(n.Attempts)-1; i++ {
		delay = time.Duration(float64(delay) * n.BackoffFactor)
		if delay > 24*time.Hour {
			delay = 24 * time.Hour
			break
		}
	}

	nextRetry := time.Now().Add(delay)
	n.NextRetryAt = &nextRetry
	n.LastError = &errorMsg
	n.ErrorCode = &errorCode

	// Guardar en historial de errores
	n.addErrorToHistory(errorMsg, errorCode)

	return nil
}

// MarkAsSent marca como enviado exitosamente
func (n *Notification) MarkAsSent(providerID string, response *string) {
	now := time.Now()
	n.Status = "sent"
	n.SentAt = &now
	n.ProviderMessageID = &providerID
	n.ProviderResponse = response
	n.LastError = nil
	n.ErrorCode = nil
	n.NextRetryAt = nil
	n.UpdatedAt = now
}

// MarkAsDelivered marca como entregado
func (n *Notification) MarkAsDelivered() {
	now := time.Now()
	n.Status = "delivered"
	n.DeliveredAt = &now
	n.UpdatedAt = now
}

// MarkAsFailed marca como fallido
func (n *Notification) MarkAsFailed(errorMsg string, errorCode string) {
	n.Status = "failed"
	n.LastError = &errorMsg
	n.ErrorCode = &errorCode
	n.UpdatedAt = time.Now()
	n.addErrorToHistory(errorMsg, errorCode)
}

// IsExpired verifica si la notificación expiró
func (n *Notification) IsExpired() bool {
	// Notificaciones expiran después de 30 días
	expiry := n.ScheduledFor.Add(30 * 24 * time.Hour)
	return time.Now().After(expiry)
}

// IsImmediate verifica si debe enviarse inmediatamente
func (n *Notification) IsImmediate() bool {
	return time.Now().After(n.ScheduledFor) || time.Now().Equal(n.ScheduledFor)
}

// GetContext obtiene el contexto como map
func (n *Notification) GetContext() (map[string]interface{}, error) {
	if n.ContextData == nil || strings.TrimSpace(*n.ContextData) == "" {
		return make(map[string]interface{}), nil
	}
	var context map[string]interface{}
	err := json.Unmarshal([]byte(*n.ContextData), &context)
	return context, err
}

// SetContext establece el contexto desde un map
func (n *Notification) SetContext(context map[string]interface{}) error {
	if context == nil {
		n.ContextData = nil
		return nil
	}
	data, err := json.Marshal(context)
	if err != nil {
		return err
	}
	str := string(data)
	n.ContextData = &str
	return nil
}

// Validate valida los campos requeridos según el canal
func (n *Notification) Validate() error {
	if n.Channel == "" {
		return errors.New("channel is required")
	}

	switch n.Channel {
	case "email":
		if n.RecipientEmail == nil || *n.RecipientEmail == "" {
			return errors.New("recipient email is required for email channel")
		}
	case "sms":
		if n.RecipientPhone == nil || *n.RecipientPhone == "" {
			return errors.New("recipient phone is required for sms channel")
		}
	case "push":
		if n.RecipientUserID == nil {
			return errors.New("recipient user ID is required for push channel")
		}
	default:
		return errors.New("invalid channel")
	}

	if n.Subject == "" && n.Channel == "email" {
		return errors.New("subject is required for email channel")
	}

	if n.Body == "" {
		return errors.New("body is required")
	}

	if n.RecipientLanguage == "" {
		n.RecipientLanguage = "es"
	}

	if n.MaxAttempts == 0 {
		n.MaxAttempts = 3
	}

	if n.RetryDelay == 0 {
		n.RetryDelay = 60 // 60 segundos
	}

	if n.BackoffFactor == 0 {
		n.BackoffFactor = 2.0
	}

	return nil
}

// addErrorToHistory añade un error al historial
func (n *Notification) addErrorToHistory(errorMsg string, errorCode string) {
	errorEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"attempt":   n.Attempts,
		"error":     errorMsg,
		"code":      errorCode,
	}

	var history []map[string]interface{}

	if n.ErrorHistory != nil {
		json.Unmarshal([]byte(*n.ErrorHistory), &history)
	}

	history = append(history, errorEntry)

	if data, err := json.Marshal(history); err == nil {
		str := string(data)
		n.ErrorHistory = &str
	}
}

// IncrementOpenCount incrementa el contador de aperturas
func (n *Notification) IncrementOpenCount() {
	n.OpenCount++
	n.UpdatedAt = time.Now()
}

// IncrementClickCount incrementa el contador de clics
func (n *Notification) IncrementClickCount() {
	n.ClickCount++
	n.UpdatedAt = time.Now()
}

// GetRetryDelaySeconds obtiene el delay actual en segundos
func (n *Notification) GetRetryDelaySeconds() int64 {
	if n.NextRetryAt == nil {
		return 0
	}
	now := time.Now()
	if n.NextRetryAt.Before(now) {
		return 0
	}
	return int64(n.NextRetryAt.Sub(now).Seconds())
}
