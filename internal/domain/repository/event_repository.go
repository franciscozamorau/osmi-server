package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// EventRepository define operaciones para eventos
type EventRepository interface {
	// CRUD básico
	Create(ctx context.Context, event *entities.Event) error
	FindByID(ctx context.Context, id int64) (*entities.Event, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Event, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Event, error)
	Update(ctx context.Context, event *entities.Event) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// Búsquedas y listados
	List(ctx context.Context, filter dto.EventFilter, pagination dto.Pagination) ([]*entities.Event, int64, error)
	FindByOrganizer(ctx context.Context, organizerID int64, pagination dto.Pagination) ([]*entities.Event, int64, error)
	FindByCategory(ctx context.Context, categoryID int64, pagination dto.Pagination) ([]*entities.Event, int64, error)
	FindByVenue(ctx context.Context, venueID int64, pagination dto.Pagination) ([]*entities.Event, int64, error)
	FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Event, int64, error)
	FindByDateRange(ctx context.Context, startDate, endDate string, pagination dto.Pagination) ([]*entities.Event, int64, error)
	FindUpcoming(ctx context.Context, limit int) ([]*entities.Event, error)
	FindFeatured(ctx context.Context, limit int) ([]*entities.Event, error)
	Search(ctx context.Context, term string, filter dto.EventFilter, pagination dto.Pagination) ([]*entities.Event, int64, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, eventID int64, status string) error
	UpdateVisibility(ctx context.Context, eventID int64, visibility string) error
	IncrementViewCount(ctx context.Context, eventID int64) error
	IncrementFavoriteCount(ctx context.Context, eventID int64) error
	DecrementFavoriteCount(ctx context.Context, eventID int64) error
	IncrementShareCount(ctx context.Context, eventID int64) error
	Publish(ctx context.Context, eventID int64) error
	Unpublish(ctx context.Context, eventID int64) error
	Cancel(ctx context.Context, eventID int64, reason string) error
	Complete(ctx context.Context, eventID int64) error
	MarkAsSoldOut(ctx context.Context, eventID int64) error

	// Estadísticas
	GetStats(ctx context.Context, eventID int64) (*dto.EventStatsResponse, error)
	GetGlobalStats(ctx context.Context) (*dto.EventGlobalStats, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByOrganizer(ctx context.Context, organizerID int64) (int64, error)
	GetRevenue(ctx context.Context, eventID int64) (float64, error)
	GetAttendanceRate(ctx context.Context, eventID int64) (float64, error)
}
