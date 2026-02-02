package dto

import "time"

type InvoiceResponse struct {
	ID              string    `json:"id"`
	InvoiceNumber   string    `json:"invoice_number"`
	InvoiceSeries   string    `json:"invoice_series,omitempty"`
	InvoiceDate     time.Time `json:"invoice_date"`
	InvoiceCurrency string    `json:"invoice_currency"`

	Order    OrderBasicInfo    `json:"order,omitempty"`
	Customer CustomerBasicInfo `json:"customer"`

	Subtotal    float64 `json:"subtotal"`
	TaxAmount   float64 `json:"tax_amount"`
	TotalAmount float64 `json:"total_amount"`

	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`

	CountrySpecificData map[string]interface{} `json:"country_specific_data,omitempty"`

	// CFDI fields (México)
	CFDIUUID           string `json:"cfdi_uuid,omitempty"`
	CFDIXML            string `json:"cfdi_xml,omitempty"`
	CFDISello          string `json:"cfdi_sello,omitempty"`
	CFDICertificado    string `json:"cfdi_certificado,omitempty"`
	CFDICadenaOriginal string `json:"cfdi_cadena_original,omitempty"`
	CFDIQRCode         string `json:"cfdi_qr_code,omitempty"`

	TaxBreakdown     []TaxBreakdownItem     `json:"tax_breakdown"`
	PaymentBreakdown []PaymentBreakdownItem `json:"payment_breakdown"`

	IssuedAt    time.Time `json:"issued_at,omitempty"`
	CancelledAt time.Time `json:"cancelled_at,omitempty"`
	PaidAt      time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderBasicInfo struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	CreatedAt   time.Time `json:"created_at"`
	TotalAmount float64   `json:"total_amount"`
}

type CustomerBasicInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	TaxID   string `json:"tax_id,omitempty"`
	TaxName string `json:"tax_name,omitempty"`
	Address string `json:"address,omitempty"`
	Country string `json:"country"`
}

type InvoiceStatsResponse struct {
	TotalInvoices     int64   `json:"total_invoices"`
	DraftInvoices     int64   `json:"draft_invoices"`
	IssuedInvoices    int64   `json:"issued_invoices"`
	PaidInvoices      int64   `json:"paid_invoices"`
	CancelledInvoices int64   `json:"cancelled_invoices"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalTax          float64 `json:"total_tax"`
	AvgInvoiceAmount  float64 `json:"avg_invoice_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
}

type TaxSummary struct {
	CountryCode  string  `json:"country_code"`
	CountryName  string  `json:"country_name"`
	TaxType      string  `json:"tax_type"`
	TaxRate      float64 `json:"tax_rate"`
	TotalBase    float64 `json:"total_base"`
	TotalTax     float64 `json:"total_tax"`
	InvoiceCount int64   `json:"invoice_count"`
}

type InvoiceListResponse struct {
	Invoices   []InvoiceResponse    `json:"invoices"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
	Stats      InvoiceStatsResponse `json:"stats"`
}
