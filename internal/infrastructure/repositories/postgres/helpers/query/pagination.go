package query

import (
	"fmt"
	"math"
	"strings"
)

// Pagination representa parámetros de paginación
type Pagination struct {
	Page     int   // Página actual (1-based)
	PageSize int   // Tamaño de página
	Total    int64 // Total de registros
}

// NewPagination crea una nueva paginación
func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50 // Valor por defecto
	}
	if pageSize > 1000 {
		pageSize = 1000 // Límite máximo
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    0,
	}
}

// Offset calcula el offset para la paginación
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit obtiene el límite
func (p *Pagination) Limit() int {
	return p.PageSize
}

// HasNext verifica si hay siguiente página
func (p *Pagination) HasNext() bool {
	if p.Total == 0 {
		return false
	}
	return int64(p.Page*p.PageSize) < p.Total
}

// HasPrev verifica si hay página anterior
func (p *Pagination) HasPrev() bool {
	return p.Page > 1
}

// TotalPages calcula el total de páginas
func (p *Pagination) TotalPages() int {
	if p.Total == 0 || p.PageSize == 0 {
		return 0
	}

	totalPages := int(math.Ceil(float64(p.Total) / float64(p.PageSize)))
	if totalPages < 1 {
		return 1
	}
	return totalPages
}

// NextPage obtiene el número de la siguiente página
func (p *Pagination) NextPage() int {
	if p.HasNext() {
		return p.Page + 1
	}
	return p.Page
}

// PrevPage obtiene el número de la página anterior
func (p *Pagination) PrevPage() int {
	if p.HasPrev() {
		return p.Page - 1
	}
	return p.Page
}

// Validate valida la paginación
func (p *Pagination) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if p.PageSize < 1 {
		return fmt.Errorf("page size must be greater than 0")
	}
	if p.PageSize > 1000 {
		return fmt.Errorf("page size cannot exceed 1000")
	}
	return nil
}

// PaginatedResult representa un resultado paginado
type PaginatedResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
	Metadata   Metadata   `json:"metadata"`
}

// Metadata información adicional del resultado paginado
type Metadata struct {
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
	Filter    string `json:"filter,omitempty"`
	Search    string `json:"search,omitempty"`
}

// NewPaginatedResult crea un nuevo resultado paginado
func NewPaginatedResult[T any](data []T, pagination Pagination) *PaginatedResult[T] {
	return &PaginatedResult[T]{
		Data:       data,
		Pagination: pagination,
	}
}

// BuildPaginatedQuery construye query paginada
func BuildPaginatedQuery(query string, pagination *Pagination) (string, []interface{}) {
	if pagination == nil {
		return query, nil
	}

	args := []interface{}{pagination.Limit(), pagination.Offset()}
	return query + " LIMIT $1 OFFSET $2", args
}

// BuildPaginatedQueryWithArgs construye query paginada con argumentos existentes
func BuildPaginatedQueryWithArgs(query string, args []interface{}, pagination *Pagination) (string, []interface{}) {
	if pagination == nil {
		return query, args
	}

	newArgs := append(args, pagination.Limit(), pagination.Offset())
	return query + fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2), newArgs
}

// BuildCountQuery construye query de conteo
func BuildCountQuery(query string) string {
	// Extraer la parte FROM en adelante
	fromIndex := indexOfCaseInsensitive(query, " FROM ")
	if fromIndex == -1 {
		return "SELECT COUNT(*) FROM (" + query + ") AS count_query"
	}

	// Remover ORDER BY, LIMIT, OFFSET
	countQuery := "SELECT COUNT(*)" + query[fromIndex:]
	countQuery = removeOrderBy(countQuery)
	countQuery = removeLimitOffset(countQuery)

	return countQuery
}

// indexOfCaseInsensitive encuentra índice ignorando mayúsculas/minúsculas
func indexOfCaseInsensitive(s, substr string) int {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Index(sLower, substrLower)
}

// removeOrderBy remueve ORDER BY de la query
func removeOrderBy(query string) string {
	orderByIndex := indexOfCaseInsensitive(query, " ORDER BY ")
	if orderByIndex != -1 {
		return query[:orderByIndex]
	}
	return query
}

// removeLimitOffset remueve LIMIT y OFFSET de la query
func removeLimitOffset(query string) string {
	limitIndex := indexOfCaseInsensitive(query, " LIMIT ")
	if limitIndex != -1 {
		return query[:limitIndex]
	}
	return query
}
