package request

import (
	"regexp"
)

// CreateCategoryRequest representa la solicitud para crear una categoría
type CreateCategoryRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100"`
	Slug            string `json:"slug" validate:"required,slug"`
	Description     string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Icon            string `json:"icon,omitempty" validate:"omitempty"`
	ColorHex        string `json:"color_hex,omitempty" validate:"omitempty,hexcolor"`
	ParentID        *int64 `json:"parent_id,omitempty" validate:"omitempty,min=1"`
	IsActive        *bool  `json:"is_active,omitempty"`
	IsFeatured      *bool  `json:"is_featured,omitempty"`
	SortOrder       *int   `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	MetaTitle       string `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription string `json:"meta_description,omitempty" validate:"omitempty,max=500"`
}

// SetDefaults establece valores por defecto para CreateCategoryRequest
func (r *CreateCategoryRequest) SetDefaults() {
	if r.IsActive == nil {
		defaultActive := true
		r.IsActive = &defaultActive
	}
	if r.IsFeatured == nil {
		defaultFeatured := false
		r.IsFeatured = &defaultFeatured
	}
	if r.SortOrder == nil {
		defaultSortOrder := 0
		r.SortOrder = &defaultSortOrder
	}
}

// UpdateCategoryRequest representa la solicitud para actualizar una categoría
type UpdateCategoryRequest struct {
	Name            *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug            *string `json:"slug,omitempty" validate:"omitempty,slug"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Icon            *string `json:"icon,omitempty" validate:"omitempty"`
	ColorHex        *string `json:"color_hex,omitempty" validate:"omitempty,hexcolor"`
	ParentID        *int64  `json:"parent_id,omitempty" validate:"omitempty,min=1"`
	IsActive        *bool   `json:"is_active,omitempty"`
	IsFeatured      *bool   `json:"is_featured,omitempty"`
	SortOrder       *int    `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	MetaTitle       *string `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription *string `json:"meta_description,omitempty" validate:"omitempty,max=500"`
}

// IsEmpty verifica si no hay campos para actualizar
func (r *UpdateCategoryRequest) IsEmpty() bool {
	return r.Name == nil && r.Slug == nil && r.Description == nil &&
		r.Icon == nil && r.ColorHex == nil && r.ParentID == nil &&
		r.IsActive == nil && r.IsFeatured == nil && r.SortOrder == nil &&
		r.MetaTitle == nil && r.MetaDescription == nil
}

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
	Page     int `json:"page,omitempty" default:"1"`
	PageSize int `json:"page_size,omitempty" default:"20"`

	// Ordenamiento
	SortBy    string `json:"sort_by,omitempty" default:"sort_order"`
	SortOrder string `json:"sort_order,omitempty" default:"asc"`
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

// Validación personalizada para slug
func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}
	// slug debe ser: letras minúsculas, números y guiones, sin guiones al inicio o final
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return slugRegex.MatchString(slug)
}
