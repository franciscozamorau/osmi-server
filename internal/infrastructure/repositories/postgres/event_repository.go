// osmi/osmi-server/internal/infrastructure/repositories/postgres/event_repository.go
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// EventRepository implementa la interfaz repository.EventRepository usando PostgreSQL
type EventRepository struct {
	db *sqlx.DB
}

// NewEventRepository crea una nueva instancia del repositorio
func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL
func (r *EventRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pqErr.Constraint, "events_slug_key") {
				return fmt.Errorf("event slug already exists")
			}
			if strings.Contains(pqErr.Constraint, "events_public_uuid_key") {
				return fmt.Errorf("event public_uuid already exists")
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced record not found: %w", err)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("event not found")
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Create inserta un nuevo evento (VERSIÓN MEJORADA CON SERIALIZACIÓN JSON)
func (r *EventRepository) Create(ctx context.Context, event *entities.Event) error {
	// Serializar campos JSON
	galleryImagesJSON, err := json.Marshal(event.GalleryImages)
	if err != nil {
		return fmt.Errorf("failed to marshal gallery images: %w", err)
	}

	tagsJSON, err := json.Marshal(event.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	settingsJSON, err := json.Marshal(event.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

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
			gen_random_uuid(), $1, $2, $3,
			$4, $5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21,
			$22, $23, $24, $25,
			$26, $27, $28, $29,
			$30, $31, $32,
			0, 0, 0,
			$33, $34, $35,
			$36, NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
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
		galleryImagesJSON,
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
		tagsJSON,
		event.AgeRestriction,
		event.RequiresApproval,
		event.AllowReservations,
		event.ReservationDuration,
		event.MetaTitle,
		event.MetaDescription,
		settingsJSON,
		event.PublishedAt,
	).Scan(&event.ID, &event.PublicID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create event")
	}

	return nil
}

// GetByID obtiene evento por ID
func (r *EventRepository) GetByID(ctx context.Context, id int64) (*entities.Event, error) {
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
	var galleryImagesJSON, tagsJSON, settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
		&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
		&event.CoverImageURL, &event.BannerImageURL, &galleryImagesJSON,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
		&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&event.MetaTitle, &event.MetaDescription, &settingsJSON,
		&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %d", id)
		}
		return nil, r.handleError(err, "failed to get event by ID")
	}

	// Deserializar JSON
	if len(galleryImagesJSON) > 0 {
		json.Unmarshal(galleryImagesJSON, &event.GalleryImages)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &event.Tags)
	}
	if len(settingsJSON) > 0 {
		json.Unmarshal(settingsJSON, &event.Settings)
	}

	return &event, nil
}

// GetByPublicID obtiene evento por UUID
func (r *EventRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.Event, error) {
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
	var galleryImagesJSON, tagsJSON, settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, publicID).Scan(
		&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
		&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
		&event.CoverImageURL, &event.BannerImageURL, &galleryImagesJSON,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
		&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&event.MetaTitle, &event.MetaDescription, &settingsJSON,
		&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %s", publicID)
		}
		return nil, r.handleError(err, "failed to get event by public ID")
	}

	// Deserializar JSON
	if len(galleryImagesJSON) > 0 {
		json.Unmarshal(galleryImagesJSON, &event.GalleryImages)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &event.Tags)
	}
	if len(settingsJSON) > 0 {
		json.Unmarshal(settingsJSON, &event.Settings)
	}

	return &event, nil
}

// GetBySlug obtiene evento por slug
func (r *EventRepository) GetBySlug(ctx context.Context, slug string) (*entities.Event, error) {
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
	var galleryImagesJSON, tagsJSON, settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
		&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
		&event.CoverImageURL, &event.BannerImageURL, &galleryImagesJSON,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
		&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&event.MetaTitle, &event.MetaDescription, &settingsJSON,
		&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %s", slug)
		}
		return nil, r.handleError(err, "failed to get event by slug")
	}

	// Deserializar JSON
	if len(galleryImagesJSON) > 0 {
		json.Unmarshal(galleryImagesJSON, &event.GalleryImages)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &event.Tags)
	}
	if len(settingsJSON) > 0 {
		json.Unmarshal(settingsJSON, &event.Settings)
	}

	return &event, nil
}

// Update actualiza evento
func (r *EventRepository) Update(ctx context.Context, event *entities.Event) error {
	exists, err := r.Exists(ctx, event.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("event not found: %d", event.ID)
	}

	// Serializar campos JSON para la actualización
	tagsJSON, err := json.Marshal(event.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	settingsJSON, err := json.Marshal(event.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE ticketing.events 
		SET slug = $1, 
			name = $2, 
			short_description = $3, 
			description = $4,
			venue_id = $5, 
			venue_name = $6, 
			address_full = $7, 
			city = $8, 
			state = $9, 
			country = $10,
			starts_at = $11, 
			ends_at = $12, 
			doors_open_at = $13, 
			doors_close_at = $14,
			status = $15, 
			visibility = $16, 
			is_featured = $17, 
			is_free = $18,
			max_attendees = $19, 
			tags = $20, 
			settings = $21,
			updated_at = NOW()
		WHERE id = $22
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
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
		tagsJSON,
		settingsJSON,
		event.ID,
	).Scan(&event.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update event")
	}

	return nil
}

// Delete elimina evento
func (r *EventRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM ticketing.events WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete event")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("event not found: %d", id)
	}

	return nil
}

// List devuelve eventos con filtros (VERSIÓN CORREGIDA CON DESERIALIZACIÓN JSON)
func (r *EventRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*entities.Event, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if val, ok := filter["name"]; ok {
		where = append(where, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+val.(string)+"%")
		argPos++
	}
	if val, ok := filter["organizer_id"]; ok {
		where = append(where, fmt.Sprintf("organizer_id = $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["status"]; ok {
		where = append(where, fmt.Sprintf("status = $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["city"]; ok {
		where = append(where, fmt.Sprintf("city = $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["country"]; ok {
		where = append(where, fmt.Sprintf("country = $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["is_featured"]; ok {
		where = append(where, fmt.Sprintf("is_featured = $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["is_free"]; ok {
		where = append(where, fmt.Sprintf("is_free = $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["date_from"]; ok {
		where = append(where, fmt.Sprintf("starts_at >= $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["date_to"]; ok {
		where = append(where, fmt.Sprintf("ends_at <= $%d", argPos))
		args = append(args, val)
		argPos++
	}
	if val, ok := filter["search"]; ok {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+val.(string)+"%", "%"+val.(string)+"%")
		argPos += 2
	}

	whereClause := strings.Join(where, " AND ")

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticketing.events WHERE %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count events")
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
	`, whereClause, argPos, argPos+1)

	queryArgs := append(args, limit, offset)

	// CORREGIDO: Usar QueryxContext en lugar de SelectContext para poder deserializar JSON manualmente
	rows, err := r.db.QueryxContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to list events")
	}
	defer rows.Close()

	var events []*entities.Event
	for rows.Next() {
		var event entities.Event
		var galleryImagesJSON, tagsJSON, settingsJSON []byte

		err = rows.Scan(
			&event.ID, &event.PublicID, &event.OrganizerID, &event.PrimaryCategoryID, &event.VenueID,
			&event.Slug, &event.Name, &event.ShortDescription, &event.Description, &event.EventType,
			&event.CoverImageURL, &event.BannerImageURL, &galleryImagesJSON,
			&event.Timezone, &event.StartsAt, &event.EndsAt, &event.DoorsOpenAt, &event.DoorsCloseAt,
			&event.VenueName, &event.AddressFull, &event.City, &event.State, &event.Country,
			&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
			&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
			&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
			&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
			&event.MetaTitle, &event.MetaDescription, &settingsJSON,
			&event.PublishedAt, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan event row")
		}

		// Deserializar JSON
		if len(galleryImagesJSON) > 0 {
			json.Unmarshal(galleryImagesJSON, &event.GalleryImages)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &event.Tags)
		}
		if len(settingsJSON) > 0 {
			json.Unmarshal(settingsJSON, &event.Settings)
		}

		events = append(events, &event)
	}

	return events, total, nil
}

// ListByOrganizer lista eventos de un organizador
func (r *EventRepository) ListByOrganizer(ctx context.Context, organizerID int64, limit, offset int) ([]*entities.Event, int64, error) {
	filter := map[string]interface{}{
		"organizer_id": organizerID,
	}
	return r.List(ctx, filter, limit, offset)
}

// ListUpcoming lista eventos próximos
func (r *EventRepository) ListUpcoming(ctx context.Context, limit int) ([]*entities.Event, error) {
	filter := map[string]interface{}{
		"date_from": time.Now(),
	}
	events, _, err := r.List(ctx, filter, limit, 0)
	return events, err
}

// ListFeatured lista eventos destacados
func (r *EventRepository) ListFeatured(ctx context.Context, limit int) ([]*entities.Event, error) {
	filter := map[string]interface{}{
		"is_featured": true,
	}
	events, _, err := r.List(ctx, filter, limit, 0)
	return events, err
}

// GetEventCategories obtiene categorías de un evento
func (r *EventRepository) GetEventCategories(ctx context.Context, eventID int64) ([]*entities.Category, error) {
	query := `
		SELECT c.*
		FROM ticketing.categories c
		JOIN ticketing.event_categories ec ON c.id = ec.category_id
		WHERE ec.event_id = $1
		ORDER BY 
			CASE WHEN ec.is_primary THEN 0 ELSE 1 END,
			c.sort_order, c.name
	`

	var categories []*entities.Category
	err := r.db.SelectContext(ctx, &categories, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to get event categories")
	}

	return categories, nil
}

// AddCategoryToEvent asocia una categoría a un evento
func (r *EventRepository) AddCategoryToEvent(ctx context.Context, eventID, categoryID int64, isPrimary bool) error {
	query := `
		INSERT INTO ticketing.event_categories (event_id, category_id, is_primary, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (event_id, category_id) 
		DO UPDATE SET is_primary = EXCLUDED.is_primary
	`

	_, err := r.db.ExecContext(ctx, query, eventID, categoryID, isPrimary)
	if err != nil {
		return r.handleError(err, "failed to add category to event")
	}
	return nil
}

// RemoveCategoryFromEvent elimina asociación evento-categoría
func (r *EventRepository) RemoveCategoryFromEvent(ctx context.Context, eventID, categoryID int64) error {
	query := `DELETE FROM ticketing.event_categories WHERE event_id = $1 AND category_id = $2`
	_, err := r.db.ExecContext(ctx, query, eventID, categoryID)
	if err != nil {
		return r.handleError(err, "failed to remove category from event")
	}
	return nil
}

// Exists verifica si existe un evento con el ID dado
func (r *EventRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM ticketing.events WHERE id = $1)`, id)
	if err != nil {
		return false, r.handleError(err, "failed to check event existence")
	}
	return exists, nil
}
