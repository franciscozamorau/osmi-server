// internal/application/handlers/grpc/event_handler.go
package grpc

import (
	"context"
	"log"
	"strings"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	eventdto "github.com/franciscozamorau/osmi-server/internal/api/dto/event"
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
	log.Println("🎯 EVENT_HANDLER: CreateEvent ENTRÓ a la función")
	log.Printf("🎯 EVENT_HANDLER: req type: %T", req)
	log.Printf("🎯 EVENT_HANDLER: req value: %+v", req)

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

	// Procesar Tags (de string a []string)
	var tags []string
	if req.Tags != "" {
		tags = strings.Split(req.Tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	log.Println("🎯 Creando DTO...")
	createReq := &eventdto.CreateEventRequest{
		Name:                req.Name,
		Slug:                req.Name,
		Description:         req.Description,
		ShortDescription:    req.ShortDescription,
		OrganizerID:         req.OrganizerId,
		VenueID:             req.VenueId,
		PrimaryCategoryID:   req.PrimaryCategoryId,
		CategoryIDs:         req.CategoryIds,
		StartsAt:            startsAt.Format(time.RFC3339),
		EndsAt:              endsAt.Format(time.RFC3339),
		DoorsOpenAt:         "",
		DoorsCloseAt:        "",
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
		MaxAttendees:        int(req.MaxAttendees),
		MinAttendees:        int(req.MinAttendees),
		Tags:                tags,
		AgeRestriction:      int(req.AgeRestriction),
		RequiresApproval:    req.RequiresApproval,
		AllowReservations:   req.AllowReservations,
		ReservationDuration: int(req.ReservationDuration),
	}
	log.Printf("🎯 DTO creado: %+v", createReq)

	log.Println("🎯 Llamando a eventService.CreateEvent...")
	event, err := h.eventService.CreateEvent(ctx, createReq)
	if err != nil {
		log.Printf("🎯 ERROR en eventService.CreateEvent: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	log.Printf("🎯 eventService.CreateEvent OK, event: %+v", event)

	log.Println("🎯 Convirtiendo a proto...")
	resp := h.eventToProto(event)
	log.Printf("🎯 Respuesta preparada: %+v", resp)

	return resp, nil
}

// GetEvent obtiene un evento por su ID público
func (h *EventHandler) GetEvent(ctx context.Context, req *osmi.GetEventRequest) (*osmi.EventResponse, error) {
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
	// ========================================================================
	// CRÍTICO: Solo crear punteros si el valor NO está vacío
	// Si está vacío, se envía nil para que PostgreSQL lo ignore
	// ========================================================================

	// Para eventStatus (renombrado para no chocar con el paquete status)
	var eventStatus *string
	if req.Status != "" {
		eventStatus = &req.Status
	}

	// Para DateFrom (opcional)
	var dateFrom *string
	if req.DateFrom != "" {
		dateFrom = &req.DateFrom
	}

	// Para DateTo (opcional)
	var dateTo *string
	if req.DateTo != "" {
		dateTo = &req.DateTo
	}

	// Para City (opcional)
	var city *string
	if req.City != "" {
		city = &req.City
	}

	// Para Country (opcional)
	var country *string
	if req.Country != "" {
		country = &req.Country
	}

	// Para OrganizerID (opcional)
	var organizerID *string
	if req.OrganizerId != "" {
		organizerID = &req.OrganizerId
	}

	// Para CategoryID (opcional)
	var categoryID *string
	if req.CategoryId != "" {
		categoryID = &req.CategoryId
	}

	// Construir filtro SOLO con valores no vacíos
	filter := eventdto.EventFilter{
		Search:      req.Name,
		Status:      eventStatus, // ✅ nil si viene vacío, renombrado para evitar conflicto
		DateFrom:    dateFrom,    // ✅ nil si viene vacío
		DateTo:      dateTo,      // ✅ nil si viene vacío
		City:        city,        // ✅ nil si viene vacío
		Country:     country,     // ✅ nil si viene vacío
		OrganizerID: organizerID, // ✅ nil si viene vacío
		CategoryID:  categoryID,  // ✅ nil si viene vacío
		IsFeatured:  &req.IsFeatured,
		IsFree:      &req.IsFree,
	}

	// Paginación
	pagination := commondto.Pagination{
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
	updateReq := &eventdto.UpdateEventRequest{
		Name:             req.Name,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		Status:           req.Status,
		Visibility:       req.Visibility,
		IsFeatured:       req.IsFeatured,
	}

	// Fechas - req.StartDate y req.EndDate son *string
	// Validar con != nil, no con != ""
	if req.StartDate != nil {
		updateReq.StartsAt = req.StartDate
	}
	if req.EndDate != nil {
		updateReq.EndsAt = req.EndDate
	}

	// Conversión de *int32 a *int
	if req.MaxAttendees != nil {
		val := int(*req.MaxAttendees)
		updateReq.MaxAttendees = &val
	}

	if req.AgeRestriction != nil {
		val := int(*req.AgeRestriction)
		updateReq.AgeRestriction = &val
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

	resp := &osmi.EventResponse{
		PublicId:         event.PublicID,
		Name:             event.Name,
		Description:      helpers.SafeStringPtr(event.Description),
		ShortDescription: helpers.SafeStringPtr(event.ShortDescription),
		StartDate:        event.StartsAt.Format(time.RFC3339),
		EndDate:          event.EndsAt.Format(time.RFC3339),
		Location:         helpers.SafeStringPtr(event.VenueName),
		VenueDetails:     helpers.SafeStringPtr(event.AddressFull),
		Category:         "",
		Tags:             []string{},
		IsActive:         event.Status != "cancelled" && event.Status != "archived",
		IsPublished:      event.Status == "published" || event.Status == "live",
		ImageUrl:         helpers.SafeStringPtr(event.CoverImageURL),
		BannerUrl:        helpers.SafeStringPtr(event.BannerImageURL),
		CreatedAt:        timestamppb.New(event.CreatedAt),
		UpdatedAt:        timestamppb.New(event.UpdatedAt),
	}

	if event.Tags != nil {
		resp.Tags = *event.Tags
	}

	if event.MaxAttendees != nil {
		resp.MaxAttendees = int32(*event.MaxAttendees)
	}

	resp.TotalTickets = 0
	resp.TicketsSold = 0

	return resp
}
