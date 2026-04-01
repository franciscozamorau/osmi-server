// internal/domain/repository/payment_repository.go
package repository

import (
	"context"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	paymentdto "github.com/franciscozamorau/osmi-server/internal/api/dto/payment"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// PaymentRepository define operaciones para pagos
type PaymentRepository interface {
	// CRUD básico
	Create(ctx context.Context, payment *entities.Payment) error
	FindByID(ctx context.Context, id int64) (*entities.Payment, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Payment, error)
	FindByTransactionID(ctx context.Context, transactionID string) (*entities.Payment, error)
	Update(ctx context.Context, payment *entities.Payment) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter paymentdto.PaymentFilter, pagination commondto.Pagination) ([]*entities.Payment, int64, error)
	FindByOrder(ctx context.Context, orderID int64) ([]*entities.Payment, error)
	FindByCustomer(ctx context.Context, customerID int64, pagination commondto.Pagination) ([]*entities.Payment, int64, error)
	FindByStatus(ctx context.Context, status string, pagination commondto.Pagination) ([]*entities.Payment, int64, error)
	FindByProvider(ctx context.Context, providerID int64, pagination commondto.Pagination) ([]*entities.Payment, int64, error)
	FindFailedPayments(ctx context.Context, hours int) ([]*entities.Payment, error)
	FindPendingPayments(ctx context.Context) ([]*entities.Payment, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, paymentID int64, status string, providerData map[string]interface{}) error
	MarkAsProcessed(ctx context.Context, paymentID int64, processedAt string) error
	MarkAsRefunded(ctx context.Context, paymentID int64, refundID int64) error
	MarkAsFailed(ctx context.Context, paymentID int64, errorMessage string, errorCode string) error
	IncrementAttempts(ctx context.Context, paymentID int64) error
	SetNextRetry(ctx context.Context, paymentID int64, nextRetryAt string) error
	RecordProviderResponse(ctx context.Context, paymentID int64, response map[string]interface{}) error
	UpdatePaymentMethod(ctx context.Context, paymentID int64, method string, details map[string]interface{}) error

	// Estadísticas
	GetStats(ctx context.Context, filter paymentdto.PaymentFilter) (*paymentdto.PaymentStatsResponse, error)
	GetProviderStats(ctx context.Context, providerID int64) (*paymentdto.ProviderStats, error)
	GetDailyPaymentVolume(ctx context.Context, days int) ([]*paymentdto.DailyVolume, error)
	GetSuccessRate(ctx context.Context, providerID *int64) (float64, error)
	GetAverageProcessingTime(ctx context.Context) (float64, error)
	GetTotalProcessedAmount(ctx context.Context, currency string) (float64, error)
}
