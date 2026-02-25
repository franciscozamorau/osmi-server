package response

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
	Tickets   []TicketResponse `json:"tickets"`
	Total     int64            `json:"total"`
	Page      int              `json:"page"`
	PageSize  int              `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}
