package entities

import "time"

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
	PaymentProviderID *int32  `json:"payment_provider_id,omitempty" db:"payment_provider_id"`

	InvoiceRequired  bool    `json:"invoice_required" db:"invoice_required"`
	InvoiceGenerated bool    `json:"invoice_generated" db:"invoice_generated"`
	InvoiceNumber    *string `json:"invoice_number,omitempty" db:"invoice_number"`

	PromotionCode *string `json:"promotion_code,omitempty" db:"promotion_code"`
	PromotionID   *int64  `json:"promotion_id,omitempty" db:"promotion_id"`

	Metadata *string `json:"metadata,omitempty" db:"metadata"`
	Notes    *string `json:"notes,omitempty" db:"notes"`

	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`

	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
