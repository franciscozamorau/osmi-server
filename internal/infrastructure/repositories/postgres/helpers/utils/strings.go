package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

// SafeStringForLog oculta información sensible en logs
func SafeStringForLog(s string) string {
	if len(s) <= 4 {
		return "***"
	}
	return s[:2] + "***" + s[len(s)-2:]
}

// SafeEmailForLog oculta parte del email en logs
func SafeEmailForLog(email string) string {
	if email == "" {
		return "***"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return SafeStringForLog(email)
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		local = "***"
	} else {
		local = local[:2] + "***"
	}

	return local + "@" + domain
}

// SafePhoneForLog oculta parte del teléfono en logs
func SafePhoneForLog(phone string) string {
	if len(phone) <= 4 {
		return "***"
	}
	return "***" + phone[len(phone)-4:]
}

// TruncateMiddle trunca un string en el medio
func TruncateMiddle(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	partLength := (maxLength - 3) / 2
	start := s[:partLength]
	end := s[len(s)-partLength:]

	return start + "..." + end
}

// GenerateSlug genera un slug a partir de un string
func GenerateSlug(s string) string {
	// Convertir a minúsculas
	s = strings.ToLower(s)

	// Remover acentos (simplificado)
	replacements := map[string]string{
		"á": "a", "é": "e", "í": "i", "ó": "o", "ú": "u",
		"à": "a", "è": "e", "ì": "i", "ò": "o", "ù": "u",
		"ä": "a", "ë": "e", "ï": "i", "ö": "o", "ü": "u",
		"ñ": "n", "ç": "c",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Remover caracteres no alfanuméricos
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) {
			result.WriteRune('-')
		}
	}

	slug := result.String()

	// Remover guiones múltiples
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remover guiones al inicio y final
	slug = strings.Trim(slug, "-")

	return slug
}

// Capitalize capitaliza la primera letra de cada palabra
func Capitalize(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}

	return strings.Join(words, " ")
}

// ToCamelCase convierte a camelCase
func ToCamelCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	if len(words) == 0 {
		return s
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(words[0]))

	for i := 1; i < len(words); i++ {
		word := words[i]
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
		}
	}

	return result.String()
}

// ToSnakeCase convierte a snake_case
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	var lastWasUpper bool

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && !lastWasUpper {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
			lastWasUpper = true
		} else {
			result.WriteRune(r)
			lastWasUpper = false
		}
	}

	return result.String()
}

// ToPascalCase convierte a PascalCase
func ToPascalCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	var result strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
		}
	}

	return result.String()
}

// RemoveAccents remueve acentos de un string
func RemoveAccents(s string) string {
	replacements := map[rune]rune{
		'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u',
		'à': 'a', 'è': 'e', 'ì': 'i', 'ò': 'o', 'ù': 'u',
		'ä': 'a', 'ë': 'e', 'ï': 'i', 'ö': 'o', 'ü': 'u',
		'â': 'a', 'ê': 'e', 'î': 'i', 'ô': 'o', 'û': 'u',
		'ã': 'a', 'ñ': 'n', 'ç': 'c',
		'Á': 'A', 'É': 'E', 'Í': 'I', 'Ó': 'O', 'Ú': 'U',
		'À': 'A', 'È': 'E', 'Ì': 'I', 'Ò': 'O', 'Ù': 'U',
		'Ä': 'A', 'Ë': 'E', 'Ï': 'I', 'Ö': 'O', 'Ü': 'U',
		'Â': 'A', 'Ê': 'E', 'Î': 'I', 'Ô': 'O', 'Û': 'U',
		'Ã': 'A', 'Ñ': 'N', 'Ç': 'C',
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, ok := replacements[r]; ok {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ContainsAny verifica si contiene alguno de los strings
func ContainsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsAll verifica si contiene todos los strings
func ContainsAll(s string, substrings []string) bool {
	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// Reverse invierte un string
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome verifica si es palíndromo
func IsPalindrome(s string) bool {
	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))
	return s == Reverse(s)
}

// GenerateRandomString genera string aleatorio
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateRandomCode genera código aleatorio
func GenerateRandomCode(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	if charset == "" {
		charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// MaskString enmascara parte de un string
func MaskString(s string, visibleStart, visibleEnd int, maskChar rune) string {
	if s == "" {
		return ""
	}

	if visibleStart < 0 {
		visibleStart = 0
	}
	if visibleEnd < 0 {
		visibleEnd = 0
	}

	if visibleStart+visibleEnd >= len(s) {
		return s
	}

	runes := []rune(s)
	var result strings.Builder

	// Parte visible al inicio
	if visibleStart > 0 {
		result.WriteString(string(runes[:visibleStart]))
	}

	// Caracteres enmascarados
	maskCount := len(runes) - visibleStart - visibleEnd
	for i := 0; i < maskCount; i++ {
		result.WriteRune(maskChar)
	}

	// Parte visible al final
	if visibleEnd > 0 {
		result.WriteString(string(runes[len(runes)-visibleEnd:]))
	}

	return result.String()
}

// ExtractEmails extrae emails de un texto
func ExtractEmails(text string) []string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	return emailRegex.FindAllString(text, -1)
}

// ExtractURLs extrae URLs de un texto
func ExtractURLs(text string) []string {
	urlRegex := regexp.MustCompile(`(https?://)?([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}(/\S*)?`)
	return urlRegex.FindAllString(text, -1)
}

// WordCount cuenta palabras en un texto
func WordCount(text string) int {
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	return len(words)
}

// CharacterCount cuenta caracteres (excluyendo espacios)
func CharacterCount(text string) int {
	if text == "" {
		return 0
	}

	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}

// LineCount cuenta líneas en un texto
func LineCount(text string) int {
	if text == "" {
		return 0
	}

	return strings.Count(text, "\n") + 1
}

// IsEmptyOrWhitespace verifica si está vacío o solo tiene espacios
func IsEmptyOrWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

// JoinNotEmpty une strings no vacías
func JoinNotEmpty(separator string, strings ...string) string {
	var nonEmpty []string
	for _, s := range strings {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, separator)
}

// DefaultIfEmpty devuelve valor por defecto si está vacío
func DefaultIfEmpty(s, defaultValue string) string {
	if IsEmptyOrWhitespace(s) {
		return defaultValue
	}
	return s
}

// PointerToString convierte string pointer a string
func PointerToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringToPointer convierte string a pointer
func StringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IntToString convierte int a string
func IntToString(i int) string {
	return fmt.Sprintf("%d", i)
}

// FloatToString convierte float a string
func FloatToString(f float64, precision int) string {
	return fmt.Sprintf("%.*f", precision, f)
}

// BoolToString convierte bool a string
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
