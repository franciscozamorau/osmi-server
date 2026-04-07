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

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrCategoryNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if strings.Contains(pgErr.ConstraintName, "unique_event_category_name") {
				return repository.ErrCategoryDuplicateName
			}
			if strings.Contains(pgErr.ConstraintName, "unique_event_category_slug") {
				return repository.ErrCategoryDuplicateSlug
			}
		}
	}

	return fmt.Errorf("%s: %w", context, err)
}

func (r *CategoryRepository) Find(ctx context.Context, filter *repository.CategoryFilter) ([]*entities.Category, int64, error) {
	baseQuery := `
        SELECT 
            id, public_uuid, event_id, name, slug, description, icon, color_hex,
            parent_id, level, path, capacity,
            total_events, total_tickets_sold, total_revenue,
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

		if filter.EventID != nil {
			conditions = append(conditions, fmt.Sprintf("event_id = @event_%d", argPos))
			args[fmt.Sprintf("event_%d", argPos)] = *filter.EventID
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
			&cat.ID, &cat.PublicID, &cat.EventID, &cat.Name, &cat.Slug,
			&description, &icon, &cat.ColorHex,
			&parentID, &cat.Level, &cat.Path, &cat.Capacity,
			&cat.TotalEvents, &cat.TotalTicketsSold, &cat.TotalRevenue,
			&cat.IsActive, &cat.IsFeatured, &cat.SortOrder,
			&metaTitle, &metaDescription,
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

func (r *CategoryRepository) GetByEventID(ctx context.Context, eventID string, isActive *bool) ([]*entities.Category, error) {
	filter := &repository.CategoryFilter{
		EventID:   &eventID,
		IsActive:  isActive,
		SortBy:    "sort_order",
		SortOrder: "ASC",
	}
	categories, _, err := r.Find(ctx, filter)
	return categories, err
}

func (r *CategoryRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.categories WHERE id = $1)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check category existence")
	}
	return exists, nil
}

func (r *CategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.categories WHERE slug = $1)`
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check slug existence")
	}
	return exists, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category *entities.Category) error {
	query := `
        INSERT INTO ticketing.categories (
            public_uuid, event_id, name, slug, description, icon, color_hex,
            parent_id, level, path, capacity,
            total_events, total_tickets_sold, total_revenue,
            is_active, is_featured, sort_order, meta_title, meta_description,
            created_at, updated_at
        ) VALUES (
            gen_random_uuid(), 
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10,
            $11, $12, $13,
            $14, $15, $16, $17, $18,
            NOW(), NOW()
        )
        RETURNING id, public_uuid, created_at, updated_at
    `

	err := r.db.QueryRow(ctx, query,
		category.EventID,
		category.Name, category.Slug, category.Description, category.Icon, category.ColorHex,
		category.ParentID, category.Level, category.Path, category.Capacity,
		category.TotalEvents, category.TotalTicketsSold, category.TotalRevenue,
		category.IsActive, category.IsFeatured, category.SortOrder,
		category.MetaTitle, category.MetaDescription,
	).Scan(&category.ID, &category.PublicID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create category")
	}
	return nil
}

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
            capacity = $9,
            is_active = $10,
            is_featured = $11,
            sort_order = $12,
            meta_title = $13,
            meta_description = $14,
            updated_at = NOW()
        WHERE id = $15
        RETURNING updated_at
    `

	err := r.db.QueryRow(ctx, query,
		category.Name, category.Slug, category.Description, category.Icon, category.ColorHex,
		category.ParentID, category.Level, category.Path, category.Capacity,
		category.IsActive, category.IsFeatured, category.SortOrder,
		category.MetaTitle, category.MetaDescription,
		category.ID,
	).Scan(&category.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update category")
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
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

func (r *CategoryRepository) IncrementEventCount(ctx context.Context, categoryID int64) error {
	query := `UPDATE ticketing.categories SET total_events = total_events + 1, updated_at = NOW() WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, categoryID)
	if err != nil {
		return r.handleError(err, "failed to increment event count")
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) DecrementEventCount(ctx context.Context, categoryID int64) error {
	query := `UPDATE ticketing.categories SET total_events = GREATEST(0, total_events - 1), updated_at = NOW() WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, categoryID)
	if err != nil {
		return r.handleError(err, "failed to decrement event count")
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCategoryNotFound
	}
	return nil
}

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

func (r *CategoryRepository) GetTree(ctx context.Context, rootID *int64) ([]*repository.CategoryNode, error) {
	var rows pgx.Rows
	var err error

	if rootID == nil {
		rows, err = r.db.Query(ctx, `
			SELECT id, public_uuid, event_id, name, slug, description, icon, color_hex,
				parent_id, level, path, capacity,
				total_events, total_tickets_sold, total_revenue,
				is_active, is_featured, sort_order, meta_title, meta_description,
				created_at, updated_at
			FROM ticketing.categories
			ORDER BY parent_id NULLS FIRST, sort_order, name
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			WITH RECURSIVE category_tree AS (
				SELECT id, public_uuid, event_id, name, slug, description, icon, color_hex,
					parent_id, level, path, capacity,
					total_events, total_tickets_sold, total_revenue,
					is_active, is_featured, sort_order, meta_title, meta_description,
					created_at, updated_at, 1 as depth
				FROM ticketing.categories
				WHERE id = $1
				UNION ALL
				SELECT c.id, c.public_uuid, c.event_id, c.name, c.slug, c.description, c.icon, c.color_hex,
					c.parent_id, c.level, c.path, c.capacity,
					c.total_events, c.total_tickets_sold, c.total_revenue,
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

	categoryMap := make(map[int64]*repository.CategoryNode)
	var roots []*repository.CategoryNode

	for rows.Next() {
		var cat entities.Category
		var description, icon, metaTitle, metaDescription *string
		var parentID *int64

		err = rows.Scan(
			&cat.ID, &cat.PublicID, &cat.EventID, &cat.Name, &cat.Slug,
			&description, &icon, &cat.ColorHex,
			&parentID, &cat.Level, &cat.Path, &cat.Capacity,
			&cat.TotalEvents, &cat.TotalTicketsSold, &cat.TotalRevenue,
			&cat.IsActive, &cat.IsFeatured, &cat.SortOrder,
			&metaTitle, &metaDescription,
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
