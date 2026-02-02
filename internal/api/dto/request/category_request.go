package dto

type CreateCategoryRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100"`
	Slug            string `json:"slug" validate:"required,slug"`
	Description     string `json:"description,omitempty"`
	Icon            string `json:"icon,omitempty" validate:"omitempty,max=50"`
	ColorHex        string `json:"color_hex,omitempty" validate:"omitempty,hexcolor"`
	ParentID        string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
	IsActive        bool   `json:"is_active,omitempty"`
	IsFeatured      bool   `json:"is_featured,omitempty"`
	SortOrder       int    `json:"sort_order,omitempty"`
	MetaTitle       string `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription string `json:"meta_description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name            string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug            string `json:"slug,omitempty" validate:"omitempty,slug"`
	Description     string `json:"description,omitempty"`
	Icon            string `json:"icon,omitempty" validate:"omitempty,max=50"`
	ColorHex        string `json:"color_hex,omitempty" validate:"omitempty,hexcolor"`
	IsActive        *bool  `json:"is_active,omitempty"`
	IsFeatured      *bool  `json:"is_featured,omitempty"`
	SortOrder       int    `json:"sort_order,omitempty"`
	MetaTitle       string `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription string `json:"meta_description,omitempty"`
}

type CategoryFilter struct {
	Search             string `json:"search,omitempty"`
	ParentID           string `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
	IsActive           *bool  `json:"is_active,omitempty"`
	IsFeatured         *bool  `json:"is_featured,omitempty"`
	Level              int    `json:"level,omitempty" validate:"omitempty,min=1"`
	IncludeDescendants bool   `json:"include_descendants,omitempty"`
}
