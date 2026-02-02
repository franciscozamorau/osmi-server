package entities

import "time"

// Category representa una categoría de eventos
type Category struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`

	Name        string  `json:"name" db:"name"`
	Slug        string  `json:"slug" db:"slug"`
	Description *string `json:"description,omitempty" db:"description"`
	Icon        *string `json:"icon,omitempty" db:"icon"`
	ColorHex    string  `json:"color_hex" db:"color_hex"`

	ParentID *int64 `json:"parent_id,omitempty" db:"parent_id"`
	Level    int32  `json:"level" db:"level"`
	Path     string `json:"path" db:"path"`

	TotalEvents      int32   `json:"total_events" db:"total_events"`
	TotalTicketsSold int64   `json:"total_tickets_sold" db:"total_tickets_sold"`
	TotalRevenue     float64 `json:"total_revenue" db:"total_revenue"`

	IsActive   bool  `json:"is_active" db:"is_active"`
	IsFeatured bool  `json:"is_featured" db:"is_featured"`
	SortOrder  int32 `json:"sort_order" db:"sort_order"`

	MetaTitle       *string `json:"meta_title,omitempty" db:"meta_title"`
	MetaDescription *string `json:"meta_description,omitempty" db:"meta_description"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EventCategory representa la relación muchos-a-muchos entre eventos y categorías
type EventCategory struct {
	ID         int64     `json:"id" db:"id"`
	EventID    int64     `json:"event_id" db:"event_id"`
	CategoryID int64     `json:"category_id" db:"category_id"`
	IsPrimary  bool      `json:"is_primary" db:"is_primary"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
