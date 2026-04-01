// internal/api/dto/common/meta.go
package common

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

// SortInfo información sobre el ordenamiento
type SortInfo struct {
	By    string `json:"by"`
	Order string `json:"order"`
}

// PaginatedResponse respuesta paginada estándar
type PaginatedResponse[T any] struct {
	Data     []T         `json:"data"`
	PageInfo PageInfo    `json:"page_info"`
	Filters  interface{} `json:"filters,omitempty"`
	Sort     *SortInfo   `json:"sort,omitempty"`
}

// CalculatePageInfo calcula información detallada de la página
func CalculatePageInfo(page, pageSize int, totalItems int64) PageInfo {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	startItem := (page-1)*pageSize + 1
	endItem := startItem + pageSize - 1
	if endItem > int(totalItems) {
		endItem = int(totalItems)
	}

	var nextPage, prevPage *int
	if page < totalPages {
		np := page + 1
		nextPage = &np
	}
	if page > 1 {
		pp := page - 1
		prevPage = &pp
	}

	return PageInfo{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		FirstPage:   1,
		LastPage:    totalPages,
		StartItem:   startItem,
		EndItem:     endItem,
	}
}

// NewPaginatedResponse crea una nueva respuesta paginada
func NewPaginatedResponse[T any](data []T, page, pageSize int, totalItems int64, filters interface{}, sort *SortInfo) PaginatedResponse[T] {
	pageInfo := CalculatePageInfo(page, pageSize, totalItems)

	return PaginatedResponse[T]{
		Data:     data,
		PageInfo: pageInfo,
		Filters:  filters,
		Sort:     sort,
	}
}
