package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/errors"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/query"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/scanner"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/types"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/utils"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/validations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// eventRepository implementa repository.EventRepository usando helpers
type eventRepository struct {
	db         *pgxpool.Pool
	converter  *types.Converter
	scanner    *scanner.RowScanner
	errHandler *errors.PostgresErrorHandler
	validator  *errors.Validator
	logger     *utils.Logger
}

// NewEventRepository crea una nueva instancia con helpers
func NewEventRepository(db *pgxpool.Pool) repository.EventRepository {
	return &eventRepository{
		db:         db,
		converter:  types.NewConverter(),
		scanner:    scanner.NewRowScanner(),
		errHandler: errors.NewPostgresErrorHandler(),
		validator:  errors.NewValidator(),
		logger:     utils.NewLogger("event-repository"),
	}
}

// Create implementa repository.EventRepository.Create usando helpers
func (r *eventRepository) Create(ctx context.Context, event *entities.Event) error {
	startTime := time.Now()

	// Validaciones usando helpers
	if err := r.validateEventForCreate(ctx, event); err != nil {
		return err
	}

	// Generar public_uuid si no existe
	if event.PublicID == "" {
		event.PublicID = uuid.New().String()
	}

	// Validar UUID usando validations
	if !validations.IsValidUUID(event.PublicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	// Validar status usando enums
	status := enums.EventStatus(event.Status)
	if !status.IsValid() {
		return fmt.Errorf("invalid event status: %s", event.Status)
	}

	// Serializar JSON usando helpers
	galleryJSON, err := r.marshalJSON(event.GalleryImages, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal gallery images: %w", err)
	}

	tagsJSON, err := r.marshalJSON(event.Tags, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	settingsJSON, err := r.marshalJSON(event.Settings, "{}")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Validar fechas usando utils
	if !r.validateEventDates(event) {
		return fmt.Errorf("invalid event dates: ends_at must be after starts_at")
	}

	query := `
		INSERT INTO ticketing.events (
			public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19, $20, $21, $22,
			$23, $24, $25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37, $38, $39, $40, $41
		)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		event.PublicID,
		r.converter.Int64Ptr(event.OrganizerID),
		r.converter.Int64Ptr(event.PrimaryCategoryID),
		r.converter.Int64Ptr(event.VenueID),
		r.converter.Text(event.Slug),
		r.converter.Text(event.Name),
		r.converter.TextPtr(event.ShortDescription),
		r.converter.TextPtr(event.Description),
		r.converter.TextPtr(event.EventType),
		r.converter.TextPtr(event.CoverImageURL),
		r.converter.TextPtr(event.BannerImageURL),
		galleryJSON,
		r.converter.TextPtr(event.Timezone),
		r.converter.TimestampPtr(event.StartsAt),
		r.converter.TimestampPtr(event.EndsAt),
		r.converter.TimestampPtr(event.DoorsOpenAt),
		r.converter.TimestampPtr(event.DoorsCloseAt),
		r.converter.TextPtr(event.VenueName),
		r.converter.TextPtr(event.AddressFull),
		r.converter.TextPtr(event.City),
		r.converter.TextPtr(event.State),
		r.converter.TextPtr(event.Country),
		r.converter.Text(event.Status),
		r.converter.TextPtr(event.Visibility),
		r.converter.BoolPtr(event.IsFeatured),
		r.converter.BoolPtr(event.IsFree),
		r.converter.Int32Ptr(event.MaxAttendees),
		r.converter.Int32Ptr(event.MinAttendees),
		tagsJSON,
		r.converter.Int32Ptr(event.AgeRestriction),
		r.converter.BoolPtr(event.RequiresApproval),
		r.converter.BoolPtr(event.AllowReservations),
		r.converter.Int32Ptr(event.ReservationDurationMinutes),
		r.converter.Int64Ptr(event.ViewCount),
		r.converter.Int64Ptr(event.FavoriteCount),
		r.converter.Int64Ptr(event.ShareCount),
		r.converter.TextPtr(event.MetaTitle),
		r.converter.TextPtr(event.MetaDescription),
		settingsJSON,
		r.converter.TimestampPtr(event.PublishedAt),
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		if r.errHandler.IsDuplicateKey(err) {
			constraint := r.errHandler.GetConstraintName(err)
			if strings.Contains(strings.ToLower(constraint), "slug") {
				return fmt.Errorf("slug already exists: %s", event.Slug)
			} else if strings.Contains(strings.ToLower(constraint), "public_uuid") {
				return fmt.Errorf("public_uuid already exists: %s", event.PublicID)
			}
		}

		if r.errHandler.IsForeignKeyViolation(err) {
			table := r.errHandler.GetTableName(err)
			if strings.Contains(strings.ToLower(table), "organizer") {
				return fmt.Errorf("organizer not found")
			} else if strings.Contains(strings.ToLower(table), "venue") {
				return fmt.Errorf("venue not found")
			} else if strings.Contains(strings.ToLower(table), "category") {
				return fmt.Errorf("primary category not found")
			}
		}

		r.logger.DatabaseLogger("INSERT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"slug":      event.Slug,
			"name":      utils.SafeStringForLog(event.Name),
			"public_id": event.PublicID,
		})

		return r.errHandler.WrapError(err, "event repository", "create event")
	}

	r.logger.DatabaseLogger("INSERT", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id": event.ID,
		"slug":     event.Slug,
		"name":     utils.SafeStringForLog(event.Name),
	})

	return nil
}

// FindByID implementa repository.EventRepository.FindByID usando scanner
func (r *eventRepository) FindByID(ctx context.Context, id int64) (*entities.Event, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	event, err := r.scanEvent(row)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Event not found", map[string]interface{}{
				"event_id": id,
			})
			return nil, fmt.Errorf("event not found")
		}

		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": id,
		})

		return nil, r.errHandler.WrapError(err, "event repository", "find event by ID")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id": id,
	})

	return event, nil
}

// FindByPublicID implementa repository.EventRepository.FindByPublicID
func (r *eventRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Event, error) {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return nil, fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE public_uuid = $1
	`

	row := r.db.QueryRow(ctx, query, publicID)
	event, err := r.scanEvent(row)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Event not found by public ID", map[string]interface{}{
				"public_id": publicID,
			})
			return nil, fmt.Errorf("event not found")
		}

		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return nil, r.errHandler.WrapError(err, "event repository", "find event by public ID")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id":  event.ID,
		"public_id": publicID,
	})

	return event, nil
}

// FindBySlug implementa repository.EventRepository.FindBySlug
func (r *eventRepository) FindBySlug(ctx context.Context, slug string) (*entities.Event, error) {
	startTime := time.Now()

	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE slug = $1
	`

	row := r.db.QueryRow(ctx, query, slug)
	event, err := r.scanEvent(row)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Event not found by slug", map[string]interface{}{
				"slug": slug,
			})
			return nil, fmt.Errorf("event not found")
		}

		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"slug": slug,
		})

		return nil, r.errHandler.WrapError(err, "event repository", "find event by slug")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id": event.ID,
		"slug":     slug,
	})

	return event, nil
}

// Update implementa repository.EventRepository.Update usando helpers
func (r *eventRepository) Update(ctx context.Context, event *entities.Event) error {
	startTime := time.Now()

	// Validaciones
	if err := r.validateEventForUpdate(ctx, event); err != nil {
		return err
	}

	// Validar status usando enums
	status := enums.EventStatus(event.Status)
	if !status.IsValid() {
		return fmt.Errorf("invalid event status: %s", event.Status)
	}

	// Serializar JSON usando helpers
	galleryJSON, err := r.marshalJSON(event.GalleryImages, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal gallery images: %w", err)
	}

	tagsJSON, err := r.marshalJSON(event.Tags, "[]")
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	settingsJSON, err := r.marshalJSON(event.Settings, "{}")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Validar fechas
	if !r.validateEventDates(event) {
		return fmt.Errorf("invalid event dates: ends_at must be after starts_at")
	}

	query := `
		UPDATE ticketing.events SET
			organizer_id = $1,
			primary_category_id = $2,
			venue_id = $3,
			slug = $4,
			name = $5,
			short_description = $6,
			description = $7,
			event_type = $8,
			cover_image_url = $9,
			banner_image_url = $10,
			gallery_images = $11,
			timezone = $12,
			starts_at = $13,
			ends_at = $14,
			doors_open_at = $15,
			doors_close_at = $16,
			venue_name = $17,
			address_full = $18,
			city = $19,
			state = $20,
			country = $21,
			status = $22,
			visibility = $23,
			is_featured = $24,
			is_free = $25,
			max_attendees = $26,
			min_attendees = $27,
			tags = $28,
			age_restriction = $29,
			requires_approval = $30,
			allow_reservations = $31,
			reservation_duration_minutes = $32,
			view_count = $33,
			favorite_count = $34,
			share_count = $35,
			meta_title = $36,
			meta_description = $37,
			settings = $38,
			published_at = $39,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $40
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		r.converter.Int64Ptr(event.OrganizerID),
		r.converter.Int64Ptr(event.PrimaryCategoryID),
		r.converter.Int64Ptr(event.VenueID),
		r.converter.Text(event.Slug),
		r.converter.Text(event.Name),
		r.converter.TextPtr(event.ShortDescription),
		r.converter.TextPtr(event.Description),
		r.converter.TextPtr(event.EventType),
		r.converter.TextPtr(event.CoverImageURL),
		r.converter.TextPtr(event.BannerImageURL),
		galleryJSON,
		r.converter.TextPtr(event.Timezone),
		r.converter.TimestampPtr(event.StartsAt),
		r.converter.TimestampPtr(event.EndsAt),
		r.converter.TimestampPtr(event.DoorsOpenAt),
		r.converter.TimestampPtr(event.DoorsCloseAt),
		r.converter.TextPtr(event.VenueName),
		r.converter.TextPtr(event.AddressFull),
		r.converter.TextPtr(event.City),
		r.converter.TextPtr(event.State),
		r.converter.TextPtr(event.Country),
		r.converter.Text(event.Status),
		r.converter.TextPtr(event.Visibility),
		r.converter.BoolPtr(event.IsFeatured),
		r.converter.BoolPtr(event.IsFree),
		r.converter.Int32Ptr(event.MaxAttendees),
		r.converter.Int32Ptr(event.MinAttendees),
		tagsJSON,
		r.converter.Int32Ptr(event.AgeRestriction),
		r.converter.BoolPtr(event.RequiresApproval),
		r.converter.BoolPtr(event.AllowReservations),
		r.converter.Int32Ptr(event.ReservationDurationMinutes),
		r.converter.Int64Ptr(event.ViewCount),
		r.converter.Int64Ptr(event.FavoriteCount),
		r.converter.Int64Ptr(event.ShareCount),
		r.converter.TextPtr(event.MetaTitle),
		r.converter.TextPtr(event.MetaDescription),
		settingsJSON,
		r.converter.TimestampPtr(event.PublishedAt),
		event.ID,
	).Scan(&event.UpdatedAt)

	if err != nil {
		if r.errHandler.IsDuplicateKey(err) {
			constraint := r.errHandler.GetConstraintName(err)
			if strings.Contains(strings.ToLower(constraint), "slug") {
				return fmt.Errorf("slug already exists: %s", event.Slug)
			}
		}

		if r.errHandler.IsForeignKeyViolation(err) {
			table := r.errHandler.GetTableName(err)
			if strings.Contains(strings.ToLower(table), "organizer") {
				return fmt.Errorf("organizer not found")
			} else if strings.Contains(strings.ToLower(table), "venue") {
				return fmt.Errorf("venue not found")
			} else if strings.Contains(strings.ToLower(table), "category") {
				return fmt.Errorf("primary category not found")
			}
		}

		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": event.ID,
			"slug":     event.Slug,
		})

		return r.errHandler.WrapError(err, "event repository", "update event")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id": event.ID,
	})

	return nil
}

// Delete implementa repository.EventRepository.Delete
func (r *eventRepository) Delete(ctx context.Context, id int64) error {
	startTime := time.Now()

	query := `DELETE FROM ticketing.events WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.DatabaseLogger("DELETE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": id,
		})

		return r.errHandler.WrapError(err, "event repository", "delete event")
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		r.logger.Debug("Event not found for deletion", map[string]interface{}{
			"event_id": id,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("DELETE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": id,
	})

	return nil
}

// SoftDelete implementa repository.EventRepository.SoftDelete
func (r *eventRepository) SoftDelete(ctx context.Context, publicID string) error {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	query := `
		UPDATE ticketing.events 
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE public_uuid = $1 AND status NOT IN ('cancelled', 'completed')
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, publicID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Event not found or already cancelled/completed", map[string]interface{}{
				"public_id": publicID,
			})
			return fmt.Errorf("event not found or already cancelled/completed")
		}

		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return r.errHandler.WrapError(err, "event repository", "soft delete event")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id":  id,
		"public_id": publicID,
	})

	return nil
}

// List implementa repository.EventRepository.List usando query builder
func (r *eventRepository) List(ctx context.Context, filter dto.EventFilter, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	startTime := time.Now()

	// Usar query builder para construir la query
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at, created_at, updated_at
		FROM ticketing.events
	`).Where("1=1", nil) // Condición inicial

	// Aplicar filtros usando query builder
	if len(filter.Status) > 0 {
		statusArgs := make([]interface{}, len(filter.Status))
		for i, status := range filter.Status {
			statusArgs[i] = status
		}
		qb.WhereIn("status", statusArgs)
	} else {
		// Por defecto, solo eventos publicados/activos
		defaultStatuses := []interface{}{"published", "live", "scheduled"}
		qb.WhereIn("status", defaultStatuses)
	}

	if filter.OrganizerID != "" {
		subquery := "(SELECT id FROM ticketing.organizers WHERE public_uuid = ?)"
		qb.Where("organizer_id = "+subquery, filter.OrganizerID)
	}

	if filter.CategoryID != "" {
		subquery := "(SELECT id FROM ticketing.categories WHERE public_uuid = ?)"
		qb.Where("primary_category_id = "+subquery, filter.CategoryID)
	}

	if filter.VenueID != "" {
		subquery := "(SELECT id FROM ticketing.venues WHERE public_uuid = ?)"
		qb.Where("venue_id = "+subquery, filter.VenueID)
	}

	if filter.EventType != "" {
		qb.Where("event_type = ?", filter.EventType)
	}

	if filter.Country != "" {
		qb.Where("country = ?", filter.Country)
	}

	if filter.City != "" {
		qb.Where("city = ?", filter.City)
	}

	if filter.State != "" {
		qb.Where("state = ?", filter.State)
	}

	if filter.IsFeatured != nil {
		qb.Where("is_featured = ?", *filter.IsFeatured)
	}

	if filter.IsFree != nil {
		qb.Where("is_free = ?", *filter.IsFree)
	}

	if filter.DateFrom != "" {
		if dateFrom, err := utils.ParseDateFromString(filter.DateFrom); err == nil {
			qb.Where("starts_at >= ?", dateFrom)
		}
	}

	if filter.DateTo != "" {
		if dateTo, err := utils.ParseDateFromString(filter.DateTo); err == nil {
			qb.Where("ends_at <= ?", dateTo)
		}
	}

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		qb.Where("(name ILIKE ? OR description ILIKE ? OR short_description ILIKE ?)",
			searchTerm, searchTerm, searchTerm)
	}

	// Ordenar por fecha de inicio
	qb.OrderBy("starts_at", false) // ASC

	// Construir query de conteo
	countQuery, countArgs := qb.BuildCount()

	// Ejecutar count
	var total int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "count",
		})

		return nil, 0, r.errHandler.WrapError(err, "event repository", "count events")
	}

	// Aplicar paginación
	limit := pagination.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := (pagination.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	qb.Limit(limit).Offset(offset)

	// Construir query principal
	queryStr, args := qb.Build()

	// Ejecutar query principal
	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "event repository", "list events")
	}
	defer rows.Close()

	// Usar scanner para procesar resultados
	events := []*entities.Event{}
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			r.logger.Error("Failed to scan event row", err, map[string]interface{}{
				"operation": "list",
			})
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "event repository", "iterate events")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), int64(len(events)), nil, map[string]interface{}{
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
		"found":     len(events),
	})

	return events, total, nil
}

// FindByOrganizer implementa repository.EventRepository.FindByOrganizer
func (r *eventRepository) FindByOrganizer(ctx context.Context, organizerID int64, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	filter := dto.EventFilter{}
	// Aquí deberías mapear organizerID a string si es necesario
	return r.List(ctx, filter, pagination)
}

// FindByCategory implementa repository.EventRepository.FindByCategory
func (r *eventRepository) FindByCategory(ctx context.Context, categoryID int64, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	filter := dto.EventFilter{}
	// Similar a FindByOrganizer
	return r.List(ctx, filter, pagination)
}

// FindByVenue implementa repository.EventRepository.FindByVenue
func (r *eventRepository) FindByVenue(ctx context.Context, venueID int64, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	filter := dto.EventFilter{}
	// Similar a FindByOrganizer
	return r.List(ctx, filter, pagination)
}

// FindByStatus implementa repository.EventRepository.FindByStatus
func (r *eventRepository) FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	filter := dto.EventFilter{Status: []string{status}}
	return r.List(ctx, filter, pagination)
}

// FindByDateRange implementa repository.EventRepository.FindByDateRange
func (r *eventRepository) FindByDateRange(ctx context.Context, startDate, endDate string, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	filter := dto.EventFilter{DateFrom: startDate, DateTo: endDate}
	return r.List(ctx, filter, pagination)
}

// FindUpcoming implementa repository.EventRepository.FindUpcoming
func (r *eventRepository) FindUpcoming(ctx context.Context, limit int) ([]*entities.Event, error) {
	startTime := time.Now()

	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE status IN ('published', 'live', 'scheduled')
		  AND starts_at > CURRENT_TIMESTAMP
		ORDER BY starts_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "find_upcoming",
			"limit":     limit,
		})

		return nil, r.errHandler.WrapError(err, "event repository", "find upcoming events")
	}
	defer rows.Close()

	events := []*entities.Event{}
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			r.logger.Error("Failed to scan event row", err, map[string]interface{}{
				"operation": "find_upcoming",
			})
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "find_upcoming",
		})

		return nil, r.errHandler.WrapError(err, "event repository", "iterate upcoming events")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), int64(len(events)), nil, map[string]interface{}{
		"limit": limit,
		"found": len(events),
	})

	return events, nil
}

// FindFeatured implementa repository.EventRepository.FindFeatured
func (r *eventRepository) FindFeatured(ctx context.Context, limit int) ([]*entities.Event, error) {
	startTime := time.Now()

	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			id, public_uuid, organizer_id, primary_category_id, venue_id,
			slug, name, short_description, description, event_type,
			cover_image_url, banner_image_url, gallery_images,
			timezone, starts_at, ends_at,
			doors_open_at, doors_close_at,
			venue_name, address_full, city, state, country,
			status, visibility, is_featured, is_free,
			max_attendees, min_attendees,
			tags, age_restriction,
			requires_approval, allow_reservations, reservation_duration_minutes,
			view_count, favorite_count, share_count,
			meta_title, meta_description,
			settings,
			published_at, created_at, updated_at
		FROM ticketing.events
		WHERE is_featured = true
		  AND status IN ('published', 'live', 'scheduled')
		  AND starts_at > CURRENT_TIMESTAMP
		ORDER BY starts_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "find_featured",
			"limit":     limit,
		})

		return nil, r.errHandler.WrapError(err, "event repository", "find featured events")
	}
	defer rows.Close()

	events := []*entities.Event{}
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			r.logger.Error("Failed to scan event row", err, map[string]interface{}{
				"operation": "find_featured",
			})
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "find_featured",
		})

		return nil, r.errHandler.WrapError(err, "event repository", "iterate featured events")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), int64(len(events)), nil, map[string]interface{}{
		"limit": limit,
		"found": len(events),
	})

	return events, nil
}

// Search implementa repository.EventRepository.Search
func (r *eventRepository) Search(ctx context.Context, term string, filter dto.EventFilter, pagination dto.Pagination) ([]*entities.Event, int64, error) {
	filter.Search = term
	return r.List(ctx, filter, pagination)
}

// UpdateStatus implementa repository.EventRepository.UpdateStatus
func (r *eventRepository) UpdateStatus(ctx context.Context, eventID int64, status string) error {
	startTime := time.Now()

	// Validar status usando enums
	eventStatus := enums.EventStatus(status)
	if !eventStatus.IsValid() {
		return fmt.Errorf("invalid event status: %s", status)
	}

	query := `
		UPDATE ticketing.events 
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": eventID,
			"status":   status,
		})

		return r.errHandler.WrapError(err, "event repository", "update event status")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found for status update", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
		"status":   status,
	})

	return nil
}

// UpdateVisibility implementa repository.EventRepository.UpdateVisibility
func (r *eventRepository) UpdateVisibility(ctx context.Context, eventID int64, visibility string) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET visibility = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, visibility, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":   eventID,
			"visibility": visibility,
		})

		return r.errHandler.WrapError(err, "event repository", "update event visibility")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found for visibility update", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id":   eventID,
		"visibility": visibility,
	})

	return nil
}

// IncrementViewCount implementa repository.EventRepository.IncrementViewCount
func (r *eventRepository) IncrementViewCount(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET view_count = view_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "increment_view_count",
		})

		return r.errHandler.WrapError(err, "event repository", "increment view count")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found for view count increment", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// IncrementFavoriteCount implementa repository.EventRepository.IncrementFavoriteCount
func (r *eventRepository) IncrementFavoriteCount(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET favorite_count = favorite_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "increment_favorite_count",
		})

		return r.errHandler.WrapError(err, "event repository", "increment favorite count")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found for favorite count increment", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// DecrementFavoriteCount implementa repository.EventRepository.DecrementFavoriteCount
func (r *eventRepository) DecrementFavoriteCount(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET favorite_count = GREATEST(0, favorite_count - 1), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "decrement_favorite_count",
		})

		return r.errHandler.WrapError(err, "event repository", "decrement favorite count")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found for favorite count decrement", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// IncrementShareCount implementa repository.EventRepository.IncrementShareCount
func (r *eventRepository) IncrementShareCount(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET share_count = share_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "increment_share_count",
		})

		return r.errHandler.WrapError(err, "event repository", "increment share count")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found for share count increment", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// Publish implementa repository.EventRepository.Publish
func (r *eventRepository) Publish(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET status = 'published', 
		    published_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('draft', 'scheduled')
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "publish",
		})

		return r.errHandler.WrapError(err, "event repository", "publish event")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found or cannot be published", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found or cannot be published")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// Unpublish implementa repository.EventRepository.Unpublish
func (r *eventRepository) Unpublish(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET status = 'draft', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'published'
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "unpublish",
		})

		return r.errHandler.WrapError(err, "event repository", "unpublish event")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found or not published", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found or not published")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// Cancel implementa repository.EventRepository.Cancel
func (r *eventRepository) Cancel(ctx context.Context, eventID int64, reason string) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status NOT IN ('cancelled', 'completed')
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "cancel",
			"reason":    utils.SafeStringForLog(reason),
		})

		return r.errHandler.WrapError(err, "event repository", "cancel event")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found or already cancelled/completed", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found or already cancelled/completed")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
		"reason":   utils.SafeStringForLog(reason),
	})

	return nil
}

// Complete implementa repository.EventRepository.Complete
func (r *eventRepository) Complete(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET status = 'completed', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('published', 'live') AND ends_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "complete",
		})

		return r.errHandler.WrapError(err, "event repository", "complete event")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found, not active, or not ended yet", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found, not active, or not ended yet")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// MarkAsSoldOut implementa repository.EventRepository.MarkAsSoldOut
func (r *eventRepository) MarkAsSoldOut(ctx context.Context, eventID int64) error {
	startTime := time.Now()

	query := `
		UPDATE ticketing.events 
		SET status = 'sold_out', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('published', 'live')
	`

	result, err := r.db.Exec(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "mark_sold_out",
		})

		return r.errHandler.WrapError(err, "event repository", "mark event as sold out")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Event not found or not active", map[string]interface{}{
			"event_id": eventID,
		})
		return fmt.Errorf("event not found or not active")
	}

	r.logger.DatabaseLogger("UPDATE", "ticketing.events", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return nil
}

// GetStats implementa repository.EventRepository.GetStats
func (r *eventRepository) GetStats(ctx context.Context, eventID int64) (*dto.EventStatsResponse, error) {
	startTime := time.Now()

	query := `
		SELECT 
			e.view_count,
			e.favorite_count,
			e.share_count,
			COALESCE(SUM(tt.sold_quantity), 0) as tickets_sold,
			COALESCE(SUM(tt.total_quantity), 0) as total_capacity,
			COALESCE(SUM(tt.sold_quantity * tt.base_price), 0) as total_revenue,
			COALESCE(AVG(tt.base_price), 0) as avg_ticket_price
		FROM ticketing.events e
		LEFT JOIN ticketing.ticket_types tt ON e.id = tt.event_id
		WHERE e.id = $1
		GROUP BY e.id, e.view_count, e.favorite_count, e.share_count
	`

	var stats dto.EventStatsResponse
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&stats.ViewsToday,
		&stats.Favorites,
		&stats.Shares,
		&stats.TicketsSold,
		&stats.TicketsAvailable,
		&stats.TotalRevenue,
		&stats.AvgTicketPrice,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Event not found for stats", map[string]interface{}{
				"event_id": eventID,
			})
			return nil, fmt.Errorf("event not found")
		}

		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "get_stats",
		})

		return nil, r.errHandler.WrapError(err, "event repository", "get event stats")
	}

	// Calcular tasas (en una implementación real, usarías datos de analytics)
	stats.ConversionRate = 0.0
	stats.CheckInRate = 0.0
	stats.SalesVelocity = 0.0

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id": eventID,
	})

	return &stats, nil
}

// GetGlobalStats implementa repository.EventRepository.GetGlobalStats
func (r *eventRepository) GetGlobalStats(ctx context.Context) (*dto.EventGlobalStats, error) {
	startTime := time.Now()

	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(CASE WHEN status IN ('published', 'live', 'scheduled') THEN 1 END) as active_events,
			COUNT(CASE WHEN status = 'sold_out' THEN 1 END) as sold_out_events,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_events,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_events,
			SUM(view_count) as total_views,
			SUM(favorite_count) as total_favorites,
			COALESCE(SUM(tt.sold_quantity), 0) as total_tickets_sold,
			COALESCE(SUM(tt.sold_quantity * tt.base_price), 0) as total_revenue
		FROM ticketing.events e
		LEFT JOIN ticketing.ticket_types tt ON e.id = tt.event_id
	`

	var stats dto.EventGlobalStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalEvents,
		&stats.ActiveEvents,
		&stats.SoldOutEvents,
		&stats.CompletedEvents,
		&stats.CancelledEvents,
		&stats.TotalViews,
		&stats.TotalFavorites,
		&stats.TotalTicketsSold,
		&stats.TotalRevenue,
	)

	if err != nil {
		r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_global_stats",
		})

		return nil, r.errHandler.WrapError(err, "event repository", "get global event stats")
	}

	r.logger.DatabaseLogger("SELECT", "ticketing.events", time.Since(startTime), 1, nil, map[string]interface{}{
		"operation": "get_global_stats",
	})

	return &stats, nil
}

// =============================================================================
// MÉTODOS PRIVADOS CON HELPERS
// =============================================================================

// scanEvent escanea una fila de evento usando scanner
func (r *eventRepository) scanEvent(row interface{}) (*entities.Event, error) {
	var event entities.Event
	var galleryJSON, tagsJSON, settingsJSON []byte
	var galleryImages []string
	var tags []string
	var settings map[string]interface{}

	// Escanear los valores básicos
	err := r.scanner.ScanRowToMap(row, []string{
		"id", "public_uuid", "organizer_id", "primary_category_id", "venue_id",
		"slug", "name", "short_description", "description", "event_type",
		"cover_image_url", "banner_image_url", "gallery_images",
		"timezone", "starts_at", "ends_at", "doors_open_at", "doors_close_at",
		"venue_name", "address_full", "city", "state", "country",
		"status", "visibility", "is_featured", "is_free",
		"max_attendees", "min_attendees", "age_restriction",
		"requires_approval", "allow_reservations", "reservation_duration_minutes",
		"view_count", "favorite_count", "share_count",
		"meta_title", "meta_description",
		"published_at", "created_at", "updated_at",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan event row: %w", err)
	}

	// Deserializar JSON usando helpers
	if len(galleryJSON) > 0 {
		if err := json.Unmarshal(galleryJSON, &galleryImages); err != nil {
			r.logger.Error("Failed to unmarshal gallery images", err)
		} else {
			event.GalleryImages = &galleryImages
		}
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &tags); err != nil {
			r.logger.Error("Failed to unmarshal tags", err)
		} else {
			event.Tags = &tags
		}
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &settings); err != nil {
			r.logger.Error("Failed to unmarshal settings", err)
		} else {
			event.Settings = settings
		}
	}

	return &event, nil
}

// validateEventForCreate valida un evento para creación
func (r *eventRepository) validateEventForCreate(ctx context.Context, event *entities.Event) error {
	// Usar validator
	r.validator.Required("name", event.Name).
		Required("slug", event.Slug).
		Required("status", event.Status).
		MinLength("slug", event.Slug, 3).
		MaxLength("slug", event.Slug, 100).
		MaxLength("name", event.Name, 200)

	if event.ShortDescription != nil {
		r.validator.MaxLength("short_description", *event.ShortDescription, 500)
	}

	if event.Description != nil {
		r.validator.MaxLength("description", *event.Description, 5000)
	}

	if event.MaxAttendees != nil && event.MinAttendees != nil {
		if *event.MaxAttendees < *event.MinAttendees {
			r.validator.Custom("max_attendees", false, "max_attendees must be greater than or equal to min_attendees")
		}
	}

	if validationErr := r.validator.Validate(); validationErr != nil {
		return validationErr
	}

	// Validar slugs duplicados
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM ticketing.events WHERE slug = $1)", event.Slug).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check slug uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("slug already exists: %s", event.Slug)
	}

	return nil
}

// validateEventForUpdate valida un evento para actualización
func (r *eventRepository) validateEventForUpdate(ctx context.Context, event *entities.Event) error {
	// Validar que el evento exista
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM ticketing.events WHERE id = $1)", event.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate event existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("event not found with ID: %d", event.ID)
	}

	// Usar validator para validaciones generales
	return r.validateEventForCreate(ctx, event)
}

// validateEventDates valida las fechas del evento usando utils
func (r *eventRepository) validateEventDates(event *entities.Event) bool {
	if event.StartsAt == nil || event.EndsAt == nil {
		return true // Fechas opcionales
	}

	return event.EndsAt.After(*event.StartsAt)
}

// marshalJSON serializa JSON con valor por defecto usando helpers
func (r *eventRepository) marshalJSON(data interface{}, defaultValue string) ([]byte, error) {
	if data == nil {
		return []byte(defaultValue), nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Si está vacío, usar valor por defecto
	if len(jsonBytes) == 0 || string(jsonBytes) == "null" {
		return []byte(defaultValue), nil
	}

	return jsonBytes, nil
}
