// internal/api/dto/ticket_type/filter.go
package ticket_type

type TicketTypeFilter struct {
	EventID   *int64   `json:"event_id,omitempty"`
	Name      string   `json:"name,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
	IsSoldOut *bool    `json:"is_sold_out,omitempty"`
	MinPrice  *float64 `json:"min_price,omitempty"`
	MaxPrice  *float64 `json:"max_price,omitempty"`
	Currency  string   `json:"currency,omitempty"`
	Search    string   `json:"search,omitempty"`
}
