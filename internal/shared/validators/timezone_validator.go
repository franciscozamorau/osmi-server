package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

func ValidateTimezone(fl validator.FieldLevel) bool {
	tz := fl.Field().String()
	if tz == "" {
		return true
	}

	_, err := time.LoadLocation(tz)
	return err == nil
}
