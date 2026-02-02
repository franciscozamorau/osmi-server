package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

func ValidateValidAge(fl validator.FieldLevel) bool {
	dobStr := fl.Field().String()
	if dobStr == "" {
		return true
	}

	dob, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		return false
	}

	now := time.Now()
	age := now.Year() - dob.Year()

	// Ajustar si el cumpleaños no ha pasado este año
	if now.YearDay() < dob.YearDay() {
		age--
	}

	// Edad válida: 13-120 años
	return age >= 13 && age <= 120
}
