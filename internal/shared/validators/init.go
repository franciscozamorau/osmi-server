package validators

import (
	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("phone", ValidatePhone)
	v.RegisterValidation("strong_password", ValidateStrongPassword)
	v.RegisterValidation("valid_age", ValidateValidAge)
	v.RegisterValidation("timezone", ValidateTimezone)
	v.RegisterValidation("iso4217", ValidateISO4217)
	v.RegisterValidation("alpha", ValidateAlpha)
	v.RegisterValidation("alphanum", ValidateAlphaNum)
	v.RegisterValidation("uuid4", ValidateUUID4)
}
