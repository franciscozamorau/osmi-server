package dto

type CreateTicketTypeRequest struct {
	EventID          string                 `json:"event_id" validate:"required,uuid4"`
	Name             string                 `json:"name" validate:"required,min=2,max=100"`
	Description      string                 `json:"description,omitempty"`
	TicketClass      string                 `json:"ticket_class" validate:"required,oneof=standard vip premium economy"`
	BasePrice        float64                `json:"base_price" validate:"required,min=0"`
	Currency         string                 `json:"currency" validate:"required,oneof=MXN USD EUR"`
	TaxRate          float64                `json:"tax_rate" validate:"required,min=0,max=1"`
	ServiceFeeType   string                 `json:"service_fee_type" validate:"required,oneof=percentage fixed none"`
	ServiceFeeValue  float64                `json:"service_fee_value" validate:"min=0"`
	TotalQuantity    int                    `json:"total_quantity" validate:"required,min=1"`
	MaxPerOrder      int                    `json:"max_per_order" validate:"required,min=1,max=20"`
	MinPerOrder      int                    `json:"min_per_order" validate:"required,min=1"`
	SaleStartsAt     string                 `json:"sale_starts_at" validate:"required,datetime"`
	SaleEndsAt       string                 `json:"sale_ends_at,omitempty" validate:"omitempty,datetime"`
	IsActive         bool                   `json:"is_active,omitempty"`
	RequiresApproval bool                   `json:"requires_approval,omitempty"`
	IsHidden         bool                   `json:"is_hidden,omitempty"`
	SalesChannel     string                 `json:"sales_channel" validate:"required,oneof=all online box_office phone"`
	Benefits         []string               `json:"benefits,omitempty"`
	AccessType       string                 `json:"access_type" validate:"required,oneof=general vip backstage meet_and_greet"`
	ValidationRules  map[string]interface{} `json:"validation_rules,omitempty"`
}

type UpdateTicketTypeRequest struct {
	Name            string                 `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description     string                 `json:"description,omitempty"`
	BasePrice       float64                `json:"base_price,omitempty" validate:"omitempty,min=0"`
	TotalQuantity   int                    `json:"total_quantity,omitempty" validate:"omitempty,min=0"`
	MaxPerOrder     int                    `json:"max_per_order,omitempty" validate:"omitempty,min=1,max=20"`
	MinPerOrder     int                    `json:"min_per_order,omitempty" validate:"omitempty,min=1"`
	SaleStartsAt    string                 `json:"sale_starts_at,omitempty" validate:"omitempty,datetime"`
	SaleEndsAt      string                 `json:"sale_ends_at,omitempty" validate:"omitempty,datetime"`
	IsActive        *bool                  `json:"is_active,omitempty"`
	IsHidden        *bool                  `json:"is_hidden,omitempty"`
	Benefits        []string               `json:"benefits,omitempty"`
	ValidationRules map[string]interface{} `json:"validation_rules,omitempty"`
}

type TicketTypeFilter struct {
	EventID     string  `json:"event_id,omitempty" validate:"omitempty,uuid4"`
	IsActive    *bool   `json:"is_active,omitempty"`
	IsSoldOut   *bool   `json:"is_sold_out,omitempty"`
	TicketClass string  `json:"ticket_class,omitempty"`
	AccessType  string  `json:"access_type,omitempty"`
	MinPrice    float64 `json:"min_price,omitempty" validate:"omitempty,min=0"`
	MaxPrice    float64 `json:"max_price,omitempty" validate:"omitempty,min=0"`
	OnSale      *bool   `json:"on_sale,omitempty"`
}
