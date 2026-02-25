package entities

import (
	"errors"
	"time"
)

// Notification representa un mensaje de notificación
// Mapea exactamente la tabla notifications.messages
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

	Attempts      int        `json:"attempts" db:"attempts"`
	MaxAttempts   int        `json:"max_attempts" db:"max_attempts"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	RetryDelay    int        `json:"retry_delay" db:"retry_delay"`
	BackoffFactor float64    `json:"backoff_factor" db:"backoff_factor"`
	LastError     *string    `json:"last_error,omitempty" db:"last_error"`
	ErrorCode     *string    `json:"error_code,omitempty" db:"error_code"`
	// CORREGIDO: error_history es JSONB
	ErrorHistory *[]map[string]interface{} `json:"error_history,omitempty" db:"error_history,type:jsonb"`

	ProviderMessageID *string `json:"provider_message_id,omitempty" db:"provider_message_id"`
	// CORREGIDO: provider_response es JSONB
	ProviderResponse *map[string]interface{} `json:"provider_response,omitempty" db:"provider_response,type:jsonb"`

	// CORREGIDO: context_data es JSONB
	ContextData *map[string]interface{} `json:"context_data,omitempty" db:"context_data,type:jsonb"`

	ScheduledFor time.Time  `json:"scheduled_for" db:"scheduled_for"`
	SentAt       *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`

	// CORREGIDO: Usamos int en lugar de int32 para INTEGER en PostgreSQL
	OpenCount  int `json:"open_count" db:"open_count"`
	ClickCount int `json:"click_count" db:"click_count"`

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
	for i := 0; i < n.Attempts-1; i++ {
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
func (n *Notification) MarkAsSent(providerID string, response map[string]interface{}) {
	now := time.Now()
	n.Status = "sent"
	n.SentAt = &now
	n.ProviderMessageID = &providerID
	n.ProviderResponse = &response
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
func (n *Notification) GetContext() map[string]interface{} {
	if n.ContextData == nil {
		return make(map[string]interface{})
	}
	return *n.ContextData
}

// SetContext establece el contexto desde un map
func (n *Notification) SetContext(context map[string]interface{}) {
	if context == nil {
		n.ContextData = nil
		return
	}
	n.ContextData = &context
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
		n.MaxAttempts = 5 // Valor por defecto en la BD
	}

	if n.RetryDelay == 0 {
		n.RetryDelay = 300 // 300 segundos (5 minutos) valor por defecto en la BD
	}

	if n.BackoffFactor == 0 {
		n.BackoffFactor = 1.5 // Valor por defecto en la BD
	}

	return nil
}

// addErrorToHistory añade un error al historial
func (n *Notification) addErrorToHistory(errorMsg string, errorCode string) {
	errorEntry := map[string]interface{}{
		"timestamp": time.Now(),
		"attempt":   n.Attempts,
		"error":     errorMsg,
		"code":      errorCode,
	}

	var history []map[string]interface{}

	if n.ErrorHistory != nil {
		history = *n.ErrorHistory
	}

	history = append(history, errorEntry)
	n.ErrorHistory = &history
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

// GetLastErrorDetails obtiene los detalles del último error
func (n *Notification) GetLastErrorDetails() map[string]interface{} {
	if n.ErrorHistory == nil || len(*n.ErrorHistory) == 0 {
		return nil
	}
	history := *n.ErrorHistory
	return history[len(history)-1]
}

// ResetForRetry reinicia el estado para un nuevo intento
func (n *Notification) ResetForRetry() {
	n.Status = "pending"
	n.LastError = nil
	n.ErrorCode = nil
	n.NextRetryAt = nil
	n.UpdatedAt = time.Now()
}
