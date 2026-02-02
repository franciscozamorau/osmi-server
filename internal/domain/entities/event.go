package entities

import "time"

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
	GalleryImages  *string `json:"gallery_images,omitempty" db:"gallery_images"`

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

	MaxAttendees *int32 `json:"max_attendees,omitempty" db:"max_attendees"`
	MinAttendees int32  `json:"min_attendees" db:"min_attendees"`

	Tags           *string `json:"tags,omitempty" db:"tags"`
	AgeRestriction *int32  `json:"age_restriction,omitempty" db:"age_restriction"`

	RequiresApproval    bool  `json:"requires_approval" db:"requires_approval"`
	AllowReservations   bool  `json:"allow_reservations" db:"allow_reservations"`
	ReservationDuration int32 `json:"reservation_duration" db:"reservation_duration_minutes"`

	ViewCount     int32 `json:"view_count" db:"view_count"`
	FavoriteCount int32 `json:"favorite_count" db:"favorite_count"`
	ShareCount    int32 `json:"share_count" db:"share_count"`

	MetaTitle       *string `json:"meta_title,omitempty" db:"meta_title"`
	MetaDescription *string `json:"meta_description,omitempty" db:"meta_description"`

	Settings *string `json:"settings,omitempty" db:"settings"`

	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
