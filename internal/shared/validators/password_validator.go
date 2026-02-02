package validators

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

func ValidateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Requerir: mayúscula, minúscula, número
	// Opcional: carácter especial
	return hasUpper && hasLower && hasNumber
}

func ValidatePasswordNotCommon(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	commonPasswords := map[string]bool{
		"password": true, "123456": true, "qwerty": true,
		"admin": true, "welcome": true, "monkey": true,
		"letmein": true, "password1": true, "abc123": true,
	}

	return !commonPasswords[password]
}
