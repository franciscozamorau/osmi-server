// internal/api/dto/event/filter.go
package event

// EventFilter para listar eventos
type EventFilter struct {
	Search      string   `json:"search,omitempty"`
	OrganizerID *string  `json:"organizer_id,omitempty" validate:"omitempty,uuid4"`
	CategoryID  *string  `json:"category_id,omitempty" validate:"omitempty,uuid4"`
	VenueID     *string  `json:"venue_id,omitempty" validate:"omitempty,uuid4"`
	EventType   *string  `json:"event_type,omitempty" validate:"omitempty,oneof=in_person virtual hybrid"`
	Status      *string  `json:"status,omitempty"`
	Country     *string  `json:"country,omitempty"`
	City        *string  `json:"city,omitempty"`
	IsFeatured  *bool    `json:"is_featured,omitempty"`
	IsFree      *bool    `json:"is_free,omitempty"`
	DateFrom    *string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo      *string  `json:"date_to,omitempty" validate:"omitempty,date"`
	Tags        []string `json:"tags,omitempty"`
}
