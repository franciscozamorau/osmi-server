package errors

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// ValidationError representa un error de validación
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implementa la interfaz error
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationErrors representa múltiples errores de validación
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implementa la interfaz error
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "no validation errors"
	}

	messages := make([]string, len(ve.Errors))
	for i, err := range ve.Errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// HasErrors verifica si hay errores
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Add añade un error de validación
func (ve *ValidationErrors) Add(field, message string, args ...interface{}) {
	code := ""
	value := interface{}(nil)

	if len(args) > 0 {
		if codeStr, ok := args[0].(string); ok {
			code = codeStr
		}
	}
	if len(args) > 1 {
		value = args[1]
	}

	ve.Errors = append(ve.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	})
}

// GetErrors devuelve los errores
func (ve *ValidationErrors) GetErrors() []ValidationError {
	return ve.Errors
}

// Clear limpia todos los errores
func (ve *ValidationErrors) Clear() {
	ve.Errors = []ValidationError{}
}

// NewValidationErrors crea nuevos ValidationErrors
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// NewValidationError crea un nuevo ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Códigos de error de validación comunes
const (
	CodeRequired      = "REQUIRED"
	CodeInvalidFormat = "INVALID_FORMAT"
	CodeTooShort      = "TOO_SHORT"
	CodeTooLong       = "TOO_LONG"
	CodeInvalidRange  = "INVALID_RANGE"
	CodeAlreadyExists = "ALREADY_EXISTS"
	CodeNotFound      = "NOT_FOUND"
	CodeInvalidValue  = "INVALID_VALUE"
	CodeMismatch      = "MISMATCH"
	CodeInvalidEmail  = "INVALID_EMAIL"
	CodeInvalidPhone  = "INVALID_PHONE"
	CodeInvalidUUID   = "INVALID_UUID"
	CodeInvalidDate   = "INVALID_DATE"
	CodeFutureDate    = "FUTURE_DATE"
	CodePastDate      = "PAST_DATE"
	CodeInvalidStatus = "INVALID_STATUS"
	CodeInvalidRole   = "INVALID_ROLE"
	CodeWeakPassword  = "WEAK_PASSWORD"
	CodeInvalidAmount = "INVALID_AMOUNT"
	CodeInvalidURL    = "INVALID_URL"
)

// Validator maneja validaciones
type Validator struct {
	errors *ValidationErrors
}

// NewValidator crea un nuevo Validator
func NewValidator() *Validator {
	return &Validator{
		errors: NewValidationErrors(),
	}
}

// Required valida que un campo sea requerido
func (v *Validator) Required(field string, value interface{}) *Validator {
	if value == nil {
		v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
		return v
	}

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
		}
	case *string:
		if val == nil || strings.TrimSpace(*val) == "" {
			v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
		}
	case int, int32, int64:
		if val == 0 {
			v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
		}
	case float32, float64:
		if val == 0.0 {
			v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
		}
	case time.Time:
		if val.IsZero() {
			v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
		}
	}
	return v
}

// Email valida formato de email
func (v *Validator) Email(field, email string) *Validator {
	if email == "" {
		return v
	}

	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	if matched, _ := regexp.MatchString(emailRegex, email); !matched {
		v.errors.Add(field, "invalid email format", CodeInvalidEmail, email)
	}
	return v
}

// Phone valida formato de teléfono
func (v *Validator) Phone(field, phone string) *Validator {
	if phone == "" {
		return v
	}

	phoneRegex := `^\+?[1-9]\d{1,14}$`
	normalized := strings.ReplaceAll(phone, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "(", "")
	normalized = strings.ReplaceAll(normalized, ")", "")

	if matched, _ := regexp.MatchString(phoneRegex, normalized); !matched {
		v.errors.Add(field, "invalid phone number", CodeInvalidPhone, phone)
	}
	return v
}

// UUID valida formato de UUID
func (v *Validator) UUID(field, uuid string) *Validator {
	if uuid == "" {
		return v
	}

	uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	if matched, _ := regexp.MatchString(uuidRegex, strings.ToLower(uuid)); !matched {
		v.errors.Add(field, "invalid UUID format", CodeInvalidUUID, uuid)
	}
	return v
}

// MinLength valida longitud mínima
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if value == "" {
		return v
	}

	if len(strings.TrimSpace(value)) < min {
		v.errors.Add(field, fmt.Sprintf("%s must be at least %d characters", field, min), CodeTooShort, value)
	}
	return v
}

// MaxLength valida longitud máxima
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if value == "" {
		return v
	}

	if len(value) > max {
		v.errors.Add(field, fmt.Sprintf("%s cannot exceed %d characters", field, max), CodeTooLong, value)
	}
	return v
}

// LengthBetween valida longitud entre rangos
func (v *Validator) LengthBetween(field, value string, min, max int) *Validator {
	if value == "" {
		return v
	}

	length := len(strings.TrimSpace(value))
	if length < min || length > max {
		v.errors.Add(field, fmt.Sprintf("%s must be between %d and %d characters", field, min, max), CodeInvalidRange, value)
	}
	return v
}

// MinValue valida valor mínimo
func (v *Validator) MinValue(field string, value, min float64) *Validator {
	if value < min {
		v.errors.Add(field, fmt.Sprintf("%s must be at least %v", field, min), CodeInvalidRange, value)
	}
	return v
}

// MaxValue valida valor máximo
func (v *Validator) MaxValue(field string, value, max float64) *Validator {
	if value > max {
		v.errors.Add(field, fmt.Sprintf("%s cannot exceed %v", field, max), CodeInvalidRange, value)
	}
	return v
}

// ValueBetween valida valor entre rangos
func (v *Validator) ValueBetween(field string, value, min, max float64) *Validator {
	if value < min || value > max {
		v.errors.Add(field, fmt.Sprintf("%s must be between %v and %v", field, min, max), CodeInvalidRange, value)
	}
	return v
}

// Positive valida que sea positivo
func (v *Validator) Positive(field string, value float64) *Validator {
	if value <= 0 {
		v.errors.Add(field, fmt.Sprintf("%s must be positive", field), CodeInvalidValue, value)
	}
	return v
}

// NonNegative valida que no sea negativo
func (v *Validator) NonNegative(field string, value float64) *Validator {
	if value < 0 {
		v.errors.Add(field, fmt.Sprintf("%s cannot be negative", field), CodeInvalidValue, value)
	}
	return v
}

// DateNotZero valida que la fecha no sea cero
func (v *Validator) DateNotZero(field string, date time.Time) *Validator {
	if date.IsZero() {
		v.errors.Add(field, fmt.Sprintf("%s is required", field), CodeRequired)
	}
	return v
}

// FutureDate valida que sea fecha futura
func (v *Validator) FutureDate(field string, date time.Time) *Validator {
	if !date.IsZero() && !date.After(time.Now()) {
		v.errors.Add(field, fmt.Sprintf("%s must be in the future", field), CodeFutureDate, date)
	}
	return v
}

// PastDate valida que sea fecha pasada
func (v *Validator) PastDate(field string, date time.Time) *Validator {
	if !date.IsZero() && !date.Before(time.Now()) {
		v.errors.Add(field, fmt.Sprintf("%s must be in the past", field), CodePastDate, date)
	}
	return v
}

// DateAfter valida que una fecha sea después de otra
func (v *Validator) DateAfter(field, afterField string, date, afterDate time.Time) *Validator {
	if !date.IsZero() && !afterDate.IsZero() && !date.After(afterDate) {
		v.errors.Add(field, fmt.Sprintf("%s must be after %s", field, afterField), CodeInvalidDate)
	}
	return v
}

// DateBefore valida que una fecha sea antes de otra
func (v *Validator) DateBefore(field, beforeField string, date, beforeDate time.Time) *Validator {
	if !date.IsZero() && !beforeDate.IsZero() && !date.Before(beforeDate) {
		v.errors.Add(field, fmt.Sprintf("%s must be before %s", field, beforeField), CodeInvalidDate)
	}
	return v
}

// OneOf valida que el valor esté en una lista de valores permitidos
func (v *Validator) OneOf(field, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}

	for _, allowedValue := range allowed {
		if strings.EqualFold(value, allowedValue) {
			return v
		}
	}

	v.errors.Add(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")), CodeInvalidValue, value)
	return v
}

// Status valida estado
func (v *Validator) Status(field, status string) *Validator {
	if status == "" {
		return v
	}

	validStatuses := []string{
		"active", "inactive", "pending", "approved", "rejected",
		"available", "sold", "used", "cancelled", "refunded",
		"draft", "published", "archived", "deleted",
	}

	return v.OneOf(field, status, validStatuses)
}

// Role valida rol
func (v *Validator) Role(field, role string) *Validator {
	if role == "" {
		return v
	}

	validRoles := []string{
		"admin", "user", "moderator", "guest", "staff",
		"customer", "organizer", "vendor", "super_admin",
	}

	return v.OneOf(field, role, validRoles)
}

// PasswordStrength valida fortaleza de contraseña
func (v *Validator) PasswordStrength(field, password string, minLength int) *Validator {
	if password == "" {
		return v
	}

	if len(password) < minLength {
		v.errors.Add(field, fmt.Sprintf("password must be at least %d characters", minLength), CodeTooShort)
		return v
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

	if count < 3 {
		v.errors.Add(field, "password must contain at least 3 of: uppercase, lowercase, number, special character", CodeWeakPassword)
	}

	return v
}

// URL valida formato de URL
func (v *Validator) URL(field, url string) *Validator {
	if url == "" {
		return v
	}

	urlRegex := `^(https?://)?([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}(/\S*)?$`
	if matched, _ := regexp.MatchString(urlRegex, url); !matched {
		v.errors.Add(field, "invalid URL format", CodeInvalidURL, url)
	}
	return v
}

// Match valida que dos valores coincidan
func (v *Validator) Match(field1, field2, value1, value2 string) *Validator {
	if value1 != value2 {
		v.errors.Add(field1, fmt.Sprintf("%s and %s do not match", field1, field2), CodeMismatch)
	}
	return v
}

// Custom valida con una función personalizada
func (v *Validator) Custom(field string, isValid bool, message string) *Validator {
	if !isValid {
		v.errors.Add(field, message, CodeInvalidValue)
	}
	return v
}

// Validate ejecuta todas las validaciones
func (v *Validator) Validate() error {
	if v.errors.HasErrors() {
		return v.errors
	}
	return nil
}

// GetErrors devuelve los errores de validación
func (v *Validator) GetErrors() *ValidationErrors {
	return v.errors
}

// ClearErrors limpia los errores
func (v *Validator) ClearErrors() {
	v.errors.Clear()
}

// ValidateStruct valida una estructura
func ValidateStruct(data interface{}, rules map[string][]func(*Validator)) error {
	validator := NewValidator()
	// Implementación simplificada - en producción usar reflection
	return validator.Validate()
}
