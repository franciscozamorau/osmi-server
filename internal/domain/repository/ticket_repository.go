package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// TicketRepository define operaciones para tickets
type TicketRepository interface {
	// CRUD básico
	Create(ctx context.Context, ticket *entities.Ticket) error
	CreateBatch(ctx context.Context, tickets []*entities.Ticket) error
	FindByID(ctx context.Context, id int64) (*entities.Ticket, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Ticket, error)
	FindByCode(ctx context.Context, code string) (*entities.Ticket, error)
	Update(ctx context.Context, ticket *entities.Ticket) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter dto.TicketFilter, pagination dto.Pagination) ([]*entities.Ticket, int64, error)
	FindByEvent(ctx context.Context, eventID int64, filter dto.TicketFilter, pagination dto.Pagination) ([]*entities.Ticket, int64, error)
	FindByCustomer(ctx context.Context, customerID int64, filter dto.TicketFilter, pagination dto.Pagination) ([]*entities.Ticket, int64, error)
	FindByOrder(ctx context.Context, orderID int64) ([]*entities.Ticket, error)
	FindByTicketType(ctx context.Context, ticketTypeID int64) ([]*entities.Ticket, error)
	FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Ticket, int64, error)
	FindReservedExpired(ctx context.Context) ([]*entities.Ticket, error)
	FindCheckedIn(ctx context.Context, eventID int64) ([]*entities.Ticket, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, ticketID int64, status string) error
	CheckIn(ctx context.Context, ticketID int64, method, location string, validatorID *int64) error
	Transfer(ctx context.Context, ticketID int64, toCustomerID int64, transferToken string) error
	Cancel(ctx context.Context, ticketID int64) error
	Refund(ctx context.Context, ticketID int64) error
	Reserve(ctx context.Context, ticketID int64, customerID int64, expiresAt string) error
	ReleaseReservation(ctx context.Context, ticketID int64) error
	Validate(ctx context.Context, ticketCode, secretHash string) (*entities.Ticket, error)
	UpdateAttendeeInfo(ctx context.Context, ticketID int64, name, email, phone string) error

	// Estadísticas
	GetStats(ctx context.Context, eventID int64) (*dto.TicketStatsResponse, error)
	CountByStatus(ctx context.Context, eventID int64, status string) (int64, error)
	CountByTicketType(ctx context.Context, ticketTypeID int64) (int64, error)
	GetCheckInRate(ctx context.Context, eventID int64) (float64, error)
	GetValidationCount(ctx context.Context, ticketID int64) (int, error)
	GetRevenueByEvent(ctx context.Context, eventID int64) (float64, error)
}
