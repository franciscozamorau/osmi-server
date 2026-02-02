package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// OrderRepository define operaciones para órdenes
type OrderRepository interface {
	// CRUD básico
	Create(ctx context.Context, order *entities.Order) error
	FindByID(ctx context.Context, id int64) (*entities.Order, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Order, error)
	Update(ctx context.Context, order *entities.Order) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter dto.OrderFilter, pagination dto.Pagination) ([]*entities.Order, int64, error)
	FindByCustomer(ctx context.Context, customerID int64, pagination dto.Pagination) ([]*entities.Order, int64, error)
	FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Order, int64, error)
	FindByEvent(ctx context.Context, eventID int64, pagination dto.Pagination) ([]*entities.Order, int64, error)
	FindByPaymentProvider(ctx context.Context, providerID int64, pagination dto.Pagination) ([]*entities.Order, int64, error)
	FindExpiredReservations(ctx context.Context) ([]*entities.Order, error)
	Search(ctx context.Context, term string, filter dto.OrderFilter, pagination dto.Pagination) ([]*entities.Order, int64, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, orderID int64, status string) error
	MarkAsPaid(ctx context.Context, orderID int64, paymentID int64, paidAt string) error
	MarkAsCancelled(ctx context.Context, orderID int64, reason string) error
	MarkAsRefunded(ctx context.Context, orderID int64, refundID int64) error
	AddOrderItem(ctx context.Context, orderID int64, item *entities.OrderItem) error
	UpdateOrderItems(ctx context.Context, orderID int64, items []*entities.OrderItem) error
	CalculateTotals(ctx context.Context, orderID int64) (*dto.OrderTotals, error)
	ApplyPromotion(ctx context.Context, orderID int64, promotionCode string) error
	RemovePromotion(ctx context.Context, orderID int64) error
	GenerateInvoice(ctx context.Context, orderID int64) (string, error)
	CancelInvoice(ctx context.Context, orderID int64) error

	// Estadísticas
	GetStats(ctx context.Context, filter dto.OrderFilter) (*dto.OrderStatsResponse, error)
	GetCustomerOrderStats(ctx context.Context, customerID int64) (*dto.CustomerOrderStats, error)
	GetEventOrderStats(ctx context.Context, eventID int64) (*dto.EventOrderStats, error)
	GetDailyRevenue(ctx context.Context, days int) ([]*dto.DailyRevenue, error)
	GetAverageOrderValue(ctx context.Context) (float64, error)
	GetConversionRate(ctx context.Context) (float64, error)
}
