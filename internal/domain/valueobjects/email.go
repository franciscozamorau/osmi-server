package valueobjects

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Email representa una dirección de email válida
type Email struct {
	value string
}

var (
	emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	maxEmailLength = 320
)

// NewEmail crea un nuevo Email validado
func NewEmail(value string) (Email, error) {
	if value == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	if len(value) > maxEmailLength {
		return Email{}, errors.New("email is too long")
	}

	// Normalizar email
	normalized := strings.ToLower(strings.TrimSpace(value))

	if !emailRegex.MatchString(normalized) {
		return Email{}, errors.New("invalid email format")
	}

	return Email{value: normalized}, nil
}

// String devuelve el email como string
func (e Email) String() string {
	return e.value
}

// LocalPart devuelve la parte local del email (antes de la @)
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// Domain devuelve el dominio del email (después de la @)
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// IsValid verifica si el email es válido
func (e Email) IsValid() bool {
	return e.value != "" && emailRegex.MatchString(e.value)
}

// Equals compara dos emails
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Masked devuelve el email enmascarado para privacidad
func (e Email) Masked() string {
	if e.value == "" {
		return ""
	}

	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return e.value
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return "***@" + domain
	}

	maskedLocal := string(local[0]) + "***" + string(local[len(local)-1])
	return maskedLocal + "@" + domain
}

// GenerateTemporaryEmail genera un email temporal único
func GenerateTemporaryEmail(prefix string) Email {
	uuid := uuid.New().String()[:8]
	return Email{value: prefix + "_" + uuid + "@temp.osmi.com"}
}

// IsTemporary verifica si es un email temporal
func (e Email) IsTemporary() bool {
	return strings.HasSuffix(e.Domain(), ".temp.osmi.com")
}
