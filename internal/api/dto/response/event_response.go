package response

import "time"

type EventResponse struct {
	ID               string           `json:"id"`
	Organizer        OrganizerInfo    `json:"organizer"`
	Venue            VenueInfo        `json:"venue,omitempty"`
	PrimaryCategory  CategoryInfo     `json:"primary_category,omitempty"`
	Categories       []CategoryInfo   `json:"categories"`
	Name             string           `json:"name"`
	Slug             string           `json:"slug"`
	ShortDescription string           `json:"short_description,omitempty"`
	Description      string           `json:"description"`
	EventType        string           `json:"event_type"`
	CoverImageURL    string           `json:"cover_image_url,omitempty"`
	BannerImageURL   string           `json:"banner_image_url,omitempty"`
	GalleryImages    []string         `json:"gallery_images"`
	Timezone         string           `json:"timezone"`
	StartsAt         time.Time        `json:"starts_at"`
	EndsAt           time.Time        `json:"ends_at"`
	DoorsOpenAt      time.Time        `json:"doors_open_at,omitempty"`
	DoorsCloseAt     time.Time        `json:"doors_close_at,omitempty"`
	VenueName        string           `json:"venue_name,omitempty"`
	AddressFull      string           `json:"address_full,omitempty"`
	City             string           `json:"city,omitempty"`
	State            string           `json:"state,omitempty"`
	Country          string           `json:"country,omitempty"`
	Status           string           `json:"status"`
	Visibility       string           `json:"visibility"`
	IsFeatured       bool             `json:"is_featured"`
	IsFree           bool             `json:"is_free"`
	MaxAttendees     int              `json:"max_attendees,omitempty"`
	MinAttendees     int              `json:"min_attendees"`
	Tags             []string         `json:"tags"`
	AgeRestriction   int              `json:"age_restriction,omitempty"`
	ViewCount        int              `json:"view_count"`
	FavoriteCount    int              `json:"favorite_count"`
	ShareCount       int              `json:"share_count"`
	TicketTypes      []TicketTypeInfo `json:"ticket_types"`
	PublishedAt      time.Time        `json:"published_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type EventStatsResponse struct {
	TicketsSold      int     `json:"tickets_sold"`
	TicketsAvailable int     `json:"tickets_available"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
	CheckInRate      float64 `json:"check_in_rate"`
	ConversionRate   float64 `json:"conversion_rate"`
	ViewsToday       int     `json:"views_today"`
	SalesVelocity    float64 `json:"sales_velocity"`
	LowTicketWarning bool    `json:"low_ticket_warning"`
	PeakSalesPeriod  string  `json:"peak_sales_period,omitempty"`
}

type EventListResponse struct {
	Events     []EventResponse `json:"events"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
	HasNext    bool            `json:"has_next"`
	HasPrev    bool            `json:"has_prev"`
	Filters    EventFilter     `json:"filters,omitempty"`
	SortBy     string          `json:"sort_by,omitempty"`
	SortOrder  string          `json:"sort_order,omitempty"`
}

type TicketTypeInfo struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
	Available      int       `json:"available"`
	Total          int       `json:"total"`
	Sold           int       `json:"sold"`
	Reserved       int       `json:"reserved"`
	IsSoldOut      bool      `json:"is_sold_out"`
	IsHidden       bool      `json:"is_hidden"`
	MaxPerOrder    int       `json:"max_per_order"`
	MinPerOrder    int       `json:"min_per_order"`
	SaleStartsAt   time.Time `json:"sale_starts_at"`
	SaleEndsAt     time.Time `json:"sale_ends_at,omitempty"`
	IsTransferable bool      `json:"is_transferable"`
	TransferFee    float64   `json:"transfer_fee,omitempty"`
	RefundPolicy   string    `json:"refund_policy,omitempty"`
	IncludesFees   bool      `json:"includes_fees"`
}

type EventFilter struct {
	CategoryIDs  []string   `json:"category_ids,omitempty"`
	OrganizerIDs []string   `json:"organizer_ids,omitempty"`
	VenueIDs     []string   `json:"venue_ids,omitempty"`
	Statuses     []string   `json:"statuses,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	PriceFrom    *float64   `json:"price_from,omitempty"`
	PriceTo      *float64   `json:"price_to,omitempty"`
	IsFeatured   *bool      `json:"is_featured,omitempty"`
	IsFree       *bool      `json:"is_free,omitempty"`
	HasTickets   *bool      `json:"has_tickets,omitempty"`
	SearchQuery  *string    `json:"search_query,omitempty"`
	City         *string    `json:"city,omitempty"`
	Country      *string    `json:"country,omitempty"`
	OnlineOnly   *bool      `json:"online_only,omitempty"`
}
