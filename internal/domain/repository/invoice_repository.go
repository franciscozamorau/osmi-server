package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// InvoiceRepository define operaciones para facturas
type InvoiceRepository interface {
	// CRUD básico
	Create(ctx context.Context, invoice *entities.Invoice) error
	FindByID(ctx context.Context, id int64) (*entities.Invoice, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Invoice, error)
	FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)
	FindByCFDIUUID(ctx context.Context, cfdiUUID string) (*entities.Invoice, error)
	Update(ctx context.Context, invoice *entities.Invoice) error
	Delete(ctx context.Context, id int64) error
	Void(ctx context.Context, invoiceID int64, reason string) error

	// Búsquedas
	List(ctx context.Context, filter dto.InvoiceFilter, pagination dto.Pagination) ([]*entities.Invoice, int64, error)
	FindByCustomer(ctx context.Context, customerID int64, pagination dto.Pagination) ([]*entities.Invoice, int64, error)
	FindByOrder(ctx context.Context, orderID int64) (*entities.Invoice, error)
	FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Invoice, int64, error)
	FindByDateRange(ctx context.Context, startDate, endDate string, pagination dto.Pagination) ([]*entities.Invoice, int64, error)
	FindUnpaid(ctx context.Context) ([]*entities.Invoice, error)
	FindOverdue(ctx context.Context) ([]*entities.Invoice, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, invoiceID int64, status string) error
	MarkAsPaid(ctx context.Context, invoiceID int64, paidAt string) error
	MarkAsSent(ctx context.Context, invoiceID int64, sentAt string) error
	UpdatePaymentStatus(ctx context.Context, invoiceID int64, paymentStatus string) error
	SetCFDIInfo(ctx context.Context, invoiceID int64, cfdiUUID, xml, sello, certificado, cadenaOriginal, qrCode string) error
	UpdateTaxBreakdown(ctx context.Context, invoiceID int64, taxBreakdown []map[string]interface{}) error
	UpdatePaymentBreakdown(ctx context.Context, invoiceID int64, paymentBreakdown []map[string]interface{}) error
	AddAttachment(ctx context.Context, invoiceID int64, attachmentURL, attachmentType string) error
	GenerateInvoiceNumber(ctx context.Context, series string) (string, error)

	// Generación de facturas
	GenerateFromOrder(ctx context.Context, orderID int64) (*entities.Invoice, error)
	Regenerate(ctx context.Context, invoiceID int64) (*entities.Invoice, error)
	CreateCreditNote(ctx context.Context, originalInvoiceID int64, reason string, amount float64) (*entities.Invoice, error)

	// Reportes
	GetMonthlyReport(ctx context.Context, year, month int) (*dto.MonthlyInvoiceReport, error)
	GetCustomerInvoiceHistory(ctx context.Context, customerID int64) ([]*dto.InvoiceHistory, error)
	GetTaxSummary(ctx context.Context, startDate, endDate string) (*dto.TaxSummary, error)

	// Estadísticas
	GetStats(ctx context.Context, filter dto.InvoiceFilter) (*dto.InvoiceStatsResponse, error)
	GetRevenueByPeriod(ctx context.Context, period string) ([]*dto.RevenueByPeriod, error)
	GetAverageInvoiceAmount(ctx context.Context) (float64, error)
	GetPaymentTermsStats(ctx context.Context) (*dto.PaymentTermsStats, error)
}
