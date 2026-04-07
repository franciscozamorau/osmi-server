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

// generateUniqueSlugForEvent genera un slug único basado en el nombre y slugs existentes del evento
func (s *CategoryService) generateUniqueSlugForEvent(ctx context.Context, eventID string, name string) (string, error) {
	existingCategories, err := s.categoryRepo.GetByEventID(ctx, eventID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get existing categories: %w", err)
	}

	existingSlugs := make([]string, 0, len(existingCategories))
	for _, cat := range existingCategories {
		existingSlugs = append(existingSlugs, cat.Slug)
	}

	baseSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	re := regexp.MustCompile(`[^a-z0-9-]`)
	baseSlug = re.ReplaceAllString(baseSlug, "")

	if baseSlug == "" {
		baseSlug = "categoria"
	}

	slug := baseSlug
	counter := 1

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

	return slug, nil
}

// CreateCategory maneja la creación de una nueva categoría para un evento específico
func (s *CategoryService) CreateCategory(ctx context.Context, req *categorydto.CreateCategoryRequest) (*entities.Category, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, req.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %s", req.EventID)
	}

	existingCategories, err := s.categoryRepo.GetByEventID(ctx, event.PublicID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing categories: %w", err)
	}

	for _, cat := range existingCategories {
		if cat.Name == req.Name {
			return nil, fmt.Errorf("category with name '%s' already exists for this event", req.Name)
		}
	}

	slug, err := s.generateUniqueSlugForEvent(ctx, event.PublicID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate slug: %w", err)
	}

	var parentID *int64
	level := 1

	if req.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found with ID: %d", *req.ParentID)
		}
		if parent.EventID != event.PublicID {
			return nil, fmt.Errorf("parent category does not belong to this event")
		}
		parentID = &parent.ID
		level = parent.Level + 1
	}

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

	req.SetDefaults()

	category := &entities.Category{
		EventID:          event.PublicID,
		Name:             req.Name,
		Slug:             slug,
		Description:      &description,
		Icon:             &icon,
		ColorHex:         req.ColorHex,
		ParentID:         parentID,
		Level:            level,
		Path:             "",
		Capacity:         0,
		TotalEvents:      0,
		TotalTicketsSold: 0,
		TotalRevenue:     0,
		IsActive:         *req.IsActive,
		IsFeatured:       *req.IsFeatured,
		SortOrder:        *req.SortOrder,
		MetaTitle:        &metaTitle,
		MetaDescription:  &metaDescription,
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

// GetCategoriesByEvent obtiene todas las categorías de un evento
func (s *CategoryService) GetCategoriesByEvent(ctx context.Context, eventID string, isActive *bool) ([]*entities.Category, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return s.categoryRepo.GetByEventID(ctx, event.PublicID, isActive)
}

// ListCategories lista categorías con filtros y paginación
func (s *CategoryService) ListCategories(ctx context.Context, filter *categorydto.CategoryFilter, page, pageSize int) ([]*entities.Category, int64, error) {
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

	if req.Name != nil && *req.Name != category.Name {
		existingCategories, err := s.categoryRepo.GetByEventID(ctx, category.EventID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing categories: %w", err)
		}
		for _, cat := range existingCategories {
			if cat.Name == *req.Name && cat.PublicID != publicID {
				return nil, fmt.Errorf("category with name '%s' already exists for this event", *req.Name)
			}
		}
		category.Name = *req.Name
	}

	if req.Slug != nil && *req.Slug != category.Slug {
		existingCategories, err := s.categoryRepo.GetByEventID(ctx, category.EventID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing slugs: %w", err)
		}
		for _, cat := range existingCategories {
			if cat.Slug == *req.Slug && cat.PublicID != publicID {
				return nil, fmt.Errorf("slug '%s' already exists for this event", *req.Slug)
			}
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

	if req.ParentID != nil {
		if *req.ParentID == 0 {
			category.ParentID = nil
			category.Level = 1
		} else {
			parent, err := s.categoryRepo.GetByID(ctx, *req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("parent category not found with ID: %d", *req.ParentID)
			}
			if parent.EventID != category.EventID {
				return nil, fmt.Errorf("parent category does not belong to this event")
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

	children, err := s.categoryRepo.GetByEventID(ctx, category.EventID, nil)
	if err == nil {
		for _, child := range children {
			if child.ParentID != nil && *child.ParentID == category.ID {
				return fmt.Errorf("cannot delete category with child categories")
			}
		}
	}

	category.IsActive = false
	category.UpdatedAt = time.Now()
	return s.categoryRepo.Update(ctx, category)
}
