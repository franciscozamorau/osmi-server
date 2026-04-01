// internal/api/dto/organizer/filter.go
package organizer

// OrganizerFilter representa los filtros aplicados en la consulta de organizadores
type OrganizerFilter struct {
	Search             string  `json:"search,omitempty"`
	Country            string  `json:"country,omitempty"`
	IsVerified         *bool   `json:"is_verified,omitempty"`
	IsActive           *bool   `json:"is_active,omitempty"`
	VerificationStatus string  `json:"verification_status,omitempty"`
	MinRating          float64 `json:"min_rating,omitempty" validate:"omitempty,min=0,max=5"`
	MinEvents          int     `json:"min_events,omitempty" validate:"omitempty,min=0"`
}
