package response

import "time"

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

type VenueAddress struct {
	AddressLine1 string  `json:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         string  `json:"city"`
	State        *string `json:"state,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
	Country      string  `json:"country"`
	FullAddress  string  `json:"full_address"`
}

type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type VenueContactInfo struct {
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Website       *string `json:"website,omitempty"`
	ContactPerson *string `json:"contact_person,omitempty"`
}

type VenueImage struct {
	URL       string  `json:"url"`
	AltText   *string `json:"alt_text,omitempty"`
	IsPrimary bool    `json:"is_primary"`
	Type      string  `json:"type"` // interior, exterior, map, seating, etc.
}

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

type TimeRange struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type SpecialHour struct {
	Date     string  `json:"date"`
	Open     *string `json:"open,omitempty"`
	Close    *string `json:"close,omitempty"`
	IsClosed bool    `json:"is_closed"`
	Reason   *string `json:"reason,omitempty"`
}

type ParkingInfo struct {
	Available  bool    `json:"available"`
	Capacity   *int    `json:"capacity,omitempty"`
	Cost       *string `json:"cost,omitempty"`
	Type       *string `json:"type,omitempty"` // underground, surface, valet, etc.
	Distance   *string `json:"distance,omitempty"`
	Accessible bool    `json:"accessible"`
}

type TransportationInfo struct {
	PublicTransport []TransportOption `json:"public_transport,omitempty"`
	Driving         *DrivingInfo      `json:"driving,omitempty"`
	Biking          *BikingInfo       `json:"biking,omitempty"`
	Rideshare       *RideshareInfo    `json:"rideshare,omitempty"`
}

type TransportOption struct {
	Type        string  `json:"type"` // metro, bus, train, etc.
	Line        *string `json:"line,omitempty"`
	Station     string  `json:"station"`
	Distance    string  `json:"distance"`
	WalkingTime string  `json:"walking_time"`
}

type DrivingInfo struct {
	DirectionsURL *string `json:"directions_url,omitempty"`
	ParkingNote   *string `json:"parking_note,omitempty"`
	DropOffZone   bool    `json:"drop_off_zone"`
}

type BikingInfo struct {
	BikeRacks bool `json:"bike_racks"`
	BikeShare bool `json:"bike_share"`
	Lockers   bool `json:"lockers"`
	Showers   bool `json:"showers"`
}

type RideshareInfo struct {
	PickupZone   bool    `json:"pickup_zone"`
	DropoffZone  bool    `json:"dropoff_zone"`
	Instructions *string `json:"instructions,omitempty"`
}

type VenueRule struct {
	Type        string `json:"type"` // age, dress_code, prohibited_items, etc.
	Title       string `json:"title"`
	Description string `json:"description"`
	Importance  string `json:"importance"` // high, medium, low
}

type VenueInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	VenueType  string   `json:"venue_type"`
	Capacity   *int     `json:"capacity,omitempty"`
	IsVerified bool     `json:"is_verified"`
	Rating     *float64 `json:"rating,omitempty"`
}

type VenueListResponse struct {
	Venues     []VenueResponse `json:"venues"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
	HasNext    bool            `json:"has_next"`
	HasPrev    bool            `json:"has_prev"`
	Filters    VenueFilter     `json:"filters,omitempty"`
	MapBounds  *MapBounds      `json:"map_bounds,omitempty"`
}

type MapBounds struct {
	NorthEast GeoLocation `json:"north_east"`
	SouthWest GeoLocation `json:"south_west"`
}

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

type VenueCityStats struct {
	City          string `json:"city"`
	Country       string `json:"country"`
	VenueCount    int    `json:"venue_count"`
	EventCount    int    `json:"event_count"`
	TotalCapacity int    `json:"total_capacity"`
}

type VenueTypeStats struct {
	Type        string  `json:"type"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	AvgCapacity float64 `json:"avg_capacity"`
}
