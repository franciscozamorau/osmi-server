package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresErrorHandler maneja errores de PostgreSQL
type PostgresErrorHandler struct{}

// NewPostgresErrorHandler crea un nuevo manejador de errores
func NewPostgresErrorHandler() *PostgresErrorHandler {
	return &PostgresErrorHandler{}
}

// PostgresError representa un error de PostgreSQL con contexto
type PostgresError struct {
	Code        string
	Message     string
	Detail      string
	Constraint  string
	Table       string
	Column      string
	OriginalErr error
}

// Error implementa la interfaz error
func (e *PostgresError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("postgres error [%s]: %s (original: %v)", e.Code, e.Message, e.OriginalErr)
	}
	return fmt.Sprintf("postgres error [%s]: %s", e.Code, e.Message)
}

// Unwrap devuelve el error original
func (e *PostgresError) Unwrap() error {
	return e.OriginalErr
}

// Códigos de error PostgreSQL comunes
const (
	UniqueViolation             = "23505"
	ForeignKeyViolation         = "23503"
	NotNullViolation            = "23502"
	CheckViolation              = "23514"
	DeadlockDetected            = "40P01"
	SerializationFailure        = "40001"
	SyntaxError                 = "42601"
	InsufficientPrivilege       = "42501"
	DuplicateDatabase           = "42P04"
	DuplicateTable              = "42P07"
	DuplicateColumn             = "42701"
	DuplicateObject             = "42710"
	UndefinedTable              = "42P01"
	UndefinedColumn             = "42703"
	UndefinedFunction           = "42883"
	UndefinedParameter          = "42P02"
	InvalidTextRepresentation   = "22P02"
	InvalidBinaryRepresentation = "22P01"
	NumericValueOutOfRange      = "22003"
	StringDataRightTruncation   = "22001"
	DatetimeFieldOverflow       = "22008"
	DivisionByZero              = "22012"
	InvalidEscapeSequence       = "22025"
)

// IsPostgresError verifica si es error de PostgreSQL
func (h *PostgresErrorHandler) IsPostgresError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr)
}

// ExtractPostgresError extrae información del error PostgreSQL
func (h *PostgresErrorHandler) ExtractPostgresError(err error) *PostgresError {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	return &PostgresError{
		Code:        pgErr.Code,
		Message:     pgErr.Message,
		Detail:      pgErr.Detail,
		Constraint:  pgErr.ConstraintName,
		Table:       pgErr.TableName,
		Column:      pgErr.ColumnName,
		OriginalErr: err,
	}
}

// IsDuplicateKey verifica si es error de clave duplicada
func (h *PostgresErrorHandler) IsDuplicateKey(err error) bool {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return false
	}

	return pgErr.Code == UniqueViolation ||
		strings.Contains(strings.ToLower(pgErr.Message), "duplicate key") ||
		strings.Contains(strings.ToLower(pgErr.Message), "already exists")
}

// IsForeignKeyViolation verifica si es error de foreign key
func (h *PostgresErrorHandler) IsForeignKeyViolation(err error) bool {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return false
	}

	return pgErr.Code == ForeignKeyViolation ||
		strings.Contains(strings.ToLower(pgErr.Message), "foreign key")
}

// IsNotNullViolation verifica si es error de NOT NULL
func (h *PostgresErrorHandler) IsNotNullViolation(err error) bool {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return false
	}

	return pgErr.Code == NotNullViolation ||
		strings.Contains(strings.ToLower(pgErr.Message), "null value") ||
		strings.Contains(strings.ToLower(pgErr.Message), "not null")
}

// IsCheckViolation verifica si es error de CHECK constraint
func (h *PostgresErrorHandler) IsCheckViolation(err error) bool {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return false
	}

	return pgErr.Code == CheckViolation ||
		strings.Contains(strings.ToLower(pgErr.Message), "check constraint")
}

// IsDeadlock verifica si es deadlock
func (h *PostgresErrorHandler) IsDeadlock(err error) bool {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return false
	}

	return pgErr.Code == DeadlockDetected ||
		strings.Contains(strings.ToLower(pgErr.Message), "deadlock")
}

// IsSerializationFailure verifica si es falla de serialización
func (h *PostgresErrorHandler) IsSerializationFailure(err error) bool {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return false
	}

	return pgErr.Code == SerializationFailure ||
		strings.Contains(strings.ToLower(pgErr.Message), "serialization")
}

// IsConnectionError verifica si es error de conexión
func (h *PostgresErrorHandler) IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	connectionKeywords := []string{
		"connection",
		"connect",
		"network",
		"timeout",
		"refused",
		"closed",
		"reset",
		"broken",
		"lost",
		"failed",
	}

	for _, keyword := range connectionKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// ShouldRetry determina si se debe reintentar la operación
func (h *PostgresErrorHandler) ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Errores que pueden reintentarse
	retryableCodes := []string{
		DeadlockDetected,
		SerializationFailure,
		"55P03", // lock_not_available
		"57014", // query_canceled
		"57P01", // admin_shutdown
		"57P02", // crash_shutdown
		"57P03", // cannot_connect_now
	}

	pgErr := h.ExtractPostgresError(err)
	if pgErr != nil {
		for _, code := range retryableCodes {
			if pgErr.Code == code {
				return true
			}
		}
	}

	// También reintentar errores de conexión
	return h.IsConnectionError(err)
}

// FormatForLog formatea error para logging
func (h *PostgresErrorHandler) FormatForLog(err error) string {
	if err == nil {
		return ""
	}

	pgErr := h.ExtractPostgresError(err)
	if pgErr != nil {
		return fmt.Sprintf("PGError[%s] Table:%s Constraint:%s Column:%s Detail:%s Message:%s",
			pgErr.Code, pgErr.Table, pgErr.Constraint, pgErr.Column,
			pgErr.Detail, pgErr.Message)
	}

	return fmt.Sprintf("Database error: %v", err)
}

// GetConstraintName obtiene el nombre del constraint del error
func (h *PostgresErrorHandler) GetConstraintName(err error) string {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return ""
	}

	if pgErr.Constraint != "" {
		return pgErr.Constraint
	}

	// Intentar extraer del mensaje
	if strings.Contains(pgErr.Message, "constraint") {
		parts := strings.Split(pgErr.Message, "\"")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	return ""
}

// GetTableName obtiene el nombre de la tabla del error
func (h *PostgresErrorHandler) GetTableName(err error) string {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return ""
	}
	return pgErr.Table
}

// GetColumnName obtiene el nombre de la columna del error
func (h *PostgresErrorHandler) GetColumnName(err error) string {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil {
		return ""
	}
	return pgErr.Column
}

// GetDuplicateValue obtiene el valor duplicado del error
func (h *PostgresErrorHandler) GetDuplicateValue(err error) string {
	pgErr := h.ExtractPostgresError(err)
	if pgErr == nil || pgErr.Detail == "" {
		return ""
	}

	// El detalle suele tener formato: "Key (column)=(value) already exists."
	if strings.Contains(pgErr.Detail, ")=(") {
		parts := strings.Split(pgErr.Detail, ")=(")
		if len(parts) > 1 {
			valuePart := strings.Split(parts[1], ")")
			if len(valuePart) > 0 {
				return valuePart[0]
			}
		}
	}

	return ""
}

// WrapError envuelve un error con contexto adicional
func (h *PostgresErrorHandler) WrapError(err error, context, operation string) error {
	if err == nil {
		return nil
	}

	pgErr := h.ExtractPostgresError(err)
	if pgErr != nil {
		return fmt.Errorf("%s: failed to %s: %w", context, operation, pgErr)
	}

	return fmt.Errorf("%s: failed to %s: %w", context, operation, err)
}

// CreateUserFriendlyError crea un error amigable para el usuario
func (h *PostgresErrorHandler) CreateUserFriendlyError(err error, entity string) error {
	if err == nil {
		return nil
	}

	if h.IsDuplicateKey(err) {
		field := h.GetConstraintName(err)
		value := h.GetDuplicateValue(err)

		if field != "" && value != "" {
			return fmt.Errorf("%s with %s '%s' already exists", entity, field, value)
		}
		return fmt.Errorf("%s already exists", entity)
	}

	if h.IsForeignKeyViolation(err) {
		return fmt.Errorf("cannot %s because related data does not exist", strings.ToLower(entity))
	}

	if h.IsNotNullViolation(err) {
		column := h.GetColumnName(err)
		if column != "" {
			return fmt.Errorf("%s is required", column)
		}
		return fmt.Errorf("required field is missing")
	}

	if h.IsCheckViolation(err) {
		return fmt.Errorf("invalid data provided for %s", strings.ToLower(entity))
	}

	// Error genérico
	return fmt.Errorf("failed to process %s: %v", strings.ToLower(entity), err)
}
