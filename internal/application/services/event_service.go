// internal/application/services/event_service.go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

type EventService struct {
	eventRepo      repository.EventRepository
	organizerRepo  repository.OrganizerRepository
	venueRepo      repository.VenueRepository
	categoryRepo   repository.CategoryRepository
	ticketTypeRepo repository.TicketTypeRepository
}

func NewEventService(
	eventRepo repository.EventRepository,
	organizerRepo repository.OrganizerRepository,
	venueRepo repository.VenueRepository,
	categoryRepo repository.CategoryRepository,
	ticketTypeRepo repository.TicketTypeRepository,
) *EventService {
	return &EventService{
		eventRepo:      eventRepo,
		organizerRepo:  organizerRepo,
		venueRepo:      venueRepo,
		categoryRepo:   categoryRepo,
		ticketTypeRepo: ticketTypeRepo,
	}
}

// CreateEvent crea un nuevo evento
func (s *EventService) CreateEvent(ctx context.Context, req *dto.CreateEventRequest) (*entities.Event, error) {
	// Validar organizador
	organizer, err := s.organizerRepo.FindByPublicID(ctx, req.OrganizerID)
	if err != nil {
		return nil, fmt.Errorf("organizer not found: %w", err)
	}

	// Validar venue si se proporciona
	var venueID *int64
	if req.VenueID != "" {
		venue, err := s.venueRepo.FindByPublicID(ctx, req.VenueID)
		if err != nil {
			return nil, fmt.Errorf("venue not found: %w", err)
		}
		venueID = &venue.ID
	}

	// Validar categoría primaria
	var primaryCategoryID *int64
	if req.PrimaryCategoryID != "" {
		category, err := s.categoryRepo.GetByPublicID(ctx, req.PrimaryCategoryID)
		if err != nil {
			return nil, fmt.Errorf("primary category not found: %w", err)
		}
		primaryCategoryID = &category.ID
	}

	// Validar fechas
	if req.EndsAt.Before(req.StartsAt) {
		return nil, errors.New("end date must be after start date")
	}

	// Procesar Tags de JSON string a []string
	var tags *[]string
	if req.Tags != "" {
		var tagsSlice []string
		// Intentar parsear como JSON array
		if err := json.Unmarshal([]byte(req.Tags), &tagsSlice); err == nil {
			tags = &tagsSlice
		} else {
			// Si no es JSON válido, tratar como un solo tag
			tagsSlice = []string{req.Tags}
			tags = &tagsSlice
		}
	}

	// Crear evento con conversiones de tipos correctas
	event := &entities.Event{
		PublicID:            uuid.New().String(),
		OrganizerID:         organizer.ID,
		PrimaryCategoryID:   primaryCategoryID,
		VenueID:             venueID,
		Name:                req.Name,
		Slug:                req.Slug,
		ShortDescription:    &req.ShortDescription,
		Description:         &req.Description,
		EventType:           req.EventType,
		CoverImageURL:       &req.CoverImageURL,
		BannerImageURL:      &req.BannerImageURL,
		GalleryImages:       nil,
		Timezone:            req.Timezone,
		StartsAt:            req.StartsAt,
		EndsAt:              req.EndsAt,
		DoorsOpenAt:         &req.DoorsOpenAt,
		DoorsCloseAt:        &req.DoorsCloseAt,
		VenueName:           &req.VenueName,
		AddressFull:         &req.AddressFull,
		City:                &req.City,
		State:               &req.State,
		Country:             &req.Country,
		Status:              string(enums.EventStatusDraft),
		Visibility:          req.Visibility,
		IsFeatured:          req.IsFeatured,
		IsFree:              req.IsFree,
		MaxAttendees:        intPtrFromInt32(req.MaxAttendees),
		MinAttendees:        int(req.MinAttendees),
		Tags:                tags,
		AgeRestriction:      intPtrFromInt32(req.AgeRestriction),
		RequiresApproval:    req.RequiresApproval,
		AllowReservations:   req.AllowReservations,
		ReservationDuration: int(req.ReservationDuration),
		Settings:            nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Asociar categorías si se proporcionan
	if len(req.CategoryIDs) > 0 {
		for _, categoryID := range req.CategoryIDs {
			category, err := s.categoryRepo.GetByPublicID(ctx, categoryID)
			if err != nil {
				// Log error pero continuar con las siguientes categorías
				continue
			}

			isPrimary := primaryCategoryID != nil && *primaryCategoryID == category.ID
			if err := s.eventRepo.AddCategoryToEvent(ctx, event.ID, category.ID, isPrimary); err != nil {
				// Aquí sí deberíamos loguear el error
				_ = err // En producción, usar logger
			}
		}
	}

	return event, nil
}

// UpdateEvent actualiza un evento existente
func (s *EventService) UpdateEvent(ctx context.Context, eventID string, req *dto.UpdateEventRequest) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Status == string(enums.EventStatusCompleted) || event.Status == string(enums.EventStatusCancelled) {
		return nil, errors.New("cannot modify completed or cancelled event")
	}

	// Actualizar campos
	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.ShortDescription != nil {
		event.ShortDescription = req.ShortDescription
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.Status != nil {
		// Validar transición de estado
		if !isValidEventStatusTransition(event.Status, *req.Status) {
			return nil, fmt.Errorf("invalid status transition from %s to %s", event.Status, *req.Status)
		}
		event.Status = *req.Status
	}
	if req.Visibility != nil {
		event.Visibility = *req.Visibility
	}
	if req.IsFeatured != nil {
		event.IsFeatured = *req.IsFeatured
	}
	if req.MaxAttendees != nil {
		event.MaxAttendees = intPtrFromInt32Ptr(req.MaxAttendees)
	}
	if req.AgeRestriction != nil {
		event.AgeRestriction = intPtrFromInt32Ptr(req.AgeRestriction)
	}
	if req.StartsAt != nil {
		event.StartsAt = *req.StartsAt
	}
	if req.EndsAt != nil {
		event.EndsAt = *req.EndsAt
	}
	if req.Timezone != nil {
		event.Timezone = *req.Timezone
	}

	event.UpdatedAt = time.Now()

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return event, nil
}

// PublishEvent publica un evento (lo hace visible para ventas)
func (s *EventService) PublishEvent(ctx context.Context, eventID string, publishAt *time.Time) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Status != string(enums.EventStatusDraft) && event.Status != string(enums.EventStatusScheduled) {
		return nil, errors.New("event is not in draft or scheduled state")
	}

	// Verificar que tenga al menos un tipo de ticket activo
	ticketTypes, err := s.ticketTypeRepo.FindByEvent(ctx, event.ID, true)
	if err != nil || len(ticketTypes) == 0 {
		return nil, errors.New("event must have at least one active ticket type to be published")
	}

	event.Status = string(enums.EventStatusPublished)
	if publishAt != nil {
		event.PublishedAt = publishAt
	} else {
		now := time.Now()
		event.PublishedAt = &now
	}
	event.UpdatedAt = time.Now()

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return event, nil
}

// CancelEvent cancela un evento
func (s *EventService) CancelEvent(ctx context.Context, eventID string, reason string) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Status == string(enums.EventStatusCompleted) || event.Status == string(enums.EventStatusCancelled) {
		return nil, errors.New("event is already completed or cancelled")
	}

	// Verificar que no tenga tickets vendidos
	ticketTypes, err := s.ticketTypeRepo.FindByEvent(ctx, event.ID, true)
	if err == nil {
		for _, tt := range ticketTypes {
			if tt.SoldQuantity > 0 {
				return nil, errors.New("cannot cancel event with sold tickets")
			}
		}
	}

	event.Status = string(enums.EventStatusCancelled)
	event.UpdatedAt = time.Now()

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to cancel event: %w", err)
	}

	return event, nil
}

// GetEvent obtiene un evento por su ID
func (s *EventService) GetEvent(ctx context.Context, eventID string) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Incrementar contador de vistas (no crítico, no detenemos la operación si falla)
	event.ViewCount++
	event.UpdatedAt = time.Now()
	_ = s.eventRepo.Update(ctx, event)

	return event, nil
}

// ListEvents lista eventos con filtros y paginación
func (s *EventService) ListEvents(ctx context.Context, filter dto.EventFilter, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	// Convertir filter a map para el repositorio
	dbFilter := make(map[string]interface{})

	if filter.Name != "" {
		dbFilter["name"] = filter.Name
	}
	if filter.OrganizerID != nil {
		dbFilter["organizer_id"] = *filter.OrganizerID
	}
	if filter.CategoryID != nil {
		dbFilter["category_id"] = *filter.CategoryID
	}
	if filter.Status != "" {
		dbFilter["status"] = filter.Status
	}
	if filter.DateFrom != "" {
		dbFilter["date_from"] = filter.DateFrom
	}
	if filter.DateTo != "" {
		dbFilter["date_to"] = filter.DateTo
	}
	if filter.City != "" {
		dbFilter["city"] = filter.City
	}
	if filter.Country != "" {
		dbFilter["country"] = filter.Country
	}
	if filter.IsFeatured != nil {
		dbFilter["is_featured"] = *filter.IsFeatured
	}
	if filter.IsFree != nil {
		dbFilter["is_free"] = *filter.IsFree
	}
	if filter.Search != "" {
		dbFilter["search"] = filter.Search
	}

	// Configurar paginación
	limit := pagination.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (pagination.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	events, total, err := s.eventRepo.List(ctx, dbFilter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}

	return events, total, nil
}

// GetEventStats obtiene estadísticas de un evento
func (s *EventService) GetEventStats(ctx context.Context, eventID string) (*dto.EventStatsResponse, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Obtener tipos de ticket activos
	ticketTypes, err := s.ticketTypeRepo.FindByEvent(ctx, event.ID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket types: %w", err)
	}

	var ticketsSold, totalRevenue float64
	var totalCapacity int64

	for _, tt := range ticketTypes {
		ticketsSold += float64(tt.SoldQuantity)
		totalRevenue += float64(tt.SoldQuantity) * tt.BasePrice
		totalCapacity += int64(tt.TotalQuantity)
	}

	avgTicketPrice := 0.0
	if ticketsSold > 0 {
		avgTicketPrice = totalRevenue / ticketsSold
	}

	// Tickets disponibles = capacidad total - vendidos
	ticketsAvailable := totalCapacity - int64(ticketsSold)
	if ticketsAvailable < 0 {
		ticketsAvailable = 0
	}

	return &dto.EventStatsResponse{
		TicketsSold:      int64(ticketsSold),
		TicketsAvailable: ticketsAvailable,
		TotalRevenue:     totalRevenue,
		AvgTicketPrice:   avgTicketPrice,
		CheckInRate:      0.0, // Requiere consulta a ticketRepo
	}, nil
}

// ============================================================================
// FUNCIONES HELPER PRIVADAS
// ============================================================================

// intPtrFromInt32 convierte int32 a *int
func intPtrFromInt32(val int32) *int {
	if val == 0 {
		return nil
	}
	result := int(val)
	return &result
}

// intPtrFromInt32Ptr convierte *int32 a *int
func intPtrFromInt32Ptr(val *int32) *int {
	if val == nil || *val == 0 {
		return nil
	}
	result := int(*val)
	return &result
}

// isValidEventStatusTransition valida transiciones de estado de evento
func isValidEventStatusTransition(current, new string) bool {
	transitions := map[string][]string{
		string(enums.EventStatusDraft):     {string(enums.EventStatusScheduled), string(enums.EventStatusPublished), string(enums.EventStatusCancelled)},
		string(enums.EventStatusScheduled): {string(enums.EventStatusPublished), string(enums.EventStatusCancelled)},
		string(enums.EventStatusPublished): {string(enums.EventStatusLive), string(enums.EventStatusCancelled), string(enums.EventStatusCompleted)},
		string(enums.EventStatusLive):      {string(enums.EventStatusCompleted), string(enums.EventStatusCancelled)},
		string(enums.EventStatusCompleted): {},
		string(enums.EventStatusCancelled): {},
		string(enums.EventStatusSoldOut):   {string(enums.EventStatusLive), string(enums.EventStatusCompleted)},
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
