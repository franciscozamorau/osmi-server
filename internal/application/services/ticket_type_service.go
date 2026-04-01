// internal/application/services/ticket_type_service.go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	tickettypedto "github.com/franciscozamorau/osmi-server/internal/api/dto/ticket_type"
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

// CreateTicketType crea un nuevo tipo de ticket
func (s *TicketTypeService) CreateTicketType(ctx context.Context, req *tickettypedto.CreateTicketTypeRequest) (*entities.TicketType, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	event, err := s.eventRepo.GetByPublicID(ctx, req.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Status != "draft" && event.Status != "published" {
		return nil, errors.New("cannot add ticket types to this event")
	}

	saleStartsAt, err := s.parseTime(req.SaleStartsAt)
	if err != nil {
		return nil, fmt.Errorf("invalid sale start date: %w", err)
	}

	var saleEndsAt *time.Time
	if req.SaleEndsAt != "" {
		endsAt, err := s.parseTime(req.SaleEndsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid sale end date: %w", err)
		}
		saleEndsAt = endsAt
	}

	if saleEndsAt != nil && saleEndsAt.Before(*saleStartsAt) {
		return nil, errors.New("sale end date must be after sale start date")
	}

	if req.MaxPerOrder < req.MinPerOrder {
		return nil, errors.New("max per order must be greater or equal than min per order")
	}

	benefits := s.parseBenefits(req.Benefits)
	validationRules := s.parseValidationRules(req.ValidationRules)

	now := time.Now()
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
		SaleStartsAt:      *saleStartsAt,
		SaleEndsAt:        saleEndsAt,
		IsActive:          req.IsActive,
		RequiresApproval:  req.RequiresApproval,
		IsHidden:          req.IsHidden,
		SalesChannel:      req.SalesChannel,
		Benefits:          benefits,
		AccessType:        req.AccessType,
		ValidationRules:   validationRules,
		AvailableQuantity: int(req.TotalQuantity),
		IsSoldOut:         false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.ticketTypeRepo.Create(ctx, ticketType); err != nil {
		return nil, fmt.Errorf("failed to create ticket type: %w", err)
	}

	return ticketType, nil
}

// UpdateTicketType actualiza un tipo de ticket existente
func (s *TicketTypeService) UpdateTicketType(ctx context.Context, ticketTypeID string, req *tickettypedto.UpdateTicketTypeRequest) (*entities.TicketType, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return nil, fmt.Errorf("ticket type not found: %w", err)
	}

	if ticketType.SoldQuantity > 0 {
		if err := s.validateUpdateWithSoldTickets(ticketType, req); err != nil {
			return nil, err
		}
	}

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
		ticketType.UpdateAvailableQuantity()
	}
	if req.MaxPerOrder != nil {
		ticketType.MaxPerOrder = int(*req.MaxPerOrder)
	}
	if req.MinPerOrder != nil {
		ticketType.MinPerOrder = int(*req.MinPerOrder)
	}
	if req.SaleStartsAt != nil {
		saleStartsAt, err := s.parseTime(*req.SaleStartsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid sale start date: %w", err)
		}
		ticketType.SaleStartsAt = *saleStartsAt
	}
	if req.SaleEndsAt != nil {
		saleEndsAt, err := s.parseTime(*req.SaleEndsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid sale end date: %w", err)
		}
		ticketType.SaleEndsAt = saleEndsAt
	}
	if req.IsActive != nil {
		ticketType.IsActive = *req.IsActive
	}
	if req.IsHidden != nil {
		ticketType.IsHidden = *req.IsHidden
	}
	if req.Benefits != nil {
		ticketType.Benefits = s.parseBenefits(*req.Benefits)
	}
	if req.ValidationRules != nil {
		ticketType.ValidationRules = s.parseValidationRules(*req.ValidationRules)
	}

	ticketType.UpdatedAt = time.Now()

	if err := s.ticketTypeRepo.Update(ctx, ticketType); err != nil {
		return nil, fmt.Errorf("failed to update ticket type: %w", err)
	}

	return ticketType, nil
}

// GetTicketType obtiene un tipo de ticket por su ID
func (s *TicketTypeService) GetTicketType(ctx context.Context, ticketTypeID string) (*entities.TicketType, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return nil, fmt.Errorf("ticket type not found: %w", err)
	}
	return ticketType, nil
}

// ListTicketTypes lista tipos de ticket con filtros y paginación
func (s *TicketTypeService) ListTicketTypes(ctx context.Context, filter *tickettypedto.TicketTypeFilter, page, pageSize int) ([]*entities.TicketType, int64, error) {
	if filter == nil {
		filter = &tickettypedto.TicketTypeFilter{}
	}

	pagination := commondto.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}

	ticketTypes, total, err := s.ticketTypeRepo.List(ctx, *filter, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list ticket types: %w", err)
	}

	return ticketTypes, total, nil
}

// GetTicketTypesByEvent obtiene todos los tipos de ticket de un evento
func (s *TicketTypeService) GetTicketTypesByEvent(ctx context.Context, eventID string) ([]*entities.TicketType, error) {
	_, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	ticketTypes, err := s.ticketTypeRepo.FindByEventPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket types: %w", err)
	}

	return ticketTypes, nil
}

// CheckAvailability verifica disponibilidad de tickets
func (s *TicketTypeService) CheckAvailability(ctx context.Context, ticketTypeID string, quantity int) (bool, error) {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return false, fmt.Errorf("ticket type not found: %w", err)
	}

	if err := ticketType.ValidateOrderQuantity(quantity); err != nil {
		return false, err
	}

	available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, quantity)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return available, nil
}

// ToggleActive activa o desactiva un tipo de ticket
func (s *TicketTypeService) ToggleActive(ctx context.Context, ticketTypeID string, active bool) error {
	ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, ticketTypeID)
	if err != nil {
		return fmt.Errorf("ticket type not found: %w", err)
	}

	if err := s.ticketTypeRepo.UpdateStatus(ctx, ticketType.ID, active); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// ============================================================================
// FUNCIONES HELPER PRIVADAS
// ============================================================================

func (s *TicketTypeService) validateCreateRequest(req *tickettypedto.CreateTicketTypeRequest) error {
	if req.EventID == "" {
		return errors.New("event_id is required")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.TotalQuantity <= 0 {
		return errors.New("total_quantity must be greater than 0")
	}
	if req.BasePrice < 0 {
		return errors.New("base_price cannot be negative")
	}
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	return nil
}

func (s *TicketTypeService) validateUpdateWithSoldTickets(ticketType *entities.TicketType, req *tickettypedto.UpdateTicketTypeRequest) error {
	if req.BasePrice != nil && *req.BasePrice != ticketType.BasePrice {
		return errors.New("cannot change price when tickets have been sold")
	}
	if req.TotalQuantity != nil && *req.TotalQuantity < (ticketType.SoldQuantity+ticketType.ReservedQuantity) {
		return errors.New("new total quantity cannot be less than sold + reserved tickets")
	}
	return nil
}

func (s *TicketTypeService) parseTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, errors.New("time string is empty")
	}
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid time format (expected RFC3339): %w", err)
	}
	return &t, nil
}

func (s *TicketTypeService) parseBenefits(benefitsStr string) []string {
	if benefitsStr == "" {
		return []string{}
	}
	var benefits []string
	if err := json.Unmarshal([]byte(benefitsStr), &benefits); err == nil {
		return benefits
	}
	return []string{benefitsStr}
}

func (s *TicketTypeService) parseValidationRules(rulesStr string) *entities.ValidationRules {
	if rulesStr == "" {
		return nil
	}
	var rules entities.ValidationRules
	if err := json.Unmarshal([]byte(rulesStr), &rules); err != nil {
		return nil
	}
	return &rules
}
