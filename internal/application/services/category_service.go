package services

import (
	"context"
	"fmt"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto/request"
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

func (s *CategoryService) CreateCategory(ctx context.Context, req *request.CreateCategoryRequest) (*entities.Category, error) {
	// Validar slug único
	if _, err := s.categoryRepo.GetBySlug(ctx, req.Slug); err == nil {
		return nil, fmt.Errorf("slug already exists: %s", req.Slug)
	}

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
		Slug:             req.Slug,
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

func (s *CategoryService) GetCategory(ctx context.Context, publicID string) (*entities.Category, error) {
	category, err := s.categoryRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", publicID)
	}
	return category, nil
}

func (s *CategoryService) GetCategoryBySlug(ctx context.Context, slug string) (*entities.Category, error) {
	category, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", slug)
	}
	return category, nil
}

func (s *CategoryService) ListCategories(ctx context.Context, filter *request.CategoryFilter, page, pageSize int) ([]*entities.Category, int64, error) {
	// CORREGIDO: Crear filtro del repositorio con los parámetros de paginación incluidos
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
		} else if filter.IncludeDescendants {
			zero := int64(0)
			repoFilter.ParentID = &zero // 0 significa categorías raíz
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

	// CORREGIDO: Find solo recibe ctx y repoFilter, la paginación ya está dentro del filtro
	return s.categoryRepo.Find(ctx, repoFilter)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, publicID string, req *request.UpdateCategoryRequest) (*entities.Category, error) {
	category, err := s.categoryRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", publicID)
	}

	// Actualizar campos
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Slug != nil && *req.Slug != category.Slug {
		if existing, _ := s.categoryRepo.GetBySlug(ctx, *req.Slug); existing != nil {
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

func (s *CategoryService) DeleteCategory(ctx context.Context, publicID string) error {
	category, err := s.categoryRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("category not found: %s", publicID)
	}

	// Verificar si tiene eventos
	events, err := s.categoryRepo.GetEventCategories(ctx, category.ID)
	if err == nil && len(events) > 0 {
		return fmt.Errorf("cannot delete category with %d associated events", len(events))
	}

	category.IsActive = false
	category.UpdatedAt = time.Now()
	return s.categoryRepo.Update(ctx, category)
}

func (s *CategoryService) AddEventToCategory(ctx context.Context, eventID, categoryID string, isPrimary bool) error {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	category, err := s.categoryRepo.GetByPublicID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %s", categoryID)
	}

	return s.categoryRepo.AddEventToCategory(ctx, event.ID, category.ID, isPrimary)
}

func (s *CategoryService) GetEventCategories(ctx context.Context, eventID string) ([]*entities.Category, error) {
	event, err := s.eventRepo.GetByPublicID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	return s.categoryRepo.GetEventCategories(ctx, event.ID)
}
