package dto

type CreateCustomerRequest struct {
	UserID          string `json:"user_id,omitempty" validate:"omitempty,uuid4"`
	FullName        string `json:"full_name" validate:"required,max=255"`
	Email           string `json:"email" validate:"required,email,max=320"`
	Phone           string `json:"phone,omitempty" validate:"omitempty,phone"`
	CompanyName     string `json:"company_name,omitempty" validate:"omitempty,max=255"`
	TaxID           string `json:"tax_id,omitempty" validate:"omitempty,max=50"`
	TaxIDType       string `json:"tax_id_type,omitempty" validate:"omitempty,oneof=rfc ein vat other"`
	RequiresInvoice bool   `json:"requires_invoice,omitempty"`
	AddressLine1    string `json:"address_line1,omitempty" validate:"omitempty,max=255"`
	AddressLine2    string `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City            string `json:"city,omitempty" validate:"omitempty,max=100"`
	State           string `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode      string `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	Country         string `json:"country,omitempty" validate:"omitempty,country_code"`
}

type UpdateCustomerRequest struct {
	FullName        string `json:"full_name,omitempty" validate:"omitempty,max=255"`
	Phone           string `json:"phone,omitempty" validate:"omitempty,phone"`
	CompanyName     string `json:"company_name,omitempty" validate:"omitempty,max=255"`
	TaxID           string `json:"tax_id,omitempty" validate:"omitempty,max=50"`
	TaxIDType       string `json:"tax_id_type,omitempty" validate:"omitempty,oneof=rfc ein vat other"`
	RequiresInvoice *bool  `json:"requires_invoice,omitempty"`
	AddressLine1    string `json:"address_line1,omitempty" validate:"omitempty,max=255"`
	AddressLine2    string `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City            string `json:"city,omitempty" validate:"omitempty,max=100"`
	State           string `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode      string `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	Country         string `json:"country,omitempty" validate:"omitempty,country_code"`
}

type CustomerFilter struct {
	Search          string `json:"search,omitempty"`
	Country         string `json:"country,omitempty"`
	IsActive        *bool  `json:"is_active,omitempty"`
	IsVIP           *bool  `json:"is_vip,omitempty"`
	CustomerSegment string `json:"customer_segment,omitempty"`
	DateFrom        string `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo          string `json:"date_to,omitempty" validate:"omitempty,date"`
}