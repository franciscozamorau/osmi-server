package repository

import (
	"context"
	"fmt"
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
            image_url, banner_url, max_attendees, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10, $11, $12,
            $13, $14, $15, $16, $17
        ) RETURNING id
    `

	now := time.Now()

	// Validaciones
	if event.PublicID == "" {
		return 0, fmt.Errorf("public_id is required")
	}

	// Validar que public_id sea UUID válido
	if _, err := uuid.Parse(event.PublicID); err != nil {
		return 0, fmt.Errorf("invalid public_id format: must be a valid UUID")
	}

	// Validar campos obligatorios
	if event.Name == "" {
		return 0, fmt.Errorf("event name is required")
	}
	if event.Location == "" {
		return 0, fmt.Errorf("location is required")
	}

	var id int64
	err := db.Pool.QueryRow(ctx, query,
		event.PublicID,
		event.Name,
		event.Description,
		event.ShortDescription,
		event.StartDate,
		event.EndDate,
		event.Location,
		event.VenueDetails,
		event.Category,
		event.Tags,
		event.IsActive,
		event.IsPublished,
		event.ImageURL,
		event.BannerURL,
		event.MaxAttendees,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error inserting event: %w", err)
	}
	return id, nil
}

func (r *EventRepository) GetEventByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	// Validar que publicID sea UUID válido
	if _, err := uuid.Parse(publicID); err != nil {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	query := `
		SELECT id, public_id, name, description, short_description, start_date, end_date,
			   location, venue_details, category, tags, is_active, is_published,
			   image_url, banner_url, max_attendees, created_at, updated_at
		FROM events WHERE public_id = $1
	`

	row := db.Pool.QueryRow(ctx, query, publicID)
	var event models.Event

	err := row.Scan(
		&event.ID,
		&event.PublicID,
		&event.Name,
		&event.Description,
		&event.ShortDescription,
		&event.StartDate,
		&event.EndDate,
		&event.Location,
		&event.VenueDetails,
		&event.Category,
		&event.Tags,
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
	return &event, nil
}

func (r *EventRepository) ListEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, name, description, short_description, start_date, end_date,
			   location, venue_details, category, tags, is_active, is_published,
			   image_url, banner_url, max_attendees, created_at, updated_at
		FROM events ORDER BY start_date ASC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Category,
			&event.Tags,
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
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetEventIDByPublicID obtiene el ID interno de un evento por su public_id
func (r *EventRepository) GetEventIDByPublicID(ctx context.Context, publicID string) (int64, error) {
	var eventID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM events WHERE public_id = $1",
		publicID).Scan(&eventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("event not found with public_id: %s", publicID)
		}
		return 0, fmt.Errorf("error getting event ID by public_id: %v", err)
	}
	return eventID, nil
}

// UpdateEvent actualiza un evento existente
func (r *EventRepository) UpdateEvent(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events 
		SET name = $1, description = $2, short_description = $3, start_date = $4, 
		    end_date = $5, location = $6, venue_details = $7, category = $8, 
		    tags = $9, is_active = $10, is_published = $11, image_url = $12, 
		    banner_url = $13, max_attendees = $14, updated_at = CURRENT_TIMESTAMP
		WHERE id = $15
	`

	result, err := db.Pool.Exec(ctx, query,
		event.Name,
		event.Description,
		event.ShortDescription,
		event.StartDate,
		event.EndDate,
		event.Location,
		event.VenueDetails,
		event.Category,
		event.Tags,
		event.IsActive,
		event.IsPublished,
		event.ImageURL,
		event.BannerURL,
		event.MaxAttendees,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with id: %d", event.ID)
	}

	return nil
}

// DeleteEvent elimina un evento
func (r *EventRepository) DeleteEvent(ctx context.Context, id int64) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found with id: %d", id)
	}

	return nil
}

// ListActiveEvents lista solo eventos activos
func (r *EventRepository) ListActiveEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, public_id, name, description, short_description, start_date, end_date,
			   location, venue_details, category, tags, is_active, is_published,
			   image_url, banner_url, max_attendees, created_at, updated_at
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
		err := rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.Name,
			&event.Description,
			&event.ShortDescription,
			&event.StartDate,
			&event.EndDate,
			&event.Location,
			&event.VenueDetails,
			&event.Category,
			&event.Tags,
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
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// Helper function
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}
