// internal/api/helpers/helpers.go
package helpers

import (
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// SafeStringPtr convierte *string a string vacío si es nil
func SafeStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SafeTimePtr convierte *time.Time a *timestamppb.Timestamp si no es nil
func SafeTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// SafeInt64Ptr convierte *int64 a int64 con valor por defecto
func SafeInt64Ptr(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// SafeStringID convierte *int64 a string, vacío si es nil
func SafeStringID(id *int64) string {
	if id == nil {
		return ""
	}
	return strconv.FormatInt(*id, 10)
}
