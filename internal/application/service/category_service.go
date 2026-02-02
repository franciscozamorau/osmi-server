package services

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
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

func (s *CategoryService) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*entities.Category, error) {
	// Validar que el slug no exista
	existing, _ := s.categoryRepo.FindBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.New("category with this slug already exists")
	}

	// Validar parent si se proporciona
	var parentID *int64
	if req.ParentID != "" {
		parent, err := s.categoryRepo.FindByPublicID(ctx, req.ParentID)
		if err != nil {
			return nil, errors.New("parent category not found")
		}
		parentID = &parent.ID
	}

	// Calcular nivel
	level := 1
	if parentID != nil {
		level = 2 // TODO: Calcular nivel recursivo
	}

	// Crear categoría
	category := &entities.Category{
		PublicID:        uuid.New().String(),
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     &req.Description,
		Icon:            &req.Icon,
		ColorHex:        req.ColorHex,
		ParentID:        parentID,
		Level:           int32(level),
		Path:            "", // TODO: Calcular path
		IsActive:        req.IsActive,
		IsFeatured:      req.IsFeatured,
		SortOrder:       int32(req.SortOrder),
		MetaTitle:       &req.MetaTitle,
		MetaDescription: &req.MetaDescription,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := s.categoryRepo.Create(ctx, category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, categoryID string, req *dto.UpdateCategoryRequest) (*entities.Category, error) {
	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	// Actualizar campos
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Slug != "" {
		// Validar que el nuevo slug no exista
		if req.Slug != category.Slug {
			existing, _ := s.categoryRepo.FindBySlug(ctx, req.Slug)
			if existing != nil {
				return nil, errors.New("category with this slug already exists")
			}
			category.Slug = req.Slug
		}
	}
	if req.Description != "" {
		category.Description = &req.Description
	}
	if req.Icon != "" {
		category.Icon = &req.Icon
	}
	if req.ColorHex != "" {
		category.ColorHex = req.ColorHex
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.IsFeatured != nil {
		category.IsFeatured = *req.IsFeatured
	}
	if req.SortOrder != 0 {
		category.SortOrder = int32(req.SortOrder)
	}
	if req.MetaTitle != "" {
		category.MetaTitle = &req.MetaTitle
	}
	if req.MetaDescription != "" {
		category.MetaDescription = &req.MetaDescription
	}

	category.UpdatedAt = time.Now()

	err = s.categoryRepo.Update(ctx, category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, categoryID string) (*entities.Category, error) {
	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}
	return category, nil
}

func (s *CategoryService) ListCategories(ctx context.Context, filter dto.CategoryFilter, pagination dto.Pagination) ([]*entities.Category, int64, error) {
	categories, err := s.categoryRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.categoryRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]*entities.Category, error) {
	return s.categoryRepo.FindCategoryTree(ctx)
}

func (s *CategoryService) GetCategoryWithDescendants(ctx context.Context, categoryID string) ([]*entities.Category, error) {
	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	return s.categoryRepo.FindDescendants(ctx, category.ID)
}

func (s *CategoryService) AddEventToCategory(ctx context.Context, eventID, categoryID string, isPrimary bool) error {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return errors.New("event not found")
	}

	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return errors.New("category not found")
	}

	eventCategory := &entities.EventCategory{
		EventID:    event.ID,
		CategoryID: category.ID,
		IsPrimary:  isPrimary,
		CreatedAt:  time.Now(),
	}

	return s.categoryRepo.AddEventToCategory(ctx, eventCategory)
}

func (s *CategoryService) RemoveEventFromCategory(ctx context.Context, eventID, categoryID string) error {
	event, err := s.eventRepo.FindByPublicID(ctx, eventID)
	if err != nil {
		return errors.New("event not found")
	}

	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return errors.New("category not found")
	}

	return s.categoryRepo.RemoveEventFromCategory(ctx, event.ID, category.ID)
}

func (s *CategoryService) GetEventsByCategory(ctx context.Context, categoryID string, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return nil, 0, errors.New("category not found")
	}

	// Obtener IDs de eventos
	eventIDs, err := s.categoryRepo.FindEventsByCategory(ctx, category.ID)
	if err != nil {
		return nil, 0, err
	}

	// Obtener eventos completos
	var events []*entities.Event
	for _, eventID := range eventIDs {
		event, err := s.eventRepo.FindByID(ctx, eventID)
		if err == nil {
			events = append(events, event)
		}
	}

	return events, int64(len(events)), nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	category, err := s.categoryRepo.FindByPublicID(ctx, categoryID)
	if err != nil {
		return errors.New("category not found")
	}

	// Verificar que no tenga eventos asociados
	events, err := s.categoryRepo.FindEventsByCategory(ctx, category.ID)
	if err == nil && len(events) > 0 {
		return errors.New("cannot delete category with associated events")
	}

	// Verificar que no tenga subcategorías
	descendants, err := s.categoryRepo.FindDescendants(ctx, category.ID)
	if err == nil && len(descendants) > 0 {
		return errors.New("cannot delete category with subcategories")
	}

	return s.categoryRepo.Delete(ctx, category.ID)
}
