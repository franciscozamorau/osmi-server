package request

// Pagination define parámetros de paginación
type Pagination struct {
	Page     int `json:"page" form:"page" query:"page"`
	PageSize int `json:"page_size" form:"page_size" query:"page_size"`
}
