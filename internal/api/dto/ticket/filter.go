// internal/api/dto/ticket/filter.go
package ticket

// TicketFilter representa filtros para listar tickets
type TicketFilter struct {
	EventID      *int64   `json:"event_id,omitempty"`
	CustomerID   *int64   `json:"customer_id,omitempty"`
	OrderID      *int64   `json:"order_id,omitempty"`
	Status       string   `json:"status,omitempty"`
	TicketTypeID *int64   `json:"ticket_type_id,omitempty"`
	DateFrom     string   `json:"date_from,omitempty"`
	DateTo       string   `json:"date_to,omitempty"`
	MinPrice     *float64 `json:"min_price,omitempty"`
	MaxPrice     *float64 `json:"max_price,omitempty"`
	Code         string   `json:"code,omitempty"`
	Search       string   `json:"search,omitempty"`
}
