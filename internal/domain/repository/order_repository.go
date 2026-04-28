// internal/domain/repository/order_repository.go
package repository

import (
	"context"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	orderdto "github.com/franciscozamorau/osmi-server/internal/api/dto/order"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/jackc/pgx/v5"
)

// OrderRepository define operaciones para órdenes
type OrderRepository interface {
	// CRUD básico
	Create(ctx context.Context, order *entities.Order) error
	FindByID(ctx context.Context, id int64) (*entities.Order, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.Order, error)
	GetByCustomerID(ctx context.Context, customerID int64) ([]*entities.Order, error)
	AddItem(ctx context.Context, item *entities.OrderItem) error
	GetItems(ctx context.Context, orderID int64) ([]*entities.OrderItem, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Order, error)
	Update(ctx context.Context, order *entities.Order) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter orderdto.OrderFilter, pagination commondto.Pagination) ([]*entities.Order, int64, error)
	FindByCustomer(ctx context.Context, customerID int64, pagination commondto.Pagination) ([]*entities.Order, int64, error)
	FindByStatus(ctx context.Context, status string, pagination commondto.Pagination) ([]*entities.Order, int64, error)
	FindByEvent(ctx context.Context, eventID int64, pagination commondto.Pagination) ([]*entities.Order, int64, error)
	FindByPaymentProvider(ctx context.Context, providerID int64, pagination commondto.Pagination) ([]*entities.Order, int64, error)
	FindExpiredReservations(ctx context.Context) ([]*entities.Order, error)
	Search(ctx context.Context, term string, filter orderdto.OrderFilter, pagination commondto.Pagination) ([]*entities.Order, int64, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, orderID int64, status string) error
	MarkAsPaid(ctx context.Context, orderID int64, paymentID int64, paidAt string) error
	MarkAsCancelled(ctx context.Context, orderID int64, reason string) error
	MarkAsRefunded(ctx context.Context, orderID int64, refundID int64) error
	AddOrderItem(ctx context.Context, orderID int64, item *entities.OrderItem) error
	UpdateOrderItems(ctx context.Context, orderID int64, items []*entities.OrderItem) error
	CalculateTotals(ctx context.Context, orderID int64) (*orderdto.OrderTotals, error)
	ApplyPromotion(ctx context.Context, orderID int64, promotionCode string) error
	RemovePromotion(ctx context.Context, orderID int64) error
	GenerateInvoice(ctx context.Context, orderID int64) (string, error)
	CancelInvoice(ctx context.Context, orderID int64) error

	// Estadísticas
	GetStats(ctx context.Context, filter orderdto.OrderFilter) (*orderdto.OrderStatsResponse, error)
	GetCustomerOrderStats(ctx context.Context, customerID int64) (*orderdto.CustomerOrderStats, error)
	GetEventOrderStats(ctx context.Context, eventID int64) (*orderdto.EventOrderStats, error)
	GetDailyRevenue(ctx context.Context, days int) ([]*orderdto.DailyRevenue, error)
	GetAverageOrderValue(ctx context.Context) (float64, error)
	GetConversionRate(ctx context.Context) (float64, error)

	FindByPublicIDForUpdate(ctx context.Context, tx pgx.Tx, publicID string) (*entities.Order, error)
}
