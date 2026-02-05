package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/errors"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/query"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/scanner"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/types"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/utils"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/validations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// categoryRepository implementa repository.CategoryRepository usando helpers
type categoryRepository struct {
	db         *pgxpool.Pool
	converter  *types.Converter
	scanner    *scanner.RowScanner
	errHandler *errors.PostgresErrorHandler
	validator  *errors.Validator
	logger     *utils.Logger
}

// NewCategoryRepository crea una nueva instancia con helpers
func NewCategoryRepository(db *pgxpool.Pool) repository.CategoryRepository {
	return &categoryRepository{
		db:         db,
		converter:  types.NewConverter(),
		scanner:    scanner.NewRowScanner(),
		errHandler: errors.NewPostgresErrorHandler(),
		validator:  errors.NewValidator(),
		logger:     utils.NewLogger("category-repository"),
	}
}

// Create implementa repository.CategoryRepository.Create usando helpers
func (r *categoryRepository) Create(ctx context.Context, category *entities.Category, benefits []string) (string, error) {
	startTime := time.Now()

	// Validaciones usando helpers
	if err := r.validateCategoryForCreate(ctx, category, benefits); err != nil {
		return "", err
	}

	// Verificar que el evento existe y está activo
	eventExists, err := r.validateEventExists(ctx, category.EventID)
	if err != nil {
		return "", err
	}
	if !eventExists {
		return "", fmt.Errorf("event not found or inactive: %d", category.EventID)
	}

	// Generar public_uuid si no existe
	if category.PublicID == "" {
		category.PublicID = uuid.New().String()
	}

	// Validar UUID usando validations
	if !validations.IsValidUUID(category.PublicID) {
		return "", fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	// Validar fechas de venta
	if !r.validateSalesDates(category) {
		return "", fmt.Errorf("invalid sales dates: sales_end must be after sales_start")
	}

	// Usar transacción con transaction manager
	tm := errors.NewTransactionManager(r.errHandler)

	var createdPublicID string

	err = tm.ExecuteInTransaction(ctx, r.db, func(tx interface{}) error {
		// Insertar categoría
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
		err := tx.(pgx.Tx).QueryRow(ctx, query,
			category.PublicID,
			category.EventID,
			r.converter.Text(category.Name),
			r.converter.TextPtr(category.Description),
			category.Price,
			r.converter.Int32(category.QuantityAvailable),
			r.converter.Int32Ptr(category.MaxTicketsPerOrder),
			r.converter.TimestampPtr(category.SalesStart),
			r.converter.TimestampPtr(category.SalesEnd),
			r.converter.BoolPtr(category.IsActive),
		).Scan(&categoryID, &createdPublicID)

		if err != nil {
			if r.errHandler.IsDuplicateKey(err) {
				constraint := r.errHandler.GetConstraintName(err)
				if strings.Contains(strings.ToLower(constraint), "public_id") {
					return fmt.Errorf("public_id already exists: %s", category.PublicID)
				}
			}
			return r.errHandler.WrapError(err, "category repository", "create category")
		}

		// Insertar beneficios si existen
		if len(benefits) > 0 {
			if err := r.createCategoryBenefitsTx(ctx, tx.(pgx.Tx), categoryID, benefits); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		r.logger.DatabaseLogger("INSERT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": category.EventID,
			"name":     utils.SafeStringForLog(category.Name),
		})
		return "", err
	}

	r.logger.DatabaseLogger("INSERT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"event_id":       category.EventID,
		"name":           utils.SafeStringForLog(category.Name),
		"benefits_count": len(benefits),
	})

	return createdPublicID, nil
}

// FindByID implementa repository.CategoryRepository.FindByID usando scanner
func (r *categoryRepository) FindByID(ctx context.Context, id int64) (*entities.Category, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_id, event_id, name, description, price, 
			quantity_available, quantity_sold, max_tickets_per_order,
			sales_start, sales_end, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, id)
	category, err := r.scanCategory(row)

	if err != nil {
		if err.Error() == "category not found" {
			r.logger.Debug("Category not found", map[string]interface{}{
				"category_id": id,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"category_id": id,
		})

		return nil, r.errHandler.WrapError(err, "category repository", "find category by ID")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"category_id": id,
	})

	return category, nil
}

// FindByPublicID implementa repository.CategoryRepository.FindByPublicID
func (r *categoryRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Category, error) {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return nil, fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	query := `
		SELECT 
			id, public_id, event_id, name, description, price, 
			quantity_available, quantity_sold, max_tickets_per_order,
			sales_start, sales_end, is_active, created_at, updated_at
		FROM categories
		WHERE public_id = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, publicID)
	category, err := r.scanCategory(row)

	if err != nil {
		if err.Error() == "category not found" {
			r.logger.Debug("Category not found by public ID", map[string]interface{}{
				"public_id": publicID,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return nil, r.errHandler.WrapError(err, "category repository", "find category by public ID")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"category_id": category.ID,
		"public_id":   publicID,
	})

	return category, nil
}

// FindByEvent implementa repository.CategoryRepository.FindByEvent
func (r *categoryRepository) FindByEvent(ctx context.Context, eventID int64) ([]*entities.Category, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_id, event_id, name, description, price, 
			quantity_available, quantity_sold, max_tickets_per_order,
			sales_start, sales_end, is_active, created_at, updated_at
		FROM categories
		WHERE event_id = $1 AND is_active = true
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id":  eventID,
			"operation": "find_by_event",
		})

		return nil, r.errHandler.WrapError(err, "category repository", "find categories by event")
	}
	defer rows.Close()

	categories := []*entities.Category{}
	for rows.Next() {
		category, err := r.scanCategory(rows)
		if err != nil {
			r.logger.Error("Failed to scan category row", err, map[string]interface{}{
				"event_id": eventID,
			})
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": eventID,
		})

		return nil, r.errHandler.WrapError(err, "category repository", "iterate categories by event")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), int64(len(categories)), nil, map[string]interface{}{
		"event_id": eventID,
		"count":    len(categories),
	})

	return categories, nil
}

// FindByEventPublicID implementa repository.CategoryRepository.FindByEventPublicID
func (r *categoryRepository) FindByEventPublicID(ctx context.Context, eventPublicID string) ([]*entities.Category, error) {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(eventPublicID) {
		return nil, fmt.Errorf("invalid event public_id: must be a valid UUID")
	}

	query := `
		SELECT 
			c.id, c.public_id, c.event_id, c.name, c.description, c.price, 
			c.quantity_available, c.quantity_sold, c.max_tickets_per_order,
			c.sales_start, c.sales_end, c.is_active, c.created_at, c.updated_at
		FROM categories c
		JOIN events e ON c.event_id = e.id
		WHERE e.public_id = $1 AND c.is_active = true AND e.is_active = true
		ORDER BY c.created_at
	`

	rows, err := r.db.Query(ctx, query, eventPublicID)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"event_public_id": eventPublicID,
			"operation":       "find_by_event_public_id",
		})

		return nil, r.errHandler.WrapError(err, "category repository", "find categories by event public ID")
	}
	defer rows.Close()

	categories := []*entities.Category{}
	for rows.Next() {
		category, err := r.scanCategory(rows)
		if err != nil {
			r.logger.Error("Failed to scan category row", err, map[string]interface{}{
				"event_public_id": eventPublicID,
			})
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"event_public_id": eventPublicID,
		})

		return nil, r.errHandler.WrapError(err, "category repository", "iterate categories by event public ID")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), int64(len(categories)), nil, map[string]interface{}{
		"event_public_id": eventPublicID,
		"count":           len(categories),
	})

	return categories, nil
}

// Update implementa repository.CategoryRepository.Update usando helpers
func (r *categoryRepository) Update(ctx context.Context, category *entities.Category, benefits []string) error {
	startTime := time.Now()

	// Validaciones
	if err := r.validateCategoryForUpdate(ctx, category, benefits); err != nil {
		return err
	}

	// Validar UUID usando validations
	if !validations.IsValidUUID(category.PublicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	// Validar fechas de venta
	if !r.validateSalesDates(category) {
		return fmt.Errorf("invalid sales dates: sales_end must be after sales_start")
	}

	// Usar transacción con transaction manager
	tm := errors.NewTransactionManager(r.errHandler)

	err := tm.ExecuteInTransaction(ctx, r.db, func(tx interface{}) error {
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
		err := tx.(pgx.Tx).QueryRow(ctx, query,
			r.converter.Text(category.Name),
			r.converter.TextPtr(category.Description),
			category.Price,
			r.converter.Int32(category.QuantityAvailable),
			r.converter.Int32Ptr(category.MaxTicketsPerOrder),
			r.converter.TimestampPtr(category.SalesStart),
			r.converter.TimestampPtr(category.SalesEnd),
			r.converter.BoolPtr(category.IsActive),
			category.PublicID,
		).Scan(&categoryID)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("category not found")
			}
			return r.errHandler.WrapError(err, "category repository", "update category")
		}

		// Actualizar beneficios (eliminar los existentes y agregar nuevos)
		deleteQuery := "DELETE FROM ticket_benefits WHERE category_id = $1"
		_, err = tx.(pgx.Tx).Exec(ctx, deleteQuery, categoryID)
		if err != nil {
			return r.errHandler.WrapError(err, "category repository", "delete old benefits")
		}

		// Insertar nuevos beneficios
		if len(benefits) > 0 {
			if err := r.createCategoryBenefitsTx(ctx, tx.(pgx.Tx), categoryID, benefits); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": category.PublicID,
			"name":      utils.SafeStringForLog(category.Name),
		})
		return err
	}

	r.logger.DatabaseLogger("UPDATE", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"public_id":      category.PublicID,
		"benefits_count": len(benefits),
	})

	return nil
}

// Delete implementa repository.CategoryRepository.Delete
func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	startTime := time.Now()

	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.DatabaseLogger("DELETE", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"category_id": id,
		})

		return r.errHandler.WrapError(err, "category repository", "delete category")
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		r.logger.Debug("Category not found for deletion", map[string]interface{}{
			"category_id": id,
		})
		return fmt.Errorf("category not found")
	}

	r.logger.DatabaseLogger("DELETE", "categories", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"category_id": id,
	})

	return nil
}

// SoftDelete implementa repository.CategoryRepository.SoftDelete
func (r *categoryRepository) SoftDelete(ctx context.Context, publicID string) error {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	query := `
		UPDATE categories 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $1 AND is_active = true
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, publicID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Category not found or already inactive", map[string]interface{}{
				"public_id": publicID,
			})
			return fmt.Errorf("category not found or already inactive")
		}

		r.logger.DatabaseLogger("UPDATE", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return r.errHandler.WrapError(err, "category repository", "soft delete category")
	}

	r.logger.DatabaseLogger("UPDATE", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"category_id": id,
		"public_id":   publicID,
	})

	return nil
}

// List implementa repository.CategoryRepository.List usando query builder
func (r *categoryRepository) List(ctx context.Context, filter dto.CategoryFilter, pagination dto.Pagination) ([]*entities.Category, int64, error) {
	startTime := time.Now()

	// Usar query builder para construir la query
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_id, event_id, name, description, price, 
			quantity_available, quantity_sold, max_tickets_per_order,
			sales_start, sales_end, is_active, created_at, updated_at
		FROM categories
	`).Where("1=1", nil) // Condición inicial

	// Aplicar filtros
	if filter.IsActive != nil {
		qb.Where("is_active = ?", *filter.IsActive)
	}

	if filter.EventID != 0 {
		qb.Where("event_id = ?", filter.EventID)
	}

	if filter.PriceFrom > 0 {
		qb.Where("price >= ?", filter.PriceFrom)
	}

	if filter.PriceTo > 0 {
		qb.Where("price <= ?", filter.PriceTo)
	}

	if filter.Search != "" {
		qb.Where("(name ILIKE ? OR description ILIKE ?)",
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.DateFrom != "" {
		if dateFrom, err := utils.ParseDateFromString(filter.DateFrom); err == nil {
			qb.Where("created_at >= ?", dateFrom)
		}
	}

	if filter.DateTo != "" {
		if dateTo, err := utils.ParseDateFromString(filter.DateTo); err == nil {
			qb.Where("created_at <= ?", dateTo)
		}
	}

	// Solo categorías con ventas activas si se solicita
	if filter.ActiveSalesOnly {
		now := time.Now()
		qb.Where("sales_start <= ?", now).
			Where("(sales_end IS NULL OR sales_end > ?)", now)
	}

	// Ordenar
	qb.OrderBy("created_at", true) // DESC

	// Construir query de conteo
	countQuery, countArgs := qb.BuildCount()

	// Ejecutar count
	var total int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "count",
		})

		return nil, 0, r.errHandler.WrapError(err, "category repository", "count categories")
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
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "category repository", "list categories")
	}
	defer rows.Close()

	// Usar scanner para procesar resultados
	categories := []*entities.Category{}
	for rows.Next() {
		category, err := r.scanCategory(rows)
		if err != nil {
			r.logger.Error("Failed to scan category row", err, map[string]interface{}{
				"operation": "list",
			})
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "category repository", "iterate categories")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), int64(len(categories)), nil, map[string]interface{}{
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
		"found":     len(categories),
	})

	return categories, total, nil
}

// GetActiveCategories implementa repository.CategoryRepository.GetActiveCategories
func (r *categoryRepository) GetActiveCategories(ctx context.Context) ([]*entities.Category, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_id, event_id, name, description, price, 
			quantity_available, quantity_sold, max_tickets_per_order,
			sales_start, sales_end, is_active, created_at, updated_at
		FROM categories
		WHERE is_active = true 
		AND sales_start <= CURRENT_TIMESTAMP
		AND (sales_end IS NULL OR sales_end > CURRENT_TIMESTAMP)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_active_categories",
		})

		return nil, r.errHandler.WrapError(err, "category repository", "get active categories")
	}
	defer rows.Close()

	categories := []*entities.Category{}
	for rows.Next() {
		category, err := r.scanCategory(rows)
		if err != nil {
			r.logger.Error("Failed to scan category row", err, map[string]interface{}{
				"operation": "get_active_categories",
			})
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_active_categories",
		})

		return nil, r.errHandler.WrapError(err, "category repository", "iterate active categories")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), int64(len(categories)), nil, map[string]interface{}{
		"operation": "get_active_categories",
		"count":     len(categories),
	})

	return categories, nil
}

// GetCategoryWithBenefits implementa repository.CategoryRepository.GetCategoryWithBenefits
func (r *categoryRepository) GetCategoryWithBenefits(ctx context.Context, publicID string) (*entities.Category, []string, error) {
	startTime := time.Now()

	// Obtener categoría
	category, err := r.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, nil, err
	}

	// Obtener beneficios
	benefits, err := r.getCategoryBenefits(ctx, publicID)
	if err != nil {
		r.logger.Error("Failed to get category benefits", err, map[string]interface{}{
			"category_id": category.ID,
		})
		// Continuar sin beneficios
		benefits = []string{}
	}

	r.logger.DatabaseLogger("SELECT", "categories+ticket_benefits", time.Since(startTime), 1, nil, map[string]interface{}{
		"category_id":    category.ID,
		"benefits_count": len(benefits),
	})

	return category, benefits, nil
}

// UpdateStatus implementa repository.CategoryRepository.UpdateStatus
func (r *categoryRepository) UpdateStatus(ctx context.Context, publicID string, isActive bool) error {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	query := `
		UPDATE categories 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $2
	`

	result, err := r.db.Exec(ctx, query, isActive, publicID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
			"is_active": isActive,
			"operation": "update_status",
		})

		return r.errHandler.WrapError(err, "category repository", "update category status")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Category not found for status update", map[string]interface{}{
			"public_id": publicID,
		})
		return fmt.Errorf("category not found")
	}

	r.logger.DatabaseLogger("UPDATE", "categories", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"public_id": publicID,
		"is_active": isActive,
	})

	return nil
}

// GetAvailability implementa repository.CategoryRepository.GetAvailability
func (r *categoryRepository) GetAvailability(ctx context.Context, publicID string) (int32, error) {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return 0, fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	var quantityAvailable, quantitySold int32

	query := `
		SELECT quantity_available, quantity_sold
		FROM categories 
		WHERE public_id = $1 AND is_active = true
	`

	err := r.db.QueryRow(ctx, query, publicID).Scan(&quantityAvailable, &quantitySold)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Category not found or inactive for availability check", map[string]interface{}{
				"public_id": publicID,
			})
			return 0, fmt.Errorf("category not found or inactive")
		}

		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
			"operation": "get_availability",
		})

		return 0, r.errHandler.WrapError(err, "category repository", "get category availability")
	}

	available := quantityAvailable - quantitySold
	if available < 0 {
		available = 0
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"public_id": publicID,
		"available": available,
	})

	return available, nil
}

// ValidateForPurchase implementa repository.CategoryRepository.ValidateForPurchase
func (r *categoryRepository) ValidateForPurchase(ctx context.Context, publicID string, quantity int32) error {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	var (
		quantityAvailable int32
		quantitySold      int32
		maxPerOrder       *int32
		salesStart        *time.Time
		salesEnd          *time.Time
		isActive          bool
	)

	query := `
		SELECT quantity_available, quantity_sold, max_tickets_per_order, 
		       sales_start, sales_end, is_active
		FROM categories 
		WHERE public_id = $1
	`

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&quantityAvailable, &quantitySold, &maxPerOrder,
		&salesStart, &salesEnd, &isActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Category not found for purchase validation", map[string]interface{}{
				"public_id": publicID,
			})
			return fmt.Errorf("category not found")
		}

		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
			"operation": "validate_for_purchase",
		})

		return r.errHandler.WrapError(err, "category repository", "validate category for purchase")
	}

	// Validaciones
	if !isActive {
		return fmt.Errorf("category is not active")
	}

	now := time.Now()
	if salesStart != nil && salesStart.After(now) {
		return fmt.Errorf("sales have not started for this category")
	}

	if salesEnd != nil && salesEnd.Before(now) {
		return fmt.Errorf("sales have ended for this category")
	}

	available := quantityAvailable - quantitySold
	if available < quantity {
		return fmt.Errorf("not enough tickets available. Available: %d, Requested: %d", available, quantity)
	}

	if maxPerOrder != nil && quantity > *maxPerOrder {
		return fmt.Errorf("maximum tickets per order is %d", *maxPerOrder)
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"public_id": publicID,
		"quantity":  quantity,
		"available": available,
	})

	return nil
}

// GetStats implementa repository.CategoryRepository.GetStats
func (r *categoryRepository) GetStats(ctx context.Context, categoryID int64) (*dto.CategoryStatsResponse, error) {
	startTime := time.Now()

	query := `
		SELECT 
			quantity_available,
			quantity_sold,
			price * quantity_sold as total_revenue,
			(quantity_sold * 100.0 / NULLIF(quantity_available + quantity_sold, 0)) as sell_rate
		FROM categories
		WHERE id = $1
	`

	var stats dto.CategoryStatsResponse
	err := r.db.QueryRow(ctx, query, categoryID).Scan(
		&stats.Available,
		&stats.Sold,
		&stats.TotalRevenue,
		&stats.SellRate,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Category not found for stats", map[string]interface{}{
				"category_id": categoryID,
			})
			return nil, fmt.Errorf("category not found")
		}

		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"category_id": categoryID,
			"operation":   "get_stats",
		})

		return nil, r.errHandler.WrapError(err, "category repository", "get category stats")
	}

	// Calcular disponibilidad
	stats.Available = stats.Available - stats.Sold
	if stats.Available < 0 {
		stats.Available = 0
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"category_id": categoryID,
	})

	return &stats, nil
}

// GetGlobalStats implementa repository.CategoryRepository.GetGlobalStats
func (r *categoryRepository) GetGlobalStats(ctx context.Context) (*dto.CategoryGlobalStats, error) {
	startTime := time.Now()

	query := `
		SELECT 
			COUNT(*) as total_categories,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_categories,
			SUM(quantity_sold) as total_tickets_sold,
			SUM(price * quantity_sold) as total_revenue,
			AVG(price) as avg_price
		FROM categories
	`

	var stats dto.CategoryGlobalStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalCategories,
		&stats.ActiveCategories,
		&stats.TotalTicketsSold,
		&stats.TotalRevenue,
		&stats.AvgPrice,
	)

	if err != nil {
		r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_global_stats",
		})

		return nil, r.errHandler.WrapError(err, "category repository", "get global category stats")
	}

	r.logger.DatabaseLogger("SELECT", "categories", time.Since(startTime), 1, nil, map[string]interface{}{
		"operation": "get_global_stats",
	})

	return &stats, nil
}

// =============================================================================
// MÉTODOS PRIVADOS CON HELPERS
// =============================================================================

// scanCategory escanea una fila de categoría usando scanner
func (r *categoryRepository) scanCategory(row interface{}) (*entities.Category, error) {
	var category entities.Category
	var description string
	var salesEnd *time.Time

	// Escanear los valores básicos usando scanner
	values, err := r.scanner.ScanRowToMap(row, []string{
		"id", "public_id", "event_id", "name", "description", "price",
		"quantity_available", "quantity_sold", "max_tickets_per_order",
		"sales_start", "sales_end", "is_active", "created_at", "updated_at",
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to scan category row: %w", err)
	}

	// Mapear valores a la estructura category
	// En implementación real usarías reflection o mapping manual
	// Esto es simplificado para el ejemplo
	category.ID = values["id"].(int64)
	category.PublicID = values["public_id"].(string)
	category.EventID = values["event_id"].(int64)
	category.Name = values["name"].(string)
	category.Price = values["price"].(float64)
	category.QuantityAvailable = values["quantity_available"].(int32)
	category.QuantitySold = values["quantity_sold"].(int32)
	category.IsActive = values["is_active"].(bool)
	category.CreatedAt = values["created_at"].(time.Time)
	category.UpdatedAt = values["updated_at"].(time.Time)

	// Manejar valores nullable
	if maxPerOrder, ok := values["max_tickets_per_order"].(int32); ok && maxPerOrder > 0 {
		category.MaxTicketsPerOrder = &maxPerOrder
	}
	if desc, ok := values["description"].(string); ok && desc != "" {
		category.Description = &desc
	}
	if salesStart, ok := values["sales_start"].(time.Time); ok && !salesStart.IsZero() {
		category.SalesStart = &salesStart
	}
	if se, ok := values["sales_end"].(time.Time); ok && !se.IsZero() {
		category.SalesEnd = &se
	}

	return &category, nil
}

// createCategoryBenefitsTx inserta los beneficios de una categoría dentro de una transacción
func (r *categoryRepository) createCategoryBenefitsTx(ctx context.Context, tx pgx.Tx, categoryID int64, benefits []string) error {
	query := `
		INSERT INTO ticket_benefits (
			public_id, category_id, name, display_order, is_active
		) VALUES ($1, $2, $3, $4, $5)
	`

	for i, benefit := range benefits {
		benefit = strings.TrimSpace(benefit)
		if benefit == "" {
			continue
		}

		benefitPublicID := uuid.New().String()
		_, err := tx.Exec(ctx, query,
			benefitPublicID,
			categoryID,
			benefit,
			i, // display_order
			true,
		)

		if err != nil {
			r.logger.Error("Failed to insert benefit", err, map[string]interface{}{
				"category_id": categoryID,
				"benefit":     utils.SafeStringForLog(benefit),
			})
			return fmt.Errorf("failed to insert benefit '%s': %w", utils.SafeStringForLog(benefit), err)
		}
	}

	return nil
}

// getCategoryBenefits obtiene los beneficios de una categoría
func (r *categoryRepository) getCategoryBenefits(ctx context.Context, categoryPublicID string) ([]string, error) {
	// Validar UUID usando helpers
	if !validations.IsValidUUID(categoryPublicID) {
		return nil, fmt.Errorf("invalid category ID format")
	}

	query := `
		SELECT tb.name
		FROM ticket_benefits tb
		JOIN categories c ON tb.category_id = c.id
		WHERE c.public_id = $1 AND tb.is_active = true
		ORDER BY tb.display_order
	`

	rows, err := r.db.Query(ctx, query, categoryPublicID)
	if err != nil {
		return nil, r.errHandler.WrapError(err, "category repository", "get category benefits")
	}
	defer rows.Close()

	var benefits []string
	for rows.Next() {
		var benefit string
		if err := rows.Scan(&benefit); err != nil {
			r.logger.Error("Failed to scan benefit", err, map[string]interface{}{
				"category_public_id": categoryPublicID,
			})
			return nil, fmt.Errorf("failed to scan benefit: %w", err)
		}
		benefits = append(benefits, benefit)
	}

	if err := rows.Err(); err != nil {
		return nil, r.errHandler.WrapError(err, "category repository", "iterate benefits")
	}

	return benefits, nil
}

// validateCategoryForCreate valida una categoría para creación
func (r *categoryRepository) validateCategoryForCreate(ctx context.Context, category *entities.Category, benefits []string) error {
	// Usar validator
	r.validator.Required("name", category.Name).
		Required("event_id", category.EventID).
		Required("price", category.Price).
		Required("quantity_available", category.QuantityAvailable).
		Positive("price", category.Price).
		Positive("quantity_available", float64(category.QuantityAvailable)).
		MaxLength("name", category.Name, 200)

	if category.Description != nil && *category.Description != "" {
		r.validator.MaxLength("description", *category.Description, 1000)
	}

	if category.MaxTicketsPerOrder != nil {
		r.validator.MinValue("max_tickets_per_order", float64(*category.MaxTicketsPerOrder), 1)
	}

	if validationErr := r.validator.Validate(); validationErr != nil {
		return validationErr
	}

	// Validar nombres duplicados para el mismo evento
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM categories WHERE event_id = $1 AND name = $2)",
		category.EventID, category.Name).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("category name '%s' already exists for this event", category.Name)
	}

	return nil
}

// validateCategoryForUpdate valida una categoría para actualización
func (r *categoryRepository) validateCategoryForUpdate(ctx context.Context, category *entities.Category, benefits []string) error {
	// Validar que la categoría exista
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE public_id = $1)", category.PublicID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate category existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("category not found with public_id: %s", category.PublicID)
	}

	// Usar validator para validaciones generales
	return r.validateCategoryForCreate(ctx, category, benefits)
}

// validateEventExists verifica que el evento existe y está activo
func (r *categoryRepository) validateEventExists(ctx context.Context, eventID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND is_active = true)",
		eventID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to validate event: %w", err)
	}

	return exists, nil
}

// validateSalesDates valida las fechas de venta de una categoría usando utils
func (r *categoryRepository) validateSalesDates(category *entities.Category) bool {
	if category.SalesStart == nil {
		return true // Fecha de inicio opcional
	}

	if category.SalesEnd == nil {
		return true // Fecha de fin opcional
	}

	return category.SalesEnd.After(*category.SalesStart)
}
