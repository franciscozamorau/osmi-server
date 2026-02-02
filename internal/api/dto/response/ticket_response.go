package dto

import "time"

type TicketResponse struct {
	ID                   string         `json:"id"`
	TicketType           TicketTypeInfo `json:"ticket_type"`
	Event                EventInfo      `json:"event"`
	Customer             CustomerInfo   `json:"customer"`
	Order                OrderInfo      `json:"order,omitempty"`
	Code                 string         `json:"code"`
	QRCodeData           string         `json:"qr_code_data,omitempty"`
	Status               string         `json:"status"`
	FinalPrice           float64        `json:"final_price"`
	Currency             string         `json:"currency"`
	TaxAmount            float64        `json:"tax_amount"`
	AttendeeName         string         `json:"attendee_name,omitempty"`
	AttendeeEmail        string         `json:"attendee_email,omitempty"`
	AttendeePhone        string         `json:"attendee_phone,omitempty"`
	CheckedInAt          time.Time      `json:"checked_in_at,omitempty"`
	CheckinMethod        string         `json:"checkin_method,omitempty"`
	CheckinLocation      string         `json:"checkin_location,omitempty"`
	ReservedAt           time.Time      `json:"reserved_at,omitempty"`
	ReservationExpiresAt time.Time      `json:"reservation_expires_at,omitempty"`
	TransferToken        string         `json:"transfer_token,omitempty"`
	TransferredFrom      CustomerInfo   `json:"transferred_from,omitempty"`
	TransferredAt        time.Time      `json:"transferred_at,omitempty"`
	ValidationCount      int            `json:"validation_count"`
	LastValidatedAt      time.Time      `json:"last_validated_at,omitempty"`
	SoldAt               time.Time      `json:"sold_at,omitempty"`
	CancelledAt          time.Time      `json:"cancelled_at,omitempty"`
	RefundedAt           time.Time      `json:"refunded_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type TicketStatsResponse struct {
	TotalTickets     int     `json:"total_tickets"`
	AvailableTickets int     `json:"available_tickets"`
	ReservedTickets  int     `json:"reserved_tickets"`
	SoldTickets      int     `json:"sold_tickets"`
	CheckedInTickets int     `json:"checked_in_tickets"`
	CancelledTickets int     `json:"cancelled_tickets"`
	RefundedTickets  int     `json:"refunded_tickets"`
	TotalRevenue     float64 `json:"total_revenue"`
	CheckInRate      float64 `json:"check_in_rate"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
}

type TicketListResponse struct {
	Tickets    []TicketResponse    `json:"tickets"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
	Stats      TicketStatsResponse `json:"stats"`
}

type EventInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
	Location string    `json:"location"`
	Venue    string    `json:"venue,omitempty"`
}

type CustomerInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
}

type OrderInfo struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}
