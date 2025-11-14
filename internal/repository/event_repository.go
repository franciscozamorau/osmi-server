package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/db"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventRepository struct{}

func NewEventRepository() *EventRepository {
	return &EventRepository{}
}

func (r *EventRepository) CreateEvent(ctx context.Context, event *models.Event) (int64, error) {
	query := `
        INSERT INTO events (
            public_id, name, description, short_description, start_date, end_date,
            location, venue_details, category, tags, is_active, is_published,
            image_url, banner_url, max_attendees, organizer_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10, $11, $12,
            $13, $14, $15, $16
        ) RETURNING id
    `

	// Validaciones mejoradas
	if strings.TrimSpace(event.Name) == "" {
		return 0, fmt.Errorf("event name is required")
	}
	if strings.TrimSpace(event.Location) == "" {
		return 0, fmt.Errorf("location is required")
	}

	// Validar que las fechas sean lógicas
	if event.EndDate.Before(event.StartDate) {
		return 0, fmt.Errorf("end date cannot be before start date")
	}

	// Validar que el evento no sea en el pasado
	if event.StartDate.Before(time.Now()) {
		log.Printf("Warning: Creating event with start date in the past: %s", event.StartDate.Format(time.RFC3339))
	}

	// Convertir tags a formato PostgreSQL array si es necesario
	var tags interface{}
	if event.Tags != nil && len(event.Tags) > 0 {
		tags = event.Tags // pgx maneja automáticamente []string como PostgreSQL array
	} else {
		tags = pgtype.FlatArray[string](nil)
	}

	var id int64
	err := db.Pool.QueryRow(ctx, query,
		event.PublicID,
		strings.TrimSpace(event.Name),
		event.Description,
		event.ShortDescription,
		event.StartDate,
		event.EndDate,
		strings.TrimSpace(event.Location),
		event.VenueDetails,
		event.Category,
		tags, // CORREGIDO: Manejo correcto de array
		event.IsActive,
		event.IsPublished,
		event.ImageURL,
		event.BannerURL,
		event.MaxAttendees,
		event.OrganizerID, // CORREGIDO: Agregado organizer_id
	).Scan(&id)

	if err != nil {
		// Manejar errores de duplicados
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return 0, fmt.Errorf("event with public_id %s already exists", event.PublicID)
		}
		return 0, fmt.Errorf("error inserting event: %w", err)
	}

	log.Printf("Event created successfully: %s (ID: %d, PublicID: %s)", event.Name, id, event.PublicID)
	return id, nil
}

func (r *EventRepository) GetEventByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	// Validar que publicID sea UUID válido
	if _, err := uuid.Parse(publicID); err != nil {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events WHERE public_id = $1
	`

	row := db.Pool.QueryRow(ctx, query, publicID)
	var event models.Event
	var tags []string

	err := row.Scan(
		&event.ID,
		&event.PublicID,
		&event.OrganizerID,
		&event.Name,
		&event.Description,
		&event.ShortDescription,
		&event.StartDate,
		&event.EndDate,
		&event.Location,
		&event.VenueDetails,
		&event.Coordinates,
		&event.Category,
		&tags, // CORREGIDO: Escanear como []string
		&event.IsActive,
		&event.IsPublished,
		&event.ImageURL,
		&event.BannerURL,
		&event.MaxAttendees,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("event not found: %s", publicID)
		}
		return nil, fmt.Errorf("error retrieving event: %w", err)
	}

	// Asignar tags escaneados al evento
	event.Tags = tags

	return &event, nil
}

func (r *EventRepository) ListEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true
		ORDER BY start_date ASC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string

		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Coordinates,
			&event.Category,
			&tags, // CORREGIDO: Escanear como []string
			&event.IsActive,
			&event.IsPublished,
			&event.ImageURL,
			&event.BannerURL,
			&event.MaxAttendees,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}

		event.Tags = tags
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	log.Printf("Retrieved %d active events", len(events))
	return events, nil
}

// GetEventIDByPublicID obtiene el ID interno de un evento por su public_id - CORREGIDO
func (r *EventRepository) GetEventIDByPublicID(ctx context.Context, publicID string) (int64, error) {
	// Validar que publicID sea UUID válido
	if _, err := uuid.Parse(publicID); err != nil {
		return 0, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	var eventID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM events WHERE public_id = $1 AND is_active = true AND is_published = true",
		publicID).Scan(&eventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("event not found or inactive: %s", publicID)
		}
		return 0, fmt.Errorf("error getting event ID by public_id: %v", err)
	}
	return eventID, nil
}

// UpdateEvent actualiza un evento existente - CORREGIDO
func (r *EventRepository) UpdateEvent(ctx context.Context, event *models.Event) error {
	// Validaciones
	if strings.TrimSpace(event.Name) == "" {
		return fmt.Errorf("event name is required")
	}
	if strings.TrimSpace(event.Location) == "" {
		return fmt.Errorf("location is required")
	}
	if event.EndDate.Before(event.StartDate) {
		return fmt.Errorf("end date cannot be before start date")
	}

	// Convertir tags a formato PostgreSQL array
	var tags interface{}
	if event.Tags != nil && len(event.Tags) > 0 {
		tags = event.Tags
	} else {
		tags = pgtype.FlatArray[string](nil)
	}

	query := `
		UPDATE events 
		SET name = $1, description = $2, short_description = $3, start_date = $4, 
		    end_date = $5, location = $6, venue_details = $7, category = $8, 
		    tags = $9, is_active = $10, is_published = $11, image_url = $12, 
		    banner_url = $13, max_attendees = $14, organizer_id = $15, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $16
	`

	result, err := db.Pool.Exec(ctx, query,
		strings.TrimSpace(event.Name),
		event.Description,
		event.ShortDescription,
		event.StartDate,
		event.EndDate,
		strings.TrimSpace(event.Location),
		event.VenueDetails,
		event.Category,
		tags, // CORREGIDO: Manejo correcto de array
		event.IsActive,
		event.IsPublished,
		event.ImageURL,
		event.BannerURL,
		event.MaxAttendees,
		event.OrganizerID, // CORREGIDO: Agregado organizer_id
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with id: %d", event.ID)
	}

	log.Printf("Event updated successfully: %s (ID: %d)", event.Name, event.ID)
	return nil
}

// DeleteEvent elimina un evento (soft delete marcando como inactivo) - MEJORADO
func (r *EventRepository) DeleteEvent(ctx context.Context, id int64) error {
	query := `UPDATE events SET is_active = false, is_published = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with id: %d", id)
	}

	log.Printf("Event marked as inactive and unpublished: ID %d", id)
	return nil
}

// ListActiveEvents lista solo eventos activos y publicados - CORREGIDO
func (r *EventRepository) ListActiveEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true
		ORDER BY start_date ASC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing active events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string

		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Coordinates,
			&event.Category,
			&tags, // CORREGIDO: Escanear como []string
			&event.IsActive,
			&event.IsPublished,
			&event.ImageURL,
			&event.BannerURL,
			&event.MaxAttendees,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}

		event.Tags = tags
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByDateRange lista eventos en un rango de fechas - CORREGIDO
func (r *EventRepository) GetEventsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true 
		  AND start_date >= $1 AND end_date <= $2
		ORDER BY start_date ASC
	`

	rows, err := db.Pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error listing events by date range: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string

		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Coordinates,
			&event.Category,
			&tags, // CORREGIDO: Escanear como []string
			&event.IsActive,
			&event.IsPublished,
			&event.ImageURL,
			&event.BannerURL,
			&event.MaxAttendees,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}

		event.Tags = tags
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetUpcomingEvents lista eventos futuros - CORREGIDO
func (r *EventRepository) GetUpcomingEvents(ctx context.Context, limit int) ([]*models.Event, error) {
	if limit <= 0 {
		limit = 10 // Valor por defecto
	}

	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events 
		WHERE is_active = true AND is_published = true 
		  AND start_date > $1
		ORDER BY start_date ASC
		LIMIT $2
	`

	rows, err := db.Pool.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("error listing upcoming events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string

		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Coordinates,
			&event.Category,
			&tags, // CORREGIDO: Escanear como []string
			&event.IsActive,
			&event.IsPublished,
			&event.ImageURL,
			&event.BannerURL,
			&event.MaxAttendees,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}

		event.Tags = tags
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByOrganizer lista eventos por organizador - NUEVO
func (r *EventRepository) GetEventsByOrganizer(ctx context.Context, organizerID int64) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events 
		WHERE organizer_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query, organizerID)
	if err != nil {
		return nil, fmt.Errorf("error listing events by organizer: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string

		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Coordinates,
			&event.Category,
			&tags, // CORREGIDO: Escanear como []string
			&event.IsActive,
			&event.IsPublished,
			&event.ImageURL,
			&event.BannerURL,
			&event.MaxAttendees,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}

		event.Tags = tags
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventsByCategory lista eventos por categoría - NUEVO
func (r *EventRepository) GetEventsByCategory(ctx context.Context, category string) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, organizer_id, name, description, short_description, 
		       start_date, end_date, location, venue_details, coordinates, category, 
		       tags, is_active, is_published, image_url, banner_url, max_attendees, 
		       created_at, updated_at
		FROM events 
		WHERE category = $1 AND is_active = true AND is_published = true
		ORDER BY start_date ASC
	`

	rows, err := db.Pool.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("error listing events by category: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string

		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.OrganizerID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Coordinates,
			&event.Category,
			&tags, // CORREGIDO: Escanear como []string
			&event.IsActive,
			&event.IsPublished,
			&event.ImageURL,
			&event.BannerURL,
			&event.MaxAttendees,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event: %w", err)
		}

		event.Tags = tags
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

/* Helper function - CORREGIDO
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: strings.TrimSpace(s), Valid: s != ""}
}*/
