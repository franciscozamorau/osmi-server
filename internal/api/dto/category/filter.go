// internal/api/dto/category/filter.go
package category

// CategoryFilter representa los filtros para listar categorías
type CategoryFilter struct {
	Search             string `json:"search,omitempty"`
	ParentID           *int64 `json:"parent_id,omitempty"`
	IsActive           *bool  `json:"is_active,omitempty"`
	IsFeatured         *bool  `json:"is_featured,omitempty"`
	MinLevel           *int   `json:"min_level,omitempty" validate:"omitempty,min=1"`
	MaxLevel           *int   `json:"max_level,omitempty" validate:"omitempty,min=1"`
	IncludeDescendants bool   `json:"include_descendants,omitempty"`
	IncludeParent      bool   `json:"include_parent,omitempty"`

	// Paginación
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	// Ordenamiento
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// SetDefaults establece valores por defecto para CategoryFilter
func (f *CategoryFilter) SetDefaults() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.SortBy == "" {
		f.SortBy = "sort_order"
	}
	if f.SortOrder == "" {
		f.SortOrder = "asc"
	}
}

// GetOffset calcula el offset para la base de datos
func (f *CategoryFilter) GetOffset() int {
	return (f.Page - 1) * f.PageSize
}

// GetLimit devuelve el límite para la consulta
func (f *CategoryFilter) GetLimit() int {
	return f.PageSize
}
