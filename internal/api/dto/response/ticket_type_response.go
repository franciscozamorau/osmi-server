package response

import "time"

type TicketTypeResponse struct {
	ID                string     `json:"id"`
	PublicID          string     `json:"public_id"`
	EventID           string     `json:"event_id"`
	EventName         string     `json:"event_name,omitempty"`
	Name              string     `json:"name"`
	Description       *string    `json:"description,omitempty"`
	TicketClass       string     `json:"ticket_class"`
	BasePrice         float64    `json:"base_price"`
	Currency          string     `json:"currency"`
	TaxRate           float64    `json:"tax_rate"`
	ServiceFeeType    string     `json:"service_fee_type"`
	ServiceFeeValue   float64    `json:"service_fee_value"`
	TotalQuantity     int32      `json:"total_quantity"`
	ReservedQuantity  int32      `json:"reserved_quantity"`
	SoldQuantity      int32      `json:"sold_quantity"`
	AvailableQuantity int32      `json:"available_quantity"`
	IsSoldOut         bool       `json:"is_sold_out"`
	MaxPerOrder       int32      `json:"max_per_order"`
	MinPerOrder       int32      `json:"min_per_order"`
	SaleStartsAt      time.Time  `json:"sale_starts_at"`
	SaleEndsAt        *time.Time `json:"sale_ends_at,omitempty"`
	IsActive          bool       `json:"is_active"`
	RequiresApproval  bool       `json:"requires_approval"`
	IsHidden          bool       `json:"is_hidden"`
	SalesChannel      string     `json:"sales_channel"`
	Benefits          *string    `json:"benefits,omitempty"`
	AccessType        string     `json:"access_type"`
	ValidationRules   *string    `json:"validation_rules,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type TicketTypeListResponse struct {
	TicketTypes []TicketTypeResponse `json:"ticket_types"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
	TotalPages  int                  `json:"total_pages"`
}

type TicketTypeStatsResponse struct {
	TotalTickets     int64   `json:"total_tickets"`
	ReservedTickets  int64   `json:"reserved_tickets"`
	SoldTickets      int64   `json:"sold_tickets"`
	AvailableTickets int64   `json:"available_tickets"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
	SellThroughRate  float64 `json:"sell_through_rate"`
}
