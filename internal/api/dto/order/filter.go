// internal/api/dto/order/filter.go
package order

// OrderFilter representa filtros para órdenes
type OrderFilter struct {
	CustomerID    string  `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	CustomerEmail string  `json:"customer_email,omitempty" validate:"omitempty,email"`
	Status        string  `json:"status,omitempty"`
	OrderType     string  `json:"order_type,omitempty"`
	DateFrom      string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo        string  `json:"date_to,omitempty" validate:"omitempty,date"`
	MinAmount     float64 `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount     float64 `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	HasInvoice    *bool   `json:"has_invoice,omitempty"`
}
