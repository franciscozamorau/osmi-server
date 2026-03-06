// internal/application/handlers/grpc/category_handler.go
package grpc

import (
	"context"
	"log"
	"strings"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto/request"
	"github.com/franciscozamorau/osmi-server/internal/api/helpers"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CategoryHandler struct {
	osmi.UnimplementedOsmiServiceServer
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// ============================================================================
// MÉTODOS QUE SÍ EXISTEN EN EL PROTO
// ============================================================================

// CreateCategory maneja la creación de una nueva categoría
// Nota: En tu proto, CreateCategoryRequest tiene campos como price, quantity_available, etc.
// que pertenecen a TicketType, no a Category. Esto es un desajuste conceptual.
func (h *CategoryHandler) CreateCategory(ctx context.Context, req *osmi.CreateCategoryRequest) (*osmi.CategoryResponse, error) {
	// Validaciones básicas
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	// Crear slug a partir del nombre
	slug := generateSlug(req.Name)

	// Valores por defecto
	isActive := true
	isFeatured := false
	sortOrder := 0

	createReq := &request.CreateCategoryRequest{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Icon:        "",
		ColorHex:    "#3498db",
		ParentID:    nil,
		IsActive:    &isActive,
		IsFeatured:  &isFeatured,
		SortOrder:   &sortOrder,
	}

	// Llamar al servicio
	// Después de crear la categoría (línea 54)
	category, err := h.categoryService.CreateCategory(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 🔴 NUEVO: Asociar la categoría al evento
	err = h.categoryService.AddEventToCategory(ctx, req.EventId, category.PublicID, false)
	if err != nil {
		// Si falla la asociación, podríamos eliminar la categoría o solo loguear
		// Por ahora, solo logueamos el error
		log.Printf("Warning: failed to associate category with event: %v", err)
	}

	// NOTA: El campo EventId en la respuesta podría no ser el más adecuado
	// ya que la relación es many-to-many. Por ahora lo dejamos como viene en el request.
	return h.categoryToResponse(category, req.EventId), nil
}

// GetEventCategories obtiene las categorías de un evento
func (h *CategoryHandler) GetEventCategories(ctx context.Context, req *osmi.EventLookup) (*osmi.CategoryListResponse, error) {
	if req.PublicId == "" {
		return nil, status.Error(codes.InvalidArgument, "event public_id is required")
	}

	// Llamar al servicio
	categories, err := h.categoryService.GetEventCategories(ctx, req.PublicId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir a respuesta
	pbCategories := make([]*osmi.CategoryResponse, len(categories))
	for i, category := range categories {
		// Nota: No tenemos event_id aquí, pero el proto lo requiere.
		// Usamos el public_id del evento que recibimos.
		pbCategories[i] = h.categoryToResponse(category, req.PublicId)
	}

	// Obtener nombre del evento (opcional, podrías obtenerlo del servicio)
	eventName := ""

	return &osmi.CategoryListResponse{
		Categories:    pbCategories,
		EventName:     eventName,
		EventPublicId: req.PublicId,
	}, nil
}

// ============================================================================
// MÉTODOS QUE NO EXISTEN EN EL PROTO (ELIMINADOS)
// ============================================================================
// Los siguientes métodos NO existen en OsmiServiceServer según tu proto:
// - GetCategory
// - ListCategories
// Por lo tanto, NO deben estar implementados aquí.

// ============================================================================
// FUNCIONES HELPER
// ============================================================================

// categoryToResponse convierte una entidad Category a proto CategoryResponse
// Recibe eventID porque el proto CategoryResponse lo requiere (aunque no sea lo ideal)
func (h *CategoryHandler) categoryToResponse(category *entities.Category, eventID string) *osmi.CategoryResponse {
	resp := &osmi.CategoryResponse{
		PublicId:           category.PublicID,
		EventId:            eventID,
		Name:               category.Name,
		Description:        helpers.SafeStringPtr(category.Description),
		Price:              0,   // No aplica - usar TicketType
		QuantityAvailable:  0,   // No aplica - usar TicketType
		QuantitySold:       0,   // No aplica - usar TicketType
		MaxTicketsPerOrder: 0,   // No aplica - usar TicketType
		SalesStart:         nil, // No aplica
		SalesEnd:           nil, // No aplica
		Benefits:           []string{},
		IsActive:           category.IsActive,
		CreatedAt:          timestamppb.New(category.CreatedAt),
		UpdatedAt:          timestamppb.New(category.UpdatedAt),
	}
	return resp
}

// generateSlug genera un slug simple (puede moverse a helpers si se usa en varios lugares)
func generateSlug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))

}
