package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// CategoryRepository define operaciones para categorías
type CategoryRepository interface {
	// CRUD básico
	Create(ctx context.Context, category *entities.Category) error
	FindByID(ctx context.Context, id int64) (*entities.Category, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Category, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// Búsquedas
	List(ctx context.Context, filter dto.CategoryFilter, pagination dto.Pagination) ([]*entities.Category, int64, error)
	ListRootCategories(ctx context.Context) ([]*entities.Category, error)
	ListByParent(ctx context.Context, parentID int64) ([]*entities.Category, error)
	ListFeatured(ctx context.Context, limit int) ([]*entities.Category, error)
	ListActive(ctx context.Context) ([]*entities.Category, error)
	Search(ctx context.Context, term string, limit int) ([]*entities.Category, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, categoryID int64, active bool) error
	UpdateParent(ctx context.Context, categoryID int64, parentID *int64) error
	UpdateOrder(ctx context.Context, categoryID int64, order int) error
	UpdateIcon(ctx context.Context, categoryID int64, icon, color string) error
	UpdateStatistics(ctx context.Context, categoryID int64, eventsCount int, ticketsSold int64, revenue float64) error
	IncrementEventCount(ctx context.Context, categoryID int64) error
	DecrementEventCount(ctx context.Context, categoryID int64) error
	AddEventToCategory(ctx context.Context, eventCategory *entities.EventCategory) error
	RemoveEventFromCategory(ctx context.Context, eventID int64, categoryID int64) error

	// Jerarquía
	GetFullPath(ctx context.Context, categoryID int64) (string, error)
	GetLevel(ctx context.Context, categoryID int64) (int, error)
	GetSubtree(ctx context.Context, categoryID int64) ([]*entities.Category, error)
	GetAncestors(ctx context.Context, categoryID int64) ([]*entities.Category, error)
	GetDescendants(ctx context.Context, categoryID int64) ([]*entities.Category, error)

	// Estadísticas
	GetStats(ctx context.Context, categoryID int64) (*dto.CategoryStatsResponse, error)
	CountEvents(ctx context.Context, categoryID int64) (int64, error)
	GetTotalRevenue(ctx context.Context, categoryID int64) (float64, error)
	GetPopularCategories(ctx context.Context, limit int) ([]*dto.PopularCategory, error)
}
