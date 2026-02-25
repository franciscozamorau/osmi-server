package enums

// UserRole representa el rol de un usuario en el sistema
// No mapea directamente a la BD, sino que combina is_staff, is_superuser y el tipo de usuario
type UserRole string

const (
	// UserRoleAdmin - Administrador del sistema (is_superuser = true)
	UserRoleAdmin UserRole = "admin"
	// UserRoleOrganizer - Organizador de eventos (usuario registrado que crea eventos)
	UserRoleOrganizer UserRole = "organizer"
	// UserRoleCustomer - Cliente registrado (tiene cuenta en auth.users)
	UserRoleCustomer UserRole = "customer"
	// UserRoleStaff - Personal del sistema (is_staff = true)
	UserRoleStaff UserRole = "staff"
	// UserRoleGuest - Usuario invitado (no tiene cuenta, solo está en crm.customers)
	UserRoleGuest UserRole = "guest"
)

// IsValid verifica si el valor del enum es válido
func (ur UserRole) IsValid() bool {
	switch ur {
	case UserRoleAdmin, UserRoleOrganizer, UserRoleCustomer, UserRoleStaff, UserRoleGuest:
		return true
	}
	return false
}

// HasPermission verifica si el rol tiene un permiso específico
func (ur UserRole) HasPermission(permission string) bool {
	permissions := GetPermissions(ur)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission verifica si el rol tiene alguno de los permisos dados
func (ur UserRole) HasAnyPermission(permissions ...string) bool {
	for _, perm := range permissions {
		if ur.HasPermission(perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions verifica si el rol tiene todos los permisos dados
func (ur UserRole) HasAllPermissions(permissions ...string) bool {
	for _, perm := range permissions {
		if !ur.HasPermission(perm) {
			return false
		}
	}
	return true
}

// CanManageEvents indica si el rol puede gestionar eventos
func (ur UserRole) CanManageEvents() bool {
	return ur == UserRoleAdmin || ur == UserRoleOrganizer || ur == UserRoleStaff
}

// CanManageUsers indica si el rol puede gestionar usuarios
func (ur UserRole) CanManageUsers() bool {
	return ur == UserRoleAdmin
}

// CanManageTickets indica si el rol puede gestionar tickets
func (ur UserRole) CanManageTickets() bool {
	return ur == UserRoleAdmin || ur == UserRoleStaff || ur == UserRoleOrganizer
}

// CanManageOrders indica si el rol puede gestionar órdenes
func (ur UserRole) CanManageOrders() bool {
	return ur == UserRoleAdmin || ur == UserRoleStaff
}

// CanManagePayments indica si el rol puede gestionar pagos
func (ur UserRole) CanManagePayments() bool {
	return ur == UserRoleAdmin || ur == UserRoleStaff
}

// CanViewReports indica si el rol puede ver reportes
func (ur UserRole) CanViewReports() bool {
	return ur == UserRoleAdmin || ur == UserRoleOrganizer || ur == UserRoleStaff
}

// CanPurchaseTickets indica si el rol puede comprar tickets
func (ur UserRole) CanPurchaseTickets() bool {
	return ur == UserRoleCustomer || ur == UserRoleGuest
}

// IsRegistered indica si el usuario está registrado (tiene cuenta)
func (ur UserRole) IsRegistered() bool {
	return ur == UserRoleAdmin || ur == UserRoleOrganizer ||
		ur == UserRoleCustomer || ur == UserRoleStaff
}

// IsGuest indica si el usuario es invitado
func (ur UserRole) IsGuest() bool {
	return ur == UserRoleGuest
}

// IsAdmin indica si el rol es administrador
func (ur UserRole) IsAdmin() bool {
	return ur == UserRoleAdmin
}

// IsStaff indica si el rol es personal
func (ur UserRole) IsStaff() bool {
	return ur == UserRoleStaff || ur == UserRoleAdmin
}

// GetBaseUserFlags devuelve los flags de usuario correspondientes en la BD
func (ur UserRole) GetBaseUserFlags() (isStaff bool, isSuperuser bool) {
	switch ur {
	case UserRoleAdmin:
		return true, true
	case UserRoleStaff:
		return true, false
	default:
		return false, false
	}
}

// String devuelve la representación string del rol
func (ur UserRole) String() string {
	return string(ur)
}

// GetPermissions devuelve los permisos asociados a un rol
func GetPermissions(role UserRole) []string {
	switch role {
	case UserRoleAdmin:
		return []string{
			"users:read", "users:write", "users:delete",
			"events:read", "events:write", "events:delete",
			"tickets:read", "tickets:write", "tickets:delete",
			"orders:read", "orders:write", "orders:delete",
			"payments:read", "payments:write",
			"reports:read", "settings:write",
		}
	case UserRoleOrganizer:
		return []string{
			"events:read", "events:write",
			"tickets:read", "tickets:write",
			"orders:read",
			"reports:read",
		}
	case UserRoleStaff:
		return []string{
			"users:read",
			"events:read", "events:write",
			"tickets:read", "tickets:write",
			"orders:read", "orders:write",
			"payments:read",
			"reports:read",
		}
	case UserRoleCustomer:
		return []string{
			"events:read",
			"tickets:read",
			"orders:read", "orders:write",
		}
	case UserRoleGuest:
		return []string{
			"events:read",
		}
	default:
		return []string{}
	}
}

// GetAllRoles devuelve todos los roles posibles
func GetAllRoles() []UserRole {
	return []UserRole{
		UserRoleAdmin,
		UserRoleOrganizer,
		UserRoleCustomer,
		UserRoleStaff,
		UserRoleGuest,
	}
}

// GetRegisteredRoles devuelve los roles de usuarios registrados
func GetRegisteredRoles() []UserRole {
	return []UserRole{
		UserRoleAdmin,
		UserRoleOrganizer,
		UserRoleCustomer,
		UserRoleStaff,
	}
}

// ParseUserRole convierte flags de BD a UserRole
func ParseUserRole(isStaff bool, isSuperuser bool, hasAccount bool) UserRole {
	switch {
	case isSuperuser:
		return UserRoleAdmin
	case isStaff:
		return UserRoleStaff
	case hasAccount:
		return UserRoleCustomer
	default:
		return UserRoleGuest
	}
}

// MarshalJSON implementa la interfaz json.Marshaler
func (ur UserRole) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(ur) + `"`), nil
}

// UnmarshalJSON implementa la interfaz json.Unmarshaler
func (ur *UserRole) UnmarshalJSON(data []byte) error {
	// Remover comillas
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	role := UserRole(str)
	if !role.IsValid() {
		return &InvalidUserRoleError{Role: str}
	}

	*ur = role
	return nil
}

// InvalidUserRoleError error para valores inválidos
type InvalidUserRoleError struct {
	Role string
}

func (e *InvalidUserRoleError) Error() string {
	return "invalid user role: " + e.Role
}
