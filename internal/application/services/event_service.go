package services

import (
	"context"
	"encoding/json"
	"errors"
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

func (s *EventService) CreateEvent(ctx context.Context, req *dto.CreateEventRequest) (*entities.Event, error) {
	// Validar organizador
	organizer, err := s.organizerRepo.FindByPublicID(ctx, req.OrganizerID)
	if err != nil {
		return nil, errors.New("organizer not found")
	}

	// Validar venue si se proporciona
	var venueID *int64
	if req.VenueID != "" {
		venue, err := s.venueRepo.FindByPublicID(ctx, req.VenueID)
		if err != nil {
			return nil, errors.New("venue not found")
		}
		venueID = &venue.ID
	}

	// Validar categoría primaria
	var primaryCategoryID *int64
	if req.PrimaryCategoryID != "" {
		category, err := s.categoryRepo.GetByPublicID(ctx, req.PrimaryCategoryID)
		if err != nil {
			return nil, errors.New("primary category not found")
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
		if err := json.Unmarshal([]byte(req.Tags), &tagsSlice); err != nil {
			// Si no es JSON válido, tratar como string único
			tagsSlice = []string{req.Tags}
		}
		tags = &tagsSlice
	}

	// Crear evento con conversiones de tipos correctas
	event := &entities.Event{
		PublicID:          uuid.New().String(),
		OrganizerID:       organizer.ID,
		PrimaryCategoryID: primaryCategoryID,
		VenueID:           venueID,
		Name:              req.Name,
		Slug:              req.Slug,
		ShortDescription:  &req.ShortDescription,
		Description:       &req.Description,
		EventType:         req.EventType,
		CoverImageURL:     &req.CoverImageURL,
		BannerImageURL:    &req.BannerImageURL,
		GalleryImages:     nil,
		Timezone:          req.Timezone,
		StartsAt:          req.StartsAt,
		EndsAt:            req.EndsAt,
		DoorsOpenAt:       &req.DoorsOpenAt,
		DoorsCloseAt:      &req.DoorsCloseAt,
		VenueName:         &req.VenueName,
		AddressFull:       &req.AddressFull,
		City:              &req.City,
		State:             &req.State,
		Country:           &req.Country,
		Status:            string(enums.EventStatusDraft),
		Visibility:        req.Visibility,
		IsFeatured:        req.IsFeatured,
		IsFree:            req.IsFree,
		// CORREGIDO (Línea 112): Convertir int32 a *int32 para la función helper
		MaxAttendees: convertInt32PtrToIntPtr(&req.MaxAttendees),
		// CORREGIDO: Convertir int32 a int
		MinAttendees: int(req.MinAttendees),
		// CORREGIDO: Usar el slice procesado
		Tags: tags,
		// CORREGIDO (Línea 118): Convertir int32 a *int32 para la función helper
		AgeRestriction:    convertInt32PtrToIntPtr(&req.AgeRestriction),
		RequiresApproval:  req.RequiresApproval,
		AllowReservations: req.AllowReservations,
		// CORREGIDO: Convertir int32 a int
		ReservationDuration: int(req.ReservationDuration),
		Settings:            nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err = s.eventRepo.Create(ctx, event)
	if err != nil {
		return nil, err
	}

	// Asociar categorías si se proporcionan
	if len(req.CategoryIDs) > 0 {
		for _, categoryID := range req.CategoryIDs {
			category, err := s.categoryRepo.GetByPublicID(ctx, categoryID)
			if err != nil {
				continue
			}
			err = s.eventRepo.AddCategoryToEvent(ctx, event.ID, category.ID, primaryCategoryID != nil && *primaryCategoryID == category.ID)
			if err != nil {
				// Log error pero continuar
			}
		}
	}

	return event, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, eventID string, req *dto.UpdateEventRequest) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	if event.Status == string(enums.EventStatusCompleted) || event.Status == string(enums.EventStatusCancelled) {
		return nil, errors.New("cannot modify completed or cancelled event")
	}

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
		event.Status = *req.Status
	}
	if req.Visibility != nil {
		event.Visibility = *req.Visibility
	}
	if req.IsFeatured != nil {
		event.IsFeatured = *req.IsFeatured
	}
	// CORREGIDO: Convertir *int32 a *int
	if req.MaxAttendees != nil {
		val := int(*req.MaxAttendees)
		event.MaxAttendees = &val
	}
	// CORREGIDO: Convertir *int32 a *int
	if req.AgeRestriction != nil {
		val := int(*req.AgeRestriction)
		event.AgeRestriction = &val
	}
	if req.StartsAt != nil {
		event.StartsAt = *req.StartsAt
	}
	if req.EndsAt != nil {
		event.EndsAt = *req.EndsAt
	}
	// CORREGIDO (Línea 195-199): Eliminado bloque de código que usaba req.Tags (no existe)

	event.UpdatedAt = time.Now()

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) PublishEvent(ctx context.Context, eventID string, publishAt *time.Time) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	if event.Status != string(enums.EventStatusDraft) && event.Status != string(enums.EventStatusScheduled) {
		return nil, errors.New("event is not in draft or scheduled state")
	}

	ticketTypes, err := s.ticketTypeRepo.FindByEvent(ctx, event.ID, true)
	if err != nil || len(ticketTypes) == 0 {
		return nil, errors.New("event must have at least one ticket type to be published")
	}

	event.Status = string(enums.EventStatusPublished)
	if publishAt != nil {
		event.PublishedAt = publishAt
	} else {
		now := time.Now()
		event.PublishedAt = &now
	}
	event.UpdatedAt = time.Now()

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) CancelEvent(ctx context.Context, eventID string, reason string) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	if event.Status == string(enums.EventStatusCompleted) || event.Status == string(enums.EventStatusCancelled) {
		return nil, errors.New("event is already completed or cancelled")
	}

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

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, eventID string) (*entities.Event, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	event.ViewCount++
	event.UpdatedAt = time.Now()
	_ = s.eventRepo.Update(ctx, event) // Ignoramos error, no crítico

	return event, nil
}

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

	// Calcular offset
	limit := pagination.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (pagination.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Llamar a List con los parámetros correctos
	events, total, err := s.eventRepo.List(ctx, dbFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (s *EventService) GetEventStats(ctx context.Context, eventID string) (*dto.EventStatsResponse, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	ticketTypes, err := s.ticketTypeRepo.FindByEvent(ctx, event.ID, true)
	if err != nil {
		return nil, err
	}

	var ticketsSold, ticketsAvailable, totalRevenue float64
	for _, tt := range ticketTypes {
		ticketsSold += float64(tt.SoldQuantity)
		ticketsAvailable += float64(tt.AvailableQuantity)
		totalRevenue += float64(tt.SoldQuantity) * tt.BasePrice
	}

	avgTicketPrice := 0.0
	if ticketsSold > 0 {
		avgTicketPrice = totalRevenue / ticketsSold
	}

	stats := &dto.EventStatsResponse{
		TicketsSold:      int64(ticketsSold),
		TicketsAvailable: int64(ticketsAvailable),
		TotalRevenue:     totalRevenue,
		AvgTicketPrice:   avgTicketPrice,
		CheckInRate:      0.0,
	}

	return stats, nil
}

// Función helper para convertir *int32 a *int
func convertInt32PtrToIntPtr(val *int32) *int {
	if val == nil {
		return nil
	}
	result := int(*val)
	return &result
}
