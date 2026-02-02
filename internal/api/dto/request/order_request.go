package dto

type CreateOrderRequest struct {
	CustomerID          string                   `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	CustomerEmail       string                   `json:"customer_email" validate:"required,email"`
	CustomerName        string                   `json:"customer_name,omitempty" validate:"omitempty,max=255"`
	CustomerPhone       string                   `json:"customer_phone,omitempty" validate:"omitempty,phone"`
	Items               []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	PromotionCode       string                   `json:"promotion_code,omitempty"`
	Currency            string                   `json:"currency" validate:"required,oneof=MXN USD EUR"`
	IsReservation       bool                     `json:"is_reservation,omitempty"`
	ReservationDuration int                      `json:"reservation_duration,omitempty" validate:"omitempty,min=1,max=1440"`
	InvoiceRequired     bool                     `json:"invoice_required,omitempty"`
	Notes               string                   `json:"notes,omitempty"`
}

type CreateOrderItemRequest struct {
	TicketTypeID string  `json:"ticket_type_id" validate:"required,uuid4"`
	Quantity     int     `json:"quantity" validate:"required,min=1,max=20"`
	UnitPrice    float64 `json:"unit_price,omitempty" validate:"omitempty,min=0"`
}

type UpdateOrderRequest struct {
	Status        string `json:"status,omitempty" validate:"omitempty,oneof=pending processing completed failed refunded"`
	PaymentMethod string `json:"payment_method,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

type ProcessPaymentRequest struct {
	OrderID              string                 `json:"order_id" validate:"required,uuid4"`
	PaymentMethod        string                 `json:"payment_method" validate:"required"`
	PaymentProvider      string                 `json:"payment_provider" validate:"required"`
	PaymentMethodDetails map[string]interface{} `json:"payment_method_details,omitempty"`
	SaveCard             bool                   `json:"save_card,omitempty"`
}

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
