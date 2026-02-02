package dto

type CreateVenueRequest struct {
	Name                  string   `json:"name" validate:"required,min=2,max=255"`
	Slug                  string   `json:"slug" validate:"required,slug"`
	Description           string   `json:"description,omitempty"`
	VenueType             string   `json:"venue_type" validate:"required,oneof=indoor outdoor stadium theater club arena conference_hall"`
	AddressLine1          string   `json:"address_line1" validate:"required,max=255"`
	AddressLine2          string   `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City                  string   `json:"city" validate:"required,max=100"`
	State                 string   `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode            string   `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	Country               string   `json:"country" validate:"required,country_code"`
	Latitude              float64  `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude             float64  `json:"longitude,omitempty" validate:"omitempty,longitude"`
	Capacity              int      `json:"capacity,omitempty" validate:"omitempty,min=1"`
	SeatingCapacity       int      `json:"seating_capacity,omitempty" validate:"omitempty,min=0"`
	StandingCapacity      int      `json:"standing_capacity,omitempty" validate:"omitempty,min=0"`
	Facilities            []string `json:"facilities,omitempty"`
	AccessibilityFeatures []string `json:"accessibility_features,omitempty"`
	ContactEmail          string   `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone          string   `json:"contact_phone,omitempty" validate:"omitempty,phone"`
	Images                []string `json:"images,omitempty"`
}

type UpdateVenueRequest struct {
	Name                  string   `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description           string   `json:"description,omitempty"`
	VenueType             string   `json:"venue_type,omitempty" validate:"omitempty,oneof=indoor outdoor stadium theater club arena conference_hall"`
	AddressLine1          string   `json:"address_line1,omitempty" validate:"omitempty,max=255"`
	AddressLine2          string   `json:"address_line2,omitempty" validate:"omitempty,max=255"`
	City                  string   `json:"city,omitempty" validate:"omitempty,max=100"`
	State                 string   `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode            string   `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	Country               string   `json:"country,omitempty" validate:"omitempty,country_code"`
	Latitude              float64  `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude             float64  `json:"longitude,omitempty" validate:"omitempty,longitude"`
	Capacity              int      `json:"capacity,omitempty" validate:"omitempty,min=1"`
	SeatingCapacity       int      `json:"seating_capacity,omitempty" validate:"omitempty,min=0"`
	StandingCapacity      int      `json:"standing_capacity,omitempty" validate:"omitempty,min=0"`
	Facilities            []string `json:"facilities,omitempty"`
	AccessibilityFeatures []string `json:"accessibility_features,omitempty"`
	ContactEmail          string   `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone          string   `json:"contact_phone,omitempty" validate:"omitempty,phone"`
	IsActive              *bool    `json:"is_active,omitempty"`
	Images                []string `json:"images,omitempty"`
}

type VenueFilter struct {
	Search              string   `json:"search,omitempty"`
	Country             string   `json:"country,omitempty"`
	City                string   `json:"city,omitempty"`
	VenueType           string   `json:"venue_type,omitempty"`
	IsActive            *bool    `json:"is_active,omitempty"`
	MinCapacity         int      `json:"min_capacity,omitempty" validate:"omitempty,min=1"`
	MaxCapacity         int      `json:"max_capacity,omitempty" validate:"omitempty,min=1"`
	HasSeating          *bool    `json:"has_seating,omitempty"`
	HasWheelchairAccess *bool    `json:"has_wheelchair_access,omitempty"`
	Facilities          []string `json:"facilities,omitempty"`
	Latitude            float64  `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude           float64  `json:"longitude,omitempty" validate:"omitempty,longitude"`
	RadiusKm            float64  `json:"radius_km,omitempty" validate:"omitempty,min=0.1,max=100"`
}
