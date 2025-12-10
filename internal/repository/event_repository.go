// event_repository.go - COMPLETO Y CORREGIDO
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

// CreateEvent crea un nuevo evento - VERSIÓN CORREGIDA
func (r *EventRepository) CreateEvent(ctx context.Context, event *models.Event) (string, error) {
	// Validaciones
	if strings.TrimSpace(event.Name) == "" {
		return "", fmt.Errorf("event name is required")
	}
	if strings.TrimSpace(event.Location) == "" {
		return "", fmt.Errorf("location is required")
	}

	// Validar que las fechas sean lógicas
	if !event.EndDate.After(event.StartDate) {
		return "", fmt.Errorf("end date cannot be before start date")
	}

	// Validar que el evento no sea en el pasado (solo warning)
	if event.StartDate.Before(time.Now()) {
		log.Printf("Warning: Creating event with start date in the past: %s",
			event.StartDate.Format("2006-01-02 15:04:05"))
	}

	// Generar public_id automáticamente si no viene
	if event.PublicID == "" {
		event.PublicID = uuid.New().String()
	}

	// Validar UUID
	if !IsValidUUID(event.PublicID) {
		return "", fmt.Errorf("invalid public_id format")
	}

	// Establecer status por defecto si no viene
	if event.Status == "" {
		event.Status = "draft"
	}

	// Validar status
	if !IsValidEventStatus(event.Status) {
		return "", fmt.Errorf("invalid event status: %s", event.Status)
	}

	// ✅ CORRECCIÓN: La consulta INSERT tiene 17 columnas y 17 valores
	query := `
        INSERT INTO events (
            public_id, organizer_id, name, description, short_description, 
            start_date, end_date, location, venue_details, coordinates, 
            category, tags, is_active, is_published, status,
            image_url, banner_url, max_attendees
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
            $13, $14, $15, $16, $17, $18
        ) RETURNING public_id
    `

	var publicID string
	err := r.db.QueryRow(ctx, query,
		event.PublicID,
		ToPgInt8FromInt64Ptr(event.OrganizerID),
		strings.TrimSpace(event.Name),
		ToPgTextFromPtr(event.Description),
		ToPgTextFromPtr(event.ShortDescription),
		event.StartDate,
		event.EndDate,
		strings.TrimSpace(event.Location),
		ToPgTextFromPtr(event.VenueDetails),
		ToPgTextFromPtr(event.Coordinates),       // ✅ $10 - coordinates
		ToPgTextFromPtr(event.Category),          // ✅ $11 - category
		toPgStringArray(event.Tags),              // ✅ $12 - tags
		event.IsActive,                           // ✅ $13 - is_active
		event.IsPublished,                        // ✅ $14 - is_published
		event.Status,                             // ✅ $15 - status
		ToPgTextFromPtr(event.ImageURL),          // ✅ $16 - image_url
		ToPgTextFromPtr(event.BannerURL),         // ✅ $17 - banner_url
		ToPgInt4FromInt32Ptr(event.MaxAttendees), // ✅ $18 - max_attendees
	).Scan(&publicID)

	if err != nil {
		if IsDuplicateKeyError(err) {
			return "", fmt.Errorf("event with public_id %s already exists", event.PublicID)
		}
		return "", fmt.Errorf("error inserting event: %w", err)
	}

	log.Printf("Event created successfully: %s (PublicID: %s, Status: %s)",
		SafeStringForLog(event.Name), publicID, event.Status)
	return publicID, nil
}

// GetEventByPublicID obtiene un evento por su UUID público
func (r *EventRepository) GetEventByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	// Validar que publicID sea UUID válido
	if !IsValidUUID(publicID) {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events WHERE public_id = $1
	`

	row := r.db.QueryRow(ctx, query, publicID)

	var event models.Event
	var dbDescription pgtype.Text
	var dbShortDescription pgtype.Text
	var dbVenueDetails pgtype.Text
	var dbCoordinates pgtype.Text
	var dbCategory pgtype.Text
	var dbImageURL pgtype.Text
	var dbBannerURL pgtype.Text
	var dbOrganizerID pgtype.Int8
	var tags pgtype.FlatArray[string]
	var dbMaxAttendees pgtype.Int4

	err := row.Scan(
		&event.ID,
		&event.PublicID,
		&dbOrganizerID,
		&event.Name,
		&dbDescription,
		&dbShortDescription,
		&event.StartDate,
		&event.EndDate,
		&event.Location,
		&dbVenueDetails,
		&dbCoordinates,
		&dbCategory,
		&tags,
		&event.IsActive,
		&event.IsPublished,
		&event.Status,
		&dbImageURL,
		&dbBannerURL,
		&dbMaxAttendees,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %s", publicID)
		}
		return nil, fmt.Errorf("error retrieving event: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	event.OrganizerID = ToInt64FromPgInt8(dbOrganizerID)
	event.Description = ToStringFromPgText(dbDescription)
	event.ShortDescription = ToStringFromPgText(dbShortDescription)
	event.VenueDetails = ToStringFromPgText(dbVenueDetails)
	event.Coordinates = ToStringFromPgText(dbCoordinates)
	event.Category = ToStringFromPgText(dbCategory)
	event.ImageURL = ToStringFromPgText(dbImageURL)
	event.BannerURL = ToStringFromPgText(dbBannerURL)
	event.MaxAttendees = ToInt32FromPgInt4(dbMaxAttendees)

	// Asignar tags escaneados al evento
	event.Tags = []string(tags)

	return &event, nil
}

// ListEvents lista todos los eventos
func (r *EventRepository) ListEvents(ctx context.Context, includeInactive bool) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE ($1 = true OR is_active = true)
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("error listing events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	log.Printf("Retrieved %d events (includeInactive: %v)", len(events), includeInactive)
	return events, nil
}

// GetEventIDByPublicID obtiene el ID interno de un evento por su public_id
func (r *EventRepository) GetEventIDByPublicID(ctx context.Context, publicID string) (int64, error) {
	// Validar que publicID sea UUID válido
	if !IsValidUUID(publicID) {
		return 0, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	var eventID int64
	err := r.db.QueryRow(ctx,
		"SELECT id FROM events WHERE public_id = $1 AND is_active = true",
		publicID).Scan(&eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("event not found or inactive: %s", publicID)
		}
		return 0, fmt.Errorf("error getting event ID by public_id: %w", err)
	}
	return eventID, nil
}

// UpdateEvent actualiza un evento existente
func (r *EventRepository) UpdateEvent(ctx context.Context, event *models.Event) error {
	// Validaciones
	if strings.TrimSpace(event.Name) == "" {
		return fmt.Errorf("event name is required")
	}
	if strings.TrimSpace(event.Location) == "" {
		return fmt.Errorf("location is required")
	}
	if !event.EndDate.After(event.StartDate) {
		return fmt.Errorf("end date cannot be before start date")
	}

	if !IsValidEventStatus(event.Status) {
		return fmt.Errorf("invalid event status: %s", event.Status)
	}

	query := `
		UPDATE events 
		SET name = $1, description = $2, short_description = $3, start_date = $4, 
		    end_date = $5, location = $6, venue_details = $7, category = $8, 
		    tags = $9, is_active = $10, is_published = $11, status = $12,
		    image_url = $13, banner_url = $14, max_attendees = $15, 
		    organizer_id = $16, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $17
	`

	result, err := r.db.Exec(ctx, query,
		strings.TrimSpace(event.Name),
		ToPgTextFromPtr(event.Description),
		ToPgTextFromPtr(event.ShortDescription),
		event.StartDate,
		event.EndDate,
		strings.TrimSpace(event.Location),
		ToPgTextFromPtr(event.VenueDetails),
		ToPgTextFromPtr(event.Category),
		toPgStringArray(event.Tags),
		event.IsActive,
		event.IsPublished,
		event.Status,
		ToPgTextFromPtr(event.ImageURL),
		ToPgTextFromPtr(event.BannerURL),
		ToPgInt4FromInt32Ptr(event.MaxAttendees),
		ToPgInt8FromInt64Ptr(event.OrganizerID),
		event.PublicID,
	)

	if err != nil {
		return fmt.Errorf("error updating event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with public_id: %s", event.PublicID)
	}

	log.Printf("Event updated successfully: %s (PublicID: %s, Status: %s)",
		SafeStringForLog(event.Name), event.PublicID, event.Status)
	return nil
}

// UpdateEventStatus actualiza solo el estado de un evento
func (r *EventRepository) UpdateEventStatus(ctx context.Context, publicID string, status string) error {
	if !IsValidUUID(publicID) {
		return fmt.Errorf("invalid event ID format")
	}

	if !IsValidEventStatus(status) {
		return fmt.Errorf("invalid event status: %s", status)
	}

	query := `
		UPDATE events 
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $2
	`

	result, err := r.db.Exec(ctx, query, status, publicID)
	if err != nil {
		return fmt.Errorf("error updating event status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with public_id: %s", publicID)
	}

	log.Printf("Event status updated: %s -> %s", publicID, status)
	return nil
}

// DeleteEvent elimina un evento (soft delete marcando como inactivo)
func (r *EventRepository) DeleteEvent(ctx context.Context, publicID string) error {
	if !IsValidUUID(publicID) {
		return fmt.Errorf("invalid event ID format")
	}

	query := `
		UPDATE events 
		SET is_active = false, is_published = false, status = 'cancelled', 
		    updated_at = CURRENT_TIMESTAMP 
		WHERE public_id = $1
	`

	result, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return fmt.Errorf("error deleting event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with public_id: %s", publicID)
	}

	log.Printf("Event marked as inactive and cancelled: %s", publicID)
	return nil
}

// ListActiveEvents lista solo eventos activos y publicados
func (r *EventRepository) ListActiveEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true AND status != 'cancelled'
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing active events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByDateRange lista eventos en un rango de fechas
func (r *EventRepository) GetEventsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Event, error) {
	if !endDate.After(startDate) {
		return nil, fmt.Errorf("invalid date range: end date cannot be before start date")
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true 
		  AND start_date >= $1 AND end_date <= $2
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error listing events by date range: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetUpcomingEvents lista eventos futuros
func (r *EventRepository) GetUpcomingEvents(ctx context.Context, limit int) ([]*models.Event, error) {
	if limit <= 0 {
		limit = 10 // Valor por defecto
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true 
		  AND start_date > $1 AND status != 'cancelled'
		ORDER BY start_date ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("error listing upcoming events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByOrganizer lista eventos por organizador
func (r *EventRepository) GetEventsByOrganizer(ctx context.Context, organizerID int64) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE organizer_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, organizerID)
	if err != nil {
		return nil, fmt.Errorf("error listing events by organizer: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByCategory lista eventos por categoría
func (r *EventRepository) GetEventsByCategory(ctx context.Context, category string) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE category = $1 AND is_active = true AND is_published = true
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("error listing events by category: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByStatus lista eventos por estado
func (r *EventRepository) GetEventsByStatus(ctx context.Context, status string) ([]*models.Event, error) {
	if !IsValidEventStatus(status) {
		return nil, fmt.Errorf("invalid event status: %s", status)
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE status = $1 AND is_active = true
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("error listing events by status: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// SearchEvents busca eventos por término
func (r *EventRepository) SearchEvents(ctx context.Context, searchTerm string, limit int) ([]*models.Event, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, status, image_url, banner_url, 
		       max_attendees, created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true 
		  AND (name ILIKE $1 OR description ILIKE $1 OR location ILIKE $1 OR category ILIKE $1)
		ORDER BY start_date ASC
		LIMIT $2
	`

	searchPattern := "%" + strings.TrimSpace(searchTerm) + "%"
	rows, err := r.db.Query(ctx, query, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("error searching events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event, err := r.scanEventRow(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventStats obtiene estadísticas de eventos
func (r *EventRepository) GetEventStats(ctx context.Context) (*models.EventStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_events,
			COUNT(CASE WHEN is_published = true THEN 1 END) as published_events,
			COUNT(CASE WHEN status = 'draft' THEN 1 END) as draft_events,
			COUNT(CASE WHEN status = 'sold_out' THEN 1 END) as sold_out_events,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_events,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_events
		FROM events
	`

	var stats models.EventStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalEvents,
		&stats.ActiveEvents,
		&stats.PublishedEvents,
		&stats.DraftEvents,
		&stats.SoldOutEvents,
		&stats.CancelledEvents,
		&stats.CompletedEvents,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting event stats: %w", err)
	}

	return &stats, nil
}

// =============================================================================
// MÉTODOS PRIVADOS
// =============================================================================

// scanEventRow escanea una fila de evento (método helper reutilizable)
func (r *EventRepository) scanEventRow(rows pgx.Row) (*models.Event, error) {
	var event models.Event
	var dbDescription pgtype.Text
	var dbShortDescription pgtype.Text
	var dbVenueDetails pgtype.Text
	var dbCoordinates pgtype.Text
	var dbCategory pgtype.Text
	var dbImageURL pgtype.Text
	var dbBannerURL pgtype.Text
	var dbOrganizerID pgtype.Int8
	var tags pgtype.FlatArray[string]
	var dbMaxAttendees pgtype.Int4

	err := rows.Scan(
		&event.ID,
		&event.PublicID,
		&dbOrganizerID,
		&event.Name,
		&dbDescription,
		&dbShortDescription,
		&event.StartDate,
		&event.EndDate,
		&event.Location,
		&dbVenueDetails,
		&dbCoordinates,
		&dbCategory,
		&tags,
		&event.IsActive,
		&event.IsPublished,
		&event.Status,
		&dbImageURL,
		&dbBannerURL,
		&dbMaxAttendees,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error scanning event: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	event.OrganizerID = ToInt64FromPgInt8(dbOrganizerID)
	event.Description = ToStringFromPgText(dbDescription)
	event.ShortDescription = ToStringFromPgText(dbShortDescription)
	event.VenueDetails = ToStringFromPgText(dbVenueDetails)
	event.Coordinates = ToStringFromPgText(dbCoordinates)
	event.Category = ToStringFromPgText(dbCategory)
	event.ImageURL = ToStringFromPgText(dbImageURL)
	event.BannerURL = ToStringFromPgText(dbBannerURL)
	event.MaxAttendees = ToInt32FromPgInt4(dbMaxAttendees)

	// Asignar tags escaneados al evento
	event.Tags = []string(tags)

	return &event, nil
}

// Helper function para convertir []string a pgtype.FlatArray[string]
func toPgStringArray(tags []string) pgtype.FlatArray[string] {
	if tags == nil || len(tags) == 0 {
		return pgtype.FlatArray[string]{}
	}
	return pgtype.FlatArray[string](tags)
}
