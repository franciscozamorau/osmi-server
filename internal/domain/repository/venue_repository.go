package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// VenueRepository define operaciones para lugares/recintos
type VenueRepository interface {
	// CRUD básico
	Create(ctx context.Context, venue *entities.Venue) error
	FindByID(ctx context.Context, id int64) (*entities.Venue, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Venue, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Venue, error)
	Update(ctx context.Context, venue *entities.Venue) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// Búsquedas
	List(ctx context.Context, filter dto.VenueFilter, pagination dto.Pagination) ([]*entities.Venue, int64, error)
	ListByCountry(ctx context.Context, countryCode string, pagination dto.Pagination) ([]*entities.Venue, int64, error)
	ListByCity(ctx context.Context, country, city string, pagination dto.Pagination) ([]*entities.Venue, int64, error)
	ListByType(ctx context.Context, venueType string, pagination dto.Pagination) ([]*entities.Venue, int64, error)
	Search(ctx context.Context, term string, filter dto.VenueFilter, pagination dto.Pagination) ([]*entities.Venue, int64, error)
	FindNearby(ctx context.Context, latitude, longitude float64, radiusKm float64, limit int) ([]*entities.Venue, error)

	// Operaciones específicas
	UpdateCapacity(ctx context.Context, venueID int64, capacity, seating, standing *int) error
	UpdateLocation(ctx context.Context, venueID int64, address, city, state, postalCode, country string) error
	UpdateCoordinates(ctx context.Context, venueID int64, latitude, longitude float64) error
	UpdateContactInfo(ctx context.Context, venueID int64, email, phone string) error
	UpdateFacilities(ctx context.Context, venueID int64, facilities []string) error
	UpdateAccessibility(ctx context.Context, venueID int64, features []string) error
	AddImage(ctx context.Context, venueID int64, imageURL string, isMain bool) error
	RemoveImage(ctx context.Context, venueID int64, imageURL string) error
	SetMainImage(ctx context.Context, venueID int64, imageURL string) error

	// Consultas geográficas
	GetDistance(ctx context.Context, venueID int64, latitude, longitude float64) (float64, error)
	GetVenuesInRadius(ctx context.Context, centerLat, centerLon float64, radiusKm float64, venueType *string) ([]*entities.Venue, error)
	GetVenuesInBounds(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]*entities.Venue, error)

	// Estadísticas
	GetStats(ctx context.Context, venueID int64) (*dto.VenueStatsResponse, error)
	CountEvents(ctx context.Context, venueID int64) (int64, error)
	GetUpcomingEvents(ctx context.Context, venueID int64, limit int) ([]*dto.VenueEvent, error)
	GetCapacityUtilization(ctx context.Context, venueID int64) (float64, error)
	GetPopularVenues(ctx context.Context, limit int) ([]*dto.PopularVenue, error)
}
