package entities

import (
	"errors"
	"time"
)

// Refund representa un reembolso en el sistema
// Mapea exactamente la tabla billing.refunds
type Refund struct {
	ID int64 `json:"id" db:"id"`
	// NOTA: En la BD, al menos uno de payment_id u order_id debe estar presente
	PaymentID *int64 `json:"payment_id,omitempty" db:"payment_id"`
	OrderID   *int64 `json:"order_id,omitempty" db:"order_id"`

	RefundReason *string `json:"refund_reason,omitempty" db:"refund_reason"`
	RefundAmount float64 `json:"refund_amount" db:"refund_amount"`
	Currency     string  `json:"currency" db:"currency"`

	Status           string  `json:"status" db:"status"` // pending, processing, completed, failed
	ProviderRefundID *string `json:"provider_refund_id,omitempty" db:"provider_refund_id"`

	RequestedBy *int64 `json:"requested_by,omitempty" db:"requested_by"`
	ApprovedBy  *int64 `json:"approved_by,omitempty" db:"approved_by"`

	RequestedAt time.Time  `json:"requested_at" db:"requested_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Refund

// IsPending verifica si el reembolso está pendiente
func (r *Refund) IsPending() bool {
	return r.Status == "pending"
}

// IsProcessing verifica si el reembolso está en proceso
func (r *Refund) IsProcessing() bool {
	return r.Status == "processing"
}

// IsCompleted verifica si el reembolso está completado
func (r *Refund) IsCompleted() bool {
	return r.Status == "completed" && r.CompletedAt != nil
}

// IsFailed verifica si el reembolso falló
func (r *Refund) IsFailed() bool {
	return r.Status == "failed"
}

// IsLinkedToPayment verifica si el reembolso está vinculado a un pago
func (r *Refund) IsLinkedToPayment() bool {
	return r.PaymentID != nil
}

// IsLinkedToOrder verifica si el reembolso está vinculado a una orden
func (r *Refund) IsLinkedToOrder() bool {
	return r.OrderID != nil
}

// MarkAsProcessing marca el reembolso como en proceso
func (r *Refund) MarkAsProcessing() {
	r.Status = "processing"
	r.UpdatedAt = time.Now()
}

// MarkAsCompleted marca el reembolso como completado
func (r *Refund) MarkAsCompleted(providerRefundID string) {
	now := time.Now()
	r.Status = "completed"
	r.CompletedAt = &now
	r.ProcessedAt = &now
	if providerRefundID != "" {
		r.ProviderRefundID = &providerRefundID
	}
	r.UpdatedAt = now
}

// MarkAsFailed marca el reembolso como fallido
func (r *Refund) MarkAsFailed() {
	r.Status = "failed"
	r.UpdatedAt = time.Now()
}

// SetProcessed marca el reembolso como procesado
func (r *Refund) SetProcessed() {
	now := time.Now()
	r.ProcessedAt = &now
	r.UpdatedAt = now
}

// Validate verifica que el reembolso sea válido
func (r *Refund) Validate() error {
	// Verificar que al menos uno de payment_id u order_id esté presente
	if r.PaymentID == nil && r.OrderID == nil {
		return errors.New("either payment_id or order_id must be provided")
	}

	if r.RefundAmount <= 0 {
		return errors.New("refund_amount must be greater than 0")
	}

	if r.Currency == "" {
		return errors.New("currency is required")
	}

	if r.Status == "" {
		return errors.New("status is required")
	}

	if r.RequestedBy == nil {
		return errors.New("requested_by is required")
	}

	return nil
}

// CanBeApproved verifica si el reembolso puede ser aprobado
func (r *Refund) CanBeApproved() bool {
	return r.IsPending() && r.ApprovedBy == nil
}

// CanBeProcessed verifica si el reembolso puede ser procesado
func (r *Refund) CanBeProcessed() bool {
	return r.IsPending() && r.ApprovedBy != nil
}

// Approve aprueba el reembolso
func (r *Refund) Approve(approvedBy int64) {
	r.ApprovedBy = &approvedBy
	r.Status = "pending" // Mantiene pending hasta que se procese
	r.UpdatedAt = time.Now()
}

// GetReferenceID obtiene el ID de referencia (payment_id u order_id)
func (r *Refund) GetReferenceID() (string, int64) {
	if r.PaymentID != nil {
		return "payment", *r.PaymentID
	}
	if r.OrderID != nil {
		return "order", *r.OrderID
	}
	return "", 0
}

// GetDuration obtiene la duración total del proceso de reembolso
func (r *Refund) GetDuration() time.Duration {
	if r.CompletedAt == nil {
		return 0
	}
	return r.CompletedAt.Sub(r.RequestedAt)
}

// GetProcessingDuration obtiene la duración del procesamiento
func (r *Refund) GetProcessingDuration() time.Duration {
	if r.ProcessedAt == nil || r.RequestedAt.IsZero() {
		return 0
	}
	return r.ProcessedAt.Sub(r.RequestedAt)
}

// HasProviderReference verifica si hay una referencia del proveedor
func (r *Refund) HasProviderReference() bool {
	return r.ProviderRefundID != nil && *r.ProviderRefundID != ""
}

// Reset reinicia el estado del reembolso
func (r *Refund) Reset() {
	r.Status = "pending"
	r.ProviderRefundID = nil
	r.ProcessedAt = nil
	r.CompletedAt = nil
	r.UpdatedAt = time.Now()
}

// Clone crea una copia del Refund
func (r *Refund) Clone() *Refund {
	clone := &Refund{
		ID:           r.ID,
		RefundAmount: r.RefundAmount,
		Currency:     r.Currency,
		Status:       r.Status,
		RequestedAt:  r.RequestedAt,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}

	// Clonar punteros
	if r.PaymentID != nil {
		paymentID := *r.PaymentID
		clone.PaymentID = &paymentID
	}
	if r.OrderID != nil {
		orderID := *r.OrderID
		clone.OrderID = &orderID
	}
	if r.RefundReason != nil {
		refundReason := *r.RefundReason
		clone.RefundReason = &refundReason
	}
	if r.ProviderRefundID != nil {
		providerRefundID := *r.ProviderRefundID
		clone.ProviderRefundID = &providerRefundID
	}
	if r.RequestedBy != nil {
		requestedBy := *r.RequestedBy
		clone.RequestedBy = &requestedBy
	}
	if r.ApprovedBy != nil {
		approvedBy := *r.ApprovedBy
		clone.ApprovedBy = &approvedBy
	}
	if r.ProcessedAt != nil {
		processedAt := *r.ProcessedAt
		clone.ProcessedAt = &processedAt
	}
	if r.CompletedAt != nil {
		completedAt := *r.CompletedAt
		clone.CompletedAt = &completedAt
	}

	return clone
}
