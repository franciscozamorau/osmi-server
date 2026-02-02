package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// EmailRegex expresión regular para validar emails
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// UUIDRegex expresión regular para validar UUID v4
var UUIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// PhoneRegex expresión regular para validar teléfonos internacionales
var PhoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// URLRegex expresión regular para validar URLs
var URLRegex = regexp.MustCompile(`^(https?://)?([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}(/\S*)?$`)

// UsernameRegex expresión regular para validar nombres de usuario
var UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,50}$`)

// NameRegex expresión regular para validar nombres personales
var NameRegex = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s'-]{2,100}$`)

// CurrencyCodeRegex expresión regular para códigos de moneda ISO 4217
var CurrencyCodeRegex = regexp.MustCompile(`^[A-Z]{3}$`)

// LanguageCodeRegex expresión regular para códigos de idioma ISO 639-1
var LanguageCodeRegex = regexp.MustCompile(`^[a-z]{2}$`)

// TimezoneRegex expresión regular para zonas horarias
var TimezoneRegex = regexp.MustCompile(`^[A-Za-z_]+/[A-Za-z_]+$`)

// IsValidEmail valida formato de email
func IsValidEmail(email string) bool {
	if strings.TrimSpace(email) == "" {
		return false
	}
	return EmailRegex.MatchString(strings.ToLower(email))
}

// IsValidEmailWithReason valida email y retorna razón
func IsValidEmailWithReason(email string) (bool, string) {
	email = strings.TrimSpace(email)

	if email == "" {
		return false, "email cannot be empty"
	}

	if len(email) > 254 {
		return false, "email cannot exceed 254 characters"
	}

	if !EmailRegex.MatchString(strings.ToLower(email)) {
		return false, "invalid email format"
	}

	// Verificar dominio
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, "invalid email format"
	}

	domain := parts[1]
	if strings.Count(domain, ".") < 1 {
		return false, "invalid domain in email"
	}

	return true, ""
}

// IsValidUUID valida formato de UUID v4
func IsValidUUID(uuid string) bool {
	if strings.TrimSpace(uuid) == "" {
		return false
	}
	return UUIDRegex.MatchString(strings.ToLower(uuid))
}

// IsValidUUIDWithReason valida UUID y retorna razón
func IsValidUUIDWithReason(uuid string) (bool, string) {
	uuid = strings.TrimSpace(uuid)

	if uuid == "" {
		return false, "uuid cannot be empty"
	}

	if !IsValidUUID(uuid) {
		return false, "invalid UUID format (must be UUID v4)"
	}

	return true, ""
}

// IsValidPhone valida formato de teléfono
func IsValidPhone(phone string) bool {
	if strings.TrimSpace(phone) == "" {
		return true // opcional
	}

	phone = NormalizePhone(phone)
	return PhoneRegex.MatchString(phone)
}

// IsValidPhoneWithReason valida teléfono y retorna razón
func IsValidPhoneWithReason(phone string) (bool, string) {
	phone = strings.TrimSpace(phone)

	if phone == "" {
		return true, "" // opcional
	}

	normalized := NormalizePhone(phone)
	if len(normalized) < 7 {
		return false, "phone number too short (minimum 7 digits)"
	}

	if len(normalized) > 20 {
		return false, "phone number too long (maximum 20 digits)"
	}

	if !PhoneRegex.MatchString(normalized) {
		return false, "invalid phone number format"
	}

	return true, ""
}

// NormalizePhone normaliza número de teléfono
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}

	hasPlus := strings.HasPrefix(phone, "+")

	// Remover todos los caracteres no numéricos excepto +
	re := regexp.MustCompile(`[^\d+]`)
	digits := re.ReplaceAllString(phone, "")

	// Si tenía + al inicio, asegurarse de mantenerlo
	if hasPlus && !strings.HasPrefix(digits, "+") {
		digits = "+" + strings.TrimPrefix(digits, "+")
	}

	// Remover ceros iniciales innecesarios después del código de país
	if strings.HasPrefix(digits, "+") && len(digits) > 3 {
		if digits[1] == '0' {
			digits = "+" + digits[2:]
		}
	}

	return digits
}

// IsValidURL valida formato de URL
func IsValidURL(url string) bool {
	if strings.TrimSpace(url) == "" {
		return true // opcional
	}
	return URLRegex.MatchString(url)
}

// IsValidUsername valida nombre de usuario
func IsValidUsername(username string) bool {
	if strings.TrimSpace(username) == "" {
		return false
	}

	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	return UsernameRegex.MatchString(username)
}

// IsValidUsernameWithReason valida username y retorna razón
func IsValidUsernameWithReason(username string) (bool, string) {
	username = strings.TrimSpace(username)

	if username == "" {
		return false, "username cannot be empty"
	}

	if len(username) < 3 {
		return false, "username too short (minimum 3 characters)"
	}

	if len(username) > 50 {
		return false, "username too long (maximum 50 characters)"
	}

	if !UsernameRegex.MatchString(username) {
		return false, "username can only contain letters, numbers, dots, underscores and hyphens"
	}

	// Verificar que no comience o termine con punto o guión
	if strings.HasPrefix(username, ".") || strings.HasPrefix(username, "-") ||
		strings.HasPrefix(username, "_") {
		return false, "username cannot start with ., - or _"
	}

	if strings.HasSuffix(username, ".") || strings.HasSuffix(username, "-") ||
		strings.HasSuffix(username, "_") {
		return false, "username cannot end with ., - or _"
	}

	// Verificar que no tenga puntos o guiones consecutivos
	if strings.Contains(username, "..") || strings.Contains(username, "--") ||
		strings.Contains(username, "__") || strings.Contains(username, ".-") ||
		strings.Contains(username, "-.") || strings.Contains(username, "._") ||
		strings.Contains(username, "_.") || strings.Contains(username, "-_") ||
		strings.Contains(username, "_-") {
		return false, "username cannot have consecutive special characters"
	}

	return true, ""
}

// IsValidName valida nombre personal
func IsValidName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}

	name = strings.TrimSpace(name)
	if len(name) > 100 {
		return false
	}

	return NameRegex.MatchString(name)
}

// IsValidNameWithReason valida nombre y retorna razón
func IsValidNameWithReason(name string) (bool, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		return false, "name cannot be empty"
	}

	if len(name) > 100 {
		return false, "name too long (maximum 100 characters)"
	}

	if !NameRegex.MatchString(name) {
		return false, "name can only contain letters, spaces, hyphens and apostrophes"
	}

	// Verificar que no tenga caracteres especiales consecutivos
	if strings.Contains(name, "  ") || strings.Contains(name, "--") ||
		strings.Contains(name, "''") || strings.Contains(name, "..") {
		return false, "name cannot have consecutive special characters"
	}

	return true, ""
}

// IsValidPassword valida fortaleza de contraseña
func IsValidPassword(password string) bool {
	return IsValidPasswordWithStrength(password, 8)
}

// IsValidPasswordWithStrength valida contraseña con fortaleza personalizada
func IsValidPasswordWithStrength(password string, minLength int) bool {
	if len(password) < minLength {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Requerir al menos 3 de los 4 tipos
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// IsValidPasswordWithReason valida contraseña y retorna razón
func IsValidPasswordWithReason(password string, minLength int) (bool, string) {
	if len(password) < minLength {
		return false, fmt.Sprintf("password too short (minimum %d characters)", minLength)
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	var requirements []string

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		requirements = append(requirements, "uppercase letter")
	}
	if !hasLower {
		requirements = append(requirements, "lowercase letter")
	}
	if !hasNumber {
		requirements = append(requirements, "number")
	}
	if !hasSpecial {
		requirements = append(requirements, "special character")
	}

	if len(requirements) > 1 {
		return false, fmt.Sprintf("password must contain at least 3 of: uppercase, lowercase, number, special character. Missing: %s",
			strings.Join(requirements, ", "))
	}

	return true, ""
}

// IsValidCurrencyCode valida código de moneda ISO 4217
func IsValidCurrencyCode(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	return CurrencyCodeRegex.MatchString(currency)
}

// IsValidLanguageCode valida código de idioma ISO 639-1
func IsValidLanguageCode(language string) bool {
	if len(language) != 2 {
		return false
	}
	return LanguageCodeRegex.MatchString(language)
}

// IsValidTimezone valida zona horaria
func IsValidTimezone(timezone string) bool {
	return TimezoneRegex.MatchString(timezone)
}

// IsValidDate valida formato de fecha
func IsValidDate(dateStr string) bool {
	if dateStr == "" {
		return false
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		_, err := time.Parse(format, dateStr)
		if err == nil {
			return true
		}
	}
	return false
}

// IsFutureDate verifica si es fecha futura
func IsFutureDate(t time.Time) bool {
	return t.After(time.Now())
}

// IsPastDate verifica si es fecha pasada
func IsPastDate(t time.Time) bool {
	return t.Before(time.Now())
}

// IsDateRangeValid verifica rango de fechas válido
func IsDateRangeValid(start, end time.Time) bool {
	return !start.IsZero() && !end.IsZero() && end.After(start)
}

// IsValidAmount valida monto
func IsValidAmount(amount float64) bool {
	return amount >= 0
}

// IsValidQuantity valida cantidad
func IsValidQuantity(quantity int) bool {
	return quantity > 0
}

// IsValidPercentage valida porcentaje
func IsValidPercentage(percentage float64) bool {
	return percentage >= 0 && percentage <= 100
}

// TruncateString trunca string a longitud máxima
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

// SanitizeString limpia string
func SanitizeString(s string) string {
	// Remover caracteres de control
	re := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	return re.ReplaceAllString(s, "")
}
