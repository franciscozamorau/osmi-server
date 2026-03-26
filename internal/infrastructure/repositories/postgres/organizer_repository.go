// internal/infrastructure/repositories/postgres/organizer_repository.go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// OrganizerRepository implementa la interfaz repository.OrganizerRepository
type OrganizerRepository struct {
	db *pgxpool.Pool
}

// NewOrganizerRepository crea una nueva instancia
func NewOrganizerRepository(db *pgxpool.Pool) *OrganizerRepository {
	return &OrganizerRepository{
		db: db,
	}
}

// handleError maneja errores de PostgreSQL
func (r *OrganizerRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("organizer not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "organizers_slug_key") {
				return fmt.Errorf("organizer slug already exists")
			}
			if strings.Contains(pgErr.ConstraintName, "organizers_public_uuid_key") {
				return fmt.Errorf("organizer public_uuid already exists")
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced record not found: %w", err)
		}
	}

	return fmt.Errorf("%s: %w", context, err)
}

// ============================================================================
// CRUD BÁSICO (TODOS ESTOS MÉTODOS ESTÁN CORRECTOS)
// ============================================================================

// Create inserta un nuevo organizador
func (r *OrganizerRepository) Create(ctx context.Context, organizer *entities.Organizer) error {
	socialLinksJSON, err := json.Marshal(organizer.SocialLinks)
	if err != nil {
		return fmt.Errorf("failed to marshal social links: %w", err)
	}

	query := `
		INSERT INTO ticketing.organizers (
			public_uuid, name, slug, description, logo_url,
			legal_name, tax_id, tax_id_type, country,
			contact_email, contact_phone,
			address_line1, address_line2, city, state, postal_code,
			is_verified, is_active, verification_status,
			total_events, total_tickets_sold, organizer_rating, rating_count,
			social_links,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18,
			0, 0, 0, 0,
			$19,
			NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		organizer.Name,
		organizer.Slug,
		organizer.Description,
		organizer.LogoURL,
		organizer.LegalName,
		organizer.TaxID,
		organizer.TaxIDType,
		organizer.Country,
		organizer.ContactEmail,
		organizer.ContactPhone,
		organizer.AddressLine1,
		organizer.AddressLine2,
		organizer.City,
		organizer.State,
		organizer.PostalCode,
		organizer.IsVerifiedField,
		organizer.IsActive,
		organizer.VerificationStatus,
		socialLinksJSON,
	).Scan(&organizer.ID, &organizer.PublicID, &organizer.CreatedAt, &organizer.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create organizer")
	}

	return nil
}

// FindByID obtiene organizador por ID numérico
func (r *OrganizerRepository) FindByID(ctx context.Context, id int64) (*entities.Organizer, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, logo_url,
			legal_name, tax_id, tax_id_type, country,
			contact_email, contact_phone,
			address_line1, address_line2, city, state, postal_code,
			is_verified, is_active, verification_status,
			total_events, total_tickets_sold, organizer_rating, rating_count,
			social_links,
			created_at, updated_at
		FROM ticketing.organizers
		WHERE id = $1
	`

	var organizer entities.Organizer
	var socialLinksJSON []byte
	var description, logoURL, legalName, taxID, taxIDType, country, addressLine1, addressLine2, city, state, postalCode *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&organizer.ID, &organizer.PublicID,
		&organizer.Name, &organizer.Slug, &description, &logoURL,
		&legalName, &taxID, &taxIDType, &country,
		&organizer.ContactEmail, &organizer.ContactPhone,
		&addressLine1, &addressLine2, &city, &state, &postalCode,
		&organizer.IsVerifiedField, &organizer.IsActive, &organizer.VerificationStatus,
		&organizer.TotalEvents, &organizer.TotalTicketsSold, &organizer.OrganizerRating, &organizer.RatingCount,
		&socialLinksJSON,
		&organizer.CreatedAt, &organizer.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get organizer by ID")
	}

	// Asignar campos NULL
	organizer.Description = description
	organizer.LogoURL = logoURL
	organizer.LegalName = legalName
	organizer.TaxID = taxID
	organizer.TaxIDType = taxIDType
	organizer.Country = country
	organizer.AddressLine1 = addressLine1
	organizer.AddressLine2 = addressLine2
	organizer.City = city
	organizer.State = state
	organizer.PostalCode = postalCode

	if len(socialLinksJSON) > 0 {
		json.Unmarshal(socialLinksJSON, &organizer.SocialLinks)
	}

	return &organizer, nil
}

// FindByPublicID obtiene organizador por UUID
func (r *OrganizerRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Organizer, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, logo_url,
			legal_name, tax_id, tax_id_type, country,
			contact_email, contact_phone,
			address_line1, address_line2, city, state, postal_code,
			is_verified, is_active, verification_status,
			total_events, total_tickets_sold, organizer_rating, rating_count,
			social_links,
			created_at, updated_at
		FROM ticketing.organizers
		WHERE public_uuid = $1
	`

	var organizer entities.Organizer
	var socialLinksJSON []byte
	var description, logoURL, legalName, taxID, taxIDType, country, addressLine1, addressLine2, city, state, postalCode *string

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&organizer.ID, &organizer.PublicID,
		&organizer.Name, &organizer.Slug, &description, &logoURL,
		&legalName, &taxID, &taxIDType, &country,
		&organizer.ContactEmail, &organizer.ContactPhone,
		&addressLine1, &addressLine2, &city, &state, &postalCode,
		&organizer.IsVerifiedField, &organizer.IsActive, &organizer.VerificationStatus,
		&organizer.TotalEvents, &organizer.TotalTicketsSold, &organizer.OrganizerRating, &organizer.RatingCount,
		&socialLinksJSON,
		&organizer.CreatedAt, &organizer.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get organizer by public ID")
	}

	// Asignar campos NULL
	organizer.Description = description
	organizer.LogoURL = logoURL
	organizer.LegalName = legalName
	organizer.TaxID = taxID
	organizer.TaxIDType = taxIDType
	organizer.Country = country
	organizer.AddressLine1 = addressLine1
	organizer.AddressLine2 = addressLine2
	organizer.City = city
	organizer.State = state
	organizer.PostalCode = postalCode

	if len(socialLinksJSON) > 0 {
		json.Unmarshal(socialLinksJSON, &organizer.SocialLinks)
	}

	return &organizer, nil
}

// FindBySlug obtiene organizador por slug
func (r *OrganizerRepository) FindBySlug(ctx context.Context, slug string) (*entities.Organizer, error) {
	query := `
		SELECT 
			id, public_uuid, name, slug, description, logo_url,
			legal_name, tax_id, tax_id_type, country,
			contact_email, contact_phone,
			address_line1, address_line2, city, state, postal_code,
			is_verified, is_active, verification_status,
			total_events, total_tickets_sold, organizer_rating, rating_count,
			social_links,
			created_at, updated_at
		FROM ticketing.organizers
		WHERE slug = $1
	`

	var organizer entities.Organizer
	var socialLinksJSON []byte
	var description, logoURL, legalName, taxID, taxIDType, country, addressLine1, addressLine2, city, state, postalCode *string

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&organizer.ID, &organizer.PublicID,
		&organizer.Name, &organizer.Slug, &description, &logoURL,
		&legalName, &taxID, &taxIDType, &country,
		&organizer.ContactEmail, &organizer.ContactPhone,
		&addressLine1, &addressLine2, &city, &state, &postalCode,
		&organizer.IsVerifiedField, &organizer.IsActive, &organizer.VerificationStatus,
		&organizer.TotalEvents, &organizer.TotalTicketsSold, &organizer.OrganizerRating, &organizer.RatingCount,
		&socialLinksJSON,
		&organizer.CreatedAt, &organizer.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get organizer by slug")
	}

	// Asignar campos NULL
	organizer.Description = description
	organizer.LogoURL = logoURL
	organizer.LegalName = legalName
	organizer.TaxID = taxID
	organizer.TaxIDType = taxIDType
	organizer.Country = country
	organizer.AddressLine1 = addressLine1
	organizer.AddressLine2 = addressLine2
	organizer.City = city
	organizer.State = state
	organizer.PostalCode = postalCode

	if len(socialLinksJSON) > 0 {
		json.Unmarshal(socialLinksJSON, &organizer.SocialLinks)
	}

	return &organizer, nil
}

// Update actualiza un organizador existente
func (r *OrganizerRepository) Update(ctx context.Context, organizer *entities.Organizer) error {
	socialLinksJSON, err := json.Marshal(organizer.SocialLinks)
	if err != nil {
		return fmt.Errorf("failed to marshal social links: %w", err)
	}

	query := `
		UPDATE ticketing.organizers SET
			name = $1,
			slug = $2,
			description = $3,
			logo_url = $4,
			legal_name = $5,
			tax_id = $6,
			tax_id_type = $7,
			country = $8,
			contact_email = $9,
			contact_phone = $10,
			address_line1 = $11,
			address_line2 = $12,
			city = $13,
			state = $14,
			postal_code = $15,
			is_verified = $16,
			is_active = $17,
			verification_status = $18,
			social_links = $19,
			updated_at = NOW()
		WHERE id = $20
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		organizer.Name,
		organizer.Slug,
		organizer.Description,
		organizer.LogoURL,
		organizer.LegalName,
		organizer.TaxID,
		organizer.TaxIDType,
		organizer.Country,
		organizer.ContactEmail,
		organizer.ContactPhone,
		organizer.AddressLine1,
		organizer.AddressLine2,
		organizer.City,
		organizer.State,
		organizer.PostalCode,
		organizer.IsVerifiedField,
		organizer.IsActive,
		organizer.VerificationStatus,
		socialLinksJSON,
		organizer.ID,
	).Scan(&organizer.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update organizer")
	}

	return nil
}

// Delete elimina permanentemente un organizador
func (r *OrganizerRepository) Delete(ctx context.Context, id int64) error {
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM ticketing.organizers WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete organizer")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// SoftDelete desactiva un organizador
func (r *OrganizerRepository) SoftDelete(ctx context.Context, publicID string) error {
	query := `UPDATE ticketing.organizers SET is_active = false, updated_at = NOW() WHERE public_uuid = $1`
	cmdTag, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return r.handleError(err, "failed to soft delete organizer")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// Exists verifica existencia por ID
func (r *OrganizerRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticketing.organizers WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check existence")
	}
	return exists, nil
}

// ============================================================================
// BÚSQUEDAS (TODOS ESTOS MÉTODOS ESTÁN CORRECTOS)
// ============================================================================

// List lista organizadores con filtros
func (r *OrganizerRepository) List(ctx context.Context, filter dto.OrganizerFilter, pagination dto.Pagination) ([]*entities.Organizer, int64, error) {
	where := []string{"1=1"}
	args := pgx.NamedArgs{}
	argPos := 1

	if filter.Name != "" {
		where = append(where, fmt.Sprintf("name ILIKE @name_%d", argPos))
		args[fmt.Sprintf("name_%d", argPos)] = "%" + filter.Name + "%"
		argPos++
	}
	if filter.IsVerified != nil {
		where = append(where, fmt.Sprintf("is_verified = @verified_%d", argPos))
		args[fmt.Sprintf("verified_%d", argPos)] = *filter.IsVerified
		argPos++
	}
	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active = @active_%d", argPos))
		args[fmt.Sprintf("active_%d", argPos)] = *filter.IsActive
		argPos++
	}
	if filter.VerificationStatus != "" {
		where = append(where, fmt.Sprintf("verification_status = @status_%d", argPos))
		args[fmt.Sprintf("status_%d", argPos)] = filter.VerificationStatus
		argPos++
	}
	if filter.Country != "" {
		where = append(where, fmt.Sprintf("country = @country_%d", argPos))
		args[fmt.Sprintf("country_%d", argPos)] = filter.Country
		argPos++
	}
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		where = append(where, fmt.Sprintf("(name ILIKE @search_%d OR description ILIKE @search_%d)", argPos, argPos))
		args[fmt.Sprintf("search_%d", argPos)] = searchTerm
		argPos++
	}
	if filter.DateFrom != "" {
		where = append(where, fmt.Sprintf("created_at >= @from_%d", argPos))
		args[fmt.Sprintf("from_%d", argPos)] = filter.DateFrom
		argPos++
	}
	if filter.DateTo != "" {
		where = append(where, fmt.Sprintf("created_at <= @to_%d", argPos))
		args[fmt.Sprintf("to_%d", argPos)] = filter.DateTo
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticketing.organizers WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count organizers")
	}

	// Obtener datos
	query := fmt.Sprintf(`
		SELECT 
			id, public_uuid, name, slug, description, logo_url,
			legal_name, tax_id, tax_id_type, country,
			contact_email, contact_phone,
			address_line1, address_line2, city, state, postal_code,
			is_verified, is_active, verification_status,
			total_events, total_tickets_sold, organizer_rating, rating_count,
			social_links,
			created_at, updated_at
		FROM ticketing.organizers
		WHERE %s
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`, whereClause)

	args["limit"] = pagination.PageSize
	args["offset"] = (pagination.Page - 1) * pagination.PageSize

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to list organizers")
	}
	defer rows.Close()

	var organizers []*entities.Organizer
	for rows.Next() {
		var org entities.Organizer
		var socialLinksJSON []byte
		var description, logoURL, legalName, taxID, taxIDType, country, addressLine1, addressLine2, city, state, postalCode *string

		err = rows.Scan(
			&org.ID, &org.PublicID,
			&org.Name, &org.Slug, &description, &logoURL,
			&legalName, &taxID, &taxIDType, &country,
			&org.ContactEmail, &org.ContactPhone,
			&addressLine1, &addressLine2, &city, &state, &postalCode,
			&org.IsVerifiedField, &org.IsActive, &org.VerificationStatus,
			&org.TotalEvents, &org.TotalTicketsSold, &org.OrganizerRating, &org.RatingCount,
			&socialLinksJSON,
			&org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan organizer")
		}

		// Asignar campos NULL
		org.Description = description
		org.LogoURL = logoURL
		org.LegalName = legalName
		org.TaxID = taxID
		org.TaxIDType = taxIDType
		org.Country = country
		org.AddressLine1 = addressLine1
		org.AddressLine2 = addressLine2
		org.City = city
		org.State = state
		org.PostalCode = postalCode

		if len(socialLinksJSON) > 0 {
			json.Unmarshal(socialLinksJSON, &org.SocialLinks)
		}
		organizers = append(organizers, &org)
	}

	return organizers, total, nil
}

// ListVerified lista organizadores verificados
func (r *OrganizerRepository) ListVerified(ctx context.Context, limit int) ([]*entities.Organizer, error) {
	verified := true
	filter := dto.OrganizerFilter{
		IsVerified: &verified,
	}
	pagination := dto.Pagination{
		Page:     1,
		PageSize: limit,
	}
	organizers, _, err := r.List(ctx, filter, pagination)
	return organizers, err
}

// ListActive lista organizadores activos
func (r *OrganizerRepository) ListActive(ctx context.Context) ([]*entities.Organizer, error) {
	active := true
	filter := dto.OrganizerFilter{
		IsActive: &active,
	}
	pagination := dto.Pagination{
		Page:     1,
		PageSize: 100,
	}
	organizers, _, err := r.List(ctx, filter, pagination)
	return organizers, err
}

// Search busca organizadores por término
func (r *OrganizerRepository) Search(ctx context.Context, term string, limit int) ([]*entities.Organizer, error) {
	filter := dto.OrganizerFilter{
		Search: term,
	}
	pagination := dto.Pagination{
		Page:     1,
		PageSize: limit,
	}
	organizers, _, err := r.List(ctx, filter, pagination)
	return organizers, err
}

// FindByCountry busca organizadores por país
func (r *OrganizerRepository) FindByCountry(ctx context.Context, countryCode string, pagination dto.Pagination) ([]*entities.Organizer, int64, error) {
	filter := dto.OrganizerFilter{
		Country: countryCode,
	}
	return r.List(ctx, filter, pagination)
}

// ============================================================================
// OPERACIONES ESPECÍFICAS (TODOS ESTOS MÉTODOS ESTÁN CORRECTOS)
// ============================================================================

// UpdateVerification actualiza estado de verificación
func (r *OrganizerRepository) UpdateVerification(ctx context.Context, organizerID int64, verified bool, status string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET is_verified = $1, verification_status = $2, updated_at = NOW()
		WHERE id = $3
	`, verified, status, organizerID)
	if err != nil {
		return r.handleError(err, "failed to update verification")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// UpdateRating actualiza calificación
func (r *OrganizerRepository) UpdateRating(ctx context.Context, organizerID int64, rating float64, reviewCount int) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET organizer_rating = $1, rating_count = $2, updated_at = NOW()
		WHERE id = $3
	`, rating, reviewCount, organizerID)
	if err != nil {
		return r.handleError(err, "failed to update rating")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// UpdateStatistics actualiza estadísticas
func (r *OrganizerRepository) UpdateStatistics(ctx context.Context, organizerID int64, eventsCount int, ticketsSold int64, revenue float64) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET total_events = $1, total_tickets_sold = $2, updated_at = NOW()
		WHERE id = $3
	`, eventsCount, ticketsSold, organizerID)
	if err != nil {
		return r.handleError(err, "failed to update statistics")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// UpdateContactInfo actualiza información de contacto
func (r *OrganizerRepository) UpdateContactInfo(ctx context.Context, organizerID int64, email, phone string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET contact_email = $1, contact_phone = $2, updated_at = NOW()
		WHERE id = $3
	`, email, phone, organizerID)
	if err != nil {
		return r.handleError(err, "failed to update contact info")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// UpdateLegalInfo actualiza información legal
func (r *OrganizerRepository) UpdateLegalInfo(ctx context.Context, organizerID int64, legalName, taxID string, country string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET legal_name = $1, tax_id = $2, country = $3, updated_at = NOW()
		WHERE id = $4
	`, legalName, taxID, country, organizerID)
	if err != nil {
		return r.handleError(err, "failed to update legal info")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// UpdateSocialLinks actualiza redes sociales
func (r *OrganizerRepository) UpdateSocialLinks(ctx context.Context, organizerID int64, socialLinks map[string]string) error {
	jsonData, err := json.Marshal(socialLinks)
	if err != nil {
		return fmt.Errorf("failed to marshal social links: %w", err)
	}
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET social_links = $1, updated_at = NOW()
		WHERE id = $2
	`, jsonData, organizerID)
	if err != nil {
		return r.handleError(err, "failed to update social links")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// AddSocialLink agrega una red social
func (r *OrganizerRepository) AddSocialLink(ctx context.Context, organizerID int64, platform, url string) error {
	// Obtener social links actuales
	var socialLinksJSON []byte
	err := r.db.QueryRow(ctx, `SELECT social_links FROM ticketing.organizers WHERE id = $1`, organizerID).Scan(&socialLinksJSON)
	if err != nil {
		return r.handleError(err, "failed to get social links")
	}

	var socialLinks map[string]string
	if len(socialLinksJSON) > 0 {
		json.Unmarshal(socialLinksJSON, &socialLinks)
	}
	if socialLinks == nil {
		socialLinks = make(map[string]string)
	}
	socialLinks[platform] = url

	return r.UpdateSocialLinks(ctx, organizerID, socialLinks)
}

// RemoveSocialLink elimina una red social
func (r *OrganizerRepository) RemoveSocialLink(ctx context.Context, organizerID int64, platform string) error {
	// Obtener social links actuales
	var socialLinksJSON []byte
	err := r.db.QueryRow(ctx, `SELECT social_links FROM ticketing.organizers WHERE id = $1`, organizerID).Scan(&socialLinksJSON)
	if err != nil {
		return r.handleError(err, "failed to get social links")
	}

	var socialLinks map[string]string
	if len(socialLinksJSON) > 0 {
		json.Unmarshal(socialLinksJSON, &socialLinks)
	}
	delete(socialLinks, platform)

	return r.UpdateSocialLinks(ctx, organizerID, socialLinks)
}

// IncrementEventCount incrementa contador de eventos
func (r *OrganizerRepository) IncrementEventCount(ctx context.Context, organizerID int64) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET total_events = total_events + 1, updated_at = NOW()
		WHERE id = $1
	`, organizerID)
	if err != nil {
		return r.handleError(err, "failed to increment event count")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// DecrementEventCount decrementa contador de eventos
func (r *OrganizerRepository) DecrementEventCount(ctx context.Context, organizerID int64) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE ticketing.organizers 
		SET total_events = GREATEST(0, total_events - 1), updated_at = NOW()
		WHERE id = $1
	`, organizerID)
	if err != nil {
		return r.handleError(err, "failed to decrement event count")
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("organizer not found")
	}
	return nil
}

// ============================================================================
// VERIFICACIONES (TODOS ESTOS MÉTODOS ESTÁN CORRECTOS)
// ============================================================================

// IsVerified verifica si un organizador está verificado
func (r *OrganizerRepository) IsVerified(ctx context.Context, organizerID int64) (bool, error) {
	var verified bool
	err := r.db.QueryRow(ctx, `SELECT is_verified FROM ticketing.organizers WHERE id = $1`, organizerID).Scan(&verified)
	if err != nil {
		return false, r.handleError(err, "failed to check verification status")
	}
	return verified, nil
}

// IsActive verifica si un organizador está activo
func (r *OrganizerRepository) IsActive(ctx context.Context, organizerID int64) (bool, error) {
	var active bool
	err := r.db.QueryRow(ctx, `SELECT is_active FROM ticketing.organizers WHERE id = $1`, organizerID).Scan(&active)
	if err != nil {
		return false, r.handleError(err, "failed to check active status")
	}
	return active, nil
}

// HasEvents verifica si tiene eventos asociados
func (r *OrganizerRepository) HasEvents(ctx context.Context, organizerID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticketing.events WHERE organizer_id = $1)`, organizerID).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check events existence")
	}
	return exists, nil
}

// internal/infrastructure/repositories/postgres/organizer_repository.go

// ============================================================================
// ESTADÍSTICAS
// ============================================================================

// CountEvents cuenta eventos de un organizador
func (r *OrganizerRepository) CountEvents(ctx context.Context, organizerID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ticketing.events WHERE organizer_id = $1`, organizerID).Scan(&count)
	if err != nil {
		return 0, r.handleError(err, "failed to count events")
	}
	return count, nil
}

// GetTotalRevenue obtiene ingresos totales
func (r *OrganizerRepository) GetTotalRevenue(ctx context.Context, organizerID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(tt.sold_quantity * tt.base_price), 0)
		FROM ticketing.events e
		JOIN ticketing.ticket_types tt ON e.id = tt.event_id
		WHERE e.organizer_id = $1
	`
	var revenue float64
	err := r.db.QueryRow(ctx, query, organizerID).Scan(&revenue)
	if err != nil {
		return 0, r.handleError(err, "failed to get total revenue")
	}
	return revenue, nil
}

// GetAverageRating obtiene calificación promedio
func (r *OrganizerRepository) GetAverageRating(ctx context.Context, organizerID int64) (float64, error) {
	var rating float64
	err := r.db.QueryRow(ctx, `SELECT organizer_rating FROM ticketing.organizers WHERE id = $1`, organizerID).Scan(&rating)
	if err != nil {
		return 0, r.handleError(err, "failed to get average rating")
	}
	return rating, nil
}

// GetStats está comentada porque OrganizerStats no tiene los campos necesarios
// Si se necesita en el futuro, crear un DTO específico
// func (r *OrganizerRepository) GetStats(ctx context.Context, organizerID int64) (*dto.OrganizerStats, error) { ... }
