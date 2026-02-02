package dto

type CreateEventRequest struct {
	OrganizerID         string   `json:"organizer_id" validate:"required,uuid4"`
	PrimaryCategoryID   string   `json:"primary_category_id,omitempty" validate:"omitempty,uuid4"`
	VenueID             string   `json:"venue_id,omitempty" validate:"omitempty,uuid4"`
	Name                string   `json:"name" validate:"required,min=3,max=255"`
	Slug                string   `json:"slug,omitempty" validate:"omitempty,slug"`
	ShortDescription    string   `json:"short_description,omitempty" validate:"omitempty,max=500"`
	Description         string   `json:"description" validate:"required,min=10"`
	EventType           string   `json:"event_type" validate:"required,oneof=in_person virtual hybrid"`
	CoverImageURL       string   `json:"cover_image_url,omitempty" validate:"omitempty,url"`
	BannerImageURL      string   `json:"banner_image_url,omitempty" validate:"omitempty,url"`
	Timezone            string   `json:"timezone" validate:"required"`
	StartsAt            string   `json:"starts_at" validate:"required,datetime"`
	EndsAt              string   `json:"ends_at" validate:"required,datetime"`
	DoorsOpenAt         string   `json:"doors_open_at,omitempty" validate:"omitempty,datetime"`
	DoorsCloseAt        string   `json:"doors_close_at,omitempty" validate:"omitempty,datetime"`
	VenueName           string   `json:"venue_name,omitempty" validate:"omitempty,max=255"`
	AddressFull         string   `json:"address_full,omitempty"`
	City                string   `json:"city,omitempty" validate:"omitempty,max=100"`
	State               string   `json:"state,omitempty" validate:"omitempty,max=100"`
	Country             string   `json:"country,omitempty" validate:"omitempty,country_code"`
	Status              string   `json:"status,omitempty" validate:"omitempty,oneof=draft scheduled published live cancelled completed sold_out archived"`
	Visibility          string   `json:"visibility,omitempty" validate:"omitempty,oneof=public private unlisted"`
	IsFeatured          bool     `json:"is_featured,omitempty"`
	IsFree              bool     `json:"is_free,omitempty"`
	MaxAttendees        int      `json:"max_attendees,omitempty" validate:"omitempty,min=1"`
	MinAttendees        int      `json:"min_attendees,omitempty" validate:"omitempty,min=0"`
	Tags                []string `json:"tags,omitempty"`
	AgeRestriction      int      `json:"age_restriction,omitempty" validate:"omitempty,min=0,max=120"`
	RequiresApproval    bool     `json:"requires_approval,omitempty"`
	AllowReservations   bool     `json:"allow_reservations,omitempty"`
	ReservationDuration int      `json:"reservation_duration,omitempty" validate:"omitempty,min=1"`
	CategoryIDs         []string `json:"category_ids,omitempty"`
}

type UpdateEventRequest struct {
	Name             string   `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	ShortDescription string   `json:"short_description,omitempty" validate:"omitempty,max=500"`
	Description      string   `json:"description,omitempty" validate:"omitempty,min=10"`
	EventType        string   `json:"event_type,omitempty" validate:"omitempty,oneof=in_person virtual hybrid"`
	CoverImageURL    string   `json:"cover_image_url,omitempty" validate:"omitempty,url"`
	BannerImageURL   string   `json:"banner_image_url,omitempty" validate:"omitempty,url"`
	Timezone         string   `json:"timezone,omitempty"`
	StartsAt         string   `json:"starts_at,omitempty" validate:"omitempty,datetime"`
	EndsAt           string   `json:"ends_at,omitempty" validate:"omitempty,datetime"`
	DoorsOpenAt      string   `json:"doors_open_at,omitempty" validate:"omitempty,datetime"`
	DoorsCloseAt     string   `json:"doors_close_at,omitempty" validate:"omitempty,datetime"`
	Status           string   `json:"status,omitempty" validate:"omitempty,oneof=draft scheduled published live cancelled completed sold_out archived"`
	Visibility       string   `json:"visibility,omitempty" validate:"omitempty,oneof=public private unlisted"`
	IsFeatured       *bool    `json:"is_featured,omitempty"`
	MaxAttendees     *int     `json:"max_attendees,omitempty" validate:"omitempty,min=1"`
	AgeRestriction   *int     `json:"age_restriction,omitempty" validate:"omitempty,min=0,max=120"`
	Tags             []string `json:"tags,omitempty"`
}

type PublishEventRequest struct {
	PublishAt string `json:"publish_at,omitempty" validate:"omitempty,datetime"`
}

type EventFilter struct {
	Search      string   `json:"search,omitempty"`
	OrganizerID string   `json:"organizer_id,omitempty" validate:"omitempty,uuid4"`
	CategoryID  string   `json:"category_id,omitempty" validate:"omitempty,uuid4"`
	VenueID     string   `json:"venue_id,omitempty" validate:"omitempty,uuid4"`
	EventType   string   `json:"event_type,omitempty" validate:"omitempty,oneof=in_person virtual hybrid"`
	Status      string   `json:"status,omitempty"`
	Country     string   `json:"country,omitempty"`
	City        string   `json:"city,omitempty"`
	IsFeatured  *bool    `json:"is_featured,omitempty"`
	IsFree      *bool    `json:"is_free,omitempty"`
	DateFrom    string   `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo      string   `json:"date_to,omitempty" validate:"omitempty,date"`
	Tags        []string `json:"tags,omitempty"`
}
