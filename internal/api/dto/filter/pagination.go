package filter

// Pagination define la paginación estándar
type Pagination struct {
	Page       int   `json:"page" validate:"min=1"`
	PageSize   int   `json:"page_size" validate:"min=1,max=100"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasNext    bool  `json:"has_next,omitempty"`
	HasPrev    bool  `json:"has_prev,omitempty"`
}

// NewPagination crea una nueva instancia de paginación con valores por defecto
func NewPagination(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset calcula el offset para consultas SQL
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit retorna el límite para consultas SQL
func (p Pagination) Limit() int {
	return p.PageSize
}

// CalculateTotals calcula los totales basados en el conteo total
func (p *Pagination) CalculateTotals(totalCount int64) {
	p.Total = totalCount
	p.TotalPages = int((totalCount + int64(p.PageSize) - 1) / int64(p.PageSize))
	p.HasNext = p.Page < p.TotalPages
	p.HasPrev = p.Page > 1
}

// CursorPagination para paginación basada en cursor (para APIs infinitas)
type CursorPagination struct {
	Cursor     string `json:"cursor,omitempty"`
	Limit      int    `json:"limit" validate:"min=1,max=100"`
	Direction  string `json:"direction" validate:"oneof=forward backward"`
	HasNext    bool   `json:"has_next,omitempty"`
	HasPrev    bool   `json:"has_prev,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// NewCursorPagination crea una nueva instancia de paginación por cursor
func NewCursorPagination(cursor string, limit int, direction string) CursorPagination {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if direction == "" {
		direction = "forward"
	}

	return CursorPagination{
		Cursor:    cursor,
		Limit:     limit,
		Direction: direction,
	}
}

// KeysetPagination para paginación basada en clave compuesta
type KeysetPagination struct {
	LastID    string      `json:"last_id,omitempty"`
	LastValue interface{} `json:"last_value,omitempty"`
	Limit     int         `json:"limit" validate:"min=1,max=100"`
	SortBy    string      `json:"sort_by"`
	SortOrder string      `json:"sort_order" validate:"oneof=asc desc"`
	HasMore   bool        `json:"has_more,omitempty"`
}

// NewKeysetPagination crea una nueva instancia de keyset pagination
func NewKeysetPagination(lastID string, lastValue interface{}, limit int, sortBy, sortOrder string) KeysetPagination {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	return KeysetPagination{
		LastID:    lastID,
		LastValue: lastValue,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}

// PageInfo contiene información sobre la página actual
type PageInfo struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
	NextPage    *int  `json:"next_page,omitempty"`
	PrevPage    *int  `json:"prev_page,omitempty"`
	FirstPage   int   `json:"first_page"`
	LastPage    int   `json:"last_page"`
	StartItem   int   `json:"start_item"`
	EndItem     int   `json:"end_item"`
}

// CalculatePageInfo calcula información detallada de la página
func CalculatePageInfo(pagination Pagination, totalItems int64) PageInfo {
	totalPages := int((totalItems + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))

	startItem := (pagination.Page-1)*pagination.PageSize + 1
	endItem := startItem + pagination.PageSize - 1
	if endItem > int(totalItems) {
		endItem = int(totalItems)
	}

	var nextPage, prevPage *int
	if pagination.Page < totalPages {
		np := pagination.Page + 1
		nextPage = &np
	}
	if pagination.Page > 1 {
		pp := pagination.Page - 1
		prevPage = &pp
	}

	return PageInfo{
		CurrentPage: pagination.Page,
		PageSize:    pagination.PageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: pagination.Page < totalPages,
		HasPrevPage: pagination.Page > 1,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		FirstPage:   1,
		LastPage:    totalPages,
		StartItem:   startItem,
		EndItem:     endItem,
	}
}

// PaginatedResponse respuesta paginada estándar
type PaginatedResponse[T any] struct {
	Data     []T         `json:"data"`
	PageInfo PageInfo    `json:"page_info"`
	Filters  interface{} `json:"filters,omitempty"`
	Sort     *SortInfo   `json:"sort,omitempty"`
}

// SortInfo información sobre el ordenamiento
type SortInfo struct {
	By    string `json:"by"`
	Order string `json:"order"`
}

// NewPaginatedResponse crea una nueva respuesta paginada
func NewPaginatedResponse[T any](data []T, pagination Pagination, totalItems int64, filters interface{}, sort *SortInfo) PaginatedResponse[T] {
	pageInfo := CalculatePageInfo(pagination, totalItems)

	return PaginatedResponse[T]{
		Data:     data,
		PageInfo: pageInfo,
		Filters:  filters,
		Sort:     sort,
	}
}

// ValidatePagination valida los parámetros de paginación
func ValidatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// GetPaginationFromQuery obtiene paginación de query parameters
func GetPaginationFromQuery(pageStr, pageSizeStr string) Pagination {
	page := 1
	pageSize := 20

	// Parsear página
	if pageStr != "" {
		// Implementar parsing seguro
	}

	// Parsear tamaño de página
	if pageSizeStr != "" {
		// Implementar parsing seguro
	}

	return NewPagination(page, pageSize)
}
