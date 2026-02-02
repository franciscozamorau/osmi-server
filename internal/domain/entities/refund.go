package entities

import "time"

type Refund struct {
	ID        int64  `json:"id" db:"id"`
	PaymentID *int64 `json:"payment_id,omitempty" db:"payment_id"`
	OrderID   *int64 `json:"order_id,omitempty" db:"order_id"`

	RefundReason *string `json:"refund_reason,omitempty" db:"refund_reason"`
	RefundAmount float64 `json:"refund_amount" db:"refund_amount"`
	Currency     string  `json:"currency" db:"currency"`

	Status           string  `json:"status" db:"status"`
	ProviderRefundID *string `json:"provider_refund_id,omitempty" db:"provider_refund_id"`

	RequestedBy *int64 `json:"requested_by,omitempty" db:"requested_by"`
	ApprovedBy  *int64 `json:"approved_by,omitempty" db:"approved_by"`

	RequestedAt time.Time  `json:"requested_at" db:"requested_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
