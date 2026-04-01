// internal/api/dto/ticket/response.go
package ticket

import "time"

// TicketResponse respuesta de ticket
type TicketResponse struct {
	ID           string    `json:"id"`
	PublicID     string    `json:"public_id"`
	TicketTypeID string    `json:"ticket_type_id"`
	EventID      string    `json:"event_id"`
	Code         string    `json:"code"`
	Status       string    `json:"status"`
	FinalPrice   float64   `json:"final_price"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TicketListResponse para listar tickets
type TicketListResponse struct {
	Tickets    []TicketResponse `json:"tickets"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// TicketStatsResponse representa estadísticas de tickets
type TicketStatsResponse struct {
	TotalTickets     int64   `json:"total_tickets"`
	AvailableTickets int64   `json:"available_tickets"`
	SoldTickets      int64   `json:"sold_tickets"`
	ReservedTickets  int64   `json:"reserved_tickets"`
	CheckedInTickets int64   `json:"checked_in_tickets"`
	CancelledTickets int64   `json:"cancelled_tickets"`
	RefundedTickets  int64   `json:"refunded_tickets"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
	CheckInRate      float64 `json:"check_in_rate"`
}
