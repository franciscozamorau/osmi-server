package services

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

type TicketTypeService struct {
	ticketTypeRepo repository.TicketTypeRepository
	eventRepo      repository.EventRepository
}

func NewTicketTypeService(
	ticketTypeRepo repository.TicketTypeRepository,
	eventRepo repository.EventRepository,
) *TicketTypeService {
	return &TicketTypeService{
		ticketTypeRepo: ticketTypeRepo,
		eventRepo:      eventRepo,
	}
}

func (s *TicketTypeService) CreateTicketType(ctx context.Context, req *dto.CreateTicketTypeRequest) (*entities.TicketType, error) {
	// Validar evento
	event, err := s.eventRepo.FindByPublicID(ctx, req.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Parsear fechas
	saleStartsAt, err := time.Parse(time.RFC3339, req.SaleStartsAt)
	if err != nil {
		return nil, errors.New("invalid sale start date format")
	}

	var saleEndsAt *time.Time
	if req.SaleEndsAt != "" {
		endsAt, err := time.Parse(time.RFC3339, req.SaleEndsAt)
		if err != nil {
			return nil, errors.New("invalid sale end date format")
		}
		saleEndsAt = &endsAt
	}

	// Validar que saleEndsAt sea después de saleStartsAt si se proporciona
	if saleEndsAt != nil && saleEndsAt.Before(saleStartsAt) {
		return nil, errors.New("sale end date must be after sale start date")
	}

	// Validar que maxPerOrder sea mayor o igual que minPerOrder
	if req.MaxPerOrder < req.MinPerOrder {
		return nil, errors.New("max per order must be greater or equal than min per order")
	}

	// Crear tipo de ticket
	ticketType := &entities.TicketType{
		PublicID:          uuid.New().String(),
		EventID:           event.ID,
		Name:              req.Name,
		Description:       &req.Description,
		TicketClass:       req.TicketClass,
		BasePrice:         req.BasePrice,
		Currency:          req.Currency,
		TaxRate:           req.TaxRate,
		ServiceFeeType:    req.ServiceFeeType,
		ServiceFeeValue:   req.ServiceFeeValue,
		TotalQuantity:     int32(req.TotalQuantity),
		ReservedQuantity:  0,
		SoldQuantity:      0,
		MaxPerOrder:       int32(req.MaxPerOrder),
		MinPerOrder:       int32(req.MinPerOrder),
		SaleStartsAt:      saleStartsAt,
		SaleEndsAt:        saleEndsAt,
		IsActive:          req.IsActive,
		RequiresApproval:  req.RequiresApproval,
		IsHidden:          req.IsHidden,
		SalesChannel:      req.SalesChannel,
		Benefits:          &req.Benefits,
		AccessType:        req.AccessType,
		ValidationRules:   &req.ValidationRules,
		AvailableQuantity: int32(req.TotalQuantity),
		IsSoldOut:         false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = s.ticketTypeRepo.Create(ctx, ticketType)
	if err != nil {
		return nil, err
	}

	return ticketType, nil
}

func (s *TicketTypeService) UpdateTicketType(ctx context.Context, ticketTypeID string, req *dto.UpdateTicketTypeRequest) (*entities.TicketType, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return nil, errors.New("ticket type not found")
	}

	// Validar que se pueda modificar
	if ticketType.SoldQuantity > 0 {
		// No permitir modificar ciertos campos si hay tickets vendidos
		if req.BasePrice != 0 && req.BasePrice != ticketType.BasePrice {
			return nil, errors.New("cannot change price when tickets have been sold")
		}
		if req.TotalQuantity != 0 && req.TotalQuantity < int(ticketType.SoldQuantity+ticketType.ReservedQuantity) {
			return nil, errors.New("new total quantity cannot be less than sold + reserved tickets")
		}
	}

	// Actualizar campos
	if req.Name != "" {
		ticketType.Name = req.Name
	}
	if req.Description != "" {
		ticketType.Description = &req.Description
	}
	if req.BasePrice != 0 {
		ticketType.BasePrice = req.BasePrice
	}
	if req.TotalQuantity != 0 {
		oldTotal := ticketType.TotalQuantity
		ticketType.TotalQuantity = int32(req.TotalQuantity)
		ticketType.AvailableQuantity = ticketType.TotalQuantity - ticketType.SoldQuantity - ticketType.ReservedQuantity
		ticketType.IsSoldOut = ticketType.AvailableQuantity <= 0

		// Actualizar estadísticas del evento
		if oldTotal != ticketType.TotalQuantity {
			// TODO: Actualizar capacidad total del evento
		}
	}
	if req.MaxPerOrder != 0 {
		ticketType.MaxPerOrder = int32(req.MaxPerOrder)
	}
	if req.MinPerOrder != 0 {
		ticketType.MinPerOrder = int32(req.MinPerOrder)
	}
	if req.SaleStartsAt != "" {
		saleStartsAt, err := time.Parse(time.RFC3339, req.SaleStartsAt)
		if err != nil {
			return nil, errors.New("invalid sale start date format")
		}
		ticketType.SaleStartsAt = saleStartsAt
	}
	if req.SaleEndsAt != "" {
		saleEndsAt, err := time.Parse(time.RFC3339, req.SaleEndsAt)
		if err != nil {
			return nil, errors.New("invalid sale end date format")
		}
		ticketType.SaleEndsAt = &saleEndsAt
	}
	if req.IsActive != nil {
		ticketType.IsActive = *req.IsActive
	}
	if req.IsHidden != nil {
		ticketType.IsHidden = *req.IsHidden
	}
	if req.Benefits != nil {
		ticketType.Benefits = &req.Benefits
	}
	if req.ValidationRules != nil {
		ticketType.ValidationRules = &req.ValidationRules
	}

	ticketType.UpdatedAt = time.Now()

	err = s.ticketTypeRepo.Update(ctx, ticketType)
	if err != nil {
		return nil, err
	}

	return ticketType, nil
}

func (s *TicketTypeService) GetTicketType(ctx context.Context, ticketTypeID string) (*entities.TicketType, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return nil, errors.New("ticket type not found")
	}
	return ticketType, nil
}

func (s *TicketTypeService) ListTicketTypes(ctx context.Context, filter dto.TicketTypeFilter, pagination dto.Pagination) ([]*entities.TicketType, int64, error) {
	ticketTypes, err := s.ticketTypeRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.ticketTypeRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return ticketTypes, total, nil
}

func (s *TicketTypeService) GetTicketTypesByEvent(ctx context.Context, eventID string) ([]*entities.TicketType, error) {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	return s.ticketTypeRepo.FindByEventID(ctx, event.ID)
}

func (s *TicketTypeService) CheckAvailability(ctx context.Context, ticketTypeID string, quantity int) (bool, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return false, errors.New("ticket type not found")
	}

	// Validar cantidad
	if quantity < int(ticketType.MinPerOrder) {
		return false, errors.New("quantity below minimum per order")
	}
	if quantity > int(ticketType.MaxPerOrder) {
		return false, errors.New("quantity exceeds maximum per order")
	}

	// Validar disponibilidad
	available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, int32(quantity))
	if err != nil {
		return false, err
	}

	return available, nil
}

func (s *TicketTypeService) UpdateSaleDates(ctx context.Context, ticketTypeID string, startsAt, endsAt time.Time) error {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return errors.New("ticket type not found")
	}

	ticketType.SaleStartsAt = startsAt
	if !endsAt.IsZero() {
		ticketType.SaleEndsAt = &endsAt
	}
	ticketType.UpdatedAt = time.Now()

	return s.ticketTypeRepo.Update(ctx, ticketType)
}

func (s *TicketTypeService) ToggleActive(ctx context.Context, ticketTypeID string, active bool) error {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return errors.New("ticket type not found")
	}

	ticketType.IsActive = active
	ticketType.UpdatedAt = time.Now()

	return s.ticketTypeRepo.Update(ctx, ticketType)
}

func (s *TicketTypeService) GetTicketTypeStats(ctx context.Context, ticketTypeID string) (*dto.TicketTypeStatsResponse, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return nil, errors.New("ticket type not found")
	}

	stats, err := s.ticketTypeRepo.GetSalesStats(ctx, ticketType.ID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
