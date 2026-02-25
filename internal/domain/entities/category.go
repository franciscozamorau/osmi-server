package entities

import "time"

// Category representa una categoría de eventos
// Mapea exactamente la tabla ticketing.categories
type Category struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`

	Name        string  `json:"name" db:"name"`
	Slug        string  `json:"slug" db:"slug"`
	Description *string `json:"description,omitempty" db:"description"`
	Icon        *string `json:"icon,omitempty" db:"icon"`
	ColorHex    string  `json:"color_hex" db:"color_hex"`

	ParentID *int64 `json:"parent_id,omitempty" db:"parent_id"`
	// Level es INTEGER en la BD, representado como int en Go
	Level int `json:"level" db:"level"`
	// Path es VARCHAR(500) con default ''
	Path string `json:"path" db:"path"`

	TotalEvents      int     `json:"total_events" db:"total_events"`             // INTEGER DEFAULT 0
	TotalTicketsSold int64   `json:"total_tickets_sold" db:"total_tickets_sold"` // BIGINT DEFAULT 0
	TotalRevenue     float64 `json:"total_revenue" db:"total_revenue"`           // DECIMAL(15,2) DEFAULT 0

	IsActive   bool `json:"is_active" db:"is_active"`     // BOOLEAN DEFAULT true
	IsFeatured bool `json:"is_featured" db:"is_featured"` // BOOLEAN DEFAULT false
	SortOrder  int  `json:"sort_order" db:"sort_order"`   // INTEGER DEFAULT 0

	MetaTitle       *string `json:"meta_title,omitempty" db:"meta_title"`             // VARCHAR(255)
	MetaDescription *string `json:"meta_description,omitempty" db:"meta_description"` // TEXT

	CreatedAt time.Time `json:"created_at" db:"created_at"` // TIMESTAMPTZ DEFAULT NOW()
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // TIMESTAMPTZ DEFAULT NOW()
}

// EventCategory representa la relación muchos-a-muchos entre eventos y categorías
// Mapea exactamente la tabla ticketing.event_categories
type EventCategory struct {
	ID         int64     `json:"id" db:"id"`
	EventID    int64     `json:"event_id" db:"event_id"`
	CategoryID int64     `json:"category_id" db:"category_id"`
	IsPrimary  bool      `json:"is_primary" db:"is_primary"` // BOOLEAN DEFAULT false
	CreatedAt  time.Time `json:"created_at" db:"created_at"` // TIMESTAMPTZ DEFAULT NOW()
}
