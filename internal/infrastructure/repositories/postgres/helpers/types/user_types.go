package types

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// UserConverter maneja conversiones específicas para usuarios
type UserConverter struct {
	*Converter
}

// NewUserConverter crea un nuevo UserConverter
func NewUserConverter() *UserConverter {
	return &UserConverter{
		Converter: NewConverter(),
	}
}

// UserEmail convierte email de usuario
func (uc *UserConverter) UserEmail(email string) pgtype.Text {
	email = strings.ToLower(strings.TrimSpace(email))
	return uc.Text(email)
}

// UserEmailPtr convierte *email de usuario
func (uc *UserConverter) UserEmailPtr(email *string) pgtype.Text {
	if email == nil {
		return pgtype.Text{Valid: false}
	}
	return uc.UserEmail(*email)
}

// UserPhone convierte teléfono de usuario
func (uc *UserConverter) UserPhone(phone string) pgtype.Text {
	if phone == "" {
		return pgtype.Text{Valid: false}
	}
	// Normalizar teléfono
	phone = normalizePhone(phone)
	return uc.Text(phone)
}

// UserPhonePtr convierte *teléfono de usuario
func (uc *UserConverter) UserPhonePtr(phone *string) pgtype.Text {
	if phone == nil {
		return pgtype.Text{Valid: false}
	}
	return uc.UserPhone(*phone)
}

// UserName convierte nombre de usuario
func (uc *UserConverter) UserName(name string) pgtype.Text {
	name = strings.TrimSpace(name)
	if name == "" {
		return pgtype.Text{Valid: false}
	}
	return uc.Text(name)
}

// UserNamePtr convierte *nombre de usuario
func (uc *UserConverter) UserNamePtr(name *string) pgtype.Text {
	if name == nil {
		return pgtype.Text{Valid: false}
	}
	return uc.UserName(*name)
}

// UserStatus convierte estado de usuario
func (uc *UserConverter) UserStatus(status string) pgtype.Text {
	validStatuses := map[string]bool{
		"active":    true,
		"inactive":  true,
		"suspended": true,
		"deleted":   true,
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if !validStatuses[status] {
		return pgtype.Text{Valid: false}
	}
	return uc.Text(status)
}

// UserRole convierte rol de usuario
func (uc *UserConverter) UserRole(role string) pgtype.Text {
	validRoles := map[string]bool{
		"admin":     true,
		"user":      true,
		"moderator": true,
		"guest":     true,
		"staff":     true,
	}

	role = strings.ToLower(strings.TrimSpace(role))
	if !validRoles[role] {
		return pgtype.Text{Valid: false}
	}
	return uc.Text(role)
}

// UserPreferences convierte preferencias de usuario a JSON
func (uc *UserConverter) UserPreferences(prefs map[string]interface{}) pgtype.Text {
	if prefs == nil || len(prefs) == 0 {
		return pgtype.Text{Valid: false}
	}
	// En implementación real, convertir a JSON
	return uc.Text("{}")
}

// UserLastLogin convierte último login
func (uc *UserConverter) UserLastLogin(t *time.Time) pgtype.Timestamp {
	return uc.TimestampPtr(t)
}

// normalizePhone normaliza número de teléfono
func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	// Implementación simplificada
	return phone
}
