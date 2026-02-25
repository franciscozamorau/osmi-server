// internal/application/services/ticket_type_service.go
package services

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/api/dto/request"
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

func (s *TicketTypeService) CreateTicketType(ctx context.Context, req *request.CreateTicketTypeRequest) (*entities.TicketType, error) {
	// Validar evento
	event, err := s.eventRepo.GetByPublicID(ctx, req.EventID)
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

	// Validar que saleEndsAt sea después de saleStartsAt
	if saleEndsAt != nil && saleEndsAt.Before(saleStartsAt) {
		return nil, errors.New("sale end date must be after sale start date")
	}

	// Validar que maxPerOrder sea mayor o igual que minPerOrder
	if req.MaxPerOrder < req.MinPerOrder {
		return nil, errors.New("max per order must be greater or equal than min per order")
	}

	// Procesar beneficios
	var benefits []string
	if req.Benefits != "" {
		// Por ahora, asumimos que es un string con un beneficio
		// En el futuro, podría ser JSON
		benefits = []string{req.Benefits}
	}

	// Crear ValidationRules
	validationRules := &entities.ValidationRules{
		RequiresID:         false,
		AgeRestriction:     0,
		RequiresMembership: false,
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
		TotalQuantity:     int(req.TotalQuantity),
		ReservedQuantity:  0,
		SoldQuantity:      0,
		MaxPerOrder:       int(req.MaxPerOrder),
		MinPerOrder:       int(req.MinPerOrder),
		SaleStartsAt:      saleStartsAt,
		SaleEndsAt:        saleEndsAt,
		IsActive:          req.IsActive,
		RequiresApproval:  req.RequiresApproval,
		IsHidden:          req.IsHidden,
		SalesChannel:      req.SalesChannel,
		Benefits:          &benefits,
		AccessType:        req.AccessType,
		ValidationRules:   validationRules,
		AvailableQuantity: int(req.TotalQuantity),
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

func (s *TicketTypeService) UpdateTicketType(ctx context.Context, ticketTypeID string, req *request.UpdateTicketTypeRequest) (*entities.TicketType, error) {
	// Usar FindByPublicID (existe en la interfaz)
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return nil, errors.New("ticket type not found")
	}

	// Validar que se pueda modificar
	if ticketType.SoldQuantity > 0 {
		// No permitir modificar ciertos campos si hay tickets vendidos
		if req.BasePrice != nil && *req.BasePrice != ticketType.BasePrice {
			return nil, errors.New("cannot change price when tickets have been sold")
		}
		if req.TotalQuantity != nil && *req.TotalQuantity < (ticketType.SoldQuantity+ticketType.ReservedQuantity) {
			return nil, errors.New("new total quantity cannot be less than sold + reserved tickets")
		}
	}

	// Actualizar campos
	if req.Name != nil {
		ticketType.Name = *req.Name
	}
	if req.Description != nil {
		ticketType.Description = req.Description
	}
	if req.BasePrice != nil {
		ticketType.BasePrice = *req.BasePrice
	}
	if req.TotalQuantity != nil {
		ticketType.TotalQuantity = int(*req.TotalQuantity)
		ticketType.AvailableQuantity = ticketType.TotalQuantity - ticketType.SoldQuantity - ticketType.ReservedQuantity
		ticketType.IsSoldOut = ticketType.AvailableQuantity <= 0
	}
	if req.MaxPerOrder != nil {
		ticketType.MaxPerOrder = int(*req.MaxPerOrder)
	}
	if req.MinPerOrder != nil {
		ticketType.MinPerOrder = int(*req.MinPerOrder)
	}
	if req.SaleStartsAt != nil {
		saleStartsAt, err := time.Parse(time.RFC3339, *req.SaleStartsAt)
		if err != nil {
			return nil, errors.New("invalid sale start date format")
		}
		ticketType.SaleStartsAt = saleStartsAt
	}
	if req.SaleEndsAt != nil {
		saleEndsAt, err := time.Parse(time.RFC3339, *req.SaleEndsAt)
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
		benefits := []string{*req.Benefits}
		ticketType.Benefits = &benefits
	}
	if req.ValidationRules != nil {
		// Por ahora, usamos valores por defecto
		rules := &entities.ValidationRules{
			RequiresID:         false,
			AgeRestriction:     0,
			RequiresMembership: false,
		}
		ticketType.ValidationRules = rules
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

func (s *TicketTypeService) ListTicketTypes(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*entities.TicketType, int64, error) {
	// Convertir el filtro genérico a dto.TicketTypeFilter
	ticketTypeFilter := dto.TicketTypeFilter{}

	// Mapear campos comunes si existen
	if eventID, ok := filter["event_id"]; ok {
		if id, ok := eventID.(int64); ok {
			ticketTypeFilter.EventID = &id
		}
	}
	if active, ok := filter["is_active"]; ok {
		if isActive, ok := active.(bool); ok {
			ticketTypeFilter.IsActive = &isActive
		}
	}
	if search, ok := filter["search"]; ok {
		if term, ok := search.(string); ok {
			ticketTypeFilter.Search = term
		}
	}

	// Configurar paginación
	pagination := dto.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return s.ticketTypeRepo.List(ctx, ticketTypeFilter, pagination)
}

func (s *TicketTypeService) GetTicketTypesByEvent(ctx context.Context, eventID string) ([]*entities.TicketType, error) {
	// Usar FindByEventPublicID (existe en la interfaz)
	return s.ticketTypeRepo.FindByEventPublicID(ctx, eventID)
}

func (s *TicketTypeService) CheckAvailability(ctx context.Context, ticketTypeID string, quantity int) (bool, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return false, errors.New("ticket type not found")
	}

	// Validar cantidad
	if quantity < ticketType.MinPerOrder {
		return false, errors.New("quantity below minimum per order")
	}
	if quantity > ticketType.MaxPerOrder {
		return false, errors.New("quantity exceeds maximum per order")
	}

	// Usar CheckAvailability (existe en la interfaz)
	return s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, quantity)
}

func (s *TicketTypeService) ToggleActive(ctx context.Context, ticketTypeID string, active bool) error {
	// Primero obtener el ticket type para tener su ID numérico
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return errors.New("ticket type not found")
	}

	// Usar UpdateStatus con el ID numérico
	return s.ticketTypeRepo.UpdateStatus(ctx, ticketType.ID, active)
}
