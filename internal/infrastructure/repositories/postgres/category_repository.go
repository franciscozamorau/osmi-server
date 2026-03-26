// internal/infrastructure/repositories/postgres/category_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// CategoryRepository implementa la interfaz repository.CategoryRepository usando PostgreSQL
type CategoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository crea una nueva instancia del repositorio
func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL a nuestros errores de dominio
func (r *CategoryRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrCategoryNotFound
	}

	// Verificar si es un error de PostgreSQL con código
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "categories_slug_key") {
				return repository.ErrCategoryDuplicateSlug
			}
		}
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Find busca categorías según los criterios del filtro
func (r *CategoryRepository) Find(ctx context.Context, filter *repository.CategoryFilter) ([]*entities.Category, int64, error) {
	baseQuery := `
        SELECT 
            id, public_uuid, name, slug, description, icon, color_hex,
            parent_id, level, path, total_events, total_tickets_sold, total_revenue,
            is_active, is_featured, sort_order, meta_title, meta_description,
            created_at, updated_at
        FROM ticketing.categories
        WHERE 1=1
    `

	countQuery := `SELECT COUNT(*) FROM ticketing.categories WHERE 1=1`

	var conditions []string
	args := pgx.NamedArgs{}
	argPos := 1

	if filter != nil {
		if len(filter.IDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("id = ANY(@id_%d)", argPos))
			args[fmt.Sprintf("id_%d", argPos)] = filter.IDs
			argPos++
		}

		if len(filter.PublicIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("public_uuid = ANY(@public_%d)", argPos))
			args[fmt.Sprintf("public_%d", argPos)] = filter.PublicIDs
			argPos++
		}

		if filter.ParentID != nil {
			if *filter.ParentID == 0 {
				conditions = append(conditions, "parent_id IS NULL")
			} else {
				conditions = append(conditions, fmt.Sprintf("parent_id = @parent_%d", argPos))
				args[fmt.Sprintf("parent_%d", argPos)] = *filter.ParentID
				argPos++
			}
		}

		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = @active_%d", argPos))
			args[fmt.Sprintf("active_%d", argPos)] = *filter.IsActive
			argPos++
		}

		if filter.IsFeatured != nil {
			conditions = append(conditions, fmt.Sprintf("is_featured = @featured_%d", argPos))
			args[fmt.Sprintf("featured_%d", argPos)] = *filter.IsFeatured
			argPos++
		}

		if filter.Slug != nil {
			conditions = append(conditions, fmt.Sprintf("slug = @slug_%d", argPos))
			args[fmt.Sprintf("slug_%d", argPos)] = *filter.Slug
			argPos++
		}

		if filter.SearchTerm != nil && *filter.SearchTerm != "" {
			searchTerm := "%" + *filter.SearchTerm + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(name ILIKE @search_%d OR slug ILIKE @search_%d OR description ILIKE @search_%d)",
				argPos, argPos, argPos,
			))
			args[fmt.Sprintf("search_%d", argPos)] = searchTerm
			argPos++
		}

		if filter.MinLevel != nil {
			conditions = append(conditions, fmt.Sprintf("level >= @min_level_%d", argPos))
			args[fmt.Sprintf("min_level_%d", argPos)] = *filter.MinLevel
			argPos++
		}
		if filter.MaxLevel != nil {
			conditions = append(conditions, fmt.Sprintf("level <= @max_level_%d", argPos))
			args[fmt.Sprintf("max_level_%d", argPos)] = *filter.MaxLevel
			argPos++
		}
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Obtener total
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count categories")
	}

	if filter != nil {
		sortBy := "sort_order"
		sortOrder := "ASC"
		if filter.SortBy != "" {
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

		if filter.Limit > 0 {
			baseQuery += " LIMIT @limit"
			args["limit"] = filter.Limit
		}
		if filter.Offset > 0 {
			baseQuery += " OFFSET @offset"
			args["offset"] = filter.Offset
		}
	}

	rows, err := r.db.Query(ctx, baseQuery, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find categories")
	}
	defer rows.Close()

	var categories []*entities.Category
	for rows.Next() {
		var cat entities.Category
		var description, icon, metaTitle, metaDescription *string
		var parentID *int64

		err = rows.Scan(
			&cat.ID, &cat.PublicID, &cat.Name, &cat.Slug, &description, &icon, &cat.ColorHex,
			&parentID, &cat.Level, &cat.Path, &cat.TotalEvents, &cat.TotalTicketsSold, &cat.TotalRevenue,
			&cat.IsActive, &cat.IsFeatured, &cat.SortOrder, &metaTitle, &metaDescription,
			&cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan category row")
		}

		cat.Description = description
		cat.Icon = icon
		cat.MetaTitle = metaTitle
		cat.MetaDescription = metaDescription
		cat.ParentID = parentID

		categories = append(categories, &cat)
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
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check category existence")
	}
	return exists, nil
}

// ExistsBySlug verifica si existe una categoría con el slug dado
func (r *CategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.categories WHERE slug = $1)`
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check slug existence")
	}
	return exists, nil
}

// Create inserta una nueva categoría
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

	err := r.db.QueryRow(ctx, query,
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

	err := r.db.QueryRow(ctx, query,
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
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ticketing.categories WHERE parent_id = $1`, id).Scan(&childCount)
	if err != nil {
		return r.handleError(err, "failed to check child categories")
	}
	if childCount > 0 {
		return repository.ErrCategoryHasChildren
	}

	cmdTag, err := r.db.Exec(ctx, `DELETE FROM ticketing.categories WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete category")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCategoryNotFound
	}

	return nil
}

// AddEventToCategory asocia un evento a una categoría
func (r *CategoryRepository) AddEventToCategory(ctx context.Context, eventID, categoryID int64, isPrimary bool) error {
	query := `
		INSERT INTO ticketing.event_categories (event_id, category_id, is_primary, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (event_id, category_id) DO UPDATE SET
			is_primary = EXCLUDED.is_primary
	`

	cmdTag, err := r.db.Exec(ctx, query, eventID, categoryID, isPrimary)
	if err != nil {
		return r.handleError(err, "failed to add event to category")
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// IncrementEventCount incrementa el contador de eventos de una categoría
func (r *CategoryRepository) IncrementEventCount(ctx context.Context, categoryID int64) error {
	query := `
		UPDATE ticketing.categories 
		SET total_events = total_events + 1, updated_at = NOW()
		WHERE id = $1
	`
	cmdTag, err := r.db.Exec(ctx, query, categoryID)
	if err != nil {
		return r.handleError(err, "failed to increment event count")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCategoryNotFound
	}
	return nil
}

// DecrementEventCount decrementa el contador de eventos de una categoría
func (r *CategoryRepository) DecrementEventCount(ctx context.Context, categoryID int64) error {
	query := `
		UPDATE ticketing.categories 
		SET total_events = GREATEST(0, total_events - 1), updated_at = NOW()
		WHERE id = $1
	`
	cmdTag, err := r.db.Exec(ctx, query, categoryID)
	if err != nil {
		return r.handleError(err, "failed to decrement event count")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCategoryNotFound
	}
	return nil
}

// UpdateEventStats actualiza las estadísticas de ventas de una categoría
func (r *CategoryRepository) UpdateEventStats(ctx context.Context, categoryID int64, ticketSold int64, revenue float64) error {
	query := `
		UPDATE ticketing.categories 
		SET total_tickets_sold = total_tickets_sold + $1,
			total_revenue = total_revenue + $2,
			updated_at = NOW()
		WHERE id = $3
	`
	cmdTag, err := r.db.Exec(ctx, query, ticketSold, revenue, categoryID)
	if err != nil {
		return r.handleError(err, "failed to update event stats")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCategoryNotFound
	}
	return nil
}

// GetEventCategories obtiene las categorías de un evento
func (r *CategoryRepository) GetEventCategories(ctx context.Context, eventID int64) ([]*entities.Category, error) {
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
		var cat entities.Category
		var description, icon, metaTitle, metaDescription *string
		var parentID *int64

		err = rows.Scan(
			&cat.ID, &cat.PublicID, &cat.Name, &cat.Slug, &description, &icon, &cat.ColorHex,
			&parentID, &cat.Level, &cat.Path, &cat.TotalEvents, &cat.TotalTicketsSold, &cat.TotalRevenue,
			&cat.IsActive, &cat.IsFeatured, &cat.SortOrder, &metaTitle, &metaDescription,
			&cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan category row")
		}

		cat.Description = description
		cat.Icon = icon
		cat.MetaTitle = metaTitle
		cat.MetaDescription = metaDescription
		cat.ParentID = parentID

		categories = append(categories, &cat)
	}

	return categories, nil
}

// RemoveEventFromCategory elimina la asociación entre un evento y una categoría
func (r *CategoryRepository) RemoveEventFromCategory(ctx context.Context, eventID, categoryID int64) error {
	query := `DELETE FROM ticketing.event_categories WHERE event_id = $1 AND category_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, eventID, categoryID)
	if err != nil {
		return r.handleError(err, "failed to remove event from category")
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("association not found")
	}
	return nil
}

// GetPrimaryCategoryForEvent obtiene la categoría principal de un evento
func (r *CategoryRepository) GetPrimaryCategoryForEvent(ctx context.Context, eventID int64) (*entities.Category, error) {
	query := `
		SELECT c.*
		FROM ticketing.categories c
		JOIN ticketing.event_categories ec ON c.id = ec.category_id
		WHERE ec.event_id = $1 AND ec.is_primary = true
		LIMIT 1
	`

	var cat entities.Category
	var description, icon, metaTitle, metaDescription *string
	var parentID *int64

	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&cat.ID, &cat.PublicID, &cat.Name, &cat.Slug, &description, &icon, &cat.ColorHex,
		&parentID, &cat.Level, &cat.Path, &cat.TotalEvents, &cat.TotalTicketsSold, &cat.TotalRevenue,
		&cat.IsActive, &cat.IsFeatured, &cat.SortOrder, &metaTitle, &metaDescription,
		&cat.CreatedAt, &cat.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No hay categoría principal
		}
		return nil, r.handleError(err, "failed to get primary category")
	}

	cat.Description = description
	cat.Icon = icon
	cat.MetaTitle = metaTitle
	cat.MetaDescription = metaDescription
	cat.ParentID = parentID

	return &cat, nil
}

// GetTree obtiene el árbol jerárquico de categorías
func (r *CategoryRepository) GetTree(ctx context.Context, rootID *int64) ([]*repository.CategoryNode, error) {
	var rows pgx.Rows
	var err error

	if rootID == nil {
		rows, err = r.db.Query(ctx, `
			SELECT id, public_uuid, name, slug, description, icon, color_hex,
				parent_id, level, path, total_events, total_tickets_sold, total_revenue,
				is_active, is_featured, sort_order, meta_title, meta_description,
				created_at, updated_at
			FROM ticketing.categories
			ORDER BY parent_id NULLS FIRST, sort_order, name
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			WITH RECURSIVE category_tree AS (
				SELECT id, public_uuid, name, slug, description, icon, color_hex,
					parent_id, level, path, total_events, total_tickets_sold, total_revenue,
					is_active, is_featured, sort_order, meta_title, meta_description,
					created_at, updated_at, 1 as depth
				FROM ticketing.categories
				WHERE id = $1
				UNION ALL
				SELECT c.id, c.public_uuid, c.name, c.slug, c.description, c.icon, c.color_hex,
					c.parent_id, c.level, c.path, c.total_events, c.total_tickets_sold, c.total_revenue,
					c.is_active, c.is_featured, c.sort_order, c.meta_title, c.meta_description,
					c.created_at, c.updated_at, ct.depth + 1
				FROM ticketing.categories c
				INNER JOIN category_tree ct ON c.parent_id = ct.id
			)
			SELECT * FROM category_tree
			ORDER BY depth, sort_order, name
		`, *rootID)
	}

	if err != nil {
		return nil, r.handleError(err, "failed to get category tree")
	}
	defer rows.Close()

	// Mapa temporal para construir el árbol
	categoryMap := make(map[int64]*repository.CategoryNode)
	var roots []*repository.CategoryNode

	for rows.Next() {
		var cat entities.Category
		var description, icon, metaTitle, metaDescription *string
		var parentID *int64

		err = rows.Scan(
			&cat.ID, &cat.PublicID, &cat.Name, &cat.Slug, &description, &icon, &cat.ColorHex,
			&parentID, &cat.Level, &cat.Path, &cat.TotalEvents, &cat.TotalTicketsSold, &cat.TotalRevenue,
			&cat.IsActive, &cat.IsFeatured, &cat.SortOrder, &metaTitle, &metaDescription,
			&cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan category row")
		}

		cat.Description = description
		cat.Icon = icon
		cat.MetaTitle = metaTitle
		cat.MetaDescription = metaDescription
		cat.ParentID = parentID

		node := &repository.CategoryNode{
			Category: &cat,
			Children: []*repository.CategoryNode{},
		}
		categoryMap[cat.ID] = node

		if cat.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := categoryMap[*cat.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots, nil
}

// GetSlugsByEventID obtiene todos los slugs de categorías asociadas a un evento
func (r *CategoryRepository) GetSlugsByEventID(ctx context.Context, eventID string) ([]string, error) {
	query := `
        SELECT DISTINCT c.slug
        FROM ticketing.event_categories ec
        JOIN ticketing.categories c ON ec.category_id = c.id
        JOIN ticketing.events e ON ec.event_id = e.id
        WHERE e.public_uuid = $1
    `

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to get slugs by event ID")
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, r.handleError(err, "failed to scan slug")
		}
		slugs = append(slugs, slug)
	}

	return slugs, nil
}
