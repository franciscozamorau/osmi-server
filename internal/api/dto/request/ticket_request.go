package request

import "github.com/go-playground/validator/v10"

// CreateTicketRequest para crear un ticket
type CreateTicketRequest struct {
	EventID    string `json:"event_id" validate:"required"`
	CustomerID string `json:"customer_id" validate:"required"`
	CategoryID string `json:"category_id" validate:"required"`
	Quantity   int32  `json:"quantity" validate:"required,min=1,max=10"`
	UserID     string `json:"user_id,omitempty"`
}

// Validate valida la estructura
func (r *CreateTicketRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// UpdateTicketStatusRequest para actualizar estado de ticket
type UpdateTicketStatusRequest struct {
	TicketID string `json:"ticket_id" validate:"required"`
	Status   string `json:"status" validate:"required,oneof=available reserved sold checked_in cancelled refunded"`
	Reason   string `json:"reason,omitempty"`
}

// Validate valida la estructura
func (r *UpdateTicketStatusRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
