package response

import "time"

type CustomerResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id,omitempty"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone,omitempty"`
	CompanyName     string    `json:"company_name,omitempty"`
	AddressLine1    string    `json:"address_line1,omitempty"`
	AddressLine2    string    `json:"address_line2,omitempty"`
	City            string    `json:"city,omitempty"`
	State           string    `json:"state,omitempty"`
	PostalCode      string    `json:"postal_code,omitempty"`
	Country         string    `json:"country,omitempty"`
	TaxID           string    `json:"tax_id,omitempty"`
	TaxIDType       string    `json:"tax_id_type,omitempty"`
	RequiresInvoice bool      `json:"requires_invoice"`
	TotalSpent      float64   `json:"total_spent"`
	TotalOrders     int       `json:"total_orders"`
	TotalTickets    int       `json:"total_tickets"`
	AvgOrderValue   float64   `json:"avg_order_value"`
	FirstOrderAt    string    `json:"first_order_at,omitempty"`
	LastOrderAt     string    `json:"last_order_at,omitempty"`
	IsActive        bool      `json:"is_active"`
	IsVIP           bool      `json:"is_vip"`
	VIPSince        string    `json:"vip_since,omitempty"`
	CustomerSegment string    `json:"customer_segment"`
	LifetimeValue   float64   `json:"lifetime_value"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CustomerStatsResponse struct {
	TotalCustomers         int64          `json:"total_customers"`
	ActiveCustomers        int64          `json:"active_customers"`
	VIPCustomers           int64          `json:"vip_customers"`
	NewCustomersLast30Days int64          `json:"new_customers_last_30_days"`
	TotalRevenue           float64        `json:"total_revenue"`
	AvgLifetimeValue       float64        `json:"avg_lifetime_value"`
	TopCountries           []CountryStats `json:"top_countries"`
}

type CountryStats struct {
	Country string  `json:"country"`
	Count   int64   `json:"count"`
	Revenue float64 `json:"revenue"`
}

type CustomerListResponse struct {
	Customers  []CustomerResponse    `json:"customers"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
	Stats      CustomerStatsResponse `json:"stats"`
}
