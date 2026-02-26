// internal/api/dto/response/payment_response.go
package response

import "time"

// PaymentProviderInfo representa información del proveedor de pago
type PaymentProviderInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PaymentError representa un error en el procesamiento de pago
type PaymentError struct {
	Attempt          int       `json:"attempt"`
	Error            string    `json:"error"`
	Code             string    `json:"code"`
	Timestamp        time.Time `json:"timestamp"`
	ProviderResponse *string   `json:"provider_response,omitempty"`
}

// PaymentMethodStats representa estadísticas por método de pago
type PaymentMethodStats struct {
	Method      string  `json:"method"`
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	SuccessRate float64 `json:"success_rate"`
}

// PaymentSummary representa resumen de pagos
type PaymentSummary struct {
	TotalAmount       float64              `json:"total_amount"`
	SuccessfulAmount  float64              `json:"successful_amount"`
	FailedAmount      float64              `json:"failed_amount"`
	PendingAmount     float64              `json:"pending_amount"`
	TotalCount        int                  `json:"total_count"`
	SuccessfulCount   int                  `json:"successful_count"`
	FailedCount       int                  `json:"failed_count"`
	PendingCount      int                  `json:"pending_count"`
	AvgAmount         float64              `json:"avg_amount"`
	SuccessRate       float64              `json:"success_rate"`
	AvgProcessingTime float64              `json:"avg_processing_time_ms"`
	TopPaymentMethods []PaymentMethodStats `json:"top_payment_methods"`
}

// PaymentResponse representa la respuesta completa de un pago
type PaymentResponse struct {
	ID                    string                 `json:"id"`
	OrderID               string                 `json:"order_id"`
	Order                 *OrderInfo             `json:"order,omitempty"`
	Customer              *CustomerInfo          `json:"customer,omitempty"`
	Provider              PaymentProviderInfo    `json:"provider"`
	ProviderTransactionID *string                `json:"provider_transaction_id,omitempty"`
	ProviderSessionID     *string                `json:"provider_session_id,omitempty"`
	Amount                float64                `json:"amount"`
	Currency              string                 `json:"currency"`
	ExchangeRate          float64                `json:"exchange_rate"`
	Status                string                 `json:"status"`
	PaymentMethod         string                 `json:"payment_method"`
	PaymentMethodDetails  map[string]interface{} `json:"payment_method_details,omitempty"`
	Attempts              int                    `json:"attempts"`
	MaxAttempts           int                    `json:"max_attempts"`
	NextRetryAt           *time.Time             `json:"next_retry_at,omitempty"`
	LastError             *string                `json:"last_error,omitempty"`
	ErrorCode             *string                `json:"error_code,omitempty"`
	ErrorHistory          []PaymentError         `json:"error_history,omitempty"`
	IPAddress             *string                `json:"ip_address,omitempty"`
	UserAgent             *string                `json:"user_agent,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	RiskScore             *float64               `json:"risk_score,omitempty"`
	RiskFlags             []string               `json:"risk_flags,omitempty"`
	ProcessorResponse     map[string]interface{} `json:"processor_response,omitempty"`
	ProcessedAt           *time.Time             `json:"processed_at,omitempty"`
	RefundedAt            *time.Time             `json:"refunded_at,omitempty"`
	CancelledAt           *time.Time             `json:"cancelled_at,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// PaymentInfo representa información resumida de un pago
type PaymentInfo struct {
	ID            string     `json:"id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
}

// PaymentListResponse representa una lista paginada de pagos
type PaymentListResponse struct {
	Payments   []PaymentResponse `json:"payments"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
	HasNext    bool              `json:"has_next"`
	HasPrev    bool              `json:"has_prev"`
	Summary    *PaymentSummary   `json:"summary,omitempty"`
	Filters    *PaymentFilter    `json:"filters,omitempty"`
}

// PaymentProcessingResponse representa respuesta de procesamiento de pago
type PaymentProcessingResponse struct {
	PaymentID            string                 `json:"payment_id"`
	Status               string                 `json:"status"`
	RequiresAction       bool                   `json:"requires_action"`
	ActionURL            *string                `json:"action_url,omitempty"`
	ActionType           *string                `json:"action_type,omitempty"`
	ProviderInstructions map[string]interface{} `json:"provider_instructions,omitempty"`
	NextSteps            []string               `json:"next_steps,omitempty"`
	EstimatedCompletion  *time.Time             `json:"estimated_completion,omitempty"`
}

// PaymentRetryResponse representa respuesta de reintento de pago
type PaymentRetryResponse struct {
	PaymentID          string     `json:"payment_id"`
	NewStatus          string     `json:"new_status"`
	RetryScheduled     bool       `json:"retry_scheduled"`
	NextRetryAt        *time.Time `json:"next_retry_at,omitempty"`
	AttemptsRemaining  int        `json:"attempts_remaining"`
	EstimatedRetryTime *time.Time `json:"estimated_retry_time,omitempty"`
}
