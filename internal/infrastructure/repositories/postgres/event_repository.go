package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) repository.EventRepository {
	return &eventRepository{db: db}
}

// Create inserta un nuevo evento
func (r *eventRepository) Create(ctx context.Context, event *entities.Event) error {
	query := `
		INSERT INTO ticketing.events (
			public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at, doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees, tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description, settings,
			published_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, 
			$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, 
			NOW(), NOW()
		)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		event.PublicID,
		event.OrganizerID,
		event.PrimaryCategoryID,
		event.VenueID,
		event.Slug,
		event.Name,
		event.ShortDescription,
		event.Description,
		event.EventType,
		event.CoverImageURL,
		event.BannerImageURL,
		event.GalleryImages,
		event.Timezone,
		event.StartsAt,
		event.EndsAt,
		event.DoorsOpenAt,
		event.DoorsCloseAt,
		event.VenueName,
		event.AddressFull,
		event.City,
		event.State,
		event.Country,
		event.Status,
		event.Visibility,
		event.IsFeatured,
		event.IsFree,
		event.MaxAttendees,
		event.MinAttendees,
		event.Tags,
		event.AgeRestriction,
		event.RequiresApproval,
		event.AllowReservations,
		event.ReservationDuration,
		event.ViewCount,
		event.FavoriteCount,
		event.ShareCount,
		event.MetaTitle,
		event.MetaDescription,
		event.Settings,
		event.PublishedAt,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetByID obtiene evento por ID
func (r *eventRepository) GetByID(ctx context.Context, id int64) (*entities.Event, error) {
	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at, doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees, tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description, settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE id = $1
	`

	var event entities.Event
	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
		&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
		&event.CoverImageURL, &event.BannerImageURL, &event.GalleryImages,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
		&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &event.Tags, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&event.MetaTitle, &event.MetaDescription, &event.Settings,
		&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

// GetByPublicID obtiene evento por UUID
func (r *eventRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.Event, error) {
	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at, doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees, tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description, settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE public_uuid = $1
	`

	var event entities.Event
	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
		&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
		&event.CoverImageURL, &event.BannerImageURL, &event.GalleryImages,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
		&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &event.Tags, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&event.MetaTitle, &event.MetaDescription, &event.Settings,
		&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", publicID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

// GetBySlug obtiene evento por slug
func (r *eventRepository) GetBySlug(ctx context.Context, slug string) (*entities.Event, error) {
	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at, doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees, tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description, settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE slug = $1
	`

	var event entities.Event
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
		&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
		&event.CoverImageURL, &event.BannerImageURL, &event.GalleryImages,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
		&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &event.Tags, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&event.MetaTitle, &event.MetaDescription, &event.Settings,
		&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", slug)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

// Update actualiza evento
func (r *eventRepository) Update(ctx context.Context, event *entities.Event) error {
	query := `
		UPDATE ticketing.events 
		SET slug = $1, name = $2, short_description = $3, description = $4,
			venue_id = $5, venue_name = $6, address_full = $7, city = $8, state = $9, country = $10,
			starts_at = $11, ends_at = $12, doors_open_at = $13, doors_close_at = $14,
			status = $15, visibility = $16, is_featured = $17, is_free = $18,
			max_attendees = $19, tags = $20, settings = $21,
			updated_at = NOW()
		WHERE id = $22
	`

	result, err := r.db.Exec(ctx, query,
		event.Slug,
		event.Name,
		event.ShortDescription,
		event.Description,
		event.VenueID,
		event.VenueName,
		event.AddressFull,
		event.City,
		event.State,
		event.Country,
		event.StartsAt,
		event.EndsAt,
		event.DoorsOpenAt,
		event.DoorsCloseAt,
		event.Status,
		event.Visibility,
		event.IsFeatured,
		event.IsFree,
		event.MaxAttendees,
		event.Tags,
		event.Settings,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %d", event.ID)
	}

	return nil
}

// Delete elimina evento
func (r *eventRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ticketing.events WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %d", id)
	}
	return nil
}

// List devuelve eventos con filtros
func (r *eventRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*entities.Event, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if val, ok := filter["status"]; ok {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := filter["organizer_id"]; ok {
		where = append(where, fmt.Sprintf("organizer_id = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := filter["is_featured"]; ok {
		where = append(where, fmt.Sprintf("is_featured = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := filter["upcoming"]; ok && val.(bool) {
		where = append(where, fmt.Sprintf("starts_at > NOW()"))
	}

	whereClause := strings.Join(where, " AND ")

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticketing.events WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Obtener datos
	query := fmt.Sprintf(`
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at, doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees, tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description, settings,
			published_at, created_at, updated_at
		FROM ticketing.events 
		WHERE %s
		ORDER BY starts_at
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []*entities.Event
	for rows.Next() {
		var e entities.Event
		err := rows.Scan(
			&e.ID, &e.PublicID, &e.OrganizerID, &e.PrimaryCategoryID, &e.VenueID,
			&e.Slug, &e.Name, &e.ShortDescription, &e.Description, &e.EventType,
			&e.CoverImageURL, &e.BannerImageURL, &e.GalleryImages,
			&e.Timezone, &e.StartsAt, &e.EndsAt, &e.DoorsOpenAt, &e.DoorsCloseAt,
			&e.VenueName, &e.AddressFull, &e.City, &e.State, &e.Country,
			&e.Status, &e.Visibility, &e.IsFeatured, &e.IsFree,
			&e.MaxAttendees, &e.MinAttendees, &e.Tags, &e.AgeRestriction,
			&e.RequiresApproval, &e.AllowReservations, &e.ReservationDuration,
			&e.ViewCount, &e.FavoriteCount, &e.ShareCount,
			&e.MetaTitle, &e.MetaDescription, &e.Settings,
			&e.PublishedAt, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &e)
	}

	return events, total, nil
}

// ListByOrganizer lista eventos de un organizador
func (r *eventRepository) ListByOrganizer(ctx context.Context, organizerID int64, limit, offset int) ([]*entities.Event, int64, error) {
	filter := map[string]interface{}{
		"organizer_id": organizerID,
	}
	return r.List(ctx, filter, limit, offset)
}

// ListUpcoming lista eventos próximos
func (r *eventRepository) ListUpcoming(ctx context.Context, limit int) ([]*entities.Event, error) {
	filter := map[string]interface{}{
		"upcoming": true,
	}
	events, _, err := r.List(ctx, filter, limit, 0)
	return events, err
}

// ListFeatured lista eventos destacados
func (r *eventRepository) ListFeatured(ctx context.Context, limit int) ([]*entities.Event, error) {
	filter := map[string]interface{}{
		"is_featured": true,
		"status":      "published",
	}
	events, _, err := r.List(ctx, filter, limit, 0)
	return events, err
}

// GetEventCategories obtiene categorías de un evento
func (r *eventRepository) GetEventCategories(ctx context.Context, eventID int64) ([]*entities.Category, error) {
	query := `
		SELECT c.*
		FROM ticketing.categories c
		JOIN ticketing.event_categories ec ON c.id = ec.category_id
		WHERE ec.event_id = $1
		ORDER BY 
			CASE WHEN ec.is_primary THEN 0 ELSE 1 END,
			c.sort_order, c.name
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event categories: %w", err)
	}
	defer rows.Close()

	var categories []*entities.Category
	for rows.Next() {
		var c entities.Category
		err := rows.Scan(
			&c.ID, &c.PublicID, &c.Name, &c.Slug,
			&c.Description, &c.Icon, &c.ColorHex,
			&c.ParentID, &c.Level, &c.Path,
			&c.IsActive, &c.IsFeatured, &c.SortOrder,
			&c.MetaTitle, &c.MetaDescription,
			&c.TotalEvents, &c.TotalTicketsSold, &c.TotalRevenue,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &c)
	}

	return categories, nil
}

// AddCategoryToEvent asocia una categoría a un evento
func (r *eventRepository) AddCategoryToEvent(ctx context.Context, eventID, categoryID int64, isPrimary bool) error {
	query := `
		INSERT INTO ticketing.event_categories (event_id, category_id, is_primary, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (event_id, category_id) 
		DO UPDATE SET is_primary = EXCLUDED.is_primary
	`

	_, err := r.db.Exec(ctx, query, eventID, categoryID, isPrimary)
	if err != nil {
		return fmt.Errorf("failed to add category to event: %w", err)
	}
	return nil
}

// RemoveCategoryFromEvent elimina asociación evento-categoría
func (r *eventRepository) RemoveCategoryFromEvent(ctx context.Context, eventID, categoryID int64) error {
	query := `DELETE FROM ticketing.event_categories WHERE event_id = $1 AND category_id = $2`
	_, err := r.db.Exec(ctx, query, eventID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to remove category from event: %w", err)
	}
	return nil
}
