// category_repository.go - COMPLETO Y CORREGIDO
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	DB *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{DB: db}
}

// CreateCategory crea una nueva categoría para un evento - VERSIÓN CORREGIDA
func (r *CategoryRepository) CreateCategory(ctx context.Context, category *models.Category, benefits []string) (string, error) {
	// Necesitamos el eventPublicID, no el EventID (int64)
	// Pero category solo tiene EventID (int64), no el public_id del evento

	// Primero necesitamos obtener el public_id del evento usando su ID
	var eventPublicID string
	err := r.DB.QueryRow(ctx,
		"SELECT public_id FROM events WHERE id = $1 AND is_active = true",
		category.EventID,
	).Scan(&eventPublicID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("event not found or inactive")
		}
		return "", fmt.Errorf("error getting event public_id: %w", err)
	}

	// Generar UUID público para la categoría
	publicID := uuid.New().String()

	// Usar transacción para asegurar consistencia
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// ✅ CORRECCIÓN: Insertar categoría usando event_id (int64), no public_id
	query := `
        INSERT INTO categories (
            public_id, event_id, name, description, price, 
            quantity_available, max_tickets_per_order, 
            sales_start, sales_end, is_active
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, public_id
    `

	var categoryID int64
	err = tx.QueryRow(ctx, query,
		publicID,
		category.EventID, // ✅ EventID (int64)
		category.Name,
		ToPgTextFromPtr(category.Description),
		category.Price,
		category.QuantityAvailable,
		category.MaxTicketsPerOrder,
		category.SalesStart,
		ToPgTimestampFromPtr(category.SalesEnd),
		category.IsActive,
	).Scan(&categoryID, &publicID)

	if err != nil {
		return "", fmt.Errorf("failed to create category: %w", err)
	}

	// Insertar beneficios si existen
	if len(benefits) > 0 {
		err = r.createCategoryBenefitsTx(ctx, tx, categoryID, benefits)
		if err != nil {
			return "", fmt.Errorf("failed to create category benefits: %w", err)
		}
	}

	// Commit de la transacción
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	return publicID, nil
}

// createCategoryBenefitsTx inserta los beneficios de una categoría dentro de una transacción
func (r *CategoryRepository) createCategoryBenefitsTx(ctx context.Context, tx pgx.Tx, categoryID int64, benefits []string) error {
	query := `
		INSERT INTO ticket_benefits (
			public_id, category_id, name, display_order, is_active
		) VALUES ($1, $2, $3, $4, $5)
	`

	for i, benefit := range benefits {
		benefitPublicID := uuid.New().String()
		_, err := tx.Exec(ctx, query,
			benefitPublicID,
			categoryID,
			benefit,
			i, // display_order
			true,
		)

		if err != nil {
			return fmt.Errorf("error inserting benefit '%s': %w", SafeStringForLog(benefit), err)
		}
	}

	return nil
}

// GetCategoryByPublicID obtiene una categoría por su UUID público
func (r *CategoryRepository) GetCategoryByPublicID(ctx context.Context, publicID string) (*models.Category, error) {
	if !IsValidUUID(publicID) {
		return nil, errors.New("invalid category ID format")
	}

	var category models.Category
	var description pgtype.Text
	var salesEnd pgtype.Timestamp

	query := `
		SELECT 
			c.id, c.public_id, c.event_id, c.name, c.description, c.price, 
			c.quantity_available, c.quantity_sold, c.max_tickets_per_order,
			c.sales_start, c.sales_end, c.is_active, c.created_at, c.updated_at
		FROM categories c
		WHERE c.public_id = $1 AND c.is_active = true
	`

	err := r.DB.QueryRow(ctx, query, publicID).Scan(
		&category.ID,
		&category.PublicID,
		&category.EventID,
		&category.Name,
		&description,
		&category.Price,
		&category.QuantityAvailable,
		&category.QuantitySold,
		&category.MaxTicketsPerOrder,
		&category.SalesStart,
		&salesEnd,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	category.Description = ToStringFromPgText(description)
	category.SalesEnd = ToTimeFromPgTimestamp(salesEnd)

	return &category, nil
}

// getCategoryBenefits obtiene los beneficios de una categoría
func (r *CategoryRepository) getCategoryBenefits(ctx context.Context, categoryPublicID string) ([]string, error) {
	if !IsValidUUID(categoryPublicID) {
		return nil, errors.New("invalid category ID format")
	}

	query := `
		SELECT tb.name
		FROM ticket_benefits tb
		JOIN categories c ON tb.category_id = c.id
		WHERE c.public_id = $1 AND tb.is_active = true
		ORDER BY tb.display_order
	`

	rows, err := r.DB.Query(ctx, query, categoryPublicID)
	if err != nil {
		return nil, fmt.Errorf("error querying benefits: %w", err)
	}
	defer rows.Close()

	var benefits []string
	for rows.Next() {
		var benefit string
		if err := rows.Scan(&benefit); err != nil {
			return nil, fmt.Errorf("error scanning benefit: %w", err)
		}
		benefits = append(benefits, benefit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating benefits: %w", err)
	}

	return benefits, nil
}

// GetCategoriesByEvent obtiene todas las categorías de un evento
func (r *CategoryRepository) GetCategoriesByEvent(ctx context.Context, eventPublicID string) ([]*models.Category, error) {
	if !IsValidUUID(eventPublicID) {
		return nil, errors.New("invalid event ID format")
	}

	// Primero verificamos que el evento existe
	var eventExists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE public_id = $1 AND is_active = true)",
		eventPublicID,
	).Scan(&eventExists)

	if err != nil {
		return nil, fmt.Errorf("error verifying event: %w", err)
	}

	if !eventExists {
		return nil, errors.New("event not found or inactive")
	}

	query := `
		SELECT 
			c.id, c.public_id, c.event_id, c.name, c.description, c.price, 
			c.quantity_available, c.quantity_sold, c.max_tickets_per_order,
			c.sales_start, c.sales_end, c.is_active, c.created_at, c.updated_at
		FROM categories c
		JOIN events e ON c.event_id = e.id
		WHERE e.public_id = $1 AND c.is_active = true
		ORDER BY c.created_at
	`

	rows, err := r.DB.Query(ctx, query, eventPublicID)
	if err != nil {
		return nil, fmt.Errorf("error querying categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var category models.Category
		var description pgtype.Text
		var salesEnd pgtype.Timestamp

		err := rows.Scan(
			&category.ID,
			&category.PublicID,
			&category.EventID,
			&category.Name,
			&description,
			&category.Price,
			&category.QuantityAvailable,
			&category.QuantitySold,
			&category.MaxTicketsPerOrder,
			&category.SalesStart,
			&salesEnd,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning category: %w", err)
		}

		// Convertir pgtype a tipos nativos usando helpers
		category.Description = ToStringFromPgText(description)
		category.SalesEnd = ToTimeFromPgTimestamp(salesEnd)
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// UpdateCategoryStatus actualiza el estado activo/inactivo de una categoría
func (r *CategoryRepository) UpdateCategoryStatus(ctx context.Context, categoryPublicID string, isActive bool) error {
	if !IsValidUUID(categoryPublicID) {
		return errors.New("invalid category ID format")
	}

	query := `
		UPDATE categories 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $2
	`

	result, err := r.DB.Exec(ctx, query, isActive, categoryPublicID)
	if err != nil {
		return fmt.Errorf("error updating category status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("category not found")
	}

	return nil
}

// GetCategoryAvailability verifica la disponibilidad de una categoría
func (r *CategoryRepository) GetCategoryAvailability(ctx context.Context, categoryPublicID string) (int32, error) {
	if !IsValidUUID(categoryPublicID) {
		return 0, errors.New("invalid category ID format")
	}

	var quantityAvailable, quantitySold int32

	query := `
		SELECT quantity_available, quantity_sold
		FROM categories 
		WHERE public_id = $1 AND is_active = true
	`

	err := r.DB.QueryRow(ctx, query, categoryPublicID).Scan(&quantityAvailable, &quantitySold)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("category not found or inactive")
		}
		return 0, fmt.Errorf("error getting category availability: %w", err)
	}

	return quantityAvailable - quantitySold, nil
}

// ValidateCategoryForPurchase valida si una categoría puede ser usada para compra
func (r *CategoryRepository) ValidateCategoryForPurchase(ctx context.Context, categoryPublicID string, quantity int32) error {
	if !IsValidUUID(categoryPublicID) {
		return errors.New("invalid category ID format")
	}

	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	var (
		quantityAvailable int32
		quantitySold      int32
		maxPerOrder       int32
		salesStart        time.Time
		salesEnd          pgtype.Timestamp
		isActive          bool
	)

	query := `
		SELECT quantity_available, quantity_sold, max_tickets_per_order, 
		       sales_start, sales_end, is_active
		FROM categories 
		WHERE public_id = $1
	`

	err := r.DB.QueryRow(ctx, query, categoryPublicID).Scan(
		&quantityAvailable, &quantitySold, &maxPerOrder,
		&salesStart, &salesEnd, &isActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("category not found")
		}
		return fmt.Errorf("error validating category: %w", err)
	}

	// Validaciones
	if !isActive {
		return errors.New("category is not active")
	}

	now := time.Now()
	if salesStart.After(now) {
		return errors.New("sales have not started for this category")
	}

	if salesEnd.Valid && salesEnd.Time.Before(now) {
		return errors.New("sales have ended for this category")
	}

	available := quantityAvailable - quantitySold
	if available < quantity {
		return fmt.Errorf("not enough tickets available. Available: %d, Requested: %d", available, quantity)
	}

	if quantity > maxPerOrder {
		return fmt.Errorf("maximum tickets per order is %d", maxPerOrder)
	}

	return nil
}

// GetCategoryWithBenefits obtiene una categoría con sus beneficios
func (r *CategoryRepository) GetCategoryWithBenefits(ctx context.Context, categoryPublicID string) (*models.Category, []string, error) {
	category, err := r.GetCategoryByPublicID(ctx, categoryPublicID)
	if err != nil {
		return nil, nil, err
	}

	benefits, err := r.getCategoryBenefits(ctx, categoryPublicID)
	if err != nil {
		return nil, nil, err
	}

	return category, benefits, nil
}

// GetActiveCategories obtiene todas las categorías activas con ventas vigentes
func (r *CategoryRepository) GetActiveCategories(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT 
			c.id, c.public_id, c.event_id, c.name, c.description, c.price, 
			c.quantity_available, c.quantity_sold, c.max_tickets_per_order,
			c.sales_start, c.sales_end, c.is_active, c.created_at, c.updated_at
		FROM categories c
		WHERE c.is_active = true 
		AND c.sales_start <= CURRENT_TIMESTAMP
		AND (c.sales_end IS NULL OR c.sales_end > CURRENT_TIMESTAMP)
		ORDER BY c.created_at DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying active categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var category models.Category
		var description pgtype.Text
		var salesEnd pgtype.Timestamp

		err := rows.Scan(
			&category.ID,
			&category.PublicID,
			&category.EventID,
			&category.Name,
			&description,
			&category.Price,
			&category.QuantityAvailable,
			&category.QuantitySold,
			&category.MaxTicketsPerOrder,
			&category.SalesStart,
			&salesEnd,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning category: %w", err)
		}

		// Convertir pgtype a tipos nativos usando helpers
		category.Description = ToStringFromPgText(description)
		category.SalesEnd = ToTimeFromPgTimestamp(salesEnd)
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// UpdateCategory actualiza una categoría existente
func (r *CategoryRepository) UpdateCategory(ctx context.Context, categoryPublicID string, category *models.Category, benefits []string) error {
	if !IsValidUUID(categoryPublicID) {
		return errors.New("invalid category ID format")
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Actualizar la categoría
	query := `
		UPDATE categories 
		SET name = $1, description = $2, price = $3, 
			quantity_available = $4, max_tickets_per_order = $5,
			sales_start = $6, sales_end = $7, is_active = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $9
		RETURNING id
	`

	var categoryID int64
	err = tx.QueryRow(ctx, query,
		category.Name,
		ToPgTextFromPtr(category.Description),
		category.Price,
		category.QuantityAvailable,
		category.MaxTicketsPerOrder,
		category.SalesStart,
		ToPgTimestampFromPtr(category.SalesEnd),
		category.IsActive,
		categoryPublicID,
	).Scan(&categoryID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("category not found")
		}
		return fmt.Errorf("error updating category: %w", err)
	}

	// Actualizar beneficios (eliminar los existentes y agregar nuevos)
	deleteQuery := "DELETE FROM ticket_benefits WHERE category_id = $1"
	_, err = tx.Exec(ctx, deleteQuery, categoryID)
	if err != nil {
		return fmt.Errorf("error deleting old benefits: %w", err)
	}

	// Insertar nuevos beneficios
	if len(benefits) > 0 {
		err = r.createCategoryBenefitsTx(ctx, tx, categoryID, benefits)
		if err != nil {
			return fmt.Errorf("failed to create category benefits: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// DeleteCategory elimina lógicamente una categoría
func (r *CategoryRepository) DeleteCategory(ctx context.Context, categoryPublicID string) error {
	if !IsValidUUID(categoryPublicID) {
		return errors.New("invalid category ID format")
	}

	query := `
		UPDATE categories 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $1
	`

	result, err := r.DB.Exec(ctx, query, categoryPublicID)
	if err != nil {
		return fmt.Errorf("error deleting category: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("category not found")
	}

	return nil
}

// GetCategoryByID obtiene una categoría por su ID interno (para uso interno)
func (r *CategoryRepository) GetCategoryByID(ctx context.Context, categoryID int64) (*models.Category, error) {
	var category models.Category
	var description pgtype.Text
	var salesEnd pgtype.Timestamp

	query := `
		SELECT 
			c.id, c.public_id, c.event_id, c.name, c.description, c.price, 
			c.quantity_available, c.quantity_sold, c.max_tickets_per_order,
			c.sales_start, c.sales_end, c.is_active, c.created_at, c.updated_at
		FROM categories c
		WHERE c.id = $1
	`

	err := r.DB.QueryRow(ctx, query, categoryID).Scan(
		&category.ID,
		&category.PublicID,
		&category.EventID,
		&category.Name,
		&description,
		&category.Price,
		&category.QuantityAvailable,
		&category.QuantitySold,
		&category.MaxTicketsPerOrder,
		&category.SalesStart,
		&salesEnd,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	category.Description = ToStringFromPgText(description)
	category.SalesEnd = ToTimeFromPgTimestamp(salesEnd)

	return &category, nil
}
