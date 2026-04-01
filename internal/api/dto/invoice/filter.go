// internal/api/dto/invoice/filter.go
package invoice

// InvoiceFilter representa filtros para facturas
type InvoiceFilter struct {
	OrderID       string  `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	CustomerID    string  `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	InvoiceNumber string  `json:"invoice_number,omitempty"`
	Status        string  `json:"status,omitempty"`
	PaymentStatus string  `json:"payment_status,omitempty"`
	DateFrom      string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo        string  `json:"date_to,omitempty" validate:"omitempty,date"`
	MinAmount     float64 `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount     float64 `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	HasCFDI       *bool   `json:"has_cfdi,omitempty"`
	TaxID         string  `json:"tax_id,omitempty"`
}
