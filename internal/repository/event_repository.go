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

	// ✅ CORREGIDO: Validaciones robustas
	if event.PublicID == "" {
		return 0, fmt.Errorf("public_id is required")
	}

	// ✅ Validar que public_id sea UUID válido
	if _, err := uuid.Parse(event.PublicID); err != nil {
		return 0, fmt.Errorf("invalid public_id format: must be a valid UUID")
	}

	// ✅ Validar campos obligatorios
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
	// ✅ CORREGIDO: Validar que publicID sea UUID válido
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

// ✅ CORREGIDO: Función helper actualizada
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}
