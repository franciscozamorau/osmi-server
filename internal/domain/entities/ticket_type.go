package entities

import "time"

type TicketType struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`
	EventID  int64  `json:"event_id" db:"event_id"`

	Name        string  `json:"name" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`
	TicketClass string  `json:"ticket_class" db:"ticket_class"`

	BasePrice       float64 `json:"base_price" db:"base_price"`
	Currency        string  `json:"currency" db:"currency"`
	TaxRate         float64 `json:"tax_rate" db:"tax_rate"`
	ServiceFeeType  string  `json:"service_fee_type" db:"service_fee_type"`
	ServiceFeeValue float64 `json:"service_fee_value" db:"service_fee_value"`

	TotalQuantity    int32 `json:"total_quantity" db:"total_quantity"`
	ReservedQuantity int32 `json:"reserved_quantity" db:"reserved_quantity"`
	SoldQuantity     int32 `json:"sold_quantity" db:"sold_quantity"`
	MaxPerOrder      int32 `json:"max_per_order" db:"max_per_order"`
	MinPerOrder      int32 `json:"min_per_order" db:"min_per_order"`

	SaleStartsAt time.Time  `json:"sale_starts_at" db:"sale_starts_at"`
	SaleEndsAt   *time.Time `json:"sale_ends_at,omitempty" db:"sale_ends_at"`

	IsActive         bool   `json:"is_active" db:"is_active"`
	RequiresApproval bool   `json:"requires_approval" db:"requires_approval"`
	IsHidden         bool   `json:"is_hidden" db:"is_hidden"`
	SalesChannel     string `json:"sales_channel" db:"sales_channel"`

	Benefits   *string `json:"benefits,omitempty" db:"benefits"`
	AccessType string  `json:"access_type" db:"access_type"`

	ValidationRules *string `json:"validation_rules,omitempty" db:"validation_rules"`

	AvailableQuantity int32 `json:"available_quantity" db:"available_quantity"`
	IsSoldOut         bool  `json:"is_sold_out" db:"is_sold_out"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
