package entities

import (
	"time"
)

// CountryConfig configuración por país
// Mapea exactamente la tabla fiscal.country_config
type CountryConfig struct {
	ID          int64  `json:"id" db:"id"`
	CountryCode string `json:"country_code" db:"country_code"`
	CountryName string `json:"country_name" db:"country_name"`

	TaxSystem           *string  `json:"tax_system,omitempty" db:"tax_system"`
	DefaultTaxRate      *float64 `json:"default_tax_rate,omitempty" db:"default_tax_rate"`
	TaxInclusiveDefault bool     `json:"tax_inclusive_default" db:"tax_inclusive_default"`

	CountrySpecificSettings map[string]interface{} `json:"country_specific_settings" db:"country_specific_settings,type:jsonb"`

	// Campos específicos para México
	MxRFCValidationRegex *string  `json:"mx_rfc_validation_regex,omitempty" db:"mx_rfc_validation_regex"`
	MxCFDIRequired       bool     `json:"mx_cfdi_required" db:"mx_cfdi_required"`
	MxCFDIFormasPago     []string `json:"mx_cfdi_formas_pago,omitempty" db:"mx_cfdi_formas_pago,type:jsonb"`

	// Campos específicos para USA
	UsEINValidationRegex *string `json:"us_ein_validation_regex,omitempty" db:"us_ein_validation_regex"`
	UsRequires1099       bool    `json:"us_requires_1099" db:"us_requires_1099"`

	// Campos específicos para UE
	EuVATValidationRegex *string `json:"eu_vat_validation_regex,omitempty" db:"eu_vat_validation_regex"`
	EuVATReverseCharge   bool    `json:"eu_vat_reverse_charge" db:"eu_vat_reverse_charge"`

	// Configuración de facturación
	InvoiceRequired       bool    `json:"invoice_required" db:"invoice_required"`
	InvoiceSequenceFormat *string `json:"invoice_sequence_format,omitempty" db:"invoice_sequence_format"`

	IsActive bool `json:"is_active" db:"is_active"`

	// CORREGIDO: time.Time en lugar de string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MXCFDISettings configuración específica para México
type MXCFDISettings struct {
	RFCValidationRegex string   `json:"rfc_validation_regex"`
	CFDIRequired       bool     `json:"cfdi_required"`
	CFDIFormasPago     []string `json:"cfdi_formas_pago"`
}

// USSettings configuración específica para USA
type USSettings struct {
	EINValidationRegex string `json:"ein_validation_regex"`
	Requires1099       bool   `json:"requires_1099"`
}

// EUSettings configuración específica para UE
type EUSettings struct {
	VATValidationRegex string `json:"vat_validation_regex"`
	VATReverseCharge   bool   `json:"vat_reverse_charge"`
}

// Métodos de utilidad para CountryConfig

// GetMXSettings obtiene la configuración de México como un struct tipado
func (c *CountryConfig) GetMXSettings() MXCFDISettings {
	return MXCFDISettings{
		RFCValidationRegex: getStringValue(c.MxRFCValidationRegex),
		CFDIRequired:       c.MxCFDIRequired,
		CFDIFormasPago:     c.MxCFDIFormasPago,
	}
}

// GetUSSettings obtiene la configuración de USA como un struct tipado
func (c *CountryConfig) GetUSSettings() USSettings {
	return USSettings{
		EINValidationRegex: getStringValue(c.UsEINValidationRegex),
		Requires1099:       c.UsRequires1099,
	}
}

// GetEUSettings obtiene la configuración de la UE como un struct tipado
func (c *CountryConfig) GetEUSettings() EUSettings {
	return EUSettings{
		VATValidationRegex: getStringValue(c.EuVATValidationRegex),
		VATReverseCharge:   c.EuVATReverseCharge,
	}
}

// IsMexico verifica si la configuración es para México
func (c *CountryConfig) IsMexico() bool {
	return c.CountryCode == "MX"
}

// IsUSA verifica si la configuración es para USA
func (c *CountryConfig) IsUSA() bool {
	return c.CountryCode == "US"
}

// IsEU verifica si la configuración es para un país de la UE
func (c *CountryConfig) IsEU() bool {
	euCountries := map[string]bool{
		"AT": true, "BE": true, "BG": true, "CY": true, "CZ": true,
		"DE": true, "DK": true, "EE": true, "ES": true, "FI": true,
		"FR": true, "GR": true, "HR": true, "HU": true, "IE": true,
		"IT": true, "LT": true, "LU": true, "LV": true, "MT": true,
		"NL": true, "PL": true, "PT": true, "RO": true, "SE": true,
		"SI": true, "SK": true,
	}
	return euCountries[c.CountryCode]
}

// Helper function
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
