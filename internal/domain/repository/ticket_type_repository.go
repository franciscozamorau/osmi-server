package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// TicketTypeRepository define operaciones para tipos de ticket
type TicketTypeRepository interface {
	// CRUD básico
	Create(ctx context.Context, ticketType *entities.TicketType) error
	FindByID(ctx context.Context, id int64) (*entities.TicketType, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.TicketType, error)
	Update(ctx context.Context, ticketType *entities.TicketType) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// Búsquedas
	List(ctx context.Context, filter dto.TicketTypeFilter, pagination dto.Pagination) ([]*entities.TicketType, int64, error)
	FindByEvent(ctx context.Context, eventID int64, activeOnly bool) ([]*entities.TicketType, error)
	FindByEventPublicID(ctx context.Context, eventPublicID string) ([]*entities.TicketType, error)
	FindAvailable(ctx context.Context, eventID int64) ([]*entities.TicketType, error)
	FindSoldOut(ctx context.Context, eventID int64) ([]*entities.TicketType, error)

	// Operaciones específicas
	UpdateQuantity(ctx context.Context, ticketTypeID int64, quantity int) error
	ReserveTickets(ctx context.Context, ticketTypeID int64, quantity int) error
	ReleaseReservation(ctx context.Context, ticketTypeID int64, quantity int) error
	SellTickets(ctx context.Context, ticketTypeID int64, quantity int) error
	CancelSoldTickets(ctx context.Context, ticketTypeID int64, quantity int) error
	RefundTickets(ctx context.Context, ticketTypeID int64, quantity int) error
	CheckAvailability(ctx context.Context, ticketTypeID int64, quantity int) (bool, error)
	GetAvailableQuantity(ctx context.Context, ticketTypeID int64) (int, error)
	UpdateSaleDates(ctx context.Context, ticketTypeID int64, startsAt, endsAt string) error
	UpdatePrice(ctx context.Context, ticketTypeID int64, price float64, currency string) error
	UpdateStatus(ctx context.Context, ticketTypeID int64, active bool) error

	// Estadísticas
	GetStats(ctx context.Context, ticketTypeID int64) (*dto.TicketTypeStatsResponse, error)
	GetEventTicketStats(ctx context.Context, eventID int64) (*dto.EventTicketStats, error)
	CountSold(ctx context.Context, ticketTypeID int64) (int, error)
	CountReserved(ctx context.Context, ticketTypeID int64) (int, error)
	GetRevenue(ctx context.Context, ticketTypeID int64) (float64, error)
	GetSalesVelocity(ctx context.Context, ticketTypeID int64) (float64, error)
}
