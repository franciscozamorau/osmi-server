package category

import (
	"regexp"
)

// CreateCategoryRequest representa la solicitud para crear una categoría
type CreateCategoryRequest struct {
	// 🔥 NUEVO CAMPO OBLIGATORIO
	EventID         string `json:"event_id" validate:"required,uuid"`
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

func (r *UpdateCategoryRequest) IsEmpty() bool {
	return r.Name == nil && r.Slug == nil && r.Description == nil &&
		r.Icon == nil && r.ColorHex == nil && r.ParentID == nil &&
		r.IsActive == nil && r.IsFeatured == nil && r.SortOrder == nil &&
		r.MetaTitle == nil && r.MetaDescription == nil
}

func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return slugRegex.MatchString(slug)
}
