// osmi/osmi-server/internal/infrastructure/repositories/postgres/event_repository.go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// EventRepository implementa la interfaz repository.EventRepository usando PostgreSQL
type EventRepository struct {
	db *pgxpool.Pool
}

// NewEventRepository crea una nueva instancia del repositorio
func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL
func (r *EventRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Para pgx, los errores son diferentes
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("event not found")
	}

	// Verificar si es un error de PostgreSQL con código
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "events_slug_key") {
				return fmt.Errorf("event slug already exists")
			}
			if strings.Contains(pgErr.ConstraintName, "events_public_uuid_key") {
				return fmt.Errorf("event public_uuid already exists")
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced record not found: %w", err)
		}
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

	err = r.db.QueryRow(ctx, query,
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
	var organizerID, primaryCategoryID, venueID *int64
	var coverImageURL, bannerImageURL, venueName, addressFull, city, state, country, metaTitle, metaDescription *string
	var shortDescription, description, eventType *string
	var doorsOpenAt, doorsCloseAt, publishedAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.PublicID, &organizerID, &primaryCategoryID, &venueID,
		&event.Slug, &event.Name, &shortDescription, &description, &eventType,
		&coverImageURL, &bannerImageURL, &galleryImagesJSON,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &doorsOpenAt, &doorsCloseAt,
		&venueName, &addressFull, &city, &state, &country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&metaTitle, &metaDescription, &settingsJSON,
		&publishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %d", id)
		}
		return nil, r.handleError(err, "failed to get event by ID")
	}

	// Asignar campos NULL
	event.OrganizerID = organizerID
	event.PrimaryCategoryID = primaryCategoryID
	event.VenueID = venueID
	event.CoverImageURL = coverImageURL
	event.BannerImageURL = bannerImageURL
	event.VenueName = venueName
	event.AddressFull = addressFull
	event.City = city
	event.State = state
	event.Country = country
	event.MetaTitle = metaTitle
	event.MetaDescription = metaDescription
	event.ShortDescription = shortDescription
	event.Description = description
	event.EventType = eventType
	event.DoorsOpenAt = doorsOpenAt
	event.DoorsCloseAt = doorsCloseAt
	event.PublishedAt = publishedAt

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
	var organizerID, primaryCategoryID, venueID *int64
	var coverImageURL, bannerImageURL, venueName, addressFull, city, state, country, metaTitle, metaDescription *string
	var shortDescription, description, eventType *string
	var doorsOpenAt, doorsCloseAt, publishedAt *time.Time

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&event.ID, &event.PublicID, &organizerID, &primaryCategoryID, &venueID,
		&event.Slug, &event.Name, &shortDescription, &description, &eventType,
		&coverImageURL, &bannerImageURL, &galleryImagesJSON,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &doorsOpenAt, &doorsCloseAt,
		&venueName, &addressFull, &city, &state, &country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&metaTitle, &metaDescription, &settingsJSON,
		&publishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %s", publicID)
		}
		return nil, r.handleError(err, "failed to get event by public ID")
	}

	// Asignar campos NULL
	event.OrganizerID = organizerID
	event.PrimaryCategoryID = primaryCategoryID
	event.VenueID = venueID
	event.CoverImageURL = coverImageURL
	event.BannerImageURL = bannerImageURL
	event.VenueName = venueName
	event.AddressFull = addressFull
	event.City = city
	event.State = state
	event.Country = country
	event.MetaTitle = metaTitle
	event.MetaDescription = metaDescription
	event.ShortDescription = shortDescription
	event.Description = description
	event.EventType = eventType
	event.DoorsOpenAt = doorsOpenAt
	event.DoorsCloseAt = doorsCloseAt
	event.PublishedAt = publishedAt

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
	var organizerID, primaryCategoryID, venueID *int64
	var coverImageURL, bannerImageURL, venueName, addressFull, city, state, country, metaTitle, metaDescription *string
	var shortDescription, description, eventType *string
	var doorsOpenAt, doorsCloseAt, publishedAt *time.Time

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&event.ID, &event.PublicID, &organizerID, &primaryCategoryID, &venueID,
		&event.Slug, &event.Name, &shortDescription, &description, &eventType,
		&coverImageURL, &bannerImageURL, &galleryImagesJSON,
		&event.Timezone, &event.StartsAt, &event.EndsAt, &doorsOpenAt, &doorsCloseAt,
		&venueName, &addressFull, &city, &state, &country,
		&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
		&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
		&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
		&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
		&metaTitle, &metaDescription, &settingsJSON,
		&publishedAt, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found: %s", slug)
		}
		return nil, r.handleError(err, "failed to get event by slug")
	}

	// Asignar campos NULL
	event.OrganizerID = organizerID
	event.PrimaryCategoryID = primaryCategoryID
	event.VenueID = venueID
	event.CoverImageURL = coverImageURL
	event.BannerImageURL = bannerImageURL
	event.VenueName = venueName
	event.AddressFull = addressFull
	event.City = city
	event.State = state
	event.Country = country
	event.MetaTitle = metaTitle
	event.MetaDescription = metaDescription
	event.ShortDescription = shortDescription
	event.Description = description
	event.EventType = eventType
	event.DoorsOpenAt = doorsOpenAt
	event.DoorsCloseAt = doorsCloseAt
	event.PublishedAt = publishedAt

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

	err = r.db.QueryRow(ctx, query,
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
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM ticketing.events WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete event")
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %d", id)
	}

	return nil
}

// List devuelve eventos con filtros
func (r *EventRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*entities.Event, int64, error) {
	where := []string{"1=1"}
	args := pgx.NamedArgs{}
	argPos := 1

	if val, ok := filter["name"]; ok {
		where = append(where, fmt.Sprintf("name ILIKE @name_%d", argPos))
		args[fmt.Sprintf("name_%d", argPos)] = "%" + val.(string) + "%"
		argPos++
	}
	if val, ok := filter["organizer_id"]; ok {
		where = append(where, fmt.Sprintf("organizer_id = @org_%d", argPos))
		args[fmt.Sprintf("org_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["status"]; ok {
		where = append(where, fmt.Sprintf("status = @status_%d", argPos))
		args[fmt.Sprintf("status_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["city"]; ok {
		where = append(where, fmt.Sprintf("city = @city_%d", argPos))
		args[fmt.Sprintf("city_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["country"]; ok {
		where = append(where, fmt.Sprintf("country = @country_%d", argPos))
		args[fmt.Sprintf("country_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["is_featured"]; ok {
		where = append(where, fmt.Sprintf("is_featured = @featured_%d", argPos))
		args[fmt.Sprintf("featured_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["is_free"]; ok {
		where = append(where, fmt.Sprintf("is_free = @free_%d", argPos))
		args[fmt.Sprintf("free_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["date_from"]; ok {
		where = append(where, fmt.Sprintf("starts_at >= @date_from_%d", argPos))
		args[fmt.Sprintf("date_from_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["date_to"]; ok {
		where = append(where, fmt.Sprintf("ends_at <= @date_to_%d", argPos))
		args[fmt.Sprintf("date_to_%d", argPos)] = val
		argPos++
	}
	if val, ok := filter["search"]; ok {
		searchTerm := "%" + val.(string) + "%"
		where = append(where, fmt.Sprintf("(name ILIKE @search_%d OR description ILIKE @search_%d)", argPos, argPos))
		args[fmt.Sprintf("search_%d", argPos)] = searchTerm
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticketing.events WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
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
		LIMIT @limit OFFSET @offset
	`, whereClause)

	args["limit"] = limit
	args["offset"] = offset

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to list events")
	}
	defer rows.Close()

	var events []*entities.Event
	for rows.Next() {
		var event entities.Event
		var galleryImagesJSON, tagsJSON, settingsJSON []byte
		var organizerID, primaryCategoryID, venueID *int64
		var coverImageURL, bannerImageURL, venueName, addressFull, city, state, country, metaTitle, metaDescription *string
		var shortDescription, description, eventType *string
		var doorsOpenAt, doorsCloseAt, publishedAt *time.Time

		err = rows.Scan(
			&event.ID, &event.PublicID, &organizerID, &primaryCategoryID, &venueID,
			&event.Slug, &event.Name, &shortDescription, &description, &eventType,
			&coverImageURL, &bannerImageURL, &galleryImagesJSON,
			&event.Timezone, &event.StartsAt, &event.EndsAt, &doorsOpenAt, &doorsCloseAt,
			&venueName, &addressFull, &city, &state, &country,
			&event.Status, &event.Visibility, &event.IsFeatured, &event.IsFree,
			&event.MaxAttendees, &event.MinAttendees, &tagsJSON, &event.AgeRestriction,
			&event.RequiresApproval, &event.AllowReservations, &event.ReservationDuration,
			&event.ViewCount, &event.FavoriteCount, &event.ShareCount,
			&metaTitle, &metaDescription, &settingsJSON,
			&publishedAt, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan event row")
		}

		// Asignar campos NULL
		event.OrganizerID = organizerID
		event.PrimaryCategoryID = primaryCategoryID
		event.VenueID = venueID
		event.CoverImageURL = coverImageURL
		event.BannerImageURL = bannerImageURL
		event.VenueName = venueName
		event.AddressFull = addressFull
		event.City = city
		event.State = state
		event.Country = country
		event.MetaTitle = metaTitle
		event.MetaDescription = metaDescription
		event.ShortDescription = shortDescription
		event.Description = description
		event.EventType = eventType
		event.DoorsOpenAt = doorsOpenAt
		event.DoorsCloseAt = doorsCloseAt
		event.PublishedAt = publishedAt

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

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to get event categories")
	}
	defer rows.Close()

	var categories []*entities.Category
	for rows.Next() {
		var category entities.Category
		err = rows.Scan(
		// Aquí necesitarías los campos de Category
		// Por simplicidad, asumimos que existe un scan completo
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan category row")
		}
		categories = append(categories, &category)
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

	_, err := r.db.Exec(ctx, query, eventID, categoryID, isPrimary)
	if err != nil {
		return r.handleError(err, "failed to add category to event")
	}
	return nil
}

// RemoveCategoryFromEvent elimina asociación evento-categoría
func (r *EventRepository) RemoveCategoryFromEvent(ctx context.Context, eventID, categoryID int64) error {
	query := `DELETE FROM ticketing.event_categories WHERE event_id = $1 AND category_id = $2`
	_, err := r.db.Exec(ctx, query, eventID, categoryID)
	if err != nil {
		return r.handleError(err, "failed to remove category from event")
	}
	return nil
}

// Exists verifica si existe un evento con el ID dado
func (r *EventRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.events WHERE id = $1)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check event existence")
	}
	return exists, nil
}
