package entities

import "time"

type Category struct {
	ID               int64     `json:"id" db:"id"`
	PublicID         string    `json:"public_id" db:"public_uuid"`
	EventID          string    `json:"event_id" db:"event_id"`
	Name             string    `json:"name" db:"name"`
	Slug             string    `json:"slug" db:"slug"`
	Description      *string   `json:"description,omitempty" db:"description"`
	Icon             *string   `json:"icon,omitempty" db:"icon"`
	ColorHex         string    `json:"color_hex" db:"color_hex"`
	ParentID         *int64    `json:"parent_id,omitempty" db:"parent_id"`
	Level            int       `json:"level" db:"level"`
	Path             string    `json:"path" db:"path"`
	Capacity         int       `json:"capacity" db:"capacity"`
	TotalEvents      int       `json:"total_events" db:"total_events"`
	TotalTicketsSold int64     `json:"total_tickets_sold" db:"total_tickets_sold"`
	TotalRevenue     float64   `json:"total_revenue" db:"total_revenue"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	IsFeatured       bool      `json:"is_featured" db:"is_featured"`
	SortOrder        int       `json:"sort_order" db:"sort_order"`
	MetaTitle        *string   `json:"meta_title,omitempty" db:"meta_title"`
	MetaDescription  *string   `json:"meta_description,omitempty" db:"meta_description"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
