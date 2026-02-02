package dto

import "time"

type TicketTypeResponse struct {
	ID                string                 `json:"id"`
	Event             EventBasicInfo         `json:"event"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	TicketClass       string                 `json:"ticket_class"`
	BasePrice         float64                `json:"base_price"`
	Currency          string                 `json:"currency"`
	TaxRate           float64                `json:"tax_rate"`
	ServiceFeeType    string                 `json:"service_fee_type"`
	ServiceFeeValue   float64                `json:"service_fee_value"`
	TotalQuantity     int                    `json:"total_quantity"`
	ReservedQuantity  int                    `json:"reserved_quantity"`
	SoldQuantity      int                    `json:"sold_quantity"`
	AvailableQuantity int                    `json:"available_quantity"`
	MaxPerOrder       int                    `json:"max_per_order"`
	MinPerOrder       int                    `json:"min_per_order"`
	SaleStartsAt      time.Time              `json:"sale_starts_at"`
	SaleEndsAt        time.Time              `json:"sale_ends_at,omitempty"`
	IsActive          bool                   `json:"is_active"`
	IsSoldOut         bool                   `json:"is_sold_out"`
	RequiresApproval  bool                   `json:"requires_approval"`
	IsHidden          bool                   `json:"is_hidden"`
	SalesChannel      string                 `json:"sales_channel"`
	Benefits          []string               `json:"benefits"`
	AccessType        string                 `json:"access_type"`
	ValidationRules   map[string]interface{} `json:"validation_rules,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type TicketTypeStatsResponse struct {
	TotalSold       int     `json:"total_sold"`
	TotalRevenue    float64 `json:"total_revenue"`
	AvgPrice        float64 `json:"avg_price"`
	ConversionRate  float64 `json:"conversion_rate"`
	ReservationRate float64 `json:"reservation_rate"`
	PeakSalesHour   string  `json:"peak_sales_hour,omitempty"`
	SalesVelocity   float64 `json:"sales_velocity"`
}

type TicketTypeListResponse struct {
	TicketTypes []TicketTypeResponse    `json:"ticket_types"`
	Total       int64                   `json:"total"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
	TotalPages  int                     `json:"total_pages"`
	Stats       TicketTypeStatsResponse `json:"stats"`
}

type EventBasicInfo struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`
	Status string    `json:"status"`
}
