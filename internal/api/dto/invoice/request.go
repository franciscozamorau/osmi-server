// internal/api/dto/invoice/request.go
package invoice

type CreateInvoiceRequest struct {
	OrderID             string                 `json:"order_id" validate:"required,uuid4"`
	CustomerID          string                 `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	InvoiceSeries       string                 `json:"invoice_series,omitempty"`
	InvoiceCurrency     string                 `json:"invoice_currency" validate:"required,oneof=MXN USD EUR"`
	CountrySpecificData map[string]interface{} `json:"country_specific_data,omitempty"`
}

type UpdateInvoiceRequest struct {
	Status           string                 `json:"status,omitempty" validate:"omitempty,oneof=draft issued cancelled paid"`
	PaymentStatus    string                 `json:"payment_status,omitempty" validate:"omitempty,oneof=pending paid partial cancelled"`
	TaxBreakdown     []TaxBreakdownItem     `json:"tax_breakdown,omitempty"`
	PaymentBreakdown []PaymentBreakdownItem `json:"payment_breakdown,omitempty"`
}

type GenerateCFDIRequest struct {
	InvoiceID      string                 `json:"invoice_id" validate:"required,uuid4"`
	PaymentMethod  string                 `json:"payment_method" validate:"required"`
	PaymentForm    string                 `json:"payment_form" validate:"required"`
	CFDIUse        string                 `json:"cfdi_use" validate:"required"`
	Exportation    string                 `json:"exportation,omitempty"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type TaxBreakdownItem struct {
	TaxType string  `json:"tax_type"`
	Rate    float64 `json:"rate"`
	Base    float64 `json:"base"`
	Amount  float64 `json:"amount"`
}

type PaymentBreakdownItem struct {
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	ExchangeRate  float64 `json:"exchange_rate,omitempty"`
}
