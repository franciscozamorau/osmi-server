// internal/api/dto/refund/filter.go
package refund

type RefundFilter struct {
	PaymentID     *string  `json:"payment_id,omitempty" validate:"omitempty,uuid4"`
	OrderID       *string  `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	CustomerID    *string  `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	Status        *string  `json:"status,omitempty" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	RefundReason  *string  `json:"refund_reason,omitempty" validate:"omitempty,max=100"`
	DateFrom      *string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo        *string  `json:"date_to,omitempty" validate:"omitempty,date"`
	MinAmount     *float64 `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount     *float64 `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	PartialOnly   *bool    `json:"partial_only,omitempty"`
	HasProviderID *bool    `json:"has_provider_id,omitempty"`
}
