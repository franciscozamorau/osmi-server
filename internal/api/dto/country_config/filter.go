// internal/api/dto/country_config/filter.go
package country_config

type CountryConfigFilter struct {
	CountryCode     string `json:"country_code,omitempty"`
	TaxSystem       string `json:"tax_system,omitempty"`
	IsActive        *bool  `json:"is_active,omitempty"`
	InvoiceRequired *bool  `json:"invoice_required,omitempty"`
}
