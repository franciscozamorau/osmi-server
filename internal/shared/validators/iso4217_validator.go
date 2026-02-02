package validators

import (
	"github.com/go-playground/validator/v10"
)

func ValidateISO4217(fl validator.FieldLevel) bool {
	currency := fl.Field().String()

	validCurrencies := map[string]bool{
		"MXN": true, "USD": true, "EUR": true, "GBP": true,
		"CAD": true, "AUD": true, "JPY": true, "CNY": true,
		"BRL": true, "ARS": true, "CLP": true, "COP": true,
		"PEN": true, "UYU": true, "VES": true,
	}

	return validCurrencies[currency]
}
