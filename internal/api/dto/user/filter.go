// internal/api/dto/user/filter.go
package user

// UserFilter representa filtros para listar usuarios
type UserFilter struct {
	Email       string `json:"email,omitempty"`
	Username    string `json:"username,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	IsStaff     *bool  `json:"is_staff,omitempty"`
	IsSuperuser *bool  `json:"is_superuser,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
	Search      string `json:"search,omitempty"`
	Role        string `json:"role,omitempty"`
}
