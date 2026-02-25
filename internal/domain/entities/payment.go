package entities

import (
	"errors"
	"time"
)

// Payment representa un pago en el sistema
// Mapea exactamente la tabla billing.payments
type Payment struct {
	ID      int64 `json:"id" db:"id"`
	OrderID int64 `json:"order_id" db:"order_id"`

	// CORREGIDO: SMALLINT en la BD -> int16 en Go
	ProviderID            int16   `json:"provider_id" db:"provider_id"`
	ProviderTransactionID *string `json:"provider_transaction_id,omitempty" db:"provider_transaction_id"`
	ProviderSessionID     *string `json:"provider_session_id,omitempty" db:"provider_session_id"`

	Amount       float64 `json:"amount" db:"amount"`
	Currency     string  `json:"currency" db:"currency"`
	ExchangeRate float64 `json:"exchange_rate" db:"exchange_rate"`

	Status        string  `json:"status" db:"status"`
	PaymentMethod *string `json:"payment_method,omitempty" db:"payment_method"`
	// CORREGIDO: JSONB en la BD
	PaymentMethodDetails *map[string]interface{} `json:"payment_method_details,omitempty" db:"payment_method_details,type:jsonb"`

	Attempts    int        `json:"attempts" db:"attempts"`
	MaxAttempts int        `json:"max_attempts" db:"max_attempts"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	LastError   *string    `json:"last_error,omitempty" db:"last_error"`
	ErrorCode   *string    `json:"error_code,omitempty" db:"error_code"`

	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`

	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Payment

// IsPending verifica si el pago está pendiente
func (p *Payment) IsPending() bool {
	return p.Status == "pending"
}

// IsProcessing verifica si el pago está en proceso
func (p *Payment) IsProcessing() bool {
	return p.Status == "processing"
}

// IsCompleted verifica si el pago está completado
func (p *Payment) IsCompleted() bool {
	return p.Status == "completed" && p.ProcessedAt != nil
}

// IsFailed verifica si el pago falló
func (p *Payment) IsFailed() bool {
	return p.Status == "failed"
}

// IsRefunded verifica si el pago fue reembolsado
func (p *Payment) IsRefunded() bool {
	return p.Status == "refunded" || p.RefundedAt != nil
}

// IsCancelled verifica si el pago fue cancelado
func (p *Payment) IsCancelled() bool {
	return p.Status == "cancelled" || p.CancelledAt != nil
}

// IsDisputed verifica si el pago está en disputa
func (p *Payment) IsDisputed() bool {
	return p.Status == "disputed"
}

// IsChargeback verifica si el pago tiene chargeback
func (p *Payment) IsChargeback() bool {
	return p.Status == "chargeback"
}

// IsExpired verifica si el pago expiró
func (p *Payment) IsExpired() bool {
	return p.Status == "expired"
}

// CanRetry verifica si se puede reintentar el pago
func (p *Payment) CanRetry() bool {
	return p.IsFailed() && p.Attempts < p.MaxAttempts
}

// ShouldRetry verifica si es momento de reintentar
func (p *Payment) ShouldRetry() bool {
	if !p.CanRetry() {
		return false
	}
	if p.NextRetryAt == nil {
		return true
	}
	return time.Now().After(*p.NextRetryAt)
}

// GetNextRetryDelay calcula el delay para el próximo reintento
func (p *Payment) GetNextRetryDelay() time.Duration {
	if p.NextRetryAt == nil {
		return 0
	}
	return p.NextRetryAt.Sub(time.Now())
}

// IncrementAttempt incrementa el contador de intentos
func (p *Payment) IncrementAttempt() {
	p.Attempts++
	p.UpdatedAt = time.Now()
}

// MarkAsProcessing marca el pago como en proceso
func (p *Payment) MarkAsProcessing() {
	p.Status = "processing"
	p.UpdatedAt = time.Now()
}

// MarkAsCompleted marca el pago como completado
func (p *Payment) MarkAsCompleted() {
	now := time.Now()
	p.Status = "completed"
	p.ProcessedAt = &now
	p.UpdatedAt = now
	p.NextRetryAt = nil
	p.LastError = nil
	p.ErrorCode = nil
}

// MarkAsFailed marca el pago como fallido
func (p *Payment) MarkAsFailed(errorMsg string, errorCode string) {
	p.Status = "failed"
	p.LastError = &errorMsg
	p.ErrorCode = &errorCode
	p.UpdatedAt = time.Now()
}

// MarkAsRefunded marca el pago como reembolsado
func (p *Payment) MarkAsRefunded() {
	now := time.Now()
	p.Status = "refunded"
	p.RefundedAt = &now
	p.UpdatedAt = now
}

// MarkAsCancelled marca el pago como cancelado
func (p *Payment) MarkAsCancelled() {
	now := time.Now()
	p.Status = "cancelled"
	p.CancelledAt = &now
	p.UpdatedAt = now
}

// ScheduleRetry programa un reintento
func (p *Payment) ScheduleRetry(delay time.Duration) {
	if !p.CanRetry() {
		return
	}
	nextRetry := time.Now().Add(delay)
	p.NextRetryAt = &nextRetry
	p.UpdatedAt = time.Now()
}

// SetPaymentMethodDetails establece los detalles del método de pago
func (p *Payment) SetPaymentMethodDetails(details map[string]interface{}) {
	if details == nil {
		p.PaymentMethodDetails = nil
		return
	}
	p.PaymentMethodDetails = &details
}

// GetPaymentMethodDetail obtiene un detalle específico del método de pago
func (p *Payment) GetPaymentMethodDetail(key string) interface{} {
	if p.PaymentMethodDetails == nil {
		return nil
	}
	return (*p.PaymentMethodDetails)[key]
}

// Validate verifica que el pago sea válido
func (p *Payment) Validate() error {
	if p.OrderID == 0 {
		return errors.New("order_id is required")
	}
	if p.ProviderID == 0 {
		return errors.New("provider_id is required")
	}
	if p.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if p.Currency == "" {
		return errors.New("currency is required")
	}
	if p.Attempts < 0 {
		return errors.New("attempts cannot be negative")
	}
	if p.MaxAttempts < 1 {
		return errors.New("max_attempts must be at least 1")
	}
	if p.Attempts > p.MaxAttempts {
		return errors.New("attempts cannot exceed max_attempts")
	}
	return nil
}

// GetPaymentMethodType obtiene el tipo de método de pago
func (p *Payment) GetPaymentMethodType() string {
	if p.PaymentMethod == nil {
		return ""
	}
	return *p.PaymentMethod
}

// GetProviderInfo obtiene información del proveedor
func (p *Payment) GetProviderInfo() map[string]string {
	info := make(map[string]string)
	if p.ProviderTransactionID != nil {
		info["transaction_id"] = *p.ProviderTransactionID
	}
	if p.ProviderSessionID != nil {
		info["session_id"] = *p.ProviderSessionID
	}
	return info
}

// HasError verifica si hay un error registrado
func (p *Payment) HasError() bool {
	return p.LastError != nil || p.ErrorCode != nil
}

// GetErrorDetails obtiene los detalles del error
func (p *Payment) GetErrorDetails() map[string]string {
	if !p.HasError() {
		return nil
	}
	details := make(map[string]string)
	if p.LastError != nil {
		details["message"] = *p.LastError
	}
	if p.ErrorCode != nil {
		details["code"] = *p.ErrorCode
	}
	return details
}

// Reset reinicia el estado del pago
func (p *Payment) Reset() {
	p.Status = "pending"
	p.Attempts = 0
	p.NextRetryAt = nil
	p.LastError = nil
	p.ErrorCode = nil
	p.ProcessedAt = nil
	p.RefundedAt = nil
	p.CancelledAt = nil
	p.UpdatedAt = time.Now()
}
