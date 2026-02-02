package valueobjects

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// UUID representa un UUID válido
type UUID struct {
	value string
}

var (
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// NewUUID crea un nuevo UUID desde un string
func NewUUID(value string) (UUID, error) {
	if value == "" {
		return UUID{}, errors.New("uuid cannot be empty")
	}

	// Normalizar a minúsculas
	normalized := regexp.MustCompile(`[A-Z]`).ReplaceAllStringFunc(value, func(s string) string {
		return strings.ToLower(s)
	})

	if !uuidRegex.MatchString(normalized) {
		return UUID{}, errors.New("invalid uuid format")
	}

	return UUID{value: normalized}, nil
}

// GenerateUUID genera un nuevo UUID aleatorio
func GenerateUUID() UUID {
	return UUID{value: uuid.New().String()}
}

// ParseUUID parsea un string a UUID
func ParseUUID(s string) (UUID, error) {
	return NewUUID(s)
}

// String devuelve el UUID como string
func (u UUID) String() string {
	return u.value
}

// IsValid verifica si el UUID es válido
func (u UUID) IsValid() bool {
	return u.value != "" && uuidRegex.MatchString(u.value)
}

// Equals compara dos UUIDs
func (u UUID) Equals(other UUID) bool {
	return u.value == other.value
}

// IsZero verifica si es un UUID cero
func (u UUID) IsZero() bool {
	return u.value == "00000000-0000-0000-0000-000000000000"
}

// Version devuelve la versión del UUID
func (u UUID) Version() int {
	if !u.IsValid() {
		return 0
	}

	// El carácter en la posición 14 indica la versión
	switch u.value[14] {
	case '1':
		return 1
	case '2':
		return 2
	case '3':
		return 3
	case '4':
		return 4
	case '5':
		return 5
	default:
		return 0
	}
}

// Variant devuelve la variante del UUID
func (u UUID) Variant() string {
	if !u.IsValid() {
		return "unknown"
	}

	// El carácter en la posición 19 indica la variante
	switch u.value[19] {
	case '8', '9', 'a', 'b':
		return "RFC 4122"
	case 'c', 'd':
		return "Microsoft"
	case 'e':
		return "Future"
	default:
		return "NCS"
	}
}

// IsV4 verifica si es un UUID v4 (aleatorio)
func (u UUID) IsV4() bool {
	return u.Version() == 4
}

// ShortID devuelve una versión corta del UUID (primeros 8 caracteres)
func (u UUID) ShortID() string {
	if !u.IsValid() {
		return ""
	}
	return u.value[:8]
}

// Format formatea el UUID con separadores diferentes
func (u UUID) Format(separator string) string {
	if !u.IsValid() {
		return ""
	}

	parts := []string{
		u.value[0:8],
		u.value[9:13],
		u.value[14:18],
		u.value[19:23],
		u.value[24:36],
	}

	return strings.Join(parts, separator)
}

// MarshalJSON implementa json.Marshaler
func (u UUID) MarshalJSON() ([]byte, error) {
	if !u.IsValid() {
		return []byte("null"), nil
	}
	return []byte(`"` + u.value + `"`), nil
}

// UnmarshalJSON implementa json.Unmarshaler
func (u *UUID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		*u = UUID{}
		return nil
	}

	// Eliminar comillas
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	newUUID, err := NewUUID(str)
	if err != nil {
		return err
	}

	*u = newUUID
	return nil
}

// MustCreate crea un UUID o panics
func MustCreate(value string) UUID {
	u, err := NewUUID(value)
	if err != nil {
		panic(err)
	}
	return u
}

// NilUUID devuelve el UUID nulo
func NilUUID() UUID {
	return UUID{value: "00000000-0000-0000-0000-000000000000"}
}

// IsNil verifica si es el UUID nulo
func (u UUID) IsNil() bool {
	return u.Equals(NilUUID())
}
