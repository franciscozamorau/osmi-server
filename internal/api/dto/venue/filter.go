// internal/api/dto/venue/filter.go
package venue

// VenueFilter representa los filtros aplicados en la consulta de venues
type VenueFilter struct {
	Name      *string
	State     *string
	City      *string
	Country   *string
	VenueType *string

	Search string

	IsActive *bool

	MinCapacity *int
	MaxCapacity *int

	HasSeating          *bool
	HasWheelchairAccess *bool

	Facilities []string

	Latitude  float64
	Longitude float64
	RadiusKm  float64
}
