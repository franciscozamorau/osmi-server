package entities

import (
	"encoding/json"
	"time"
)

// Invoice representa una factura en el sistema fiscal
// Mapea exactamente la tabla fiscal.invoices
type Invoice struct {
	ID int64 `json:"id" db:"id"`
	// CORREGIDO: El campo en la BD se llama invoice_uuid
	InvoiceUUID string `json:"invoice_uuid" db:"invoice_uuid"`

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

	// CORREGIDO: country_specific_data es JSONB
	CountrySpecificData *map[string]interface{} `json:"country_specific_data,omitempty" db:"country_specific_data,type:jsonb"`

	// CORREGIDO: Los campos MX tienen prefijo mx_ en la BD
	CFDIUUID           *string `json:"cfdi_uuid,omitempty" db:"mx_cfdi_uuid"`
	CFDIXML            *string `json:"cfdi_xml,omitempty" db:"mx_cfdi_xml"`
	CFDISello          *string `json:"cfdi_sello,omitempty" db:"mx_cfdi_sello"`
	CFDICertificado    *string `json:"cfdi_certificado,omitempty" db:"mx_cfdi_certificado"`
	CFDICadenaOriginal *string `json:"cfdi_cadena_original,omitempty" db:"mx_cfdi_cadena_original"`
	CFDIQRCode         *string `json:"cfdi_qr_code,omitempty" db:"mx_cfdi_qr_code"`

	// CORREGIDO: tax_breakdown y payment_breakdown son JSONB
	TaxBreakdown     *[]TaxBreakdownItem     `json:"tax_breakdown,omitempty" db:"tax_breakdown,type:jsonb"`
	PaymentBreakdown *[]PaymentBreakdownItem `json:"payment_breakdown,omitempty" db:"payment_breakdown,type:jsonb"`

	IssuedAt    *time.Time `json:"issued_at,omitempty" db:"issued_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// TaxBreakdownItem representa un item del desglose de impuestos
type TaxBreakdownItem struct {
	TaxType    string  `json:"tax_type"`    // e.g., "VAT", "IVA", "GST"
	TaxRate    float64 `json:"tax_rate"`    // e.g., 0.16 for 16%
	Taxable    float64 `json:"taxable"`     // Base imponible
	TaxAmount  float64 `json:"tax_amount"`  // Monto del impuesto
	IsWithheld bool    `json:"is_withheld"` // Si es retención
}

// PaymentBreakdownItem representa un item del desglose de pagos
type PaymentBreakdownItem struct {
	PaymentMethod string    `json:"payment_method"`
	Amount        float64   `json:"amount"`
	Reference     *string   `json:"reference,omitempty"`
	PaidAt        time.Time `json:"paid_at"`
}

// Métodos de utilidad para Invoice

// IsIssued verifica si la factura ha sido emitida
func (i *Invoice) IsIssued() bool {
	return i.IssuedAt != nil
}

// IsCancelled verifica si la factura ha sido cancelada
func (i *Invoice) IsCancelled() bool {
	return i.CancelledAt != nil
}

// IsPaid verifica si la factura ha sido pagada
func (i *Invoice) IsPaid() bool {
	return i.PaidAt != nil || i.PaymentStatus == "paid" || i.PaymentStatus == "completed"
}

// IsDraft verifica si la factura está en borrador
func (i *Invoice) IsDraft() bool {
	return i.Status == "draft"
}

// IsFinal verifica si la factura es definitiva (emitida y no cancelada)
func (i *Invoice) IsFinal() bool {
	return i.IsIssued() && !i.IsCancelled()
}

// GetOutstandingAmount obtiene el monto pendiente de pago
func (i *Invoice) GetOutstandingAmount() float64 {
	if i.IsPaid() {
		return 0
	}

	// Si hay breakdown de pagos, calcular pagado
	if i.PaymentBreakdown != nil {
		var paidAmount float64
		for _, payment := range *i.PaymentBreakdown {
			paidAmount += payment.Amount
		}
		if paidAmount >= i.TotalAmount {
			return 0
		}
		return i.TotalAmount - paidAmount
	}

	return i.TotalAmount
}

// AddTaxBreakdownItem añade un item al desglose de impuestos
func (i *Invoice) AddTaxBreakdownItem(item TaxBreakdownItem) {
	if i.TaxBreakdown == nil {
		i.TaxBreakdown = &[]TaxBreakdownItem{}
	}

	*i.TaxBreakdown = append(*i.TaxBreakdown, item)

	// Recalcular tax amount basado en el breakdown
	var totalTax float64
	for _, tax := range *i.TaxBreakdown {
		totalTax += tax.TaxAmount
	}
	i.TaxAmount = totalTax
	i.TotalAmount = i.Subtotal + i.TaxAmount
}

// AddPaymentBreakdownItem añade un item al desglose de pagos
func (i *Invoice) AddPaymentBreakdownItem(item PaymentBreakdownItem) {
	if i.PaymentBreakdown == nil {
		i.PaymentBreakdown = &[]PaymentBreakdownItem{}
	}

	*i.PaymentBreakdown = append(*i.PaymentBreakdown, item)

	// Verificar si está completamente pagado
	var totalPaid float64
	for _, payment := range *i.PaymentBreakdown {
		totalPaid += payment.Amount
	}

	if totalPaid >= i.TotalAmount {
		now := time.Now()
		i.PaidAt = &now
		i.PaymentStatus = "paid"
	}
}

// IsMexicanCFDI verifica si es una factura mexicana con CFDI
func (i *Invoice) IsMexicanCFDI() bool {
	return i.CFDIUUID != nil
}

// GetCFDIStatus obtiene el estado del CFDI
func (i *Invoice) GetCFDIStatus() string {
	if i.CFDIUUID == nil {
		return "not_applicable"
	}
	if i.CancelledAt != nil {
		return "cancelled"
	}
	if i.IssuedAt != nil && i.CFDIXML != nil {
		return "issued"
	}
	return "pending"
}

// SetCountrySpecificData establece datos específicos por país
func (i *Invoice) SetCountrySpecificData(data map[string]interface{}) {
	i.CountrySpecificData = &data
}

// GetCountrySpecificData obtiene datos específicos por país
func (i *Invoice) GetCountrySpecificData() map[string]interface{} {
	if i.CountrySpecificData == nil {
		return make(map[string]interface{})
	}
	return *i.CountrySpecificData
}

// MarshalJSON implementa la interfaz json.Marshaler para serialización personalizada
func (i *Invoice) MarshalJSON() ([]byte, error) {
	type Alias Invoice
	return json.Marshal(&struct {
		*Alias
		OutstandingAmount float64 `json:"outstanding_amount"`
		IsPaid            bool    `json:"is_paid"`
		IsIssued          bool    `json:"is_issued"`
		IsCancelled       bool    `json:"is_cancelled"`
		CFDIStatus        string  `json:"cfdi_status,omitempty"`
	}{
		Alias:             (*Alias)(i),
		OutstandingAmount: i.GetOutstandingAmount(),
		IsPaid:            i.IsPaid(),
		IsIssued:          i.IsIssued(),
		IsCancelled:       i.IsCancelled(),
		CFDIStatus:        i.GetCFDIStatus(),
	})
}
