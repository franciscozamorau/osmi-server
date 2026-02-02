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

	// Validar categoría primaria si se proporciona
	var primaryCategoryID *int64
	if req.PrimaryCategoryID != "" {
		category, err := s.categoryRepo.FindByPublicID(ctx, req.PrimaryCategoryID)
		if err != nil {
			return nil, errors.New("primary category not found")
		}
		primaryCategoryID = &category.ID
	}

	// Parsear fechas
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}

	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		return nil, errors.New("invalid end date format")
	}

	// Validar que endsAt sea después de startsAt
	if endsAt.Before(startsAt) {
		return nil, errors.New("end date must be after start date")
	}

	// Parsear fechas opcionales
	var doorsOpenAt, doorsCloseAt *time.Time
	if req.DoorsOpenAt != "" {
		openAt, err := time.Parse(time.RFC3339, req.DoorsOpenAt)
		if err != nil {
			return nil, errors.New("invalid doors open date format")
		}
		doorsOpenAt = &openAt
	}

	if req.DoorsCloseAt != "" {
		closeAt, err := time.Parse(time.RFC3339, req.DoorsCloseAt)
		if err != nil {
			return nil, errors.New("invalid doors close date format")
		}
		doorsCloseAt = &closeAt
	}

	// Crear evento
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
		Timezone:            req.Timezone,
		StartsAt:            startsAt,
		EndsAt:              endsAt,
		DoorsOpenAt:         doorsOpenAt,
		DoorsCloseAt:        doorsCloseAt,
		VenueName:           &req.VenueName,
		AddressFull:         &req.AddressFull,
		City:                &req.City,
		State:               &req.State,
		Country:             &req.Country,
		Status:              string(enums.EventStatusDraft),
		Visibility:          req.Visibility,
		IsFeatured:          req.IsFeatured,
		IsFree:              req.IsFree,
		MaxAttendees:        &req.MaxAttendees,
		MinAttendees:        req.MinAttendees,
		Tags:                req.Tags,
		AgeRestriction:      &req.AgeRestriction,
		RequiresApproval:    req.RequiresApproval,
		AllowReservations:   req.AllowReservations,
		ReservationDuration: req.ReservationDuration,
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
			category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
			if err != nil {
				continue // Saltar categorías no encontradas
			}

			eventCategory := &entities.EventCategory{
				EventID:    event.ID,
				CategoryID: category.ID,
				IsPrimary:  primaryCategoryID != nil && *primaryCategoryID == category.ID,
				CreatedAt:  time.Now(),
			}

			err = s.categoryRepo.AddEventToCategory(ctx, eventCategory)
			if err != nil {
				// Log error pero continuar
			}
		}
	}

	return event, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, eventID string, req *dto.UpdateEventRequest) (*entities.Event, error) {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Validar que se pueda modificar
	if event.Status == string(enums.EventStatusCompleted) || event.Status == string(enums.EventStatusCancelled) {
		return nil, errors.New("cannot modify completed or cancelled event")
	}

	// Actualizar campos
	if req.Name != "" {
		event.Name = req.Name
	}
	if req.ShortDescription != "" {
		event.ShortDescription = &req.ShortDescription
	}
	if req.Description != "" {
		event.Description = &req.Description
	}
	if req.EventType != "" {
		event.EventType = req.EventType
	}
	if req.CoverImageURL != "" {
		event.CoverImageURL = &req.CoverImageURL
	}
	if req.BannerImageURL != "" {
		event.BannerImageURL = &req.BannerImageURL
	}
	if req.Timezone != "" {
		event.Timezone = req.Timezone
	}
	if req.StartsAt != "" {
		startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
		if err != nil {
			return nil, errors.New("invalid start date format")
		}
		event.StartsAt = startsAt
	}
	if req.EndsAt != "" {
		endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			return nil, errors.New("invalid end date format")
		}
		event.EndsAt = endsAt
	}
	if req.Status != "" {
		event.Status = req.Status
	}
	if req.Visibility != "" {
		event.Visibility = req.Visibility
	}
	if req.IsFeatured != nil {
		event.IsFeatured = *req.IsFeatured
	}
	if req.MaxAttendees != nil {
		event.MaxAttendees = req.MaxAttendees
	}
	if req.AgeRestriction != nil {
		event.AgeRestriction = req.AgeRestriction
	}
	if req.Tags != nil {
		event.Tags = req.Tags
	}

	event.UpdatedAt = time.Now()

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) PublishEvent(ctx context.Context, eventID string, publishAt *time.Time) (*entities.Event, error) {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Validar que el evento esté en estado draft o scheduled
	if event.Status != string(enums.EventStatusDraft) && event.Status != string(enums.EventStatusScheduled) {
		return nil, errors.New("event is not in draft or scheduled state")
	}

	// Validar que tenga al menos un tipo de ticket
	ticketTypes, err := s.ticketTypeRepo.FindByEventID(ctx, event.ID)
	if err != nil || len(ticketTypes) == 0 {
		return nil, errors.New("event must have at least one ticket type to be published")
	}

	// Actualizar estado
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
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Validar que se pueda cancelar
	if event.Status == string(enums.EventStatusCompleted) || event.Status == string(enums.EventStatusCancelled) {
		return nil, errors.New("event is already completed or cancelled")
	}

	// Verificar si hay tickets vendidos
	ticketTypes, err := s.ticketTypeRepo.FindByEventID(ctx, event.ID)
	if err == nil {
		for _, tt := range ticketTypes {
			if tt.SoldQuantity > 0 {
				return nil, errors.New("cannot cancel event with sold tickets")
			}
		}
	}

	// Cancelar evento
	event.Status = string(enums.EventStatusCancelled)
	event.UpdatedAt = time.Now()

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, eventID string) (*entities.Event, error) {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Incrementar contador de vistas
	go s.eventRepo.IncrementViewCount(context.Background(), event.ID)

	return event, nil
}

func (s *EventService) ListEvents(ctx context.Context, filter dto.EventFilter, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	events, err := s.eventRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.eventRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (s *EventService) GetEventStats(ctx context.Context, eventID string) (*dto.EventStatsResponse, error) {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Obtener tipos de ticket para calcular estadísticas
	ticketTypes, err := s.ticketTypeRepo.FindByEventID(ctx, event.ID)
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
		TicketsSold:      int(ticketsSold),
		TicketsAvailable: int(ticketsAvailable),
		TotalRevenue:     totalRevenue,
		AvgTicketPrice:   avgTicketPrice,
		CheckInRate:      0.0, // TODO: Calcular del registro de check-ins
		ConversionRate:   0.0, // TODO: Calcular de analytics
		ViewsToday:       event.ViewCount,
		SalesVelocity:    0.0, // TODO: Calcular de analytics
	}

	return stats, nil
}
