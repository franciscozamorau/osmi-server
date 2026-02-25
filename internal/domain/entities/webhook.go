package entities

import (
	"errors"
	"time"
)

// Webhook representa un webhook configurado para integraciones
// Mapea exactamente la tabla integration.webhooks
type Webhook struct {
	ID int64 `json:"id" db:"id"`
	// CORREGIDO: Mantenemos WebhookID para JSON pero db usa public_uuid
	WebhookID string `json:"webhook_id" db:"public_uuid"`

	Provider  string `json:"provider" db:"provider"`     // stripe, paypal, etc.
	EventType string `json:"event_type" db:"event_type"` // ticket.purchased, order.completed, etc.
	TargetURL string `json:"target_url" db:"target_url"`

	SecretToken     *string `json:"-" db:"secret_token"` // Nunca se expone en JSON
	SignatureHeader *string `json:"signature_header,omitempty" db:"signature_header"`

	IsActive        bool       `json:"is_active" db:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty" db:"last_triggered_at"`

	// CORREGIDO: config es JSONB
	Config *map[string]interface{} `json:"config,omitempty" db:"config,type:jsonb"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Webhook

// IsEnabled verifica si el webhook está activo
func (w *Webhook) IsEnabled() bool {
	return w.IsActive
}

// Enable activa el webhook
func (w *Webhook) Enable() {
	w.IsActive = true
	w.UpdatedAt = time.Now()
}

// Disable desactiva el webhook
func (w *Webhook) Disable() {
	w.IsActive = false
	w.UpdatedAt = time.Now()
}

// Trigger registra que el webhook fue ejecutado
func (w *Webhook) Trigger() {
	now := time.Now()
	w.LastTriggeredAt = &now
	w.UpdatedAt = now
}

// SetConfigValue establece un valor de configuración
func (w *Webhook) SetConfigValue(key string, value interface{}) {
	if w.Config == nil {
		w.Config = &map[string]interface{}{}
	}
	(*w.Config)[key] = value
	w.UpdatedAt = time.Now()
}

// GetConfigValue obtiene un valor de configuración
func (w *Webhook) GetConfigValue(key string) interface{} {
	if w.Config == nil {
		return nil
	}
	return (*w.Config)[key]
}

// DeleteConfigValue elimina un valor de configuración
func (w *Webhook) DeleteConfigValue(key string) {
	if w.Config == nil {
		return
	}
	delete(*w.Config, key)
	if len(*w.Config) == 0 {
		w.Config = nil
	}
	w.UpdatedAt = time.Now()
}

// HasConfigKey verifica si existe una clave de configuración
func (w *Webhook) HasConfigKey(key string) bool {
	if w.Config == nil {
		return false
	}
	_, exists := (*w.Config)[key]
	return exists
}

// GetConfigString obtiene un valor de configuración como string
func (w *Webhook) GetConfigString(key string, defaultValue string) string {
	val := w.GetConfigValue(key)
	if val == nil {
		return defaultValue
	}
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

// GetConfigInt obtiene un valor de configuración como int
func (w *Webhook) GetConfigInt(key string, defaultValue int) int {
	val := w.GetConfigValue(key)
	if val == nil {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return defaultValue
	}
}

// GetConfigBool obtiene un valor de configuración como bool
func (w *Webhook) GetConfigBool(key string, defaultValue bool) bool {
	val := w.GetConfigValue(key)
	if val == nil {
		return defaultValue
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// Validate verifica que el webhook sea válido
func (w *Webhook) Validate() error {
	if w.Provider == "" {
		return errors.New("provider is required")
	}
	if w.EventType == "" {
		return errors.New("event_type is required")
	}
	if w.TargetURL == "" {
		return errors.New("target_url is required")
	}

	// Validar que la URL sea válida (básico)
	if !isValidURL(w.TargetURL) {
		return errors.New("target_url must be a valid URL")
	}

	return nil
}

// GetHeaders obtiene los headers a incluir en la petición
func (w *Webhook) GetHeaders() map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	if w.SignatureHeader != nil && *w.SignatureHeader != "" {
		headers["X-Signature"] = *w.SignatureHeader
	}

	// Añadir headers personalizados de la configuración
	if w.Config != nil {
		if customHeaders, ok := (*w.Config)["headers"]; ok {
			if headersMap, ok := customHeaders.(map[string]interface{}); ok {
				for k, v := range headersMap {
					if strVal, ok := v.(string); ok {
						headers[k] = strVal
					}
				}
			}
		}
	}

	return headers
}

// GetRetryPolicy obtiene la política de reintentos de la configuración
func (w *Webhook) GetRetryPolicy() (maxRetries int, backoffFactor float64) {
	maxRetries = w.GetConfigInt("max_retries", 3)
	backoffFactor = 1.0

	if val := w.GetConfigValue("backoff_factor"); val != nil {
		if f, ok := val.(float64); ok {
			backoffFactor = f
		}
	}

	return maxRetries, backoffFactor
}

// GetTimeout obtiene el timeout de la configuración
func (w *Webhook) GetTimeout() int {
	return w.GetConfigInt("timeout_seconds", 30) // Default 30 segundos
}

// Helper function para validación básica de URL
func isValidURL(url string) bool {
	// Implementación básica - en producción usar net/url.Parse
	return len(url) > 0 && (len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://"))
}

// Clone crea una copia del Webhook
func (w *Webhook) Clone() *Webhook {
	clone := &Webhook{
		ID:              w.ID,
		WebhookID:       w.WebhookID,
		Provider:        w.Provider,
		EventType:       w.EventType,
		TargetURL:       w.TargetURL,
		IsActive:        w.IsActive,
		LastTriggeredAt: w.LastTriggeredAt,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
	}

	// Clonar SecretToken
	if w.SecretToken != nil {
		secretToken := *w.SecretToken
		clone.SecretToken = &secretToken
	}

	// Clonar SignatureHeader
	if w.SignatureHeader != nil {
		signatureHeader := *w.SignatureHeader
		clone.SignatureHeader = &signatureHeader
	}

	// Clonar Config
	if w.Config != nil {
		config := make(map[string]interface{})
		for k, v := range *w.Config {
			config[k] = v
		}
		clone.Config = &config
	}

	return clone
}
