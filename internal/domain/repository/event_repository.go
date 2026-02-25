package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

type EventRepository interface {
	// CRUD básico
	Create(ctx context.Context, event *entities.Event) error
	GetByID(ctx context.Context, id int64) (*entities.Event, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.Event, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Event, error)
	Update(ctx context.Context, event *entities.Event) error
	Delete(ctx context.Context, id int64) error

	// Listados con filtros
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*entities.Event, int64, error)

	// Búsquedas específicas (las que realmente usas)
	ListByOrganizer(ctx context.Context, organizerID int64, limit, offset int) ([]*entities.Event, int64, error)
	ListUpcoming(ctx context.Context, limit int) ([]*entities.Event, error)
	ListFeatured(ctx context.Context, limit int) ([]*entities.Event, error)

	// Relaciones
	GetEventCategories(ctx context.Context, eventID int64) ([]*entities.Category, error)
	AddCategoryToEvent(ctx context.Context, eventID, categoryID int64, isPrimary bool) error
	RemoveCategoryFromEvent(ctx context.Context, eventID, categoryID int64) error
}
