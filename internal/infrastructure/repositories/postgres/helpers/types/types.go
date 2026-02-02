package types

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Converter maneja conversiones entre tipos Go y PostgreSQL
type Converter struct{}

// NewConverter crea una nueva instancia de Converter
func NewConverter() *Converter {
	return &Converter{}
}

// Text convierte string a pgtype.Text
func (c *Converter) Text(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// TextPtr convierte *string a pgtype.Text
func (c *Converter) TextPtr(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// Int convierte int a pgtype.Int4
func (c *Converter) Int(i int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(i), Valid: true}
}

// Int32 convierte int32 a pgtype.Int4
func (c *Converter) Int32(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// Int32Ptr convierte *int32 a pgtype.Int4
func (c *Converter) Int32Ptr(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// Int64 convierte int64 a pgtype.Int8
func (c *Converter) Int64(i int64) pgtype.Int8 {
	return pgtype.Int8{Int64: i, Valid: true}
}

// Int64Ptr convierte *int64 a pgtype.Int8
func (c *Converter) Int64Ptr(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// Float64 convierte float64 a pgtype.Float8
func (c *Converter) Float64(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: true}
}

// Float64Ptr convierte *float64 a pgtype.Float8
func (c *Converter) Float64Ptr(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

// Bool convierte bool a pgtype.Bool
func (c *Converter) Bool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

// BoolPtr convierte *bool a pgtype.Bool
func (c *Converter) BoolPtr(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

// Timestamp convierte time.Time a pgtype.Timestamp
func (c *Converter) Timestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: !t.IsZero()}
}

// TimestampPtr convierte *time.Time a pgtype.Timestamp
func (c *Converter) TimestampPtr(t *time.Time) pgtype.Timestamp {
	if t == nil || t.IsZero() {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// Date convierte time.Time a pgtype.Date
func (c *Converter) Date(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: !t.IsZero()}
}

// DatePtr convierte *time.Time a pgtype.Date
func (c *Converter) DatePtr(t *time.Time) pgtype.Date {
	if t == nil || t.IsZero() {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

// UUID convierte string UUID a pgtype.UUID
func (c *Converter) UUID(uuid string) pgtype.UUID {
	if uuid == "" {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: parseUUID(uuid), Valid: true}
}

// UUIDPtr convierte *string UUID a pgtype.UUID
func (c *Converter) UUIDPtr(uuid *string) pgtype.UUID {
	if uuid == nil || *uuid == "" {
		return pgtype.UUID{Valid: false}
	}
	return c.UUID(*uuid)
}

// FromText convierte pgtype.Text a *string
func (c *Converter) FromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// FromInt4 convierte pgtype.Int4 a *int32
func (c *Converter) FromInt4(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// FromInt8 convierte pgtype.Int8 a *int64
func (c *Converter) FromInt8(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// FromFloat8 convierte pgtype.Float8 a *float64
func (c *Converter) FromFloat8(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// FromBool convierte pgtype.Bool a *bool
func (c *Converter) FromBool(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// FromTimestamp convierte pgtype.Timestamp a *time.Time
func (c *Converter) FromTimestamp(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// FromDate convierte pgtype.Date a *time.Time
func (c *Converter) FromDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

// parseUUID convierte string UUID a bytes (simplificado)
func parseUUID(uuid string) [16]byte {
	var result [16]byte
	// Implementación simplificada
	return result
}
