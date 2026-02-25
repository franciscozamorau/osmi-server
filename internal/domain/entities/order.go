package entities

import (
	"errors"
	"time"
)

// Order representa una orden en el sistema de facturación
// Mapea exactamente la tabla billing.orders
type Order struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`

	CustomerID    *int64  `json:"customer_id,omitempty" db:"customer_id"`
	CustomerEmail string  `json:"customer_email" db:"customer_email"`
	CustomerName  *string `json:"customer_name,omitempty" db:"customer_name"`
	CustomerPhone *string `json:"customer_phone,omitempty" db:"customer_phone"`

	Subtotal         float64 `json:"subtotal" db:"subtotal"`
	TaxAmount        float64 `json:"tax_amount" db:"tax_amount"`
	ServiceFeeAmount float64 `json:"service_fee_amount" db:"service_fee_amount"`
	DiscountAmount   float64 `json:"discount_amount" db:"discount_amount"`
	TotalAmount      float64 `json:"total_amount" db:"total_amount"`
	Currency         string  `json:"currency" db:"currency"`

	Status    string `json:"status" db:"status"`
	OrderType string `json:"order_type" db:"order_type"`

	IsReservation        bool       `json:"is_reservation" db:"is_reservation"`
	ReservationExpiresAt *time.Time `json:"reservation_expires_at,omitempty" db:"reservation_expires_at"`

	PaymentMethod     *string `json:"payment_method,omitempty" db:"payment_method"`
	PaymentProviderID *int    `json:"payment_provider_id,omitempty" db:"payment_provider_id"`

	InvoiceRequired  bool    `json:"invoice_required" db:"invoice_required"`
	InvoiceGenerated bool    `json:"invoice_generated" db:"invoice_generated"`
	InvoiceNumber    *string `json:"invoice_number,omitempty" db:"invoice_number"`

	PromotionCode *string `json:"promotion_code,omitempty" db:"promotion_code"`
	PromotionID   *int64  `json:"promotion_id,omitempty" db:"promotion_id"`

	Metadata *map[string]interface{} `json:"metadata,omitempty" db:"metadata,type:jsonb"`
	Notes    *string                 `json:"notes,omitempty" db:"notes"`

	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`

	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// IsPending verifica si la orden está pendiente
func (o *Order) IsPending() bool {
	return o.Status == "pending"
}

// IsCompleted verifica si la orden está completada
func (o *Order) IsCompleted() bool {
	return o.Status == "completed"
}

// IsFailed verifica si la orden falló
func (o *Order) IsFailed() bool {
	return o.Status == "failed"
}

// IsRefunded verifica si la orden fue reembolsada
func (o *Order) IsRefunded() bool {
	return o.Status == "refunded" || o.RefundedAt != nil
}

// IsCancelled verifica si la orden fue cancelada
func (o *Order) IsCancelled() bool {
	return o.Status == "cancelled" || o.CancelledAt != nil
}

// IsExpired verifica si la orden expiró
func (o *Order) IsExpired() bool {
	if o.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*o.ExpiresAt)
}

// IsDisputed verifica si la orden está en disputa
func (o *Order) IsDisputed() bool {
	return o.Status == "disputed"
}

// IsChargeback verifica si la orden tiene chargeback
func (o *Order) IsChargeback() bool {
	return o.Status == "chargeback"
}

// IsActive verifica si la orden está activa (no cancelada, no expirada, no reembolsada)
func (o *Order) IsActive() bool {
	return !o.IsCancelled() && !o.IsExpired() && !o.IsRefunded() && o.Status != "failed"
}

// CanBePaid verifica si la orden puede ser pagada
func (o *Order) CanBePaid() bool {
	return o.IsPending() && !o.IsExpired() && !o.IsCancelled()
}

// CanBeCancelled verifica si la orden puede ser cancelada
func (o *Order) CanBeCancelled() bool {
	return o.IsPending() && !o.IsExpired()
}

// MarkAsPaid marca la orden como pagada
func (o *Order) MarkAsPaid() {
	now := time.Now()
	o.Status = "completed"
	o.PaidAt = &now
	o.UpdatedAt = now
}

// MarkAsFailed marca la orden como fallida
func (o *Order) MarkAsFailed() {
	now := time.Now()
	o.Status = "failed"
	o.UpdatedAt = now
}

// MarkAsCancelled marca la orden como cancelada
func (o *Order) MarkAsCancelled() {
	now := time.Now()
	o.Status = "cancelled"
	o.CancelledAt = &now
	o.UpdatedAt = now
}

// MarkAsRefunded marca la orden como reembolsada
func (o *Order) MarkAsRefunded() {
	now := time.Now()
	o.Status = "refunded"
	o.RefundedAt = &now
	o.UpdatedAt = now
}

// CalculateTotals calcula los totales basado en los componentes
func (o *Order) CalculateTotals() {
	o.TotalAmount = o.Subtotal + o.TaxAmount + o.ServiceFeeAmount - o.DiscountAmount
}

// Validate verifica que la orden sea válida
func (o *Order) Validate() error {
	if o.CustomerEmail == "" {
		return errors.New("customer_email is required")
	}
	if o.Subtotal < 0 {
		return errors.New("subtotal cannot be negative")
	}
	if o.TaxAmount < 0 {
		return errors.New("tax_amount cannot be negative")
	}
	if o.ServiceFeeAmount < 0 {
		return errors.New("service_fee_amount cannot be negative")
	}
	if o.DiscountAmount < 0 {
		return errors.New("discount_amount cannot be negative")
	}
	if o.DiscountAmount > o.Subtotal {
		return errors.New("discount_amount cannot exceed subtotal")
	}
	if o.Currency == "" {
		return errors.New("currency is required")
	}

	// Verificar la relación entre total y componentes
	calculatedTotal := o.Subtotal + o.TaxAmount + o.ServiceFeeAmount - o.DiscountAmount
	if o.TotalAmount != calculatedTotal {
		return errors.New("total_amount does not match calculated total")
	}

	return nil
}

// SetMetadata establece un valor en metadata
func (o *Order) SetMetadata(key string, value interface{}) {
	if o.Metadata == nil {
		o.Metadata = &map[string]interface{}{}
	}
	(*o.Metadata)[key] = value
}

// GetMetadata obtiene un valor de metadata
func (o *Order) GetMetadata(key string) interface{} {
	if o.Metadata == nil {
		return nil
	}
	return (*o.Metadata)[key]
}

// DeleteMetadata elimina una clave de metadata
func (o *Order) DeleteMetadata(key string) {
	if o.Metadata == nil {
		return
	}
	delete(*o.Metadata, key)
	if len(*o.Metadata) == 0 {
		o.Metadata = nil
	}
}

// HasPromotion verifica si la orden tiene una promoción aplicada
func (o *Order) HasPromotion() bool {
	return o.PromotionID != nil || (o.PromotionCode != nil && *o.PromotionCode != "")
}

// RequiresInvoice verifica si la orden requiere factura
func (o *Order) RequiresInvoice() bool {
	return o.InvoiceRequired
}

// IsInvoiceGenerated verifica si ya se generó la factura
func (o *Order) IsInvoiceGenerated() bool {
	return o.InvoiceGenerated && o.InvoiceNumber != nil
}

// GetPaymentProviderID obtiene el ID del proveedor de pago
func (o *Order) GetPaymentProviderID() int {
	if o.PaymentProviderID == nil {
		return 0
	}
	return *o.PaymentProviderID
}
