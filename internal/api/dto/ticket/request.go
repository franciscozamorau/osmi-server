// internal/api/dto/ticket/request.go
package ticket

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// CreateTicketRequest para crear un ticket
type CreateTicketRequest struct {
	EventID      string `json:"event_id" validate:"required"`
	CustomerID   string `json:"customer_id" validate:"required"`
	TicketTypeID string `json:"ticketTypeId" validate:"required"`
	Quantity     int32  `json:"quantity" validate:"required,min=1,max=10"`
	UserID       string `json:"user_id,omitempty"`
}

// Validate valida la estructura
func (r *CreateTicketRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// UpdateTicketRequest para actualizar un ticket
type UpdateTicketRequest struct {
	AttendeeName  *string `json:"attendee_name,omitempty"`
	AttendeeEmail *string `json:"attendee_email,omitempty"`
	AttendeePhone *string `json:"attendee_phone,omitempty"`
	Status        *string `json:"status,omitempty"`
}

// UpdateTicketStatusRequest para actualizar estado de ticket
type UpdateTicketStatusRequest struct {
	TicketID string `json:"ticket_id" validate:"required"`
	Status   string `json:"status" validate:"required,oneof=available reserved sold checked_in cancelled transferred refunded expired"`
	Reason   string `json:"reason,omitempty"`
}

// Validate valida la estructura
func (r *UpdateTicketStatusRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// ReserveTicketRequest para reservar un ticket
type ReserveTicketRequest struct {
	TicketID  string    `json:"ticket_id" validate:"required"`
	UserID    string    `json:"user_id" validate:"required"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// CheckInTicketRequest para marcar ticket como usado
type CheckInTicketRequest struct {
	TicketID  string `json:"ticket_id" validate:"required"`
	CheckedBy string `json:"checked_by" validate:"required"`
	Method    string `json:"method,omitempty"`
	Location  string `json:"location,omitempty"`
}

// TransferTicketRequest para transferir un ticket
type TransferTicketRequest struct {
	TicketID       string `json:"ticket_id" validate:"required"`
	FromCustomerID string `json:"from_customer_id" validate:"required"`
	ToCustomerID   string `json:"to_customer_id" validate:"required"`
	Token          string `json:"token,omitempty"`
}

// PurchaseTicketRequest para comprar un ticket reservado
type PurchaseTicketRequest struct {
	TicketID   string `json:"ticket_id" validate:"required"`
	CustomerID string `json:"customer_id" validate:"required"`
}
