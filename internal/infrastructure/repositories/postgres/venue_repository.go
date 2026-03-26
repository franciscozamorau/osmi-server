// internal/infrastructure/repositories/postgres/venue_repository.go
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

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// VenueRepository implementa la interfaz repository.VenueRepository
type VenueRepository struct {
	db *pgxpool.Pool
}

// NewVenueRepository crea una nueva instancia
func NewVenueRepository(db *pgxpool.Pool) *VenueRepository {
	return &VenueRepository{
		db: db,
	}
}

// handleError maneja errores de PostgreSQL
func (r *VenueRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("venue not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "venues_slug_key") {
				return fmt.Errorf("venue slug already exists")
			}
			if strings.Contains(pgErr.ConstraintName, "venues_public_uuid_key") {
				return fmt.Errorf("venue public_uuid already exists")
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced record not found: %w", err)
		}
	}

	return fmt.Errorf("%s: %w", context, err)
}

// ============================================================================
// CRUD BÁSICO
// ============================================================================

// Create inserta un nuevo venue
func (r *VenueRepository) Create(ctx context.Context, venue *entities.Venue) error {
	facilitiesJSON, err := json.Marshal(venue.Facilities)
	if err != nil {
		return fmt.Errorf("failed to marshal facilities: %w", err)
	}

	accessibilityJSON, err := json.Marshal(venue.AccessibilityFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal accessibility features: %w", err)
	}

	imagesJSON, err := json.Marshal(venue.Images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	query := `
		INSERT INTO ticketing.venues (
			public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10,
			$11, $12,
			$13, $14, $15,
			$16, $17,
			$18, $19,
			$20, $21,
			NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		venue.Name,
		venue.Slug,
		venue.Description,
		venue.VenueType,
		venue.AddressLine1,
		venue.AddressLine2,
		venue.City,
		venue.State,
		venue.PostalCode,
		venue.Country,
		venue.Latitude,
		venue.Longitude,
		venue.Capacity,
		venue.SeatingCapacity,
		venue.StandingCapacity,
		facilitiesJSON,
		accessibilityJSON,
		venue.ContactEmail,
		venue.ContactPhone,
		imagesJSON,
		venue.IsActive,
	).Scan(&venue.ID, &venue.PublicID, &venue.CreatedAt, &venue.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create venue")
	}

	return nil
}

// FindByID obtiene venue por ID numérico
func (r *VenueRepository) FindByID(ctx context.Context, id int64) (*entities.Venue, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at
		FROM ticketing.venues
		WHERE id = $1
	`

	var venue entities.Venue
	var facilitiesJSON, accessibilityJSON, imagesJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&venue.ID, &venue.PublicID,
		&venue.Name, &venue.Slug, &venue.Description, &venue.VenueType,
		&venue.AddressLine1, &venue.AddressLine2, &venue.City, &venue.State, &venue.PostalCode, &venue.Country,
		&venue.Latitude, &venue.Longitude,
		&venue.Capacity, &venue.SeatingCapacity, &venue.StandingCapacity,
		&facilitiesJSON, &accessibilityJSON,
		&venue.ContactEmail, &venue.ContactPhone,
		&imagesJSON,
		&venue.IsActive,
		&venue.CreatedAt, &venue.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get venue by ID")
	}

	// Deserializar JSON
	if len(facilitiesJSON) > 0 {
		json.Unmarshal(facilitiesJSON, &venue.Facilities)
	}
	if len(accessibilityJSON) > 0 {
		json.Unmarshal(accessibilityJSON, &venue.AccessibilityFeatures)
	}
	if len(imagesJSON) > 0 {
		json.Unmarshal(imagesJSON, &venue.Images)
	}

	return &venue, nil
}

// FindByPublicID obtiene venue por UUID
func (r *VenueRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Venue, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at
		FROM ticketing.venues
		WHERE public_uuid = $1
	`

	var venue entities.Venue
	var facilitiesJSON, accessibilityJSON, imagesJSON []byte

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&venue.ID, &venue.PublicID,
		&venue.Name, &venue.Slug, &venue.Description, &venue.VenueType,
		&venue.AddressLine1, &venue.AddressLine2, &venue.City, &venue.State, &venue.PostalCode, &venue.Country,
		&venue.Latitude, &venue.Longitude,
		&venue.Capacity, &venue.SeatingCapacity, &venue.StandingCapacity,
		&facilitiesJSON, &accessibilityJSON,
		&venue.ContactEmail, &venue.ContactPhone,
		&imagesJSON,
		&venue.IsActive,
		&venue.CreatedAt, &venue.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get venue by public ID")
	}

	// Deserializar JSON
	if len(facilitiesJSON) > 0 {
		json.Unmarshal(facilitiesJSON, &venue.Facilities)
	}
	if len(accessibilityJSON) > 0 {
		json.Unmarshal(accessibilityJSON, &venue.AccessibilityFeatures)
	}
	if len(imagesJSON) > 0 {
		json.Unmarshal(imagesJSON, &venue.Images)
	}

	return &venue, nil
}

// FindBySlug obtiene venue por slug
func (r *VenueRepository) FindBySlug(ctx context.Context, slug string) (*entities.Venue, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at
		FROM ticketing.venues
		WHERE slug = $1
	`

	var venue entities.Venue
	var facilitiesJSON, accessibilityJSON, imagesJSON []byte

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&venue.ID, &venue.PublicID,
		&venue.Name, &venue.Slug, &venue.Description, &venue.VenueType,
		&venue.AddressLine1, &venue.AddressLine2, &venue.City, &venue.State, &venue.PostalCode, &venue.Country,
		&venue.Latitude, &venue.Longitude,
		&venue.Capacity, &venue.SeatingCapacity, &venue.StandingCapacity,
		&facilitiesJSON, &accessibilityJSON,
		&venue.ContactEmail, &venue.ContactPhone,
		&imagesJSON,
		&venue.IsActive,
		&venue.CreatedAt, &venue.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get venue by slug")
	}

	// Deserializar JSON
	if len(facilitiesJSON) > 0 {
		json.Unmarshal(facilitiesJSON, &venue.Facilities)
	}
	if len(accessibilityJSON) > 0 {
		json.Unmarshal(accessibilityJSON, &venue.AccessibilityFeatures)
	}
	if len(imagesJSON) > 0 {
		json.Unmarshal(imagesJSON, &venue.Images)
	}

	return &venue, nil
}

// Update actualiza un venue existente
func (r *VenueRepository) Update(ctx context.Context, venue *entities.Venue) error {
	exists, err := r.Exists(ctx, venue.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("venue not found")
	}

	facilitiesJSON, err := json.Marshal(venue.Facilities)
	if err != nil {
		return fmt.Errorf("failed to marshal facilities: %w", err)
	}

	accessibilityJSON, err := json.Marshal(venue.AccessibilityFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal accessibility features: %w", err)
	}

	imagesJSON, err := json.Marshal(venue.Images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	query := `
		UPDATE ticketing.venues SET
			name = $1,
			slug = $2,
			description = $3,
			venue_type = $4,
			address_line1 = $5,
			address_line2 = $6,
			city = $7,
			state = $8,
			postal_code = $9,
			country = $10,
			latitude = $11,
			longitude = $12,
			capacity = $13,
			seating_capacity = $14,
			standing_capacity = $15,
			facilities = $16,
			accessibility_features = $17,
			contact_email = $18,
			contact_phone = $19,
			images = $20,
			is_active = $21,
			updated_at = NOW()
		WHERE id = $22
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		venue.Name,
		venue.Slug,
		venue.Description,
		venue.VenueType,
		venue.AddressLine1,
		venue.AddressLine2,
		venue.City,
		venue.State,
		venue.PostalCode,
		venue.Country,
		venue.Latitude,
		venue.Longitude,
		venue.Capacity,
		venue.SeatingCapacity,
		venue.StandingCapacity,
		facilitiesJSON,
		accessibilityJSON,
		venue.ContactEmail,
		venue.ContactPhone,
		imagesJSON,
		venue.IsActive,
		venue.ID,
	).Scan(&venue.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update venue")
	}

	return nil
}

// Delete elimina permanentemente un venue
func (r *VenueRepository) Delete(ctx context.Context, id int64) error {
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM ticketing.venues WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete venue")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// SoftDelete desactiva un venue
func (r *VenueRepository) SoftDelete(ctx context.Context, publicID string) error {
	query := `UPDATE ticketing.venues SET is_active = false, updated_at = NOW() WHERE public_uuid = $1`
	cmdTag, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return r.handleError(err, "failed to soft delete venue")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// Exists verifica existencia por ID
func (r *VenueRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticketing.venues WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check existence")
	}
	return exists, nil
}

// ============================================================================
// BÚSQUEDAS
// ============================================================================

// List lista venues con filtros
func (r *VenueRepository) List(ctx context.Context, filter dto.VenueFilter, pagination dto.Pagination) ([]*entities.Venue, int64, error) {
	where := []string{"1=1"}
	args := pgx.NamedArgs{}
	argPos := 1

	if filter.Name != "" {
		where = append(where, fmt.Sprintf("name ILIKE @name_%d", argPos))
		args[fmt.Sprintf("name_%d", argPos)] = "%" + filter.Name + "%"
		argPos++
	}
	if filter.City != "" {
		where = append(where, fmt.Sprintf("city ILIKE @city_%d", argPos))
		args[fmt.Sprintf("city_%d", argPos)] = "%" + filter.City + "%"
		argPos++
	}
	if filter.State != "" {
		where = append(where, fmt.Sprintf("state ILIKE @state_%d", argPos))
		args[fmt.Sprintf("state_%d", argPos)] = "%" + filter.State + "%"
		argPos++
	}
	if filter.Country != "" {
		where = append(where, fmt.Sprintf("country = @country_%d", argPos))
		args[fmt.Sprintf("country_%d", argPos)] = filter.Country
		argPos++
	}
	if filter.VenueType != "" {
		where = append(where, fmt.Sprintf("venue_type = @type_%d", argPos))
		args[fmt.Sprintf("type_%d", argPos)] = filter.VenueType
		argPos++
	}
	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active = @active_%d", argPos))
		args[fmt.Sprintf("active_%d", argPos)] = *filter.IsActive
		argPos++
	}
	if filter.MinCapacity != nil {
		where = append(where, fmt.Sprintf("capacity >= @min_cap_%d", argPos))
		args[fmt.Sprintf("min_cap_%d", argPos)] = *filter.MinCapacity
		argPos++
	}
	if filter.MaxCapacity != nil {
		where = append(where, fmt.Sprintf("capacity <= @max_cap_%d", argPos))
		args[fmt.Sprintf("max_cap_%d", argPos)] = *filter.MaxCapacity
		argPos++
	}
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		where = append(where, fmt.Sprintf("(name ILIKE @search_%d OR description ILIKE @search_%d OR city ILIKE @search_%d)", argPos, argPos, argPos))
		args[fmt.Sprintf("search_%d", argPos)] = searchTerm
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticketing.venues WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count venues")
	}

	// Obtener datos
	query := fmt.Sprintf(`
		SELECT 
			id, public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at
		FROM ticketing.venues
		WHERE %s
		ORDER BY name
		LIMIT @limit OFFSET @offset
	`, whereClause)

	args["limit"] = pagination.PageSize
	args["offset"] = (pagination.Page - 1) * pagination.PageSize

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to list venues")
	}
	defer rows.Close()

	var venues []*entities.Venue
	for rows.Next() {
		var venue entities.Venue
		var facilitiesJSON, accessibilityJSON, imagesJSON []byte

		err = rows.Scan(
			&venue.ID, &venue.PublicID,
			&venue.Name, &venue.Slug, &venue.Description, &venue.VenueType,
			&venue.AddressLine1, &venue.AddressLine2, &venue.City, &venue.State, &venue.PostalCode, &venue.Country,
			&venue.Latitude, &venue.Longitude,
			&venue.Capacity, &venue.SeatingCapacity, &venue.StandingCapacity,
			&facilitiesJSON, &accessibilityJSON,
			&venue.ContactEmail, &venue.ContactPhone,
			&imagesJSON,
			&venue.IsActive,
			&venue.CreatedAt, &venue.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan venue")
		}

		// Deserializar JSON
		if len(facilitiesJSON) > 0 {
			json.Unmarshal(facilitiesJSON, &venue.Facilities)
		}
		if len(accessibilityJSON) > 0 {
			json.Unmarshal(accessibilityJSON, &venue.AccessibilityFeatures)
		}
		if len(imagesJSON) > 0 {
			json.Unmarshal(imagesJSON, &venue.Images)
		}

		venues = append(venues, &venue)
	}

	return venues, total, nil
}

// ListByCountry lista venues por país
func (r *VenueRepository) ListByCountry(ctx context.Context, countryCode string, pagination dto.Pagination) ([]*entities.Venue, int64, error) {
	filter := dto.VenueFilter{
		Country: countryCode,
	}
	return r.List(ctx, filter, pagination)
}

// ListByCity lista venues por ciudad
func (r *VenueRepository) ListByCity(ctx context.Context, country, city string, pagination dto.Pagination) ([]*entities.Venue, int64, error) {
	filter := dto.VenueFilter{
		Country: country,
		City:    city,
	}
	return r.List(ctx, filter, pagination)
}

// ListByType lista venues por tipo
func (r *VenueRepository) ListByType(ctx context.Context, venueType string, pagination dto.Pagination) ([]*entities.Venue, int64, error) {
	filter := dto.VenueFilter{
		VenueType: venueType,
	}
	return r.List(ctx, filter, pagination)
}

// Search busca venues por término
func (r *VenueRepository) Search(ctx context.Context, term string, filter dto.VenueFilter, pagination dto.Pagination) ([]*entities.Venue, int64, error) {
	filter.Search = term
	return r.List(ctx, filter, pagination)
}

// FindNearby encuentra venues cercanos a una ubicación
func (r *VenueRepository) FindNearby(ctx context.Context, latitude, longitude float64, radiusKm float64, limit int) ([]*entities.Venue, error) {
	// Aproximación simple usando el teorema de Pitágoras
	// En producción usarías PostGIS
	query := `
		SELECT 
			id, public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at,
			SQRT(POW(($1 - latitude), 2) + POW(($2 - longitude), 2)) * 111 as distance_km
		FROM ticketing.venues
		WHERE is_active = true
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
		HAVING distance_km <= $3
		ORDER BY distance_km
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, latitude, longitude, radiusKm, limit)
	if err != nil {
		return nil, r.handleError(err, "failed to find nearby venues")
	}
	defer rows.Close()

	var venues []*entities.Venue
	for rows.Next() {
		var venue entities.Venue
		var facilitiesJSON, accessibilityJSON, imagesJSON []byte
		var distance float64

		err = rows.Scan(
			&venue.ID, &venue.PublicID,
			&venue.Name, &venue.Slug, &venue.Description, &venue.VenueType,
			&venue.AddressLine1, &venue.AddressLine2, &venue.City, &venue.State, &venue.PostalCode, &venue.Country,
			&venue.Latitude, &venue.Longitude,
			&venue.Capacity, &venue.SeatingCapacity, &venue.StandingCapacity,
			&facilitiesJSON, &accessibilityJSON,
			&venue.ContactEmail, &venue.ContactPhone,
			&imagesJSON,
			&venue.IsActive,
			&venue.CreatedAt, &venue.UpdatedAt,
			&distance,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan nearby venue")
		}

		// Deserializar JSON
		if len(facilitiesJSON) > 0 {
			json.Unmarshal(facilitiesJSON, &venue.Facilities)
		}
		if len(accessibilityJSON) > 0 {
			json.Unmarshal(accessibilityJSON, &venue.AccessibilityFeatures)
		}
		if len(imagesJSON) > 0 {
			json.Unmarshal(imagesJSON, &venue.Images)
		}

		venues = append(venues, &venue)
	}

	return venues, nil
}

// GetVenuesInRadius obtiene venues en un radio
func (r *VenueRepository) GetVenuesInRadius(ctx context.Context, centerLat, centerLon float64, radiusKm float64, venueType *string) ([]*entities.Venue, error) {
	// Usar FindNearby con radio
	return r.FindNearby(ctx, centerLat, centerLon, radiusKm, 100)
}

// GetVenuesInBounds obtiene venues dentro de un rectángulo geográfico
func (r *VenueRepository) GetVenuesInBounds(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]*entities.Venue, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, venue_type,
			address_line1, address_line2, city, state, postal_code, country,
			latitude, longitude,
			capacity, seating_capacity, standing_capacity,
			facilities, accessibility_features,
			contact_email, contact_phone,
			images, is_active,
			created_at, updated_at
		FROM ticketing.venues
		WHERE latitude BETWEEN $1 AND $3
			AND longitude BETWEEN $2 AND $4
			AND is_active = true
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, minLat, minLon, maxLat, maxLon)
	if err != nil {
		return nil, r.handleError(err, "failed to get venues in bounds")
	}
	defer rows.Close()

	var venues []*entities.Venue
	for rows.Next() {
		var venue entities.Venue
		var facilitiesJSON, accessibilityJSON, imagesJSON []byte

		err = rows.Scan(
			&venue.ID, &venue.PublicID,
			&venue.Name, &venue.Slug, &venue.Description, &venue.VenueType,
			&venue.AddressLine1, &venue.AddressLine2, &venue.City, &venue.State, &venue.PostalCode, &venue.Country,
			&venue.Latitude, &venue.Longitude,
			&venue.Capacity, &venue.SeatingCapacity, &venue.StandingCapacity,
			&facilitiesJSON, &accessibilityJSON,
			&venue.ContactEmail, &venue.ContactPhone,
			&imagesJSON,
			&venue.IsActive,
			&venue.CreatedAt, &venue.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan venue")
		}

		// Deserializar JSON
		if len(facilitiesJSON) > 0 {
			json.Unmarshal(facilitiesJSON, &venue.Facilities)
		}
		if len(accessibilityJSON) > 0 {
			json.Unmarshal(accessibilityJSON, &venue.AccessibilityFeatures)
		}
		if len(imagesJSON) > 0 {
			json.Unmarshal(imagesJSON, &venue.Images)
		}

		venues = append(venues, &venue)
	}

	return venues, nil
}

// ============================================================================
// OPERACIONES ESPECÍFICAS
// ============================================================================

// UpdateCapacity actualiza la capacidad del venue
func (r *VenueRepository) UpdateCapacity(ctx context.Context, venueID int64, capacity, seating, standing *int) error {
	query := `
		UPDATE ticketing.venues 
		SET capacity = COALESCE($1, capacity),
			seating_capacity = COALESCE($2, seating_capacity),
			standing_capacity = COALESCE($3, standing_capacity),
			updated_at = NOW()
		WHERE id = $4
	`
	cmdTag, err := r.db.Exec(ctx, query, capacity, seating, standing, venueID)
	if err != nil {
		return r.handleError(err, "failed to update capacity")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// UpdateLocation actualiza la ubicación del venue
func (r *VenueRepository) UpdateLocation(ctx context.Context, venueID int64, address, city, state, postalCode, country string) error {
	query := `
		UPDATE ticketing.venues 
		SET address_line1 = $1,
			city = $2,
			state = $3,
			postal_code = $4,
			country = $5,
			updated_at = NOW()
		WHERE id = $6
	`
	cmdTag, err := r.db.Exec(ctx, query, address, city, state, postalCode, country, venueID)
	if err != nil {
		return r.handleError(err, "failed to update location")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// UpdateCoordinates actualiza las coordenadas del venue
func (r *VenueRepository) UpdateCoordinates(ctx context.Context, venueID int64, latitude, longitude float64) error {
	query := `
		UPDATE ticketing.venues 
		SET latitude = $1,
			longitude = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	cmdTag, err := r.db.Exec(ctx, query, latitude, longitude, venueID)
	if err != nil {
		return r.handleError(err, "failed to update coordinates")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// UpdateContactInfo actualiza información de contacto
func (r *VenueRepository) UpdateContactInfo(ctx context.Context, venueID int64, email, phone string) error {
	query := `
		UPDATE ticketing.venues 
		SET contact_email = $1,
			contact_phone = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	cmdTag, err := r.db.Exec(ctx, query, email, phone, venueID)
	if err != nil {
		return r.handleError(err, "failed to update contact info")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// UpdateFacilities actualiza las instalaciones del venue
func (r *VenueRepository) UpdateFacilities(ctx context.Context, venueID int64, facilities []string) error {
	jsonData, err := json.Marshal(facilities)
	if err != nil {
		return fmt.Errorf("failed to marshal facilities: %w", err)
	}
	cmdTag, err := r.db.Exec(ctx, `UPDATE ticketing.venues SET facilities = $1, updated_at = NOW() WHERE id = $2`, jsonData, venueID)
	if err != nil {
		return r.handleError(err, "failed to update facilities")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// UpdateAccessibility actualiza las características de accesibilidad
func (r *VenueRepository) UpdateAccessibility(ctx context.Context, venueID int64, features []string) error {
	jsonData, err := json.Marshal(features)
	if err != nil {
		return fmt.Errorf("failed to marshal accessibility features: %w", err)
	}
	cmdTag, err := r.db.Exec(ctx, `UPDATE ticketing.venues SET accessibility_features = $1, updated_at = NOW() WHERE id = $2`, jsonData, venueID)
	if err != nil {
		return r.handleError(err, "failed to update accessibility features")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}
	return nil
}

// ============================================================================
// CONSULTAS GEOGRÁFICAS
// ============================================================================

// GetDistance obtiene la distancia entre el venue y un punto
func (r *VenueRepository) GetDistance(ctx context.Context, venueID int64, latitude, longitude float64) (float64, error) {
	var distance float64
	query := `
		SELECT SQRT(POW(($1 - latitude), 2) + POW(($2 - longitude), 2)) * 111
		FROM ticketing.venues
		WHERE id = $3 AND latitude IS NOT NULL AND longitude IS NOT NULL
	`
	err := r.db.QueryRow(ctx, query, latitude, longitude, venueID).Scan(&distance)
	if err != nil {
		return 0, r.handleError(err, "failed to calculate distance")
	}
	return distance, nil
}

// ============================================================================
// ESTADÍSTICAS
// ============================================================================

// GetStats obtiene estadísticas de un venue
func (r *VenueRepository) GetStats(ctx context.Context, venueID int64) (*dto.VenueStatsResponse, error) {
	// Contar eventos en este venue
	var eventCount int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ticketing.events WHERE venue_id = $1`, venueID).Scan(&eventCount)
	if err != nil {
		return nil, r.handleError(err, "failed to count events")
	}

	// Calcular capacidad promedio de los eventos
	var avgCapacity float64
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(max_attendees), 0)
		FROM ticketing.events
		WHERE venue_id = $1 AND max_attendees IS NOT NULL
	`, venueID).Scan(&avgCapacity)
	if err != nil {
		return nil, r.handleError(err, "failed to calculate avg capacity")
	}

	stats := &dto.VenueStatsResponse{
		TotalEvents:      eventCount,
		AvgCapacity:      avgCapacity,
		ActiveVenues:     1,
		TotalVenues:      1,
		TotalTicketsSold: 0,
	}

	return stats, nil
}

// CountEvents cuenta eventos en un venue
func (r *VenueRepository) CountEvents(ctx context.Context, venueID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ticketing.events WHERE venue_id = $1`, venueID).Scan(&count)
	if err != nil {
		return 0, r.handleError(err, "failed to count events")
	}
	return count, nil
}

// GetUpcomingEvents está comentado porque dto.VenueEvent no existe
// Si se necesita en el futuro, crear el DTO correspondiente
/*
func (r *VenueRepository) GetUpcomingEvents(ctx context.Context, venueID int64, limit int) ([]*dto.VenueEvent, error) {
	query := `
		SELECT
			id, public_uuid, name, slug, start_date, end_date
		FROM ticketing.events
		WHERE venue_id = $1 AND start_date > NOW()
		ORDER BY start_date
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, venueID, limit)
	if err != nil {
		return nil, r.handleError(err, "failed to get upcoming events")
	}
	defer rows.Close()

	var events []*dto.VenueEvent
	for rows.Next() {
		var event dto.VenueEvent
		err = rows.Scan(
			&event.ID,
			&event.PublicID,
			&event.Name,
			&event.Slug,
			&event.StartDate,
			&event.EndDate,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan event")
		}
		events = append(events, &event)
	}

	return events, nil
}
*/

// GetCapacityUtilization obtiene el porcentaje de utilización de capacidad
func (r *VenueRepository) GetCapacityUtilization(ctx context.Context, venueID int64) (float64, error) {
	var utilization float64
	query := `
		SELECT 
			COALESCE(
				(SELECT SUM(max_attendees) FROM ticketing.events WHERE venue_id = $1 AND start_date > NOW()) * 1.0 /
				(SELECT capacity FROM ticketing.venues WHERE id = $1),
			0)
	`
	err := r.db.QueryRow(ctx, query, venueID).Scan(&utilization)
	if err != nil {
		return 0, r.handleError(err, "failed to calculate capacity utilization")
	}
	return utilization, nil
}

// GetPopularVenues está comentado porque dto.PopularVenue no existe
// Si se necesita en el futuro, crear el DTO correspondiente
/*
func (r *VenueRepository) GetPopularVenues(ctx context.Context, limit int) ([]*dto.PopularVenue, error) {
	query := `
		SELECT
			v.id, v.name, v.slug, v.city, v.country,
			COUNT(DISTINCT e.id) as event_count,
			COALESCE(SUM(tt.sold_quantity), 0) as tickets_sold
		FROM ticketing.venues v
		LEFT JOIN ticketing.events e ON v.id = e.venue_id
		LEFT JOIN ticketing.ticket_types tt ON e.id = tt.event_id
		WHERE v.is_active = true
		GROUP BY v.id, v.name, v.slug, v.city, v.country
		ORDER BY tickets_sold DESC, event_count DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, r.handleError(err, "failed to get popular venues")
	}
	defer rows.Close()

	var venues []*dto.PopularVenue
	for rows.Next() {
		var venue dto.PopularVenue
		err = rows.Scan(
			&venue.ID,
			&venue.Name,
			&venue.Slug,
			&venue.City,
			&venue.Country,
			&venue.EventCount,
			&venue.TicketsSold,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan popular venue")
		}
		venues = append(venues, &venue)
	}

	return venues, nil
}
*/

// ============================================================================
// OPERACIONES CON IMÁGENES
// ============================================================================

// AddImage agrega una imagen al venue
func (r *VenueRepository) AddImage(ctx context.Context, venueID int64, image entities.VenueImage) error {
	// Obtener imágenes actuales
	var imagesJSON []byte
	err := r.db.QueryRow(ctx, `SELECT images FROM ticketing.venues WHERE id = $1`, venueID).Scan(&imagesJSON)
	if err != nil {
		return r.handleError(err, "failed to get images")
	}

	var images []entities.VenueImage
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &images); err != nil {
			return fmt.Errorf("failed to unmarshal images: %w", err)
		}
	}

	// Verificar si ya existe
	for i, img := range images {
		if img.URL == image.URL {
			return nil
		}
		if image.IsPrimary {
			images[i].IsPrimary = false
		}
	}

	// Configurar metadatos
	image.UploadedAt = time.Now()
	images = append(images, image)

	// Guardar
	newImagesJSON, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`UPDATE ticketing.venues SET images = $1, updated_at = NOW() WHERE id = $2`,
		newImagesJSON, venueID)
	return r.handleError(err, "failed to add image")
}

// RemoveImage elimina una imagen del venue
func (r *VenueRepository) RemoveImage(ctx context.Context, venueID int64, imageURL string) error {
	// Obtener imágenes actuales
	var imagesJSON []byte
	err := r.db.QueryRow(ctx, `SELECT images FROM ticketing.venues WHERE id = $1`, venueID).Scan(&imagesJSON)
	if err != nil {
		return r.handleError(err, "failed to get images")
	}

	var images []entities.VenueImage
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &images); err != nil {
			return fmt.Errorf("failed to unmarshal images: %w", err)
		}
	}

	// Filtrar la imagen a eliminar
	newImages := []entities.VenueImage{}
	for _, img := range images {
		if img.URL != imageURL {
			newImages = append(newImages, img)
		}
	}

	// Si eliminamos la imagen principal y quedan imágenes, establecer la primera como principal
	if len(newImages) > 0 {
		hasPrimary := false
		for _, img := range newImages {
			if img.IsPrimary {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary {
			newImages[0].IsPrimary = true
		}
	}

	// Guardar
	newImagesJSON, err := json.Marshal(newImages)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`UPDATE ticketing.venues SET images = $1, updated_at = NOW() WHERE id = $2`,
		newImagesJSON, venueID)
	return r.handleError(err, "failed to remove image")
}

// SetMainImage establece una imagen como principal
func (r *VenueRepository) SetMainImage(ctx context.Context, venueID int64, imageURL string) error {
	// Obtener imágenes actuales
	var imagesJSON []byte
	err := r.db.QueryRow(ctx, `SELECT images FROM ticketing.venues WHERE id = $1`, venueID).Scan(&imagesJSON)
	if err != nil {
		return r.handleError(err, "failed to get images")
	}

	var images []entities.VenueImage
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &images); err != nil {
			return fmt.Errorf("failed to unmarshal images: %w", err)
		}
	}

	// Actualizar flags
	found := false
	for i := range images {
		if images[i].URL == imageURL {
			images[i].IsPrimary = true
			found = true
		} else {
			images[i].IsPrimary = false
		}
	}

	if !found {
		return fmt.Errorf("image not found")
	}

	// Guardar
	newImagesJSON, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`UPDATE ticketing.venues SET images = $1, updated_at = NOW() WHERE id = $2`,
		newImagesJSON, venueID)
	return r.handleError(err, "failed to set main image")
}

// GetImages obtiene todas las imágenes de un venue
func (r *VenueRepository) GetImages(ctx context.Context, venueID int64) ([]entities.VenueImage, error) {
	var imagesJSON []byte
	err := r.db.QueryRow(ctx, `SELECT images FROM ticketing.venues WHERE id = $1`, venueID).Scan(&imagesJSON)
	if err != nil {
		return nil, r.handleError(err, "failed to get images")
	}

	var images []entities.VenueImage
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &images); err != nil {
			return nil, fmt.Errorf("failed to unmarshal images: %w", err)
		}
	}

	return images, nil
}

// UpdateImageMetadata actualiza los metadatos de una imagen específica
func (r *VenueRepository) UpdateImageMetadata(ctx context.Context, venueID int64, imageURL string, caption *string, imageType *string) error {
	// Obtener imágenes actuales
	var imagesJSON []byte
	err := r.db.QueryRow(ctx, `SELECT images FROM ticketing.venues WHERE id = $1`, venueID).Scan(&imagesJSON)
	if err != nil {
		return r.handleError(err, "failed to get images")
	}

	var images []entities.VenueImage
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &images); err != nil {
			return fmt.Errorf("failed to unmarshal images: %w", err)
		}
	}

	// Actualizar metadatos
	found := false
	for i := range images {
		if images[i].URL == imageURL {
			if caption != nil {
				images[i].Caption = caption
			}
			if imageType != nil {
				images[i].Type = *imageType
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("image not found")
	}

	// Guardar
	newImagesJSON, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`UPDATE ticketing.venues SET images = $1, updated_at = NOW() WHERE id = $2`,
		newImagesJSON, venueID)
	return r.handleError(err, "failed to update image metadata")
}
