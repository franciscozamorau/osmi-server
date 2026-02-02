package scanner

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// RowScanner escanea una fila de resultados
type RowScanner struct{}

// NewRowScanner crea un nuevo RowScanner
func NewRowScanner() *RowScanner {
	return &RowScanner{}
}

// ScanNullableString escanea string nullable
func (rs *RowScanner) ScanNullableString(row pgx.Row) (*string, error) {
	var value pgtype.Text
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan string: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.String, nil
}

// ScanNullableInt32 escanea int32 nullable
func (rs *RowScanner) ScanNullableInt32(row pgx.Row) (*int32, error) {
	var value pgtype.Int4
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan int32: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Int32, nil
}

// ScanNullableInt64 escanea int64 nullable
func (rs *RowScanner) ScanNullableInt64(row pgx.Row) (*int64, error) {
	var value pgtype.Int8
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan int64: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Int64, nil
}

// ScanNullableFloat64 escanea float64 nullable
func (rs *RowScanner) ScanNullableFloat64(row pgx.Row) (*float64, error) {
	var value pgtype.Float8
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan float64: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Float64, nil
}

// ScanNullableBool escanea bool nullable
func (rs *RowScanner) ScanNullableBool(row pgx.Row) (*bool, error) {
	var value pgtype.Bool
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan bool: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Bool, nil
}

// ScanNullableTime escanea time nullable
func (rs *RowScanner) ScanNullableTime(row pgx.Row) (*time.Time, error) {
	var value pgtype.Timestamp
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan time: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Time, nil
}

// ScanNullableDate escanea date nullable
func (rs *RowScanner) ScanNullableDate(row pgx.Row) (*time.Time, error) {
	var value pgtype.Date
	if err := row.Scan(&value); err != nil {
		return nil, fmt.Errorf("failed to scan date: %w", err)
	}
	if !value.Valid {
		return nil, nil
	}
	return &value.Time, nil
}

// ScanRequiredString escanea string requerido
func (rs *RowScanner) ScanRequiredString(row pgx.Row) (string, error) {
	var value string
	if err := row.Scan(&value); err != nil {
		return "", fmt.Errorf("failed to scan required string: %w", err)
	}
	return value, nil
}

// ScanRequiredInt32 escanea int32 requerido
func (rs *RowScanner) ScanRequiredInt32(row pgx.Row) (int32, error) {
	var value int32
	if err := row.Scan(&value); err != nil {
		return 0, fmt.Errorf("failed to scan required int32: %w", err)
	}
	return value, nil
}

// ScanRequiredInt64 escanea int64 requerido
func (rs *RowScanner) ScanRequiredInt64(row pgx.Row) (int64, error) {
	var value int64
	if err := row.Scan(&value); err != nil {
		return 0, fmt.Errorf("failed to scan required int64: %w", err)
	}
	return value, nil
}

// ScanRequiredTime escanea time requerido
func (rs *RowScanner) ScanRequiredTime(row pgx.Row) (time.Time, error) {
	var value time.Time
	if err := row.Scan(&value); err != nil {
		return time.Time{}, fmt.Errorf("failed to scan required time: %w", err)
	}
	return value, nil
}

// ScanRowToMap escanea una fila completa a mapa
func (rs *RowScanner) ScanRowToMap(row pgx.Row, columns []string) (map[string]interface{}, error) {
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := row.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	result := make(map[string]interface{})
	for i, col := range columns {
		result[col] = values[i]
	}

	return result, nil
}

// ConvertSQLNullable convierte tipos sql.Null* a pointers
func (rs *RowScanner) ConvertSQLNullable(nullString sql.NullString) *string {
	if nullString.Valid {
		return &nullString.String
	}
	return nil
}

func (rs *RowScanner) ConvertSQLNullableInt32(nullInt sql.NullInt32) *int32 {
	if nullInt.Valid {
		return &nullInt.Int32
	}
	return nil
}

func (rs *RowScanner) ConvertSQLNullableInt64(nullInt sql.NullInt64) *int64 {
	if nullInt.Valid {
		return &nullInt.Int64
	}
	return nil
}

func (rs *RowScanner) ConvertSQLNullableFloat64(nullFloat sql.NullFloat64) *float64 {
	if nullFloat.Valid {
		return &nullFloat.Float64
	}
	return nil
}

func (rs *RowScanner) ConvertSQLNullableBool(nullBool sql.NullBool) *bool {
	if nullBool.Valid {
		return &nullBool.Bool
	}
	return nil
}

func (rs *RowScanner) ConvertSQLNullableTime(nullTime sql.NullTime) *time.Time {
	if nullTime.Valid {
		return &nullTime.Time
	}
	return nil
}
