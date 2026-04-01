// internal/api/dto/payment/filter.go
package payment

// PaymentFilter representa filtros para listar pagos
type PaymentFilter struct {
	OrderID         string  `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	Status          string  `json:"status,omitempty"`
	PaymentMethod   string  `json:"payment_method,omitempty"`
	PaymentProvider string  `json:"payment_provider,omitempty"`
	DateFrom        string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo          string  `json:"date_to,omitempty" validate:"omitempty,date"`
	MinAmount       float64 `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount       float64 `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	Attempts        int     `json:"attempts,omitempty" validate:"omitempty,min=0"`
}
