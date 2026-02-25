package request

type CreateOrganizerRequest struct {
	Name         string            `json:"name" validate:"required,min=2,max=255"`
	Slug         string            `json:"slug" validate:"required,slug"`
	Description  string            `json:"description,omitempty"`
	LogoURL      string            `json:"logo_url,omitempty" validate:"omitempty,url"`
	LegalName    string            `json:"legal_name,omitempty" validate:"omitempty,max=255"`
	TaxID        string            `json:"tax_id,omitempty" validate:"omitempty,max=50"`
	TaxIDType    string            `json:"tax_id_type,omitempty" validate:"omitempty,oneof=rfc ein vat other"`
	Country      string            `json:"country,omitempty" validate:"omitempty,country_code"`
	ContactEmail string            `json:"contact_email" validate:"required,email"`
	ContactPhone string            `json:"contact_phone,omitempty" validate:"omitempty,phone"`
	AddressLine1 string            `json:"address_line1,omitempty" validate:"omitempty,max=255"`
	AddressLine2 string            `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City         string            `json:"city,omitempty" validate:"omitempty,max=100"`
	State        string            `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode   string            `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	SocialLinks  map[string]string `json:"social_links,omitempty"`
}

type UpdateOrganizerRequest struct {
	Name         string            `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  string            `json:"description,omitempty"`
	LogoURL      string            `json:"logo_url,omitempty" validate:"omitempty,url"`
	ContactEmail string            `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone string            `json:"contact_phone,omitempty" validate:"omitempty,phone"`
	AddressLine1 string            `json:"address_line1,omitempty" validate:"omitempty,max=255"`
	AddressLine2 string            `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City         string            `json:"city,omitempty" validate:"omitempty,max=100"`
	State        string            `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode   string            `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	IsActive     *bool             `json:"is_active,omitempty"`
	SocialLinks  map[string]string `json:"social_links,omitempty"`
}

type VerifyOrganizerRequest struct {
	VerificationStatus string `json:"verification_status" validate:"required,oneof=verified rejected pending"`
	Notes              string `json:"notes,omitempty"`
}

type OrganizerFilter struct {
	Search             string  `json:"search,omitempty"`
	Country            string  `json:"country,omitempty"`
	IsVerified         *bool   `json:"is_verified,omitempty"`
	IsActive           *bool   `json:"is_active,omitempty"`
	VerificationStatus string  `json:"verification_status,omitempty"`
	MinRating          float64 `json:"min_rating,omitempty" validate:"omitempty,min=0,max=5"`
	MinEvents          int     `json:"min_events,omitempty" validate:"omitempty,min=0"`
}
