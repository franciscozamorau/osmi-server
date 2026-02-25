package response

import "time"

type CountryConfigResponse struct {
	ID                      int32                  `json:"id"`
	CountryCode             string                 `json:"country_code"`
	CountryName             string                 `json:"country_name"`
	TaxSystem               string                 `json:"tax_system"`
	DefaultTaxRate          float64                `json:"default_tax_rate"`
	TaxInclusiveDefault     bool                   `json:"tax_inclusive_default"`
	InvoiceRequired         bool                   `json:"invoice_required"`
	InvoiceSequenceFormat   string                 `json:"invoice_sequence_format,omitempty"`
	CountrySpecificSettings map[string]interface{} `json:"country_specific_settings"`
	IsActive                bool                   `json:"is_active"`
	CreatedAt               time.Time              `json:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at"`

	// Configuración específica por país
	MXConfig *MXCountryConfig `json:"mx_config,omitempty"`
	USConfig *USCountryConfig `json:"us_config,omitempty"`
	EUConfig *EUCountryConfig `json:"eu_config,omitempty"`
}

type MXCountryConfig struct {
	RFCValidationRegex     string   `json:"rfc_validation_regex,omitempty"`
	CFDIRequired           bool     `json:"cfdi_required"`
	CFDIFormasPago         []string `json:"cfdi_formas_pago"`
	RequiresSATCertificate bool     `json:"requires_sat_certificate"`
}

type USCountryConfig struct {
	EINValidationRegex string     `json:"ein_validation_regex,omitempty"`
	Requires1099       bool       `json:"requires_1099"`
	StateTaxes         []StateTax `json:"state_taxes,omitempty"`
}

type EUCountryConfig struct {
	VATValidationRegex string             `json:"vat_validation_regex,omitempty"`
	VATReverseCharge   bool               `json:"vat_reverse_charge"`
	VATRates           map[string]float64 `json:"vat_rates,omitempty"`
}

type StateTax struct {
	StateCode string  `json:"state_code"`
	StateName string  `json:"state_name"`
	TaxRate   float64 `json:"tax_rate"`
}

type CountryConfigListResponse struct {
	Configs    []CountryConfigResponse `json:"configs"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

type TaxSummaryResponse struct {
	CountryCode  string  `json:"country_code"`
	CountryName  string  `json:"country_name"`
	TotalTax     float64 `json:"total_tax"`
	TotalRevenue float64 `json:"total_revenue"`
	TaxRate      float64 `json:"tax_rate"`
	InvoiceCount int64   `json:"invoice_count"`
}
