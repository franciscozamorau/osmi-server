package request

type CreateTicketTypeRequest struct {
	EventID          string  `json:"event_id" validate:"required,uuid4"`
	Name             string  `json:"name" validate:"required,min=3,max=100"`
	Description      string  `json:"description,omitempty"`
	TicketClass      string  `json:"ticket_class" validate:"required,oneof=standard vip early_bird group"`
	BasePrice        float64 `json:"base_price" validate:"required,min=0"`
	Currency         string  `json:"currency" validate:"required,len=3"`
	TaxRate          float64 `json:"tax_rate" validate:"min=0,max=1"`
	ServiceFeeType   string  `json:"service_fee_type" validate:"oneof=percentage fixed"`
	ServiceFeeValue  float64 `json:"service_fee_value" validate:"min=0"`
	TotalQuantity    int     `json:"total_quantity" validate:"required,min=1"`
	MaxPerOrder      int     `json:"max_per_order" validate:"required,min=1"`
	MinPerOrder      int     `json:"min_per_order" validate:"required,min=1"`
	SaleStartsAt     string  `json:"sale_starts_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	SaleEndsAt       string  `json:"sale_ends_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	IsActive         bool    `json:"is_active"`
	RequiresApproval bool    `json:"requires_approval"`
	IsHidden         bool    `json:"is_hidden"`
	SalesChannel     string  `json:"sales_channel" validate:"oneof=all online offline"`
	Benefits         string  `json:"benefits,omitempty"` // JSON string
	AccessType       string  `json:"access_type" validate:"oneof=general vip backstage"`
	ValidationRules  string  `json:"validation_rules,omitempty"` // JSON string
}

type UpdateTicketTypeRequest struct {
	Name             *string  `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description      *string  `json:"description,omitempty"`
	BasePrice        *float64 `json:"base_price,omitempty" validate:"omitempty,min=0"`
	TaxRate          *float64 `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=1"`
	ServiceFeeType   *string  `json:"service_fee_type,omitempty" validate:"omitempty,oneof=percentage fixed"`
	ServiceFeeValue  *float64 `json:"service_fee_value,omitempty" validate:"omitempty,min=0"`
	TotalQuantity    *int     `json:"total_quantity,omitempty" validate:"omitempty,min=1"`
	MaxPerOrder      *int     `json:"max_per_order,omitempty" validate:"omitempty,min=1"`
	MinPerOrder      *int     `json:"min_per_order,omitempty" validate:"omitempty,min=1"`
	SaleStartsAt     *string  `json:"sale_starts_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	SaleEndsAt       *string  `json:"sale_ends_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	IsActive         *bool    `json:"is_active,omitempty"`
	RequiresApproval *bool    `json:"requires_approval,omitempty"`
	IsHidden         *bool    `json:"is_hidden,omitempty"`
	Benefits         *string  `json:"benefits,omitempty"`
	AccessType       *string  `json:"access_type,omitempty" validate:"omitempty,oneof=general vip backstage"`
	ValidationRules  *string  `json:"validation_rules,omitempty"`
}
