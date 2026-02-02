package entities

import "time"

type Customer struct {
	ID          int64   `json:"id" db:"id"`
	PublicID    string  `json:"public_id" db:"public_uuid"`
	UserID      *int64  `json:"user_id,omitempty" db:"user_id"`
	FullName    string  `json:"full_name" db:"full_name"`
	Email       string  `json:"email" db:"email"`
	Phone       *string `json:"phone,omitempty" db:"phone"`
	CompanyName *string `json:"company_name,omitempty" db:"company_name"`

	AddressLine1 *string `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty" db:"address_line2"`
	City         *string `json:"city,omitempty" db:"city"`
	State        *string `json:"state,omitempty" db:"state"`
	PostalCode   *string `json:"postal_code,omitempty" db:"postal_code"`
	Country      *string `json:"country,omitempty" db:"country"`

	TaxID           *string `json:"tax_id,omitempty" db:"tax_id"`
	TaxIDType       *string `json:"tax_id_type,omitempty" db:"tax_id_type"`
	TaxName         *string `json:"tax_name,omitempty" db:"tax_name"`
	RequiresInvoice bool    `json:"requires_invoice" db:"requires_invoice"`

	CommunicationPreferences *string `json:"communication_preferences,omitempty" db:"communication_preferences"`

	TotalSpent    float64 `json:"total_spent" db:"total_spent"`
	TotalOrders   int32   `json:"total_orders" db:"total_orders"`
	TotalTickets  int32   `json:"total_tickets" db:"total_tickets"`
	AvgOrderValue float64 `json:"avg_order_value" db:"avg_order_value"`

	FirstOrderAt   *time.Time `json:"first_order_at,omitempty" db:"first_order_at"`
	LastOrderAt    *time.Time `json:"last_order_at,omitempty" db:"last_order_at"`
	LastPurchaseAt *time.Time `json:"last_purchase_at,omitempty" db:"last_purchase_at"`

	IsActive bool       `json:"is_active" db:"is_active"`
	IsVIP    bool       `json:"is_vip" db:"is_vip"`
	VIPSince *time.Time `json:"vip_since,omitempty" db:"vip_since"`

	CustomerSegment string  `json:"customer_segment" db:"customer_segment"`
	LifetimeValue   float64 `json:"lifetime_value" db:"lifetime_value"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
