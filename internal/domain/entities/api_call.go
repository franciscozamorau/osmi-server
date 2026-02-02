package entities

import (
	"encoding/json"
	"time"
)

type ApiCall struct {
	ID              int64                  `json:"id" db:"id"`
	Provider        string                 `json:"provider" db:"provider"`
	Endpoint        string                 `json:"endpoint" db:"endpoint"`
	Method          string                 `json:"method" db:"method"`
	RequestBody     *string                `json:"request_body,omitempty" db:"request_body"`
	RequestHeaders  map[string]interface{} `json:"request_headers,omitempty" db:"request_headers"`
	ResponseBody    *string                `json:"response_body,omitempty" db:"response_body"`
	ResponseHeaders map[string]interface{} `json:"response_headers,omitempty" db:"response_headers"`
	ResponseStatus  *int                   `json:"response_status,omitempty" db:"response_status"`
	ResponseTimeMs  *int                   `json:"response_time_ms,omitempty" db:"response_time_ms"`
	RetryCount      int                    `json:"retry_count" db:"retry_count"`
	Success         bool                   `json:"success" db:"success"`
	ErrorMessage    *string                `json:"error_message,omitempty" db:"error_message"`
	UserID          *int64                 `json:"user_id,omitempty" db:"user_id"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// Validate valida los campos requeridos
func (a *ApiCall) Validate() error {
	if a.Provider == "" {
		return NewValidationError("provider is required")
	}
	if a.Endpoint == "" {
		return NewValidationError("endpoint is required")
	}
	if !isValidMethod(a.Method) {
		return NewValidationError("invalid HTTP method")
	}
	return nil
}

// SetRequestHeaders establece los headers de la petición
func (a *ApiCall) SetRequestHeaders(headers map[string]string) error {
	if headers == nil {
		a.RequestHeaders = make(map[string]interface{})
		return nil
	}

	// Convertir map[string]string a map[string]interface{}
	interfaceHeaders := make(map[string]interface{})
	for k, v := range headers {
		interfaceHeaders[k] = v
	}
	a.RequestHeaders = interfaceHeaders
	return nil
}

// SetResponseHeaders establece los headers de la respuesta
func (a *ApiCall) SetResponseHeaders(headers map[string]string) error {
	if headers == nil {
		a.ResponseHeaders = make(map[string]interface{})
		return nil
	}

	interfaceHeaders := make(map[string]interface{})
	for k, v := range headers {
		interfaceHeaders[k] = v
	}
	a.ResponseHeaders = interfaceHeaders
	return nil
}

// MarkSuccess marca la llamada como exitosa
func (a *ApiCall) MarkSuccess(status int, responseTimeMs int, body *string, headers map[string]string) {
	a.Success = true
	a.ResponseStatus = &status
	a.ResponseTimeMs = &responseTimeMs
	a.ResponseBody = body
	a.SetResponseHeaders(headers)
}

// MarkFailure marca la llamada como fallida
func (a *ApiCall) MarkFailure(errorMessage string, retryCount int) {
	a.Success = false
	a.ErrorMessage = &errorMessage
	a.RetryCount = retryCount
}

// ShouldRetry determina si se debe reintentar
func (a *ApiCall) ShouldRetry(maxRetries int) bool {
	if a.Success {
		return false
	}

	// Reintentar solo para errores 5xx o timeout
	if a.ResponseStatus != nil && *a.ResponseStatus >= 500 {
		return a.RetryCount < maxRetries
	}

	return false
}

// GetResponseBodyAsJSON obtiene el cuerpo de respuesta como JSON
func (a *ApiCall) GetResponseBodyAsJSON() (map[string]interface{}, error) {
	if a.ResponseBody == nil || *a.ResponseBody == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(*a.ResponseBody), &result)
	return result, err
}

// GetRequestBodyAsJSON obtiene el cuerpo de petición como JSON
func (a *ApiCall) GetRequestBodyAsJSON() (map[string]interface{}, error) {
	if a.RequestBody == nil || *a.RequestBody == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(*a.RequestBody), &result)
	return result, err
}

// GetLatencyCategory categoriza la latencia
func (a *ApiCall) GetLatencyCategory() string {
	if a.ResponseTimeMs == nil {
		return "unknown"
	}

	ms := *a.ResponseTimeMs
	switch {
	case ms < 100:
		return "fast"
	case ms < 500:
		return "normal"
	case ms < 2000:
		return "slow"
	default:
		return "very_slow"
	}
}

// GetStatusCategory categoriza el estado HTTP
func (a *ApiCall) GetStatusCategory() string {
	if a.ResponseStatus == nil {
		return "unknown"
	}

	status := *a.ResponseStatus
	switch {
	case status >= 200 && status < 300:
		return "success"
	case status >= 300 && status < 400:
		return "redirect"
	case status >= 400 && status < 500:
		return "client_error"
	case status >= 500:
		return "server_error"
	default:
		return "unknown"
	}
}

// Helper functions
func isValidMethod(method string) bool {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	return validMethods[method]
}

type ValidationError struct {
	Message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

func (e *ValidationError) Error() string {
	return e.Message
}
