package enums

type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleOrganizer UserRole = "organizer"
	UserRoleCustomer  UserRole = "customer"
	UserRoleStaff     UserRole = "staff"
	UserRoleGuest     UserRole = "guest"
)

func (ur UserRole) IsValid() bool {
	switch ur {
	case UserRoleAdmin, UserRoleOrganizer, UserRoleCustomer, UserRoleStaff, UserRoleGuest:
		return true
	}
	return false
}

func (ur UserRole) HasPermission(permission string) bool {
	permissions := GetPermissions(ur)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (ur UserRole) CanManageEvents() bool {
	return ur == UserRoleAdmin || ur == UserRoleOrganizer || ur == UserRoleStaff
}

func (ur UserRole) CanManageUsers() bool {
	return ur == UserRoleAdmin
}

func (ur UserRole) CanManageTickets() bool {
	return ur == UserRoleAdmin || ur == UserRoleStaff
}

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
