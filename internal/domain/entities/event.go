package entities

import (
	"encoding/json"
	"time"
)

// Event representa un evento en el sistema de ticketing
// Mapea exactamente la tabla ticketing.events
type Event struct {
	ID                int64  `json:"id" db:"id"`
	PublicID          string `json:"public_id" db:"public_uuid"`
	OrganizerID       int64  `json:"organizer_id" db:"organizer_id"`
	PrimaryCategoryID *int64 `json:"primary_category_id,omitempty" db:"primary_category_id"`
	VenueID           *int64 `json:"venue_id,omitempty" db:"venue_id"`

	Name             string  `json:"name" db:"name"`
	Slug             string  `json:"slug" db:"slug"`
	ShortDescription *string `json:"short_description,omitempty" db:"short_description"`
	Description      *string `json:"description,omitempty" db:"description"`
	EventType        string  `json:"event_type" db:"event_type"`

	CoverImageURL  *string `json:"cover_image_url,omitempty" db:"cover_image_url"`
	BannerImageURL *string `json:"banner_image_url,omitempty" db:"banner_image_url"`
	// CORREGIDO: gallery_images es JSONB, no string
	GalleryImages *[]string `json:"gallery_images,omitempty" db:"gallery_images,type:jsonb"`

	Timezone     string     `json:"timezone" db:"timezone"`
	StartsAt     time.Time  `json:"starts_at" db:"starts_at"`
	EndsAt       time.Time  `json:"ends_at" db:"ends_at"`
	DoorsOpenAt  *time.Time `json:"doors_open_at,omitempty" db:"doors_open_at"`
	DoorsCloseAt *time.Time `json:"doors_close_at,omitempty" db:"doors_close_at"`

	VenueName   *string `json:"venue_name,omitempty" db:"venue_name"`
	AddressFull *string `json:"address_full,omitempty" db:"address_full"`
	City        *string `json:"city,omitempty" db:"city"`
	State       *string `json:"state,omitempty" db:"state"`
	Country     *string `json:"country,omitempty" db:"country"`

	Status     string `json:"status" db:"status"`
	Visibility string `json:"visibility" db:"visibility"`
	IsFeatured bool   `json:"is_featured" db:"is_featured"`
	IsFree     bool   `json:"is_free" db:"is_free"`

	// CORREGIDO: Usamos int en lugar de int32 para INTEGER en PostgreSQL
	MaxAttendees *int `json:"max_attendees,omitempty" db:"max_attendees"`
	MinAttendees int  `json:"min_attendees" db:"min_attendees"`

	// CORREGIDO: tags es JSONB, no string
	Tags           *[]string `json:"tags,omitempty" db:"tags,type:jsonb"`
	AgeRestriction *int      `json:"age_restriction,omitempty" db:"age_restriction"`

	RequiresApproval    bool `json:"requires_approval" db:"requires_approval"`
	AllowReservations   bool `json:"allow_reservations" db:"allow_reservations"`
	ReservationDuration int  `json:"reservation_duration" db:"reservation_duration_minutes"`

	ViewCount     int `json:"view_count" db:"view_count"`
	FavoriteCount int `json:"favorite_count" db:"favorite_count"`
	ShareCount    int `json:"share_count" db:"share_count"`

	MetaTitle       *string `json:"meta_title,omitempty" db:"meta_title"`
	MetaDescription *string `json:"meta_description,omitempty" db:"meta_description"`

	// CORREGIDO: settings es JSONB con estructura definida, no string
	Settings *EventSettings `json:"settings,omitempty" db:"settings,type:jsonb"`

	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// EventSettings representa la configuración JSONB del evento
type EventSettings struct {
	AllowCancellations        bool   `json:"allow_cancellations"`
	CancellationDeadlineHours int    `json:"cancellation_deadline_hours"`
	AllowTransfers            bool   `json:"allow_transfers"`
	RequireID                 bool   `json:"require_id"`
	CheckinMethod             string `json:"checkin_method"` // qr_code, manual, rfid
}

// Métodos de utilidad para Event

// IsPublished verifica si el evento está publicado
func (e *Event) IsPublished() bool {
	return e.Status == "published" || e.Status == "live"
}

// IsLive verifica si el evento está en vivo
func (e *Event) IsLive() bool {
	return e.Status == "live"
}

// IsCancelled verifica si el evento está cancelado
func (e *Event) IsCancelled() bool {
	return e.Status == "cancelled"
}

// IsSoldOut verifica si el evento está agotado
func (e *Event) IsSoldOut() bool {
	return e.Status == "sold_out"
}

// IsDraft verifica si el evento está en borrador
func (e *Event) IsDraft() bool {
	return e.Status == "draft"
}

// IsCompleted verifica si el evento ha finalizado
func (e *Event) IsCompleted() bool {
	return e.Status == "completed" || time.Now().After(e.EndsAt)
}

// IsArchived verifica si el evento está archivado
func (e *Event) IsArchived() bool {
	return e.Status == "archived"
}

// IsUpcoming verifica si el evento es futuro
func (e *Event) IsUpcoming() bool {
	return time.Now().Before(e.StartsAt)
}

// IsOngoing verifica si el evento está ocurriendo ahora
func (e *Event) IsOngoing() bool {
	now := time.Now()
	return now.After(e.StartsAt) && now.Before(e.EndsAt)
}

// IsPast verifica si el evento ya pasó
func (e *Event) IsPast() bool {
	return time.Now().After(e.EndsAt)
}

// GetDuration calcula la duración del evento
func (e *Event) GetDuration() time.Duration {
	return e.EndsAt.Sub(e.StartsAt)
}

// GetDefaultSettings obtiene la configuración por defecto
func GetDefaultSettings() EventSettings {
	return EventSettings{
		AllowCancellations:        true,
		CancellationDeadlineHours: 24,
		AllowTransfers:            true,
		RequireID:                 false,
		CheckinMethod:             "qr_code",
	}
}

// GetSettings obtiene la configuración del evento, con valores por defecto si es nil
func (e *Event) GetSettings() EventSettings {
	if e.Settings == nil {
		return GetDefaultSettings()
	}
	return *e.Settings
}

// AddTag añade una etiqueta al evento
func (e *Event) AddTag(tag string) {
	if e.Tags == nil {
		e.Tags = &[]string{}
	}

	// Verificar si ya existe
	for _, t := range *e.Tags {
		if t == tag {
			return
		}
	}

	*e.Tags = append(*e.Tags, tag)
}

// RemoveTag elimina una etiqueta del evento
func (e *Event) RemoveTag(tag string) {
	if e.Tags == nil {
		return
	}

	newTags := []string{}
	for _, t := range *e.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	if len(newTags) == 0 {
		e.Tags = nil
	} else {
		*e.Tags = newTags
	}
}

// HasTag verifica si el evento tiene una etiqueta específica
func (e *Event) HasTag(tag string) bool {
	if e.Tags == nil {
		return false
	}

	for _, t := range *e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddGalleryImage añade una imagen a la galería
func (e *Event) AddGalleryImage(imageURL string) {
	if e.GalleryImages == nil {
		e.GalleryImages = &[]string{}
	}

	*e.GalleryImages = append(*e.GalleryImages, imageURL)
}

// RemoveGalleryImage elimina una imagen de la galería
func (e *Event) RemoveGalleryImage(imageURL string) {
	if e.GalleryImages == nil {
		return
	}

	newImages := []string{}
	for _, img := range *e.GalleryImages {
		if img != imageURL {
			newImages = append(newImages, img)
		}
	}

	if len(newImages) == 0 {
		e.GalleryImages = nil
	} else {
		*e.GalleryImages = newImages
	}
}

// IncrementViewCount incrementa el contador de vistas
func (e *Event) IncrementViewCount() {
	e.ViewCount++
}

// IncrementFavoriteCount incrementa el contador de favoritos
func (e *Event) IncrementFavoriteCount() {
	e.FavoriteCount++
}

// IncrementShareCount incrementa el contador de compartidos
func (e *Event) IncrementShareCount() {
	e.ShareCount++
}

// MarshalJSON implementa la interfaz json.Marshaler para serialización personalizada
func (e *Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		*Alias
		IsUpcoming bool `json:"is_upcoming"`
		IsOngoing  bool `json:"is_ongoing"`
		IsPast     bool `json:"is_past"`
	}{
		Alias:      (*Alias)(e),
		IsUpcoming: e.IsUpcoming(),
		IsOngoing:  e.IsOngoing(),
		IsPast:     e.IsPast(),
	})
}
