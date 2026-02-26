// internal/api/dto/response/venue_response.go
package response

import "time"

// VenueAddress representa la dirección del venue
type VenueAddress struct {
	AddressLine1 string  `json:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         string  `json:"city"`
	State        *string `json:"state,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
	Country      string  `json:"country"`
	FullAddress  string  `json:"full_address"`
}

// VenueContactInfo representa información de contacto del venue
type VenueContactInfo struct {
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Website       *string `json:"website,omitempty"`
	ContactPerson *string `json:"contact_person,omitempty"`
}

// VenueImage representa una imagen del venue
type VenueImage struct {
	URL       string  `json:"url"`
	AltText   *string `json:"alt_text,omitempty"`
	IsPrimary bool    `json:"is_primary"`
	Type      string  `json:"type"` // interior, exterior, map, seating, etc.
}

// TimeRange representa un rango de tiempo (para horarios)
type TimeRange struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

// SpecialHour representa horario especial (feriados, eventos, etc.)
type SpecialHour struct {
	Date     string  `json:"date"`
	Open     *string `json:"open,omitempty"`
	Close    *string `json:"close,omitempty"`
	IsClosed bool    `json:"is_closed"`
	Reason   *string `json:"reason,omitempty"`
}

// OperatingHours representa los horarios de operación del venue
type OperatingHours struct {
	Monday       *TimeRange    `json:"monday,omitempty"`
	Tuesday      *TimeRange    `json:"tuesday,omitempty"`
	Wednesday    *TimeRange    `json:"wednesday,omitempty"`
	Thursday     *TimeRange    `json:"thursday,omitempty"`
	Friday       *TimeRange    `json:"friday,omitempty"`
	Saturday     *TimeRange    `json:"saturday,omitempty"`
	Sunday       *TimeRange    `json:"sunday,omitempty"`
	SpecialHours []SpecialHour `json:"special_hours,omitempty"`
	Notes        *string       `json:"notes,omitempty"`
}

// ParkingInfo representa información de estacionamiento
type ParkingInfo struct {
	Available  bool    `json:"available"`
	Capacity   *int    `json:"capacity,omitempty"`
	Cost       *string `json:"cost,omitempty"`
	Type       *string `json:"type,omitempty"` // underground, surface, valet, etc.
	Distance   *string `json:"distance,omitempty"`
	Accessible bool    `json:"accessible"`
}

// TransportOption representa una opción de transporte público
type TransportOption struct {
	Type        string  `json:"type"` // metro, bus, train, etc.
	Line        *string `json:"line,omitempty"`
	Station     string  `json:"station"`
	Distance    string  `json:"distance"`
	WalkingTime string  `json:"walking_time"`
}

// DrivingInfo representa información para conductores
type DrivingInfo struct {
	DirectionsURL *string `json:"directions_url,omitempty"`
	ParkingNote   *string `json:"parking_note,omitempty"`
	DropOffZone   bool    `json:"drop_off_zone"`
}

// BikingInfo representa información para ciclistas
type BikingInfo struct {
	BikeRacks bool `json:"bike_racks"`
	BikeShare bool `json:"bike_share"`
	Lockers   bool `json:"lockers"`
	Showers   bool `json:"showers"`
}

// RideshareInfo representa información para servicios de rideshare
type RideshareInfo struct {
	PickupZone   bool    `json:"pickup_zone"`
	DropoffZone  bool    `json:"dropoff_zone"`
	Instructions *string `json:"instructions,omitempty"`
}

// TransportationInfo representa información completa de transporte
type TransportationInfo struct {
	PublicTransport []TransportOption `json:"public_transport,omitempty"`
	Driving         *DrivingInfo      `json:"driving,omitempty"`
	Biking          *BikingInfo       `json:"biking,omitempty"`
	Rideshare       *RideshareInfo    `json:"rideshare,omitempty"`
}

// VenueRule representa una regla del venue
type VenueRule struct {
	Type        string `json:"type"` // age, dress_code, prohibited_items, etc.
	Title       string `json:"title"`
	Description string `json:"description"`
	Importance  string `json:"importance"` // high, medium, low
}

// VenueFilter representa los filtros aplicados en la consulta de venues
type VenueFilter struct {
	Search        *string  `json:"search,omitempty"`
	City          *string  `json:"city,omitempty"`
	Country       *string  `json:"country,omitempty"`
	VenueType     *string  `json:"venue_type,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
	IsVerified    *bool    `json:"is_verified,omitempty"`
	MinCapacity   *int     `json:"min_capacity,omitempty"`
	MaxCapacity   *int     `json:"max_capacity,omitempty"`
	HasEvents     *bool    `json:"has_events,omitempty"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	RadiusKm      *float64 `json:"radius_km,omitempty"`
	Accessibility *bool    `json:"accessibility,omitempty"`
	Parking       *bool    `json:"parking,omitempty"`
}

// VenueCityStats representa estadísticas por ciudad
type VenueCityStats struct {
	City          string `json:"city"`
	Country       string `json:"country"`
	VenueCount    int    `json:"venue_count"`
	EventCount    int    `json:"event_count"`
	TotalCapacity int    `json:"total_capacity"`
}

// VenueTypeStats representa estadísticas por tipo de venue
type VenueTypeStats struct {
	Type        string  `json:"type"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	AvgCapacity float64 `json:"avg_capacity"`
}

// VenueStatsResponse representa estadísticas globales de venues
type VenueStatsResponse struct {
	TotalVenues      int              `json:"total_venues"`
	ActiveVenues     int              `json:"active_venues"`
	VerifiedVenues   int              `json:"verified_venues"`
	TotalCapacity    int              `json:"total_capacity"`
	AvgCapacity      float64          `json:"avg_capacity"`
	VenuesWithEvents int              `json:"venues_with_events"`
	TopCities        []VenueCityStats `json:"top_cities"`
	VenueTypes       []VenueTypeStats `json:"venue_types"`
	GrowthRate       float64          `json:"growth_rate"`
}

// VenueResponse representa la respuesta completa de un venue
type VenueResponse struct {
	ID                    string              `json:"id"`
	Name                  string              `json:"name"`
	Slug                  string              `json:"slug"`
	Description           *string             `json:"description,omitempty"`
	VenueType             string              `json:"venue_type"`
	Address               VenueAddress        `json:"address"`
	GeoLocation           *GeoLocation        `json:"geo_location,omitempty"`
	Capacity              *int                `json:"capacity,omitempty"`
	SeatingCapacity       *int                `json:"seating_capacity,omitempty"`
	StandingCapacity      *int                `json:"standing_capacity,omitempty"`
	Facilities            []string            `json:"facilities"`
	AccessibilityFeatures []string            `json:"accessibility_features"`
	ContactInfo           *VenueContactInfo   `json:"contact_info,omitempty"`
	Images                []VenueImage        `json:"images"`
	IsActive              bool                `json:"is_active"`
	IsVerified            bool                `json:"is_verified"`
	VerificationStatus    string              `json:"verification_status"`
	TotalEvents           int                 `json:"total_events"`
	UpcomingEvents        []EventInfo         `json:"upcoming_events,omitempty"`
	Rating                *float64            `json:"rating,omitempty"`
	ReviewCount           int                 `json:"review_count"`
	OperatingHours        *OperatingHours     `json:"operating_hours,omitempty"`
	ParkingInfo           *ParkingInfo        `json:"parking_info,omitempty"`
	TransportationInfo    *TransportationInfo `json:"transportation_info,omitempty"`
	Rules                 []VenueRule         `json:"rules,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

// VenueListResponse representa una lista paginada de venues
type VenueListResponse struct {
	Venues     []VenueResponse `json:"venues"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
	HasNext    bool            `json:"has_next"`
	HasPrev    bool            `json:"has_prev"`
	Filters    *VenueFilter    `json:"filters,omitempty"`
	MapBounds  *MapBounds      `json:"map_bounds,omitempty"`
}
