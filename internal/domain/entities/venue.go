package entities

import (
	"errors"
	"time"
)

// Venue recinto para eventos
// Mapea exactamente la tabla ticketing.venues
type Venue struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`

	Name        string  `json:"name" db:"name"`
	Slug        string  `json:"slug" db:"slug"`
	Description *string `json:"description,omitempty" db:"description"`
	VenueType   string  `json:"venue_type" db:"venue_type"` // indoor, outdoor, stadium, theater, etc.

	AddressLine1 string  `json:"address_line1" db:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty" db:"address_line2"`
	City         string  `json:"city" db:"city"`
	State        *string `json:"state,omitempty" db:"state"`
	PostalCode   *string `json:"postal_code,omitempty" db:"postal_code"`
	Country      string  `json:"country" db:"country"`

	// CORREGIDO: Latitude y Longitude pueden ser NULL en la BD
	Latitude  *float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude *float64 `json:"longitude,omitempty" db:"longitude"`
	// NOTA: geolocation es tipo GEOGRAPHY en PostGIS, se maneja aparte o mediante funciones

	// Capacidades - pueden ser NULL
	Capacity         *int `json:"capacity,omitempty" db:"capacity"`
	SeatingCapacity  *int `json:"seating_capacity,omitempty" db:"seating_capacity"`
	StandingCapacity *int `json:"standing_capacity,omitempty" db:"standing_capacity"`

	// CORREGIDO: facilities es JSONB
	Facilities *[]string `json:"facilities,omitempty" db:"facilities,type:jsonb"`
	// CORREGIDO: accessibility_features es JSONB
	AccessibilityFeatures *[]string `json:"accessibility_features,omitempty" db:"accessibility_features,type:jsonb"`

	ContactEmail *string `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone *string `json:"contact_phone,omitempty" db:"contact_phone"`

	// CORREGIDO: images es JSONB
	Images *[]string `json:"images,omitempty" db:"images,type:jsonb"`

	IsActive bool `json:"is_active" db:"is_active"`

	// CORREGIDO: time.Time en lugar de string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Venue

// HasFullAddress verifica si tiene dirección completa
func (v *Venue) HasFullAddress() bool {
	return v.AddressLine1 != "" &&
		v.City != "" &&
		v.Country != "" &&
		v.PostalCode != nil && *v.PostalCode != ""
}

// GetFullAddress obtiene la dirección completa formateada
func (v *Venue) GetFullAddress() string {
	address := v.AddressLine1
	if v.AddressLine2 != nil && *v.AddressLine2 != "" {
		address += ", " + *v.AddressLine2
	}
	address += ", " + v.City
	if v.State != nil && *v.State != "" {
		address += ", " + *v.State
	}
	if v.PostalCode != nil && *v.PostalCode != "" {
		address += " " + *v.PostalCode
	}
	address += ", " + v.Country
	return address
}

// HasCoordinates verifica si tiene coordenadas
func (v *Venue) HasCoordinates() bool {
	return v.Latitude != nil && v.Longitude != nil
}

// GetCoordinates obtiene las coordenadas como un par (lat, lng)
func (v *Venue) GetCoordinates() (float64, float64) {
	if !v.HasCoordinates() {
		return 0, 0
	}
	return *v.Latitude, *v.Longitude
}

// GetTotalCapacity obtiene la capacidad total
func (v *Venue) GetTotalCapacity() int {
	if v.Capacity != nil {
		return *v.Capacity
	}

	// Si no hay capacidad definida, sumar seating + standing si existen
	total := 0
	if v.SeatingCapacity != nil {
		total += *v.SeatingCapacity
	}
	if v.StandingCapacity != nil {
		total += *v.StandingCapacity
	}
	return total
}

// HasSeating verifica si tiene asientos
func (v *Venue) HasSeating() bool {
	return v.SeatingCapacity != nil && *v.SeatingCapacity > 0
}

// HasStanding verifica si tiene espacio de pie
func (v *Venue) HasStanding() bool {
	return v.StandingCapacity != nil && *v.StandingCapacity > 0
}

// AddFacility añade una instalación
func (v *Venue) AddFacility(facility string) {
	if v.Facilities == nil {
		v.Facilities = &[]string{}
	}

	// Verificar si ya existe
	for _, f := range *v.Facilities {
		if f == facility {
			return
		}
	}

	*v.Facilities = append(*v.Facilities, facility)
	v.UpdatedAt = time.Now()
}

// RemoveFacility elimina una instalación
func (v *Venue) RemoveFacility(facility string) {
	if v.Facilities == nil {
		return
	}

	newFacilities := []string{}
	for _, f := range *v.Facilities {
		if f != facility {
			newFacilities = append(newFacilities, f)
		}
	}

	if len(newFacilities) == 0 {
		v.Facilities = nil
	} else {
		*v.Facilities = newFacilities
	}
	v.UpdatedAt = time.Now()
}

// HasFacility verifica si tiene una instalación específica
func (v *Venue) HasFacility(facility string) bool {
	if v.Facilities == nil {
		return false
	}

	for _, f := range *v.Facilities {
		if f == facility {
			return true
		}
	}
	return false
}

// AddAccessibilityFeature añade una característica de accesibilidad
func (v *Venue) AddAccessibilityFeature(feature string) {
	if v.AccessibilityFeatures == nil {
		v.AccessibilityFeatures = &[]string{}
	}

	// Verificar si ya existe
	for _, f := range *v.AccessibilityFeatures {
		if f == feature {
			return
		}
	}

	*v.AccessibilityFeatures = append(*v.AccessibilityFeatures, feature)
	v.UpdatedAt = time.Now()
}

// RemoveAccessibilityFeature elimina una característica de accesibilidad
func (v *Venue) RemoveAccessibilityFeature(feature string) {
	if v.AccessibilityFeatures == nil {
		return
	}

	newFeatures := []string{}
	for _, f := range *v.AccessibilityFeatures {
		if f != feature {
			newFeatures = append(newFeatures, f)
		}
	}

	if len(newFeatures) == 0 {
		v.AccessibilityFeatures = nil
	} else {
		*v.AccessibilityFeatures = newFeatures
	}
	v.UpdatedAt = time.Now()
}

// HasAccessibilityFeature verifica si tiene una característica de accesibilidad
func (v *Venue) HasAccessibilityFeature(feature string) bool {
	if v.AccessibilityFeatures == nil {
		return false
	}

	for _, f := range *v.AccessibilityFeatures {
		if f == feature {
			return true
		}
	}
	return false
}

// AddImage añade una imagen
func (v *Venue) AddImage(imageURL string) {
	if v.Images == nil {
		v.Images = &[]string{}
	}

	// Verificar si ya existe
	for _, img := range *v.Images {
		if img == imageURL {
			return
		}
	}

	*v.Images = append(*v.Images, imageURL)
	v.UpdatedAt = time.Now()
}

// RemoveImage elimina una imagen
func (v *Venue) RemoveImage(imageURL string) {
	if v.Images == nil {
		return
	}

	newImages := []string{}
	for _, img := range *v.Images {
		if img != imageURL {
			newImages = append(newImages, img)
		}
	}

	if len(newImages) == 0 {
		v.Images = nil
	} else {
		*v.Images = newImages
	}
	v.UpdatedAt = time.Now()
}

// HasImage verifica si tiene una imagen específica
func (v *Venue) HasImage(imageURL string) bool {
	if v.Images == nil {
		return false
	}

	for _, img := range *v.Images {
		if img == imageURL {
			return true
		}
	}
	return false
}

// GetPrimaryImage obtiene la imagen principal (primera)
func (v *Venue) GetPrimaryImage() string {
	if v.Images == nil || len(*v.Images) == 0 {
		return ""
	}
	return (*v.Images)[0]
}

// Validate verifica que el venue sea válido
func (v *Venue) Validate() error {
	if v.Name == "" {
		return errors.New("name is required")
	}
	if v.Slug == "" {
		return errors.New("slug is required")
	}
	if v.AddressLine1 == "" {
		return errors.New("address_line1 is required")
	}
	if v.City == "" {
		return errors.New("city is required")
	}
	if v.Country == "" {
		return errors.New("country is required")
	}
	if v.VenueType == "" {
		return errors.New("venue_type is required")
	}

	// Validar coordenadas si están presentes
	if v.Latitude != nil && (*v.Latitude < -90 || *v.Latitude > 90) {
		return errors.New("latitude must be between -90 and 90")
	}
	if v.Longitude != nil && (*v.Longitude < -180 || *v.Longitude > 180) {
		return errors.New("longitude must be between -180 and 180")
	}

	return nil
}

// IsIndoor verifica si es indoor
func (v *Venue) IsIndoor() bool {
	return v.VenueType == "indoor"
}

// IsOutdoor verifica si es outdoor
func (v *Venue) IsOutdoor() bool {
	return v.VenueType == "outdoor"
}

// IsStadium verifica si es un estadio
func (v *Venue) IsStadium() bool {
	return v.VenueType == "stadium"
}

// IsTheater verifica si es un teatro
func (v *Venue) IsTheater() bool {
	return v.VenueType == "theater"
}

// GetContactInfo obtiene información de contacto
func (v *Venue) GetContactInfo() map[string]string {
	info := make(map[string]string)
	if v.ContactEmail != nil {
		info["email"] = *v.ContactEmail
	}
	if v.ContactPhone != nil {
		info["phone"] = *v.ContactPhone
	}
	return info
}
