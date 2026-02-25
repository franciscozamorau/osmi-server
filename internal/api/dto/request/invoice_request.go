package request

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

type InvoiceFilter struct {
	OrderID       string  `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	CustomerID    string  `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	InvoiceNumber string  `json:"invoice_number,omitempty"`
	Status        string  `json:"status,omitempty"`
	PaymentStatus string  `json:"payment_status,omitempty"`
	DateFrom      string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo        string  `json:"date_to,omitempty" validate:"omitempty,date"`
	MinAmount     float64 `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount     float64 `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	HasCFDI       *bool   `json:"has_cfdi,omitempty"`
	TaxID         string  `json:"tax_id,omitempty"`
}
