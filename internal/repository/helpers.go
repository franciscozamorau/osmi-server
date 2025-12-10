// helpers.go - COMPLETO Y CORREGIDO
package repository

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// =============================================================================
// CONVERSIONES DE TIPOS - CORREGIDAS
// =============================================================================

// ToPgText convierte un string a pgtype.Text
func ToPgText(s string) pgtype.Text {
	return pgtype.Text{String: strings.TrimSpace(s), Valid: s != ""}
}

// ToPgTextFromPtr convierte *string a pgtype.Text
func ToPgTextFromPtr(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: strings.TrimSpace(*s), Valid: true}
}

// ToPgInt4 convierte un int a pgtype.Int4
func ToPgInt4(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// ToPgInt4FromInt convierte int a pgtype.Int4
func ToPgInt4FromInt(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// ToPgInt4FromInt32 convierte int32 a pgtype.Int4
func ToPgInt4FromInt32(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// ToPgInt4FromPtr convierte un *int a pgtype.Int4
func ToPgInt4FromPtr(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

// ToPgInt4FromInt32Ptr convierte *int32 a pgtype.Int4
func ToPgInt4FromInt32Ptr(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// ToPgInt8FromInt64 convierte int64 a pgtype.Int8
func ToPgInt8FromInt64(i int64) pgtype.Int8 {
	return pgtype.Int8{Int64: i, Valid: true}
}

// ToPgInt8FromInt64Ptr convierte *int64 a pgtype.Int8
func ToPgInt8FromInt64Ptr(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// ToPgFloat8 convierte un float64 a pgtype.Float8
func ToPgFloat8(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: true}
}

// ToPgFloat8FromPtr convierte *float64 a pgtype.Float8
func ToPgFloat8FromPtr(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

// ToPgTimestamp convierte time.Time a pgtype.Timestamp
func ToPgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: !t.IsZero()}
}

// ToPgTimestampFromPtr convierte *time.Time a pgtype.Timestamp
func ToPgTimestampFromPtr(t *time.Time) pgtype.Timestamp {
	if t == nil || t.IsZero() {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// ToPgDateFromPtr convierte *time.Time a pgtype.Date
func ToPgDateFromPtr(t *time.Time) pgtype.Date {
	if t == nil || t.IsZero() {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

// ToTimeFromPgTimestamp convierte pgtype.Timestamp a *time.Time
func ToTimeFromPgTimestamp(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// ToTimeFromPgDate convierte pgtype.Date a *time.Time
func ToTimeFromPgDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

// ToInt32FromPgInt4 convierte pgtype.Int4 a *int32
func ToInt32FromPgInt4(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// ToInt64FromPgInt8 convierte pgtype.Int8 a *int64
func ToInt64FromPgInt8(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// ToStringFromPgText convierte pgtype.Text a *string
func ToStringFromPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// ToFloat64FromPgFloat8 convierte pgtype.Float8 a *float64
func ToFloat64FromPgFloat8(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// =============================================================================
// FUNCIONES DE CONVERSIÓN ESPECÍFICAS PARA TICKETS
// =============================================================================

// ToPgInt4FromInt64 convierte int64 a pgtype.Int4
func ToPgInt4FromInt64(i int64) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// ToInt64FromPgInt4 convierte pgtype.Int4 a int64
func ToInt64FromPgInt4(i pgtype.Int4) int64 {
	if !i.Valid {
		return 0
	}
	return int64(i.Int32)
}

// ToInt64PtrFromPgInt4 convierte pgtype.Int4 a *int64
func ToInt64PtrFromPgInt4(i pgtype.Int4) *int64 {
	if !i.Valid {
		return nil
	}
	val := int64(i.Int32)
	return &val
}

// ToPgInt4FromInt64Ptr convierte *int64 a pgtype.Int4
func ToPgInt4FromInt64Ptr(i *int64) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

// =============================================================================
// VALIDACIONES
// =============================================================================

// IsValidEmail valida formato de email
func IsValidEmail(email string) bool {
	if strings.TrimSpace(email) == "" {
		return false
	}
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

// IsValidUUID valida formato de UUID
func IsValidUUID(uuid string) bool {
	if strings.TrimSpace(uuid) == "" {
		return false
	}
	uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidRegex, strings.ToLower(uuid))
	return matched
}

// IsValidPhone valida formato básico de teléfono
func IsValidPhone(phone string) bool {
	if strings.TrimSpace(phone) == "" {
		return true // opcional
	}
	phoneRegex := `^[\+]?[0-9\s\-\(\)]{10,}$`
	matched, _ := regexp.MatchString(phoneRegex, phone)
	return matched
}

// =============================================================================
// MANEJO DE ERRORES DE POSTGRES
// =============================================================================

// IsDuplicateKeyError verifica si el error es por violación de unique constraint
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}

	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate key") ||
		strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "unique constraint") ||
		strings.Contains(errorMsg, "23505")
}

// IsForeignKeyError verifica si el error es por violación de foreign key
func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503" // foreign_key_violation
	}

	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "foreign key") ||
		strings.Contains(errorMsg, "23503")
}

// IsNotNullViolationError verifica si el error es por violación de NOT NULL
func IsNotNullViolationError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23502" // not_null_violation
	}

	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "null value") ||
		strings.Contains(errorMsg, "not null") ||
		strings.Contains(errorMsg, "23502")
}

// GetPostgresErrorCode obtiene el código de error PostgreSQL
func GetPostgresErrorCode(err error) string {
	if err == nil {
		return ""
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// =============================================================================
// MANEJO DE STRINGS Y FORMATOS
// =============================================================================

// NormalizeEmail normaliza un email (minúsculas, trim)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizePhone normaliza un teléfono (solo dígitos)
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}

	hasPlus := strings.HasPrefix(phone, "+")
	digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	if hasPlus {
		return "+" + digits
	}
	return digits
}

// TruncateString trunca un string a una longitud máxima
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

// =============================================================================
// VALIDACIONES DE NEGOCIO
// =============================================================================

// IsValidTicketStatus verifica si el estado del ticket es válido
func IsValidTicketStatus(status string) bool {
	validStatuses := []string{
		"available", "reserved", "sold", "used",
		"cancelled", "transferred", "refunded",
	}

	for _, validStatus := range validStatuses {
		if strings.EqualFold(status, validStatus) {
			return true
		}
	}
	return false
}

// IsValidCustomerType verifica si el tipo de cliente es válido
func IsValidCustomerType(customerType string) bool {
	validTypes := []string{"registered", "guest", "corporate"}

	for _, validType := range validTypes {
		if strings.EqualFold(customerType, validType) {
			return true
		}
	}
	return false
}

// IsValidEventStatus verifica si el estado del evento es válido
func IsValidEventStatus(status string) bool {
	validStatuses := []string{
		"draft", "published", "cancelled", "completed", "sold_out",
	}

	for _, validStatus := range validStatuses {
		if strings.EqualFold(status, validStatus) {
			return true
		}
	}
	return false
}

// IsValidUserRole verifica si el rol de usuario es válido
func IsValidUserRole(role string) bool {
	validRoles := []string{"admin", "customer", "organizer", "guest"}

	for _, validRole := range validRoles {
		if strings.EqualFold(role, validRole) {
			return true
		}
	}
	return false
}

// =============================================================================
// FUNCIONES DE FECHA Y TIEMPO
// =============================================================================

// IsFutureDate verifica si una fecha es futura
func IsFutureDate(t time.Time) bool {
	return t.After(time.Now())
}

// IsPastDate verifica si una fecha es pasada
func IsPastDate(t time.Time) bool {
	return t.Before(time.Now())
}

// IsDateRangeValid verifica si un rango de fechas es válido
func IsDateRangeValid(start, end time.Time) bool {
	return !start.IsZero() && !end.IsZero() && end.After(start)
}

// FormatDateForDB formatea fecha para consultas SQL
func FormatDateForDB(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseDateFromString parsea fecha de string
func ParseDateFromString(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de fecha no válido: %s", dateStr)
}

// =============================================================================
// FUNCIONES DE LOGGING Y DEBUG
// =============================================================================

// SafeStringForLog oculta información sensible en logs
func SafeStringForLog(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return s[:2] + "***" + s[len(s)-2:]
}

// FormatPgError formatea error de PostgreSQL para logging
func FormatPgError(err error) string {
	if err == nil {
		return ""
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Sprintf("PGError[%s]: %s (Detail: %s, Where: %s)",
			pgErr.Code, pgErr.Message, pgErr.Detail, pgErr.Where)
	}
	return err.Error()
}
