// internal/domain/repository/ticket_repository.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
)

// TicketFilter encapsula TODOS los criterios de búsqueda para tickets
type TicketFilter struct {
	// Filtros por ID
	IDs          []int64
	PublicIDs    []string
	EventID      *int64
	TicketTypeID *int64
	CustomerID   *int64
	OrderID      *int64

	// Filtros por código/estado
	Code   *string
	Status []enums.TicketStatus

	// Filtros de rango de fechas
	CreatedFrom   *time.Time
	CreatedTo     *time.Time
	SoldFrom      *time.Time
	SoldTo        *time.Time
	CheckedInFrom *time.Time
	CheckedInTo   *time.Time

	// Filtros específicos
	HasCheckedIn   *bool
	HasReservation *bool
	TransferToken  *string

	// Paginación y ordenamiento
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// TicketStats representa estadísticas de tickets para un evento
type TicketStats struct {
	TotalTickets     int64   `json:"total_tickets"`
	AvailableTickets int64   `json:"available_tickets"`
	ReservedTickets  int64   `json:"reserved_tickets"`
	SoldTickets      int64   `json:"sold_tickets"`
	CheckedInTickets int64   `json:"checked_in_tickets"`
	CancelledTickets int64   `json:"cancelled_tickets"`
	RefundedTickets  int64   `json:"refunded_tickets"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
}

// Errores específicos del repositorio
var (
	ErrTicketNotFound      = errors.New("ticket not found")
	ErrTicketAlreadyExists = errors.New("ticket already exists")
	ErrInvalidTicketStatus = errors.New("invalid ticket status transition")
	ErrTicketNotAvailable  = errors.New("ticket not available for this operation")
	ErrTicketDuplicateCode = errors.New("ticket code already exists")
)

type TicketRepository interface {
	// --- Operaciones de Escritura ---
	Create(ctx context.Context, ticket *entities.Ticket) error
	CreateBatch(ctx context.Context, tickets []*entities.Ticket) error
	Update(ctx context.Context, ticket *entities.Ticket) error
	Delete(ctx context.Context, id int64) error

	// --- Operaciones de Lectura (Flexibles) ---
	Find(ctx context.Context, filter *TicketFilter) ([]*entities.Ticket, int64, error)

	// Atajos
	GetByID(ctx context.Context, id int64) (*entities.Ticket, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.Ticket, error)
	GetByCode(ctx context.Context, code string) (*entities.Ticket, error)

	// --- Operaciones de Verificación ---
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)

	// --- Operaciones de Estado ---
	UpdateStatus(ctx context.Context, ticketID int64, status enums.TicketStatus) error
	CheckIn(ctx context.Context, ticketID int64, method, location string, checkedBy *int64) error
	Reserve(ctx context.Context, ticketID int64, reservedBy int64, expiresAt time.Time) error
	ReleaseReservation(ctx context.Context, ticketID int64) error
	Transfer(ctx context.Context, ticketID int64, toCustomerID int64, transferToken string) error
	Cancel(ctx context.Context, ticketID int64) error
	Refund(ctx context.Context, ticketID int64) error

	// --- Operaciones Específicas de Negocio ---
	ValidateTicket(ctx context.Context, code, secretHash string) (*entities.Ticket, error)
	GetEventStats(ctx context.Context, eventID int64) (*TicketStats, error)
	GetReservedExpired(ctx context.Context) ([]*entities.Ticket, error)
}
