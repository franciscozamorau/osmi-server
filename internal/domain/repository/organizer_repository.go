package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// OrganizerRepository define operaciones para organizadores
type OrganizerRepository interface {
	// CRUD básico
	Create(ctx context.Context, organizer *entities.Organizer) error
	FindByID(ctx context.Context, id int64) (*entities.Organizer, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Organizer, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Organizer, error)
	Update(ctx context.Context, organizer *entities.Organizer) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// Búsquedas
	List(ctx context.Context, filter dto.OrganizerFilter, pagination dto.Pagination) ([]*entities.Organizer, int64, error)
	ListVerified(ctx context.Context, limit int) ([]*entities.Organizer, error)
	ListActive(ctx context.Context) ([]*entities.Organizer, error)
	Search(ctx context.Context, term string, limit int) ([]*entities.Organizer, error)
	FindByCountry(ctx context.Context, countryCode string, pagination dto.Pagination) ([]*entities.Organizer, int64, error)

	// Operaciones específicas
	UpdateVerification(ctx context.Context, organizerID int64, verified bool, status string) error
	UpdateRating(ctx context.Context, organizerID int64, rating float64, reviewCount int) error
	UpdateStatistics(ctx context.Context, organizerID int64, eventsCount int, ticketsSold int64, revenue float64) error
	UpdateContactInfo(ctx context.Context, organizerID int64, email, phone string) error
	UpdateLegalInfo(ctx context.Context, organizerID int64, legalName, taxID string, country string) error
	UpdateSocialLinks(ctx context.Context, organizerID int64, socialLinks map[string]string) error
	AddSocialLink(ctx context.Context, organizerID int64, platform, url string) error
	RemoveSocialLink(ctx context.Context, organizerID int64, platform string) error
	IncrementEventCount(ctx context.Context, organizerID int64) error
	DecrementEventCount(ctx context.Context, organizerID int64) error

	// Verificaciones
	IsVerified(ctx context.Context, organizerID int64) (bool, error)
	IsActive(ctx context.Context, organizerID int64) (bool, error)
	HasEvents(ctx context.Context, organizerID int64) (bool, error)

	// Estadísticas
	GetStats(ctx context.Context, organizerID int64) (*dto.OrganizerStatsResponse, error)
	GetGlobalStats(ctx context.Context) (*dto.OrganizerGlobalStats, error)
	CountEvents(ctx context.Context, organizerID int64) (int64, error)
	GetTotalRevenue(ctx context.Context, organizerID int64) (float64, error)
	GetAverageRating(ctx context.Context, organizerID int64) (float64, error)
	GetTopOrganizers(ctx context.Context, limit int) ([]*dto.TopOrganizer, error)
}
