// internal/application/services/category_service.go
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	categorydto "github.com/franciscozamorau/osmi-server/internal/api/dto/category"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

type CategoryService struct {
	categoryRepo repository.CategoryRepository
	eventRepo    repository.EventRepository
}

func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	eventRepo repository.EventRepository,
) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		eventRepo:    eventRepo,
	}
}

// ============================================================================
// FUNCIÓN HELPER PARA GENERAR SLUGS ÚNICOS
// ============================================================================

// generateUniqueSlug genera un slug único basado en el nombre y slugs existentes
func generateUniqueSlug(name string, existingSlugs []string) string {
	// Convertir a slug básico (ej: "Zona VIP" -> "zona-vip")
	baseSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	// Limpiar caracteres especiales (solo letras, números y guiones)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	baseSlug = re.ReplaceAllString(baseSlug, "")

	// Si después de limpiar queda vacío, usar un default
	if baseSlug == "" {
		baseSlug = "categoria"
	}

	slug := baseSlug
	counter := 1

	// Verificar si el slug ya existe
	for {
		exists := false
		for _, existing := range existingSlugs {
			if existing == slug {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		counter++
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
	}

	return slug
}

// ============================================================================
// MÉTODOS PRINCIPALES
// ============================================================================

// CreateCategory maneja la creación de una nueva categoría
func (s *CategoryService) CreateCategory(ctx context.Context, req *categorydto.CreateCategoryRequest) (*entities.Category, error) {
	// Obtener slugs existentes para generar uno único
	categories, _, err := s.categoryRepo.Find(ctx, &repository.CategoryFilter{
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing categories: %w", err)
	}

	// Extraer slugs existentes
	existingSlugs := make([]string, 0, len(categories))
	for _, cat := range categories {
		existingSlugs = append(existingSlugs, cat.Slug)
	}

	// Generar slug único
	slug := generateUniqueSlug(req.Name, existingSlugs)

	// Determinar nivel y padre
	var parentID *int64
	level := 1

	if req.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found with ID: %d", *req.ParentID)
		}
		parentID = &parent.ID
		level = parent.Level + 1
	}

	// Manejo de campos opcionales
	description := ""
	if req.Description != "" {
		description = req.Description
	}

	icon := ""
	if req.Icon != "" {
		icon = req.Icon
	}

	metaTitle := ""
	if req.MetaTitle != "" {
		metaTitle = req.MetaTitle
	}

	metaDescription := ""
	if req.MetaDescription != "" {
		metaDescription = req.MetaDescription
	}

	// Valores por defecto
	req.SetDefaults()

	// Crear entidad
	category := &entities.Category{
		Name:             req.Name,
		Slug:             slug,
		Description:      &description,
		Icon:             &icon,
		ColorHex:         req.ColorHex,
		ParentID:         parentID,
		Level:            level,
		Path:             "",
		IsActive:         *req.IsActive,
		IsFeatured:       *req.IsFeatured,
		SortOrder:        *req.SortOrder,
		MetaTitle:        &metaTitle,
		MetaDescription:  &metaDescription,
		TotalEvents:      0,
		TotalTicketsSold: 0,
		TotalRevenue:     0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetCategory obtiene una categoría por su ID público
func (s *CategoryService) GetCategory(ctx context.Context, publicID string) (*entities.Category, error) {
	category, err := s.categoryRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", publicID)
	}
	return category, nil
}

// GetCategoryBySlug obtiene una categoría por su slug
func (s *CategoryService) GetCategoryBySlug(ctx context.Context, slug string) (*entities.Category, error) {
	category, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", slug)
	}
	return category, nil
}

// ListCategories lista categorías con filtros y paginación
func (s *CategoryService) ListCategories(ctx context.Context, filter *categorydto.CategoryFilter, page, pageSize int) ([]*entities.Category, int64, error) {
	// Crear filtro del repositorio
	repoFilter := &repository.CategoryFilter{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	if filter != nil {
		filter.SetDefaults()

		if filter.IsActive != nil {
			repoFilter.IsActive = filter.IsActive
		}
		if filter.IsFeatured != nil {
			repoFilter.IsFeatured = filter.IsFeatured
		}
		if filter.ParentID != nil {
			repoFilter.ParentID = filter.ParentID
		}
		if filter.MinLevel != nil {
			repoFilter.MinLevel = filter.MinLevel
		}
		if filter.MaxLevel != nil {
			repoFilter.MaxLevel = filter.MaxLevel
		}
		if filter.Search != "" {
			repoFilter.SearchTerm = &filter.Search
		}
		if filter.SortBy != "" {
			repoFilter.SortBy = filter.SortBy
		}
		if filter.SortOrder != "" {
			repoFilter.SortOrder = filter.SortOrder
		}
	}

	return s.categoryRepo.Find(ctx, repoFilter)
}

// UpdateCategory actualiza una categoría existente
func (s *CategoryService) UpdateCategory(ctx context.Context, publicID string, req *categorydto.UpdateCategoryRequest) (*entities.Category, error) {
	category, err := s.categoryRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", publicID)
	}

	// Actualizar campos
	if req.Name != nil {
		category.Name = *req.Name
	}

	// Manejar actualización de slug
	if req.Slug != nil && *req.Slug != category.Slug {
		if existing, _ := s.categoryRepo.GetBySlug(ctx, *req.Slug); existing != nil && existing.PublicID != publicID {
			return nil, fmt.Errorf("slug already exists: %s", *req.Slug)
		}
		category.Slug = *req.Slug
	}

	if req.Description != nil {
		category.Description = req.Description
	}
	if req.Icon != nil {
		category.Icon = req.Icon
	}
	if req.ColorHex != nil {
		category.ColorHex = *req.ColorHex
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.IsFeatured != nil {
		category.IsFeatured = *req.IsFeatured
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.MetaTitle != nil {
		category.MetaTitle = req.MetaTitle
	}
	if req.MetaDescription != nil {
		category.MetaDescription = req.MetaDescription
	}

	// Actualizar ParentID
	if req.ParentID != nil {
		if *req.ParentID == 0 {
			category.ParentID = nil
			category.Level = 1
		} else {
			parent, err := s.categoryRepo.GetByID(ctx, *req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("parent category not found with ID: %d", *req.ParentID)
			}
			category.ParentID = &parent.ID
			category.Level = parent.Level + 1
		}
	}

	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// DeleteCategory elimina (desactiva) una categoría
func (s *CategoryService) DeleteCategory(ctx context.Context, publicID string) error {
	category, err := s.categoryRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("category not found: %s", publicID)
	}

	// Verificar si tiene eventos asociados
	events, err := s.categoryRepo.GetEventCategories(ctx, category.ID)
	if err == nil && len(events) > 0 {
		return fmt.Errorf("cannot delete category with %d associated events", len(events))
	}

	// Soft delete
	category.IsActive = false
	category.UpdatedAt = time.Now()
	return s.categoryRepo.Update(ctx, category)
}

// AddEventToCategory asocia un evento a una categoría
func (s *CategoryService) AddEventToCategory(ctx context.Context, eventID, categoryID string, isPrimary bool) error {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	category, err := s.categoryRepo.GetByPublicID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %s", categoryID)
	}

	// Actualizar contadores
	if isPrimary {
		category.TotalEvents++
		if err := s.categoryRepo.Update(ctx, category); err != nil {
			return fmt.Errorf("failed to update category event count: %w", err)
		}
	}

	return s.categoryRepo.AddEventToCategory(ctx, event.ID, category.ID, isPrimary)
}

// GetEventCategories obtiene las categorías de un evento
func (s *CategoryService) GetEventCategories(ctx context.Context, eventID string) ([]*entities.Category, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return s.categoryRepo.GetEventCategories(ctx, event.ID)
}
