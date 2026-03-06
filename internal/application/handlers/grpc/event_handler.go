// internal/application/handlers/grpc/event_handler.go
package grpc

import (
	"context"
	"log"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/api/helpers"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventHandler struct {
	osmi.UnimplementedOsmiServiceServer
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// ============================================================================
// MÉTODOS PRINCIPALES
//============================================================================

// CreateEvent maneja la creación de un nuevo evento
func (h *EventHandler) CreateEvent(ctx context.Context, req *osmi.CreateEventRequest) (*osmi.EventResponse, error) {
	// LOG 1: Entrada a la función
	log.Println("🎯 EVENT_HANDLER: CreateEvent ENTRÓ a la función")
	log.Printf("🎯 EVENT_HANDLER: req type: %T", req)
	log.Printf("🎯 EVENT_HANDLER: req value: %+v", req)

	// LOG 2: Validar campos requeridos UNO POR UNO
	log.Printf("🎯 Validando Name: %q", req.Name)
	if req.Name == "" {
		log.Println("🎯 ERROR: Name vacío")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	log.Printf("🎯 Validando OrganizerId: %q", req.OrganizerId)
	if req.OrganizerId == "" {
		log.Println("🎯 ERROR: OrganizerId vacío")
		return nil, status.Error(codes.InvalidArgument, "organizer_id is required")
	}

	log.Printf("🎯 Validando StartDate: %q", req.StartDate)
	if req.StartDate == "" {
		log.Println("🎯 ERROR: StartDate vacío")
		return nil, status.Error(codes.InvalidArgument, "start_date is required")
	}

	log.Printf("🎯 Validando EndDate: %q", req.EndDate)
	if req.EndDate == "" {
		log.Println("🎯 ERROR: EndDate vacío")
		return nil, status.Error(codes.InvalidArgument, "end_date is required")
	}

	// LOG 3: Parsear fechas
	log.Println("🎯 Parseando fechas...")
	startsAt, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		log.Printf("🎯 ERROR parseando start_date: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid start_date format (use RFC3339)")
	}

	endsAt, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		log.Printf("🎯 ERROR parseando end_date: %v", err)
		return nil, status.Error(codes.InvalidArgument, "invalid end_date format (use RFC3339)")
	}
	log.Printf("🎯 Fechas parseadas: startsAt=%v, endsAt=%v", startsAt, endsAt)

	// LOG 4: Crear DTO
	log.Println("🎯 Creando DTO...")
	createReq := &dto.CreateEventRequest{
		Name:                req.Name,
		Slug:                req.Name,
		Description:         req.Description,
		ShortDescription:    req.ShortDescription,
		OrganizerID:         req.OrganizerId,
		VenueID:             req.VenueId,
		PrimaryCategoryID:   req.PrimaryCategoryId,
		CategoryIDs:         req.CategoryIds,
		StartsAt:            startsAt,
		EndsAt:              endsAt,
		DoorsOpenAt:         time.Time{},
		DoorsCloseAt:        time.Time{},
		Timezone:            req.Timezone,
		EventType:           req.EventType,
		CoverImageURL:       req.CoverImageUrl,
		BannerImageURL:      req.BannerImageUrl,
		VenueName:           req.VenueName,
		AddressFull:         req.AddressFull,
		City:                req.City,
		State:               req.State,
		Country:             req.Country,
		Visibility:          req.Visibility,
		IsFeatured:          req.IsFeatured,
		IsFree:              req.IsFree,
		MaxAttendees:        req.MaxAttendees,
		MinAttendees:        req.MinAttendees,
		Tags:                req.Tags,
		AgeRestriction:      req.AgeRestriction,
		RequiresApproval:    req.RequiresApproval,
		AllowReservations:   req.AllowReservations,
		ReservationDuration: req.ReservationDuration,
	}
	log.Printf("🎯 DTO creado: %+v", createReq)

	// LOG 5: Llamar al servicio
	log.Println("🎯 Llamando a eventService.CreateEvent...")
	event, err := h.eventService.CreateEvent(ctx, createReq)
	if err != nil {
		log.Printf("🎯 ERROR en eventService.CreateEvent: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	log.Printf("🎯 eventService.CreateEvent OK, event: %+v", event)

	// LOG 6: Convertir a proto
	log.Println("🎯 Convirtiendo a proto...")
	resp := h.eventToProto(event)
	log.Printf("🎯 Respuesta preparada: %+v", resp)

	return resp, nil
}

// GetEvent obtiene un evento por su ID
func (h *EventHandler) GetEvent(ctx context.Context, req *osmi.EventLookup) (*osmi.EventResponse, error) {
	if req.PublicId == "" {
		return nil, status.Error(codes.InvalidArgument, "event public_id is required")
	}

	event, err := h.eventService.GetEvent(ctx, req.PublicId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return h.eventToProto(event), nil
}

// ListEvents lista eventos con filtros y paginación
func (h *EventHandler) ListEvents(ctx context.Context, req *osmi.ListEventsRequest) (*osmi.EventListResponse, error) {
	// Convertir filtros
	filter := dto.EventFilter{
		Name:       req.Name,
		Status:     req.Status,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		City:       req.City,
		Country:    req.Country,
		IsFeatured: &req.IsFeatured,
		IsFree:     &req.IsFree,
		Search:     req.Search,
	}

	// Usar GetOrganizerId() para obtener el valor
	organizerId := req.GetOrganizerId()
	if organizerId != "" {
		// Nota: Necesitarías convertir organizer_id a int64
		// Por ahora se omite
	}

	if req.CategoryId != "" {
		// Nota: Necesitarías convertir category_id a int64
		// Por ahora se omite
	}

	// Paginación
	pagination := dto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}

	// Llamar al servicio
	events, total, err := h.eventService.ListEvents(ctx, filter, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convertir entidades a protobuf
	pbEvents := make([]*osmi.EventResponse, len(events))
	for i, event := range events {
		pbEvents[i] = h.eventToProto(event)
	}

	// Calcular total de páginas
	totalPages := int32(0)
	if pagination.PageSize > 0 {
		totalPages = int32((int(total) + pagination.PageSize - 1) / pagination.PageSize)
	}

	return &osmi.EventListResponse{
		Events:     pbEvents,
		TotalCount: int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
		TotalPages: totalPages,
	}, nil
}

// UpdateEvent actualiza un evento existente
func (h *EventHandler) UpdateEvent(ctx context.Context, req *osmi.UpdateEventRequest) (*osmi.EventResponse, error) {
	if req.PublicId == "" {
		return nil, status.Error(codes.InvalidArgument, "event public_id is required")
	}

	// Convertir protobuf a DTO
	updateReq := &dto.UpdateEventRequest{
		Name:             req.Name,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		Status:           req.Status,
		Visibility:       req.Visibility,
		IsFeatured:       req.IsFeatured,
		IsPublished:      req.IsPublished,
	}

	// Parsear fechas si vienen
	if req.StartDate != nil {
		startsAt, err := time.Parse(time.RFC3339, *req.StartDate)
		if err == nil {
			updateReq.StartsAt = &startsAt
		}
	}
	if req.EndDate != nil {
		endsAt, err := time.Parse(time.RFC3339, *req.EndDate)
		if err == nil {
			updateReq.EndsAt = &endsAt
		}
	}
	if req.MaxAttendees != nil {
		updateReq.MaxAttendees = req.MaxAttendees
	}
	if req.AgeRestriction != nil {
		updateReq.AgeRestriction = req.AgeRestriction
	}

	// Llamar al servicio
	event, err := h.eventService.UpdateEvent(ctx, req.PublicId, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.eventToProto(event), nil
}

// ============================================================================
// FUNCIÓN HELPER PARA CONVERSIÓN
// ============================================================================

// eventToProto convierte una entidad Event a protobuf EventResponse
func (h *EventHandler) eventToProto(event *entities.Event) *osmi.EventResponse {
	if event == nil {
		return nil
	}

	// Construir respuesta base
	resp := &osmi.EventResponse{
		PublicId:         event.PublicID,
		Name:             event.Name,
		Description:      helpers.SafeStringPtr(event.Description),
		ShortDescription: helpers.SafeStringPtr(event.ShortDescription),
		StartDate:        event.StartsAt.Format(time.RFC3339),
		EndDate:          event.EndsAt.Format(time.RFC3339),
		Location:         helpers.SafeStringPtr(event.VenueName),
		VenueDetails:     helpers.SafeStringPtr(event.AddressFull),
		Category:         "", // Requiere consulta adicional
		Tags:             []string{},
		IsActive:         event.Status != "cancelled" && event.Status != "archived",
		IsPublished:      event.Status == "published" || event.Status == "live",
		ImageUrl:         helpers.SafeStringPtr(event.CoverImageURL),
		BannerUrl:        helpers.SafeStringPtr(event.BannerImageURL),
		CreatedAt:        timestamppb.New(event.CreatedAt),
		UpdatedAt:        timestamppb.New(event.UpdatedAt),
	}

	// Agregar tags si existen
	if event.Tags != nil {
		resp.Tags = *event.Tags
	}

	// Agregar max_attendees si existe
	if event.MaxAttendees != nil {
		resp.MaxAttendees = int32(*event.MaxAttendees)
	}

	// Nota: total_tickets y tickets_sold requieren consultas adicionales
	resp.TotalTickets = 0
	resp.TicketsSold = 0

	return resp
}
