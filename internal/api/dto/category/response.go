// internal/api/dto/category/response.go
package category

import "time"

// CategoryResponse representa la respuesta completa de una categoría
type CategoryResponse struct {
	ID               string         `json:"id"`
	PublicID         string         `json:"public_id,omitempty"`
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

// CategoryListResponse representa una lista paginada de categorías
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	HasNext    bool               `json:"has_next"`
	HasPrev    bool               `json:"has_prev"`
	Filters    interface{}        `json:"filters,omitempty"`
}

// CategoryStatsResponse representa estadísticas de categorías
type CategoryStatsResponse struct {
	TotalCategories         int               `json:"total_categories"`
	ActiveCategories        int               `json:"active_categories"`
	InactiveCategories      int               `json:"inactive_categories"`
	FeaturedCategories      int               `json:"featured_categories"`
	CategoriesWithEvents    int               `json:"categories_with_events"`
	CategoriesWithoutEvents int               `json:"categories_without_events"`
	TopCategories           []CategoryRevenue `json:"top_categories"`
	RevenueByCategory       []CategoryRevenue `json:"revenue_by_category"`
	GrowthRate              float64           `json:"growth_rate,omitempty"`
	Period                  string            `json:"period,omitempty"`
}

// CategoryRevenue representa ingresos por categoría
type CategoryRevenue struct {
	CategoryID     string  `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	CategorySlug   string  `json:"category_slug,omitempty"`
	EventCount     int     `json:"event_count"`
	TicketsSold    int64   `json:"tickets_sold"`
	Revenue        float64 `json:"revenue"`
	AvgTicketPrice float64 `json:"avg_ticket_price"`
	Percentage     float64 `json:"percentage,omitempty"`
}

// CategoryTreeResponse representa la estructura jerárquica de categorías
type CategoryTreeResponse struct {
	Categories []CategoryNode `json:"categories"`
	Depth      int            `json:"depth"`
	TotalNodes int            `json:"total_nodes"`
	RootCount  int            `json:"root_count"`
}

// CategoryNode representa un nodo en el árbol de categorías
type CategoryNode struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Icon        *string        `json:"icon,omitempty"`
	ColorHex    string         `json:"color_hex"`
	Level       int            `json:"level"`
	Path        string         `json:"path,omitempty"`
	TotalEvents int            `json:"total_events,omitempty"`
	IsActive    bool           `json:"is_active"`
	IsFeatured  bool           `json:"is_featured,omitempty"`
	Children    []CategoryNode `json:"children,omitempty"`
	HasChildren bool           `json:"has_children"`
}

// CategoryInfo representa información básica de una categoría
type CategoryInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Icon        *string `json:"icon,omitempty"`
	ColorHex    string  `json:"color_hex"`
	TotalEvents int     `json:"total_events"`
	IsActive    bool    `json:"is_active"`
	IsFeatured  bool    `json:"is_featured,omitempty"`
}

// CalculatePagination calcula campos de paginación
func (r *CategoryListResponse) CalculatePagination() {
	if r.PageSize > 0 {
		r.TotalPages = int((r.Total + int64(r.PageSize) - 1) / int64(r.PageSize))
		r.HasNext = r.Page < r.TotalPages
		r.HasPrev = r.Page > 1
	}
}

// AddChild añade un hijo a un CategoryNode
func (n *CategoryNode) AddChild(child CategoryNode) {
	n.Children = append(n.Children, child)
	n.HasChildren = true
}

// ToCategoryInfo convierte CategoryNode a CategoryInfo
func (n *CategoryNode) ToCategoryInfo() CategoryInfo {
	return CategoryInfo{
		ID:          n.ID,
		Name:        n.Name,
		Slug:        n.Slug,
		Icon:        n.Icon,
		ColorHex:    n.ColorHex,
		TotalEvents: n.TotalEvents,
		IsActive:    n.IsActive,
		IsFeatured:  n.IsFeatured,
	}
}

// CalculatePercentage calcula el porcentaje del total
func (r *CategoryRevenue) CalculatePercentage(total float64) {
	if total > 0 {
		r.Percentage = (r.Revenue / total) * 100
	}
}

// NewCategoryListResponse crea una nueva respuesta de lista con paginación
func NewCategoryListResponse(categories []CategoryResponse, total int64, page, pageSize int) *CategoryListResponse {
	resp := &CategoryListResponse{
		Categories: categories,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}
	resp.CalculatePagination()
	return resp
}

// NewCategoryTreeResponse crea una nueva respuesta de árbol
func NewCategoryTreeResponse(nodes []CategoryNode) *CategoryTreeResponse {
	totalNodes := countNodes(nodes)
	maxDepth := calculateMaxDepth(nodes, 1)

	return &CategoryTreeResponse{
		Categories: nodes,
		Depth:      maxDepth,
		TotalNodes: totalNodes,
		RootCount:  len(nodes),
	}
}

// countNodes cuenta recursivamente el número total de nodos
func countNodes(nodes []CategoryNode) int {
	count := len(nodes)
	for _, node := range nodes {
		count += countNodes(node.Children)
	}
	return count
}

// calculateMaxDepth calcula la profundidad máxima del árbol
func calculateMaxDepth(nodes []CategoryNode, currentDepth int) int {
	if len(nodes) == 0 {
		return currentDepth - 1
	}
	maxDepth := currentDepth
	for _, node := range nodes {
		depth := calculateMaxDepth(node.Children, currentDepth+1)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}
