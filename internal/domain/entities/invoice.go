package entities

import "time"

type Invoice struct {
	ID        int64  `json:"id" db:"id"`
	InvoiceID string `json:"invoice_id" db:"invoice_uuid"`

	OrderID    *int64 `json:"order_id,omitempty" db:"order_id"`
	CustomerID *int64 `json:"customer_id,omitempty" db:"customer_id"`

	InvoiceNumber   string    `json:"invoice_number" db:"invoice_number"`
	InvoiceSeries   *string   `json:"invoice_series,omitempty" db:"invoice_series"`
	InvoiceDate     time.Time `json:"invoice_date" db:"invoice_date"`
	InvoiceCurrency string    `json:"invoice_currency" db:"invoice_currency"`

	Subtotal    float64 `json:"subtotal" db:"subtotal"`
	TaxAmount   float64 `json:"tax_amount" db:"tax_amount"`
	TotalAmount float64 `json:"total_amount" db:"total_amount"`

	Status        string `json:"status" db:"status"`
	PaymentStatus string `json:"payment_status" db:"payment_status"`

	CountrySpecificData *string `json:"country_specific_data,omitempty" db:"country_specific_data"`

	CFDIUUID           *string `json:"cfdi_uuid,omitempty" db:"mx_cfdi_uuid"`
	CFDIXML            *string `json:"cfdi_xml,omitempty" db:"mx_cfdi_xml"`
	CFDISello          *string `json:"cfdi_sello,omitempty" db:"mx_cfdi_sello"`
	CFDICertificado    *string `json:"cfdi_certificado,omitempty" db:"mx_cfdi_certificado"`
	CFDICadenaOriginal *string `json:"cfdi_cadena_original,omitempty" db:"mx_cfdi_cadena_original"`
	CFDIQRCode         *string `json:"cfdi_qr_code,omitempty" db:"mx_cfdi_qr_code"`

	TaxBreakdown     *string `json:"tax_breakdown,omitempty" db:"tax_breakdown"`
	PaymentBreakdown *string `json:"payment_breakdown,omitempty" db:"payment_breakdown"`

	IssuedAt    *time.Time `json:"issued_at,omitempty" db:"issued_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
