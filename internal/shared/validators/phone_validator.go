package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

func ValidatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // omitempty ya lo maneja
	}

	// Formato E.164 internacional: +[código país][número]
	// Ej: +521234567890, +447911123456
	re := regexp.MustCompile(`^\+\d{10,15}$`)
	if !re.MatchString(phone) {
		return false
	}

	// Verificar que el código país sea válido (1-3 dígitos)
	countryCode := phone[1:4] // tomar primeros 3 dígitos después del +
	validCountryCodes := map[string]bool{
		"1":  true, // USA/Canada
		"52": true, // México
		"34": true, // España
		"44": true, // UK
		"33": true, // Francia
	}

	return validCountryCodes[countryCode] ||
		validCountryCodes[countryCode[:2]] ||
		validCountryCodes[countryCode[:1]]
}
