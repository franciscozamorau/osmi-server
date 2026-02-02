package dto

type CountryConfigRequest struct {
	CountryCode             string                 `json:"country_code" validate:"required,country_code"`
	CountryName             string                 `json:"country_name" validate:"required"`
	TaxSystem               string                 `json:"tax_system" validate:"required"`
	DefaultTaxRate          float64                `json:"default_tax_rate" validate:"required,min=0,max=1"`
	TaxInclusiveDefault     bool                   `json:"tax_inclusive_default"`
	InvoiceRequired         bool                   `json:"invoice_required"`
	InvoiceSequenceFormat   string                 `json:"invoice_sequence_format,omitempty"`
	CountrySpecificSettings map[string]interface{} `json:"country_specific_settings"`
	IsActive                bool                   `json:"is_active"`
}

type UpdateCountryConfigRequest struct {
	CountryName             string                  `json:"country_name,omitempty"`
	TaxSystem               string                  `json:"tax_system,omitempty"`
	DefaultTaxRate          float64                 `json:"default_tax_rate,omitempty" validate:"omitempty,min=0,max=1"`
	TaxInclusiveDefault     *bool                   `json:"tax_inclusive_default,omitempty"`
	InvoiceRequired         *bool                   `json:"invoice_required,omitempty"`
	InvoiceSequenceFormat   string                  `json:"invoice_sequence_format,omitempty"`
	CountrySpecificSettings *map[string]interface{} `json:"country_specific_settings,omitempty"`
	IsActive                *bool                   `json:"is_active,omitempty"`
}

type CountryConfigFilter struct {
	CountryCode     string `json:"country_code,omitempty"`
	TaxSystem       string `json:"tax_system,omitempty"`
	IsActive        *bool  `json:"is_active,omitempty"`
	InvoiceRequired *bool  `json:"invoice_required,omitempty"`
}
