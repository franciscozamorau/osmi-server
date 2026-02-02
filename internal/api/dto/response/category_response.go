package response

import "time"

type CategoryResponse struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Description      *string        `json:"description,omitempty"`
	Icon             *string        `json:"icon,omitempty"`
	ColorHex         string         `json:"color_hex"`
	ParentID         *string        `json:"parent_id,omitempty"`
	ParentCategory   *CategoryInfo  `json:"parent_category,omitempty"`
	Level            int            `json:"level"`
	Path             string         `json:"path"`
	TotalEvents      int            `json:"total_events"`
	TotalTicketsSold int64          `json:"total_tickets_sold"`
	TotalRevenue     float64        `json:"total_revenue"`
	IsActive         bool           `json:"is_active"`
	IsFeatured       bool           `json:"is_featured"`
	SortOrder        int            `json:"sort_order"`
	Children         []CategoryInfo `json:"children,omitempty"`
	MetaTitle        *string        `json:"meta_title,omitempty"`
	MetaDescription  *string        `json:"meta_description,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type CategoryInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Icon        *string `json:"icon,omitempty"`
	ColorHex    string  `json:"color_hex"`
	TotalEvents int     `json:"total_events"`
	IsActive    bool    `json:"is_active"`
	IsFeatured  bool    `json:"is_featured"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	HasNext    bool               `json:"has_next"`
	HasPrev    bool               `json:"has_prev"`
	Filters    CategoryFilter     `json:"filters,omitempty"`
}

type CategoryStatsResponse struct {
	TotalCategories      int               `json:"total_categories"`
	ActiveCategories     int               `json:"active_categories"`
	FeaturedCategories   int               `json:"featured_categories"`
	CategoriesWithEvents int               `json:"categories_with_events"`
	TopCategories        []CategoryRevenue `json:"top_categories"`
	RevenueByCategory    []CategoryRevenue `json:"revenue_by_category"`
	GrowthRate           float64           `json:"growth_rate"`
}

type CategoryRevenue struct {
	CategoryID     string  `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	EventCount     int     `json:"event_count"`
	TicketsSold    int64   `json:"tickets_sold"`
	Revenue        float64 `json:"revenue"`
	AvgTicketPrice float64 `json:"avg_ticket_price"`
}

type CategoryTreeResponse struct {
	Categories []CategoryNode `json:"categories"`
	Depth      int            `json:"depth"`
	TotalNodes int            `json:"total_nodes"`
}

type CategoryNode struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Slug     string         `json:"slug"`
	Icon     *string        `json:"icon,omitempty"`
	ColorHex string         `json:"color_hex"`
	Level    int            `json:"level"`
	Children []CategoryNode `json:"children,omitempty"`
	IsActive bool           `json:"is_active"`
}
