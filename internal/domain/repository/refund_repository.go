package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// RefundRepository define operaciones para reembolsos
type RefundRepository interface {
	// CRUD básico
	Create(ctx context.Context, refund *entities.Refund) error
	FindByID(ctx context.Context, id int64) (*entities.Refund, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Refund, error)
	FindByProviderRefundID(ctx context.Context, providerRefundID string) (*entities.Refund, error)
	Update(ctx context.Context, refund *entities.Refund) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter dto.RefundFilter, pagination dto.Pagination) ([]*entities.Refund, int64, error)
	FindByOrder(ctx context.Context, orderID int64) ([]*entities.Refund, error)
	FindByPayment(ctx context.Context, paymentID int64) ([]*entities.Refund, error)
	FindByCustomer(ctx context.Context, customerID int64, pagination dto.Pagination) ([]*entities.Refund, int64, error)
	FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Refund, int64, error)
	FindByRequester(ctx context.Context, requesterID int64, pagination dto.Pagination) ([]*entities.Refund, int64, error)
	FindByApprover(ctx context.Context, approverID int64, pagination dto.Pagination) ([]*entities.Refund, int64, error)
	FindPendingRefunds(ctx context.Context) ([]*entities.Refund, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, refundID int64, status string, providerData map[string]interface{}) error
	MarkAsProcessed(ctx context.Context, refundID int64, processedAt string) error
	MarkAsCompleted(ctx context.Context, refundID int64, completedAt string) error
	Approve(ctx context.Context, refundID int64, approverID int64) error
	Reject(ctx context.Context, refundID int64, reason string) error
	SetProviderRefundID(ctx context.Context, refundID int64, providerRefundID string) error
	UpdateAmount(ctx context.Context, refundID int64, amount float64, currency string) error
	AddNote(ctx context.Context, refundID int64, note string) error

	// Validaciones
	CanRefundOrder(ctx context.Context, orderID int64) (bool, error)
	CalculateRefundableAmount(ctx context.Context, orderID int64) (float64, error)
	IsRefundWithinPolicy(ctx context.Context, orderID int64, refundAmount float64) (bool, error)
	HasPreviousRefunds(ctx context.Context, orderID int64) (bool, error)

	// Estadísticas
	GetStats(ctx context.Context, filter dto.RefundFilter) (*dto.RefundStatsResponse, error)
	GetRefundRate(ctx context.Context, eventID *int64) (float64, error)
	GetAverageRefundAmount(ctx context.Context) (float64, error)
	GetRefundReasons(ctx context.Context, limit int) ([]*dto.RefundReasonStats, error)
	GetProcessingTimeStats(ctx context.Context) (*dto.ProcessingTimeStats, error)
}
