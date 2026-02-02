package services

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
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

func (s *TicketService) CreateTicket(ctx context.Context, req *dto.CreateTicketRequest) (*entities.Ticket, error) {
	// Validar tipo de ticket
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, req.TicketTypeID)
	if err != nil {
		return nil, errors.New("ticket type not found")
	}

	// Validar disponibilidad
	available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, 1)
	if err != nil || !available {
		return nil, errors.New("ticket type not available")
	}

	// Validar cliente
	customer, err := s.customerRepo.FindByPublicID(ctx, req.CustomerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	// Validar evento
	event, err := s.eventRepo.FindByID(ctx, ticketType.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Validar que el evento esté activo
	if event.Status != string(enums.EventStatusPublished) && event.Status != string(enums.EventStatusLive) {
		return nil, errors.New("event is not active for ticket sales")
	}

	// Crear ticket
	ticket := &entities.Ticket{
		PublicID:      uuid.New().String(),
		TicketTypeID:  ticketType.ID,
		EventID:       event.ID,
		CustomerID:    &customer.ID,
		Code:          generateTicketCode(),
		SecretHash:    uuid.New().String(),
		Status:        string(enums.TicketStatusSold),
		FinalPrice:    ticketType.BasePrice,
		Currency:      ticketType.Currency,
		TaxAmount:     ticketType.BasePrice * ticketType.TaxRate,
		AttendeeName:  &req.AttendeeName,
		AttendeeEmail: &req.AttendeeEmail,
		AttendeePhone: &req.AttendeePhone,
		SoldAt:        &time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Si hay orderID, asociarlo
	if req.OrderID != "" {
		order, err := s.orderRepo.FindByPublicID(ctx, req.OrderID)
		if err == nil {
			ticket.OrderID = &order.ID
		}
	}

	err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}

	// Actualizar contador de tickets vendidos
	err = s.ticketTypeRepo.SellTickets(ctx, ticketType.ID, 1)
	if err != nil {
		// Rollback: eliminar ticket creado
		s.ticketRepo.Delete(ctx, ticket.ID)
		return nil, err
	}

	// Actualizar estadísticas del cliente
	s.customerRepo.UpdateStats(ctx, customer.ID, ticket.FinalPrice)

	return ticket, nil
}

func (s *TicketService) ReserveTicket(ctx context.Context, req *dto.ReserveTicketRequest) (*entities.Ticket, error) {
	// Validar tipo de ticket
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, req.TicketTypeID)
	if err != nil {
		return nil, errors.New("ticket type not found")
	}

	// Validar disponibilidad
	available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, req.Quantity)
	if err != nil || !available {
		return nil, errors.New("not enough tickets available")
	}

	// Validar cliente
	customer, err := s.customerRepo.FindByPublicID(ctx, req.CustomerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	// Validar evento
	event, err := s.eventRepo.FindByID(ctx, ticketType.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Validar que el evento permita reservas
	if !event.AllowReservations {
		return nil, errors.New("event does not allow reservations")
	}

	// Calcular duración de reserva
	duration := req.DurationMinutes
	if duration == 0 {
		duration = int(event.ReservationDuration)
	}
	if duration == 0 {
		duration = 15 // Default 15 minutos
	}

	reservationExpiresAt := time.Now().Add(time.Duration(duration) * time.Minute)

	// Crear ticket reservado
	ticket := &entities.Ticket{
		PublicID:             uuid.New().String(),
		TicketTypeID:         ticketType.ID,
		EventID:              event.ID,
		CustomerID:           &customer.ID,
		Code:                 generateTicketCode(),
		SecretHash:           uuid.New().String(),
		Status:               string(enums.TicketStatusReserved),
		FinalPrice:           ticketType.BasePrice,
		Currency:             ticketType.Currency,
		TaxAmount:            ticketType.BasePrice * ticketType.TaxRate,
		ReservedAt:           &time.Now(),
		ReservationExpiresAt: &reservationExpiresAt,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}

	// Reservar tickets en el tipo de ticket
	err = s.ticketTypeRepo.ReserveTickets(ctx, ticketType.ID, req.Quantity)
	if err != nil {
		// Rollback: eliminar ticket creado
		s.ticketRepo.Delete(ctx, ticket.ID)
		return nil, err
	}

	return ticket, nil
}

func (s *TicketService) CheckInTicket(ctx context.Context, req *dto.CheckInTicketRequest) (*entities.Ticket, error) {
	// Obtener ticket
	ticket, err := s.ticketRepo.FindByPublicID(ctx, req.TicketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// Validar estado
	if ticket.Status != string(enums.TicketStatusSold) {
		return nil, errors.New("ticket is not valid for check-in")
	}

	// Validar que no esté ya usado
	if ticket.CheckedInAt != nil {
		return nil, errors.New("ticket already checked in")
	}

	// Validar evento
	event, err := s.eventRepo.FindByID(ctx, ticket.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Validar fechas del evento
	now := time.Now()
	if now.Before(event.StartsAt.Add(-1 * time.Hour)) {
		return nil, errors.New("check-in not available yet")
	}

	if now.After(event.EndsAt.Add(2 * time.Hour)) {
		return nil, errors.New("check-in period has ended")
	}

	// Obtener validador si se proporciona
	var validatorID *int64
	if req.ValidatorID != "" {
		// TODO: Validar que el validador exista y tenga permisos
	}

	// Realizar check-in
	ticket.CheckedInAt = &now
	ticket.CheckedInBy = validatorID
	ticket.CheckinMethod = &req.CheckinMethod
	ticket.CheckinLocation = &req.CheckinLocation
	ticket.Status = string(enums.TicketStatusCheckedIn)
	ticket.ValidationCount++
	ticket.LastValidatedAt = &now
	ticket.UpdatedAt = now

	err = s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *TicketService) TransferTicket(ctx context.Context, req *dto.TransferTicketRequest) (*entities.Ticket, error) {
	// Obtener ticket
	ticket, err := s.ticketRepo.FindByPublicID(ctx, req.TicketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// Validar que pertenece al cliente que lo transfiere
	fromCustomer, err := s.customerRepo.FindByPublicID(ctx, req.FromCustomerID)
	if err != nil {
		return nil, errors.New("sender customer not found")
	}

	if ticket.CustomerID == nil || *ticket.CustomerID != fromCustomer.ID {
		return nil, errors.New("ticket does not belong to sender")
	}

	// Validar que el ticket se puede transferir
	if ticket.Status != string(enums.TicketStatusSold) {
		return nil, errors.New("only sold tickets can be transferred")
	}

	// Obtener o crear cliente destino
	var toCustomer *entities.Customer
	if req.ToCustomerID != "" {
		toCustomer, err = s.customerRepo.FindByPublicID(ctx, req.ToCustomerID)
		if err != nil {
			return nil, errors.New("recipient customer not found")
		}
	} else {
		// Crear cliente invitado
		toCustomer = &entities.Customer{
			PublicID:  uuid.New().String(),
			Email:     req.ToEmail,
			FullName:  req.ToName,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = s.customerRepo.Create(ctx, toCustomer)
		if err != nil {
			return nil, errors.New("could not create recipient customer")
		}
	}

	// Realizar transferencia
	now := time.Now()
	ticket.CustomerID = &toCustomer.ID
	ticket.TransferredFrom = &fromCustomer.ID
	ticket.TransferredAt = &now
	transferToken := uuid.New().String()
	ticket.TransferToken = &transferToken
	ticket.UpdatedAt = now

	err = s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *TicketService) GetTicket(ctx context.Context, ticketID string) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.FindByPublicID(ctx, ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}
	return ticket, nil
}

func (s *TicketService) ListTickets(ctx context.Context, filter dto.TicketFilter, pagination dto.Pagination) ([]*entities.Ticket, int64, error) {
	tickets, err := s.ticketRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.ticketRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (s *TicketService) UpdateTicket(ctx context.Context, ticketID string, req *dto.UpdateTicketRequest) (*entities.Ticket, error) {
	ticket, err := s.ticketRepo.FindByPublicID(ctx, ticketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	// Actualizar campos
	if req.Status != "" {
		// Validar transición de estado
		if !isValidStatusTransition(ticket.Status, req.Status) {
			return nil, errors.New("invalid status transition")
		}
		ticket.Status = req.Status

		// Actualizar timestamps según el estado
		now := time.Now()
		switch req.Status {
		case string(enums.TicketStatusCancelled):
			ticket.CancelledAt = &now
		case string(enums.TicketStatusRefunded):
			ticket.RefundedAt = &now
		}
	}

	if req.AttendeeName != "" {
		ticket.AttendeeName = &req.AttendeeName
	}
	if req.AttendeeEmail != "" {
		ticket.AttendeeEmail = &req.AttendeeEmail
	}
	if req.AttendeePhone != "" {
		ticket.AttendeePhone = &req.AttendeePhone
	}

	ticket.UpdatedAt = time.Now()

	err = s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *TicketService) GetTicketStats(ctx context.Context, eventID string) (*dto.TicketStatsResponse, error) {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Obtener todos los tickets del evento
	tickets, err := s.ticketRepo.FindByEventID(ctx, event.ID)
	if err != nil {
		return nil, err
	}

	// Calcular estadísticas
	stats := &dto.TicketStatsResponse{
		TotalTickets:     len(tickets),
		AvailableTickets: 0,
		ReservedTickets:  0,
		SoldTickets:      0,
		CheckedInTickets: 0,
		CancelledTickets: 0,
		RefundedTickets:  0,
		TotalRevenue:     0,
		CheckInRate:      0,
		AvgTicketPrice:   0,
	}

	var totalRevenue, checkedInCount float64
	for _, ticket := range tickets {
		switch ticket.Status {
		case string(enums.TicketStatusSold):
			stats.SoldTickets++
			totalRevenue += ticket.FinalPrice
		case string(enums.TicketStatusReserved):
			stats.ReservedTickets++
		case string(enums.TicketStatusCheckedIn):
			stats.CheckedInTickets++
			checkedInCount++
			totalRevenue += ticket.FinalPrice
		case string(enums.TicketStatusCancelled):
			stats.CancelledTickets++
		case string(enums.TicketStatusRefunded):
			stats.RefundedTickets++
		}
	}

	stats.TotalRevenue = totalRevenue
	if stats.SoldTickets > 0 {
		stats.AvgTicketPrice = totalRevenue / float64(stats.SoldTickets)
	}
	if stats.SoldTickets > 0 {
		stats.CheckInRate = checkedInCount / float64(stats.SoldTickets)
	}

	return stats, nil
}

// Helper functions
func generateTicketCode() string {
	return uuid.New().String()[:8]
}

func isValidStatusTransition(current, new string) bool {
	transitions := map[string][]string{
		string(enums.TicketStatusAvailable): {string(enums.TicketStatusReserved), string(enums.TicketStatusSold)},
		string(enums.TicketStatusReserved):  {string(enums.TicketStatusSold), string(enums.TicketStatusCancelled)},
		string(enums.TicketStatusSold):      {string(enums.TicketStatusCheckedIn), string(enums.TicketStatusCancelled), string(enums.TicketStatusRefunded)},
		string(enums.TicketStatusCheckedIn): {},
		string(enums.TicketStatusCancelled): {},
		string(enums.TicketStatusRefunded):  {},
	}

	allowed, ok := transitions[current]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}

	return false
}
