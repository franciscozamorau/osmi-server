package dto

type CreateTicketRequest struct {
	TicketTypeID  string `json:"ticket_type_id" validate:"required,uuid4"`
	CustomerID    string `json:"customer_id" validate:"required,uuid4"`
	OrderID       string `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	AttendeeName  string `json:"attendee_name,omitempty" validate:"omitempty,max=255"`
	AttendeeEmail string `json:"attendee_email,omitempty" validate:"omitempty,email"`
	AttendeePhone string `json:"attendee_phone,omitempty" validate:"omitempty,phone"`
}

type ReserveTicketRequest struct {
	TicketTypeID    string `json:"ticket_type_id" validate:"required,uuid4"`
	CustomerID      string `json:"customer_id" validate:"required,uuid4"`
	Quantity        int    `json:"quantity" validate:"required,min=1,max=10"`
	DurationMinutes int    `json:"duration_minutes,omitempty" validate:"omitempty,min=1,max=1440"`
}

type UpdateTicketRequest struct {
	Status        string `json:"status,omitempty" validate:"omitempty,oneof=reserved sold checked_in cancelled refunded"`
	AttendeeName  string `json:"attendee_name,omitempty" validate:"omitempty,max=255"`
	AttendeeEmail string `json:"attendee_email,omitempty" validate:"omitempty,email"`
	AttendeePhone string `json:"attendee_phone,omitempty" validate:"omitempty,phone"`
}

type CheckInTicketRequest struct {
	TicketID        string `json:"ticket_id" validate:"required,uuid4"`
	CheckinMethod   string `json:"checkin_method" validate:"required,oneof=qr_code manual facial"`
	CheckinLocation string `json:"checkin_location,omitempty" validate:"omitempty,max=100"`
	ValidatorID     string `json:"validator_id,omitempty" validate:"omitempty,uuid4"`
}

type TransferTicketRequest struct {
	TicketID       string `json:"ticket_id" validate:"required,uuid4"`
	FromCustomerID string `json:"from_customer_id" validate:"required,uuid4"`
	ToCustomerID   string `json:"to_customer_id,omitempty" validate:"omitempty,uuid4"`
	ToEmail        string `json:"to_email" validate:"required,email"`
	ToName         string `json:"to_name,omitempty" validate:"omitempty,max=255"`
}

type TicketFilter struct {
	EventID      string `json:"event_id,omitempty" validate:"omitempty,uuid4"`
	CustomerID   string `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	OrderID      string `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	Status       string `json:"status,omitempty"`
	TicketTypeID string `json:"ticket_type_id,omitempty" validate:"omitempty,uuid4"`
	Code         string `json:"code,omitempty"`
	DateFrom     string `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo       string `json:"date_to,omitempty" validate:"omitempty,date"`
	CheckedIn    *bool  `json:"checked_in,omitempty"`
}
