package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// CategoryRepository implementa la interfaz repository.CategoryRepository usando PostgreSQL
type CategoryRepository struct {
	db *sqlx.DB
	// Opcional: logger si lo necesitas
	// logger logger.Logger
}

// NewCategoryRepository crea una nueva instancia del repositorio
func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// Helper privado para mapear errores de PostgreSQL a nuestros errores de dominio
func (r *CategoryRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Errores específicos de PostgreSQL
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pqErr.Constraint, "categories_slug_key") {
				return repository.ErrCategoryDuplicateSlug
			}
			// Podríamos tener más constraints aquí
		}
	}

	// Wrap el error para dar contexto, pero mantener el error original si es uno nuestro
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrCategoryNotFound
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Find busca categorías según los criterios del filtro.
// Retorna las categorías y el total de registros (ignorando paginación).
func (r *CategoryRepository) Find(ctx context.Context, filter *repository.CategoryFilter) ([]*entities.Category, int64, error) {
	// 1. Construir la query base
	baseQuery := `
        SELECT 
            id, public_uuid, name, slug, description, icon, color_hex,
            parent_id, level, path, total_events, total_tickets_sold, total_revenue,
            is_active, is_featured, sort_order, meta_title, meta_description,
            created_at, updated_at
        FROM ticketing.categories
        WHERE 1=1
    `

	// Query para contar el total (sin paginación)
	countQuery := `SELECT COUNT(*) FROM ticketing.categories WHERE 1=1`

	// 2. Acumuladores para condiciones y argumentos
	var conditions []string
	var args []interface{}
	argPos := 1

	// 3. Aplicar filtros si existen
	if filter != nil {
		// Filtro por IDs específicos
		if len(filter.IDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("id = ANY($%d)", argPos))
			args = append(args, pq.Array(filter.IDs))
			argPos++
		}

		// Filtro por PublicIDs
		if len(filter.PublicIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("public_uuid = ANY($%d)", argPos))
			args = append(args, pq.Array(filter.PublicIDs))
			argPos++
		}

		// Filtro por ParentID
		// Si ParentID es nil: no filtramos por padre (traemos todos)
		// Si ParentID apunta a 0: buscamos categorías raíz (parent_id IS NULL)
		// Si ParentID apunta a un valor > 0: buscamos hijos de ese padre
		if filter.ParentID != nil {
			if *filter.ParentID == 0 {
				conditions = append(conditions, "parent_id IS NULL")
			} else {
				conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argPos))
				args = append(args, *filter.ParentID)
				argPos++
			}
		}

		// Filtro por IsActive
		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = $%d", argPos))
			args = append(args, *filter.IsActive)
			argPos++
		}

		// Filtro por IsFeatured
		if filter.IsFeatured != nil {
			conditions = append(conditions, fmt.Sprintf("is_featured = $%d", argPos))
			args = append(args, *filter.IsFeatured)
			argPos++
		}

		// Filtro por Slug exacto
		if filter.Slug != nil {
			conditions = append(conditions, fmt.Sprintf("slug = $%d", argPos))
			args = append(args, *filter.Slug)
			argPos++
		}

		// Filtro de búsqueda por texto (en múltiples campos)
		if filter.SearchTerm != nil && *filter.SearchTerm != "" {
			searchTerm := "%" + *filter.SearchTerm + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(name ILIKE $%d OR slug ILIKE $%d OR description ILIKE $%d)",
				argPos, argPos, argPos,
			))
			args = append(args, searchTerm)
			argPos++
		}

		// Filtros por nivel
		if filter.MinLevel != nil {
			conditions = append(conditions, fmt.Sprintf("level >= $%d", argPos))
			args = append(args, *filter.MinLevel)
			argPos++
		}
		if filter.MaxLevel != nil {
			conditions = append(conditions, fmt.Sprintf("level <= $%d", argPos))
			args = append(args, *filter.MaxLevel)
			argPos++
		}
	}

	// 4. Unir todas las condiciones
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// 5. Ejecutar count query para obtener el total
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count categories")
	}

	// 6. Añadir ordenamiento y paginación a la query base
	if filter != nil {
		// Ordenamiento
		sortBy := "sort_order"
		sortOrder := "ASC"
		if filter.SortBy != "" {
			// Validar que sortBy sea una columna permitida para evitar inyección SQL
			allowedSortColumns := map[string]bool{
				"name": true, "created_at": true, "total_events": true,
				"sort_order": true, "level": true,
			}
			if allowedSortColumns[filter.SortBy] {
				sortBy = filter.SortBy
			}
		}
		if filter.SortOrder != "" {
			if strings.ToUpper(filter.SortOrder) == "DESC" {
				sortOrder = "DESC"
			}
		}
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

		// Paginación
		if filter.Limit > 0 {
			baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
			args = append(args, filter.Limit)
			argPos++
		}
		if filter.Offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET $%d", argPos)
			args = append(args, filter.Offset)
			argPos++
		}
	}

	// 7. Ejecutar query principal
	var categories []*entities.Category
	err = r.db.SelectContext(ctx, &categories, baseQuery, args...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find categories")
	}

	return categories, total, nil
}

// GetByID obtiene una categoría por su ID numérico
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*entities.Category, error) {
	filter := &repository.CategoryFilter{
		IDs:   []int64{id},
		Limit: 1,
	}

	categories, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(categories) == 0 {
		return nil, repository.ErrCategoryNotFound
	}

	return categories[0], nil
}

// GetByPublicID obtiene una categoría por su UUID público
func (r *CategoryRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.Category, error) {
	filter := &repository.CategoryFilter{
		PublicIDs: []string{publicID},
		Limit:     1,
	}

	categories, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(categories) == 0 {
		return nil, repository.ErrCategoryNotFound
	}

	return categories[0], nil
}

// GetBySlug obtiene una categoría por su slug
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*entities.Category, error) {
	filter := &repository.CategoryFilter{
		Slug:  &slug,
		Limit: 1,
	}

	categories, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(categories) == 0 {
		return nil, repository.ErrCategoryNotFound
	}

	return categories[0], nil
}

// Exists verifica si existe una categoría con el ID dado
func (r *CategoryRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.categories WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, r.handleError(err, "failed to check category existence")
	}
	return exists, nil
}

// ExistsBySlug verifica si existe una categoría con el slug dado
func (r *CategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.categories WHERE slug = $1)`
	err := r.db.GetContext(ctx, &exists, query, slug)
	if err != nil {
		return false, r.handleError(err, "failed to check slug existence")
	}
	return exists, nil
}

// Create inserta una nueva categoría en la base de datos
func (r *CategoryRepository) Create(ctx context.Context, category *entities.Category) error {
	query := `
        INSERT INTO ticketing.categories (
            public_uuid, name, slug, description, icon, color_hex,
            parent_id, level, path, total_events, total_tickets_sold, total_revenue,
            is_active, is_featured, sort_order, meta_title, meta_description,
            created_at, updated_at
        ) VALUES (
            gen_random_uuid(), $1, $2, $3, $4, $5,
            $6, $7, $8, $9, $10, $11,
            $12, $13, $14, $15, $16,
            NOW(), NOW()
        )
        RETURNING id, public_uuid, created_at, updated_at
    `

	err := r.db.QueryRowContext(
		ctx, query,
		category.Name, category.Slug, category.Description, category.Icon, category.ColorHex,
		category.ParentID, category.Level, category.Path,
		category.TotalEvents, category.TotalTicketsSold, category.TotalRevenue,
		category.IsActive, category.IsFeatured, category.SortOrder,
		category.MetaTitle, category.MetaDescription,
	).Scan(&category.ID, &category.PublicID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create category")
	}

	return nil
}

// Update actualiza una categoría existente
func (r *CategoryRepository) Update(ctx context.Context, category *entities.Category) error {
	// Primero verificamos que la categoría existe
	exists, err := r.Exists(ctx, category.ID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrCategoryNotFound
	}

	query := `
        UPDATE ticketing.categories SET
            name = $1,
            slug = $2,
            description = $3,
            icon = $4,
            color_hex = $5,
            parent_id = $6,
            level = $7,
            path = $8,
            is_active = $9,
            is_featured = $10,
            sort_order = $11,
            meta_title = $12,
            meta_description = $13,
            updated_at = NOW()
        WHERE id = $14
        RETURNING updated_at
    `

	err = r.db.QueryRowContext(
		ctx, query,
		category.Name, category.Slug, category.Description, category.Icon, category.ColorHex,
		category.ParentID, category.Level, category.Path,
		category.IsActive, category.IsFeatured, category.SortOrder,
		category.MetaTitle, category.MetaDescription,
		category.ID,
	).Scan(&category.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update category")
	}

	return nil
}

// Delete elimina una categoría por su ID
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	// Verificar si tiene hijos
	var childCount int
	err := r.db.GetContext(ctx, &childCount,
		`SELECT COUNT(*) FROM ticketing.categories WHERE parent_id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to check child categories")
	}
	if childCount > 0 {
		return repository.ErrCategoryHasChildren
	}

	result, err := r.db.ExecContext(ctx, `DELETE FROM ticketing.categories WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete category")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCategoryNotFound
	}

	return nil
}
