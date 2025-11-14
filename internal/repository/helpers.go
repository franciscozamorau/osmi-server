package repository

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// toPgText convierte un string a pgtype.Text
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: strings.TrimSpace(s), Valid: s != ""}
}

// isValidEmail valida formato de email
func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

// isDuplicateKeyError verifica si el error es por violación de unique constraint
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}

	// Fallback: verificar el mensaje de error
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "duplicate key") ||
		strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "unique constraint") ||
		strings.Contains(errorMsg, "23505")
}
