package validations

import (
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// DomainValidator valida entidades de dominio
type DomainValidator struct{}

// NewDomainValidator crea un nuevo DomainValidator
func NewDomainValidator() *DomainValidator {
	return &DomainValidator{}
}

// ValidateUser valida entidad User
func (dv *DomainValidator) ValidateUser(user *entities.User) (bool, []string) {
	var errors []string

	// Validar email
	if user.Email == "" {
		errors = append(errors, "email is required")
	} else if !IsValidEmail(user.Email) {
		errors = append(errors, "invalid email format")
	}

	// Validar username si está presente
	if user.Username != nil && *user.Username != "" {
		if !IsValidUsername(*user.Username) {
			errors = append(errors, "invalid username format")
		}
	}

	// Validar teléfono si está presente
	if user.Phone != nil && *user.Phone != "" {
		if !IsValidPhone(*user.Phone) {
			errors = append(errors, "invalid phone number")
		}
	}

	// Validar nombres
	if user.FirstName != nil && *user.FirstName != "" {
		if !IsValidName(*user.FirstName) {
			errors = append(errors, "invalid first name")
		}
	}

	if user.LastName != nil && *user.LastName != "" {
		if !IsValidName(*user.LastName) {
			errors = append(errors, "invalid last name")
		}
	}

	// Validar fecha de nacimiento si está presente
	if user.DateOfBirth != nil && !user.DateOfBirth.IsZero() {
		if user.DateOfBirth.After(time.Now()) {
			errors = append(errors, "date of birth cannot be in the future")
		}
	}

	// Validar zona horaria si está presente
	if user.Timezone != "" && !IsValidTimezone(user.Timezone) {
		errors = append(errors, "invalid timezone")
	}

	// Validar moneda preferida si está presente
	if user.PreferredCurrency != "" && !IsValidCurrencyCode(user.PreferredCurrency) {
		errors = append(errors, "invalid currency code")
	}

	// Validar idioma preferido si está presente
	if user.PreferredLanguage != "" && !IsValidLanguageCode(user.PreferredLanguage) {
		errors = append(errors, "invalid language code")
	}

	return len(errors) == 0, errors
}

// ValidateTicket valida entidad Ticket
func (dv *DomainValidator) ValidateTicket(ticket *entities.Ticket) (bool, []string) {
	var errors []string

	// Validar código
	if ticket.Code == "" {
		errors = append(errors, "ticket code is required")
	} else if len(ticket.Code) < 6 || len(ticket.Code) > 50 {
		errors = append(errors, "ticket code must be between 6 and 50 characters")
	}

	// Validar estado
	if ticket.Status == "" {
		errors = append(errors, "ticket status is required")
	} else if !IsValidTicketStatus(ticket.Status) {
		validStatuses := []string{"available", "reserved", "sold", "checked_in", "cancelled", "refunded"}
		errors = append(errors, fmt.Sprintf("invalid ticket status. Must be one of: %s", strings.Join(validStatuses, ", ")))
	}

	// Validar precio
	if ticket.FinalPrice < 0 {
		errors = append(errors, "ticket price cannot be negative")
	}

	if ticket.FinalPrice > 1000000 {
		errors = append(errors, "ticket price cannot exceed 1,000,000")
	}

	// Validar moneda
	if ticket.Currency == "" {
		errors = append(errors, "currency is required")
	} else if !IsValidCurrencyCode(ticket.Currency) {
		errors = append(errors, "invalid currency code")
	}

	// Validar tax
	if ticket.TaxAmount < 0 {
		errors = append(errors, "tax amount cannot be negative")
	}

	return len(errors) == 0, errors
}
