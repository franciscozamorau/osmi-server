// internal/application/services/ticket_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	ticketdto "github.com/franciscozamorau/osmi-server/internal/api/dto/ticket"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

type TicketService struct {
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	eventRepo      repository.EventRepository
	customerRepo   repository.CustomerRepository
	orderRepo      repository.OrderRepository
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	eventRepo repository.EventRepository,
	customerRepo repository.CustomerRepository,
	orderRepo repository.OrderRepository,
) *TicketService {
	return &TicketService{
		ticketRepo:     ticketRepo,
		ticketTypeRepo: ticketTypeRepo,
		eventRepo:      eventRepo,
		customerRepo:   customerRepo,
		orderRepo:      orderRepo,
	}
}

// CreateTicket crea un nuevo ticket vendido (flujo directo - temporal)
func (s *TicketService) CreateTicket(ctx context.Context, req *ticketdto.CreateTicketRequest) (*entities.Ticket, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, req.TicketTypeID)
	if err != nil {
		return nil, fmt.Errorf("ticket type not found: %w", err)
	}

	available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, int(req.Quantity))
	if err != nil {
		return nil, fmt.Errorf("error checking availability: %w", err)
	}
	if !available {
		return nil, errors.New("ticket type not available")
	}

	customer, err := s.customerRepo.GetByPublicID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	event, err := s.eventRepo.GetByID(ctx, ticketType.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Status != string(enums.EventStatusPublished) && event.Status != string(enums.EventStatusLive) {
		return nil, errors.New("event is not active for ticket sales")
	}

	finalPrice := ticketType.GetFinalPrice()
	taxAmount := ticketType.BasePrice * ticketType.TaxRate

	now := time.Now()
	ticket := &entities.Ticket{
		PublicID:      uuid.New().String(),
		TicketTypeID:  ticketType.ID,
		EventID:       event.ID,
		CustomerID:    &customer.ID,
		Code:          s.generateTicketCode(event.ID, ticketType.ID, 0),
		SecretHash:    uuid.New().String(),
		Status:        string(enums.TicketStatusSold),
		FinalPrice:    finalPrice,
		Currency:      ticketType.Currency,
		TaxAmount:     taxAmount,
		AttendeeName:  nil,
		AttendeeEmail: nil,
		AttendeePhone: nil,
		SoldAt:        &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := ticket.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ticket: %w", err)
	}

	err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	err = s.ticketTypeRepo.SellTicketsDirect(ctx, ticketType.ID, int(req.Quantity))
	if err != nil {
		_ = s.ticketRepo.Delete(ctx, ticket.ID)
		return nil, fmt.Errorf("failed to update ticket type sales: %w", err)
	}

	go s.customerRepo.UpdateStats(ctx, customer.ID, finalPrice)

	return ticket, nil
}

// ReserveTicket reserva un ticket - CORREGIDO (usando los campos del DTO original)
func (s *TicketService) ReserveTicket(ctx context.Context, req *ticketdto.ReserveTicketRequest) (*entities.Ticket, error) {
	// 🔥 Usar TicketID (que es el ID del ticket type)
	if req.TicketID == "" {
		return nil, errors.New("ticket_type_id is required")
	}

	// 🔥 Por ahora, reservamos 1 ticket por llamada (quantity = 1)
	// Si necesitas quantity, habrá que agregarlo al DTO y al handler
	quantity := 1

	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("ticket type not found: %w", err)
	}

	available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, quantity)
	if err != nil {
		return nil, fmt.Errorf("error checking availability: %w", err)
	}
	if !available {
		return nil, errors.New("ticket type not available")
	}

	event, err := s.eventRepo.GetByID(ctx, ticketType.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if !event.AllowReservations {
		return nil, errors.New("event does not allow reservations")
	}

	// Calcular expiración
	duration := req.ExpiresAt.Sub(time.Now())
	if duration <= 0 {
		reservationDuration := time.Duration(event.ReservationDuration) * time.Minute
		if reservationDuration <= 0 {
			reservationDuration = 15 * time.Minute
		}
		duration = reservationDuration
	}
	reservationExpiresAt := time.Now().Add(duration)

	now := time.Now()

	// Crear ticket reservado
	ticket := &entities.Ticket{
		PublicID:             uuid.New().String(),
		TicketTypeID:         ticketType.ID,
		EventID:              event.ID,
		CustomerID:           nil,
		Code:                 s.generateTicketCode(event.ID, ticketType.ID, 0),
		SecretHash:           uuid.New().String(),
		Status:               string(enums.TicketStatusReserved),
		FinalPrice:           ticketType.GetFinalPrice(),
		Currency:             ticketType.Currency,
		TaxAmount:            ticketType.BasePrice * ticketType.TaxRate,
		ReservedAt:           &now,
		ReservationExpiresAt: &reservationExpiresAt,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := ticket.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ticket: %w", err)
	}

	err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	err = s.ticketTypeRepo.ReserveTickets(ctx, ticketType.ID, quantity)
	if err != nil {
		_ = s.ticketRepo.Delete(ctx, ticket.ID)
		return nil, fmt.Errorf("failed to reserve tickets: %w", err)
	}

	return ticket, nil
}

// CheckInTicket marca un ticket como usado
func (s *TicketService) CheckInTicket(ctx context.Context, req *ticketdto.CheckInTicketRequest) (*entities.Ticket, error) {
	if req.TicketID == "" {
		return nil, errors.New("ticket_id is required")
	}

	ticket, err := s.ticketRepo.GetByPublicID(ctx, req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	if ticket.Status != string(enums.TicketStatusSold) {
		return nil, errors.New("ticket is not valid for check-in")
	}

	if ticket.CheckedInAt != nil {
		return nil, errors.New("ticket already checked in")
	}

	event, err := s.eventRepo.GetByID(ctx, ticket.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	now := time.Now()
	if now.Before(event.StartsAt.Add(-1 * time.Hour)) {
		return nil, errors.New("check-in not available yet")
	}
	if now.After(event.EndsAt.Add(2 * time.Hour)) {
		return nil, errors.New("check-in period has ended")
	}

	var validatorID *int64
	if req.CheckedBy != "" {
		// TODO: Validar validador cuando exista auth
	}

	err = s.ticketRepo.CheckIn(ctx, ticket.ID, req.Method, req.Location, validatorID)
	if err != nil {
		return nil, fmt.Errorf("check-in failed: %w", err)
	}

	updatedTicket, err := s.ticketRepo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("ticket checked in but retrieval failed: %w", err)
	}

	return updatedTicket, nil
}

// TransferTicket transfiere un ticket
func (s *TicketService) TransferTicket(ctx context.Context, req *ticketdto.TransferTicketRequest) (*entities.Ticket, error) {
	if req.TicketID == "" {
		return nil, errors.New("ticket_id is required")
	}
	if req.ToCustomerID == "" {
		return nil, errors.New("to_customer_id is required")
	}

	ticket, err := s.ticketRepo.GetByPublicID(ctx, req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	// Si se proporciona from_customer_id, validar ownership
	if req.FromCustomerID != "" {
		fromCustomer, err := s.customerRepo.GetByPublicID(ctx, req.FromCustomerID)
		if err != nil {
			return nil, fmt.Errorf("sender customer not found: %w", err)
		}
		if ticket.CustomerID == nil || *ticket.CustomerID != fromCustomer.ID {
			return nil, errors.New("ticket does not belong to sender")
		}
	}

	if !ticket.CanBeTransferred() {
		return nil, errors.New("ticket cannot be transferred")
	}

	toCustomer, err := s.customerRepo.GetByPublicID(ctx, req.ToCustomerID)
	if err != nil {
		return nil, fmt.Errorf("recipient customer not found: %w", err)
	}

	transferToken := req.Token
	if transferToken == "" {
		transferToken = uuid.New().String()
	}

	err = s.ticketRepo.Transfer(ctx, ticket.ID, toCustomer.ID, transferToken)
	if err != nil {
		return nil, fmt.Errorf("transfer failed: %w", err)
	}

	updatedTicket, err := s.ticketRepo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("ticket transferred but retrieval failed: %w", err)
	}

	return updatedTicket, nil
}

// GetTicket obtiene un ticket por su ID público
func (s *TicketService) GetTicket(ctx context.Context, ticketID string) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.GetByPublicID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}
	return ticket, nil
}

// GetTicketByCode obtiene un ticket por su código
func (s *TicketService) GetTicketByCode(ctx context.Context, code string) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("ticket not found with code %s: %w", code, err)
	}
	return ticket, nil
}

// ListTickets lista tickets con filtros y paginación
func (s *TicketService) ListTickets(ctx context.Context, filter *ticketdto.TicketFilter, pagination commondto.Pagination) ([]*entities.Ticket, int64, error) {
	repoFilter := &repository.TicketFilter{
		Limit:  pagination.PageSize,
		Offset: (pagination.Page - 1) * pagination.PageSize,
	}

	if filter != nil {
		if filter.EventID != nil {
			repoFilter.EventID = filter.EventID
		}
		if filter.CustomerID != nil {
			repoFilter.CustomerID = filter.CustomerID
		}
		if filter.OrderID != nil {
			repoFilter.OrderID = filter.OrderID
		}
		if filter.Status != "" {
			status := enums.TicketStatus(filter.Status)
			if status.IsValid() {
				repoFilter.Status = []enums.TicketStatus{status}
			}
		}
		if filter.TicketTypeID != nil {
			repoFilter.TicketTypeID = filter.TicketTypeID
		}
		if filter.DateFrom != "" {
			if t, err := time.Parse(time.RFC3339, filter.DateFrom); err == nil {
				repoFilter.CreatedFrom = &t
			}
		}
		if filter.DateTo != "" {
			if t, err := time.Parse(time.RFC3339, filter.DateTo); err == nil {
				repoFilter.CreatedTo = &t
			}
		}
		if filter.Code != "" {
			repoFilter.Code = &filter.Code
		}
	}

	return s.ticketRepo.Find(ctx, repoFilter)
}

// GetTicketsByEvent obtiene todos los tickets de un evento
func (s *TicketService) GetTicketsByEvent(ctx context.Context, eventID string) ([]*entities.Ticket, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	filter := &repository.TicketFilter{
		EventID: &event.ID,
	}
	tickets, _, err := s.ticketRepo.Find(ctx, filter)
	return tickets, err
}

// GetTicketsByCustomer obtiene tickets de un cliente
func (s *TicketService) GetTicketsByCustomer(ctx context.Context, customerID string, filter *ticketdto.TicketFilter, pagination commondto.Pagination) ([]*entities.Ticket, int64, error) {
	customer, err := s.customerRepo.GetByPublicID(ctx, customerID)
	if err != nil {
		return nil, 0, fmt.Errorf("customer not found: %w", err)
	}

	repoFilter := &repository.TicketFilter{
		CustomerID: &customer.ID,
		Limit:      pagination.PageSize,
		Offset:     (pagination.Page - 1) * pagination.PageSize,
	}

	if filter != nil {
		if filter.Status != "" {
			status := enums.TicketStatus(filter.Status)
			if status.IsValid() {
				repoFilter.Status = []enums.TicketStatus{status}
			}
		}
		if filter.DateFrom != "" {
			if t, err := time.Parse(time.RFC3339, filter.DateFrom); err == nil {
				repoFilter.CreatedFrom = &t
			}
		}
		if filter.DateTo != "" {
			if t, err := time.Parse(time.RFC3339, filter.DateTo); err == nil {
				repoFilter.CreatedTo = &t
			}
		}
	}

	return s.ticketRepo.Find(ctx, repoFilter)
}

// UpdateTicket actualiza información de un ticket (incluyendo status)
func (s *TicketService) UpdateTicket(ctx context.Context, ticketID string, req *ticketdto.UpdateTicketRequest) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.GetByPublicID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	if req.AttendeeName != nil {
		ticket.AttendeeName = req.AttendeeName
	}
	if req.AttendeeEmail != nil {
		ticket.AttendeeEmail = req.AttendeeEmail
	}
	if req.AttendeePhone != nil {
		ticket.AttendeePhone = req.AttendeePhone
	}

	// 🔥 Manejar cambio de status (reserved → sold)
	if req.Status != nil && *req.Status != ticket.Status {
		// Validar transición permitida
		if ticket.Status == string(enums.TicketStatusReserved) && *req.Status == string(enums.TicketStatusSold) {
			// Transición válida: reserved → sold
			now := time.Now()
			ticket.Status = *req.Status
			ticket.SoldAt = &now

			// Confirmar reserva en inventario
			err = s.ticketTypeRepo.ConfirmReservation(ctx, ticket.TicketTypeID, 1)
			if err != nil {
				return nil, fmt.Errorf("failed to confirm reservation: %w", err)
			}
		} else if !enums.CanTransitionTicket(enums.TicketStatus(ticket.Status), enums.TicketStatus(*req.Status)) {
			return nil, fmt.Errorf("invalid status transition from %s to %s", ticket.Status, *req.Status)
		} else {
			now := time.Now()
			switch enums.TicketStatus(*req.Status) {
			case enums.TicketStatusCancelled:
				ticket.CancelledAt = &now
			case enums.TicketStatusRefunded:
				ticket.RefundedAt = &now
			}
			ticket.Status = *req.Status
		}
	}

	ticket.UpdatedAt = time.Now()

	err = s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return ticket, nil
}

// CancelTicket cancela un ticket
func (s *TicketService) CancelTicket(ctx context.Context, ticketID string) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.GetByPublicID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	if !ticket.CanBeCancelled() {
		return nil, errors.New("ticket cannot be cancelled")
	}

	err = s.ticketRepo.Cancel(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel ticket: %w", err)
	}

	updatedTicket, err := s.ticketRepo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("ticket cancelled but retrieval failed: %w", err)
	}

	return updatedTicket, nil
}

// RefundTicket reembolsa un ticket
func (s *TicketService) RefundTicket(ctx context.Context, ticketID string) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.GetByPublicID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	if !ticket.CanBeRefunded() {
		return nil, errors.New("ticket cannot be refunded")
	}

	err = s.ticketRepo.Refund(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to refund ticket: %w", err)
	}

	updatedTicket, err := s.ticketRepo.GetByID(ctx, ticket.ID)
	if err != nil {
		return nil, fmt.Errorf("ticket refunded but retrieval failed: %w", err)
	}

	return updatedTicket, nil
}

// ValidateTicket valida un ticket por código y hash
func (s *TicketService) ValidateTicket(ctx context.Context, code, secretHash string) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.ValidateTicket(ctx, code, secretHash)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket: %w", err)
	}
	return ticket, nil
}

// GetTicketStats obtiene estadísticas de tickets para un evento
func (s *TicketService) GetTicketStats(ctx context.Context, eventID string) (*ticketdto.TicketStatsResponse, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	stats, err := s.ticketRepo.GetEventStats(ctx, event.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket stats: %w", err)
	}

	checkInRate := 0.0
	if stats.SoldTickets > 0 {
		checkInRate = float64(stats.CheckedInTickets) / float64(stats.SoldTickets)
	}

	return &ticketdto.TicketStatsResponse{
		TotalTickets:     stats.TotalTickets,
		AvailableTickets: stats.AvailableTickets,
		SoldTickets:      stats.SoldTickets,
		ReservedTickets:  stats.ReservedTickets,
		CheckedInTickets: stats.CheckedInTickets,
		CancelledTickets: stats.CancelledTickets,
		RefundedTickets:  stats.RefundedTickets,
		TotalRevenue:     stats.TotalRevenue,
		AvgTicketPrice:   stats.AvgTicketPrice,
		CheckInRate:      checkInRate,
	}, nil
}

// ReleaseExpiredReservations libera reservas expiradas
func (s *TicketService) ReleaseExpiredReservations(ctx context.Context) (int, error) {
	expiredTickets, err := s.ticketRepo.GetReservedExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired reservations: %w", err)
	}

	count := 0
	for _, ticket := range expiredTickets {
		err = s.ticketRepo.ReleaseReservation(ctx, ticket.ID)
		if err != nil {
			continue
		}

		err = s.ticketTypeRepo.ReleaseReservation(ctx, ticket.TicketTypeID, 1)
		if err != nil {
			continue
		}
		count++
	}

	return count, nil
}

// generateTicketCode genera un código único para el ticket
func (s *TicketService) generateTicketCode(eventID, ticketTypeID int64, attempt int) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TKT-%d-%d-%d-%d", eventID, ticketTypeID, timestamp, attempt)
}

// PurchaseTicket convierte una reserva en venta
func (s *TicketService) PurchaseTicket(ctx context.Context, req *ticketdto.PurchaseTicketRequest) (*entities.Ticket, error) {
	if req.TicketID == "" {
		return nil, errors.New("ticket_id is required")
	}
	if req.CustomerID == "" {
		return nil, errors.New("customer_id is required")
	}

	ticket, err := s.ticketRepo.GetByPublicID(ctx, req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	// Debe estar reservado
	if ticket.Status != string(enums.TicketStatusReserved) {
		return nil, errors.New("ticket is not reserved")
	}

	// Validar expiración
	if ticket.ReservationExpiresAt != nil && time.Now().After(*ticket.ReservationExpiresAt) {
		return nil, errors.New("reservation expired")
	}

	customer, err := s.customerRepo.GetByPublicID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	now := time.Now()

	// Actualizar ticket
	ticket.Status = string(enums.TicketStatusSold)
	ticket.CustomerID = &customer.ID
	ticket.SoldAt = &now
	ticket.UpdatedAt = now

	// 🔥 CRÍTICO: Limpiar campos de reserva
	ticket.ReservedAt = nil
	ticket.ReservedBy = nil
	ticket.ReservationExpiresAt = nil // 🔥 ESTO ES CLAVE

	err = s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to purchase ticket: %w", err)
	}

	// Confirmar reserva en inventario
	err = s.ticketTypeRepo.ConfirmReservation(ctx, ticket.TicketTypeID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm reservation: %w", err)
	}

	go s.customerRepo.UpdateStats(ctx, customer.ID, ticket.FinalPrice)

	return ticket, nil
}
