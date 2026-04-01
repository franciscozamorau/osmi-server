// internal/api/dto/common/pagination.go
package common

// Pagination define la paginación estándar
type Pagination struct {
	Page     int `json:"page" form:"page" query:"page"`
	PageSize int `json:"page_size" form:"page_size" query:"page_size"`
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
