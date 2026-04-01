// internal/application/handlers/grpc/category_handler.go
package grpc

import (
	"context"
	"log"
	"strings"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	categorydto "github.com/franciscozamorau/osmi-server/internal/api/dto/category"
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

	createReq := &categorydto.CreateCategoryRequest{
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
	category, err := h.categoryService.CreateCategory(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Asociar la categoría al evento
	err = h.categoryService.AddEventToCategory(ctx, req.EventId, category.PublicID, false)
	if err != nil {
		log.Printf("Warning: failed to associate category with event: %v", err)
	}

	return h.categoryToResponse(category, req.EventId), nil
}

// GetEventCategories obtiene las categorías de un evento
func (h *CategoryHandler) GetEventCategories(ctx context.Context, req *osmi.GetEventCategoriesRequest) (*osmi.CategoryListResponse, error) {
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
		pbCategories[i] = h.categoryToResponse(category, req.PublicId)
	}

	eventName := ""

	return &osmi.CategoryListResponse{
		Categories:    pbCategories,
		EventName:     eventName,
		EventPublicId: req.PublicId,
	}, nil
}

// ============================================================================
// FUNCIONES HELPER
// ============================================================================

// categoryToResponse convierte una entidad Category a proto CategoryResponse
func (h *CategoryHandler) categoryToResponse(category *entities.Category, eventID string) *osmi.CategoryResponse {
	resp := &osmi.CategoryResponse{
		PublicId:    category.PublicID,
		EventId:     eventID,
		Name:        category.Name,
		Description: helpers.SafeStringPtr(category.Description),
		IsActive:    category.IsActive,
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}
	return resp
}

// generateSlug genera un slug simple
func generateSlug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
