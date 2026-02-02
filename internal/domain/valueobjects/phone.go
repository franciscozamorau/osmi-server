package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

// Phone representa un número de teléfono válido
type Phone struct {
	value string
}

var (
	// Regex para validar números internacionales (E.164)
	e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	// Regex para números locales (sin código de país)
	localRegex = regexp.MustCompile(`^\d{8,15}$`)
)

// NewPhone crea un nuevo Phone validado
func NewPhone(value string) (Phone, error) {
	if value == "" {
		return Phone{}, nil // Phone puede ser opcional
	}

	cleaned := cleanPhoneNumber(value)

	// Validar formato internacional o local
	if !e164Regex.MatchString(cleaned) && !localRegex.MatchString(cleaned) {
		return Phone{}, errors.New("invalid phone number format")
	}

	return Phone{value: cleaned}, nil
}

// cleanPhoneNumber limpia el número de teléfono
func cleanPhoneNumber(phone string) string {
	// Eliminar espacios, guiones, paréntesis
	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"(", "",
		")", "",
		".", "",
	)
	return replacer.Replace(strings.TrimSpace(phone))
}

// String devuelve el teléfono como string
func (p Phone) String() string {
	return p.value
}

// IsValid verifica si el teléfono es válido
func (p Phone) IsValid() bool {
	if p.value == "" {
		return true // Phone puede estar vacío
	}
	return e164Regex.MatchString(p.value) || localRegex.MatchString(p.value)
}

// IsInternational verifica si es un número internacional
func (p Phone) IsInternational() bool {
	return strings.HasPrefix(p.value, "+")
}

// IsLocal verifica si es un número local
func (p Phone) IsLocal() bool {
	return !p.IsInternational() && p.value != ""
}

// CountryCode extrae el código de país (si es internacional)
func (p Phone) CountryCode() string {
	if !p.IsInternational() {
		return ""
	}

	// El formato es +CCXXXXXXXXX
	// Buscamos el código de país (1-3 dígitos después del +)
	for i := 1; i <= 3 && i < len(p.value); i++ {
		if !isDigit(p.value[i]) {
			break
		}
		// Verificar que el siguiente carácter sea dígito o fin de string
		if i+1 >= len(p.value) || isDigit(p.value[i+1]) {
			return p.value[1 : i+1]
		}
	}
	return ""
}

// LocalNumber extrae el número local (sin código de país)
func (p Phone) LocalNumber() string {
	if p.value == "" {
		return ""
	}

	if p.IsInternational() {
		countryCode := p.CountryCode()
		if countryCode != "" {
			return p.value[len(countryCode)+1:] // +CC + número
		}
	}
	return p.value
}

// Format formatea el número para visualización
func (p Phone) Format() string {
	if p.value == "" {
		return ""
	}

	if p.IsInternational() {
		countryCode := p.CountryCode()
		localNumber := p.LocalNumber()

		// Formatear según la longitud
		switch len(localNumber) {
		case 10: // Formato: +CC (XXX) XXX-XXXX
			return p.value[:len(countryCode)+1] + " (" + localNumber[:3] + ") " + localNumber[3:6] + "-" + localNumber[6:]
		case 9: // Formato: +CC XXX XXX XXX
			return p.value[:len(countryCode)+1] + " " + formatChunks(localNumber, 3)
		default:
			return p.value
		}
	}

	// Formato local
	switch len(p.value) {
	case 10: // Formato: (XXX) XXX-XXXX
		return "(" + p.value[:3] + ") " + p.value[3:6] + "-" + p.value[6:]
	case 8: // Formato: XXXX-XXXX
		return p.value[:4] + "-" + p.value[4:]
	default:
		return p.value
	}
}

// formatChunks divide un string en chunks
func formatChunks(s string, chunkSize int) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && i%chunkSize == 0 {
			result.WriteString(" ")
		}
		result.WriteRune(r)
	}
	return result.String()
}

// isDigit verifica si un carácter es dígito
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// Equals compara dos números de teléfono
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}

// Masked devuelve el teléfono enmascarado para privacidad
func (p Phone) Masked() string {
	if p.value == "" {
		return ""
	}

	if len(p.value) <= 4 {
		return "****"
	}

	visibleDigits := 4
	masked := strings.Repeat("*", len(p.value)-visibleDigits) + p.value[len(p.value)-visibleDigits:]

	if p.IsInternational() {
		countryCode := p.CountryCode()
		if countryCode != "" {
			return "+" + countryCode + " " + masked[len(countryCode)+1:]
		}
	}
	return masked
}
