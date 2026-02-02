package validations

import (
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/valueobjects"
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

	// Validar roles
	if user.Role != "" && !IsValidUserRole(user.Role) {
		validRoles := []string{"admin", "customer", "organizer", "guest", "staff"}
		errors = append(errors, fmt.Sprintf("invalid role. Must be one of: %s", strings.Join(validRoles, ", ")))
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

	// Validar código si está presente
	if ticket.Code != nil && *ticket.Code != "" {
		if len(*ticket.Code) < 6 || len(*ticket.Code) > 50 {
			errors = append(errors, "ticket code must be between 6 and 50 characters")
		}
	}

	// Validar tipo
	if ticket.Type == "" {
		errors = append(errors, "ticket type is required")
	} else if !IsValidTicketType(ticket.Type) {
		validTypes := []string{"general", "vip", "premium", "student", "senior"}
		errors = append(errors, fmt.Sprintf("invalid ticket type. Must be one of: %s", strings.Join(validTypes, ", ")))
	}

	// Validar estado
	if ticket.Status == "" {
		errors = append(errors, "ticket status is required")
	} else if !IsValidTicketStatus(ticket.Status) {
		validStatuses := []string{"available", "reserved", "sold", "used", "cancelled"}
		errors = append(errors, fmt.Sprintf("invalid ticket status. Must be one of: %s", strings.Join(validStatuses, ", ")))
	}

	// Validar precio
	if ticket.Price < 0 {
		errors = append(errors, "ticket price cannot be negative")
	}

	if ticket.Price > 1000000 {
		errors = append(errors, "ticket price cannot exceed 1,000,000")
	}

	// Validar moneda
	if ticket.Currency == "" {
		errors = append(errors, "currency is required")
	} else if !IsValidCurrencyCode(ticket.Currency) {
		errors = append(errors, "invalid currency code")
	}

	// Validar fechas
	if ticket.PurchaseDate != nil && !ticket.PurchaseDate.IsZero() {
		if ticket.PurchaseDate.After(time.Now()) {
			errors = append(errors, "purchase date cannot be in the future")
		}
	}

	if ticket.ValidUntil != nil && !ticket.ValidUntil.IsZero() {
		if ticket.ValidUntil.Before(time.Now()) {
			errors = append(errors, "ticket is already expired")
		}
	}

	return len(errors) == 0, errors
}

// ValidateEvent valida entidad Event
func (dv *DomainValidator) ValidateEvent(event *entities.Event) (bool, []string) {
	var errors []string

	// Validar nombre
	if event.Name == "" {
		errors = append(errors, "event name is required")
	} else if len(event.Name) > 200 {
		errors = append(errors, "event name cannot exceed 200 characters")
	}

	// Validar descripción si está presente
	if event.Description != nil && len(*event.Description) > 5000 {
		errors = append(errors, "event description cannot exceed 5000 characters")
	}

	// Validar fechas
	if event.StartDate.IsZero() {
		errors = append(errors, "event start date is required")
	}

	if event.EndDate.IsZero() {
		errors = append(errors, "event end date is required")
	}

	if !event.StartDate.IsZero() && !event.EndDate.IsZero() {
		if !event.EndDate.After(event.StartDate) {
			errors = append(errors, "event end date must be after start date")
		}

		if event.StartDate.Before(time.Now()) {
			errors = append(errors, "event cannot start in the past")
		}
	}

	// Validar estado
	if event.Status == "" {
		errors = append(errors, "event status is required")
	} else if !IsValidEventStatus(event.Status) {
		validStatuses := []string{"draft", "published", "cancelled", "completed"}
		errors = append(errors, fmt.Sprintf("invalid event status. Must be one of: %s", strings.Join(validStatuses, ", ")))
	}

	// Validar tipo
	if event.Type == "" {
		errors = append(errors, "event type is required")
	} else if !IsValidEventType(event.Type) {
		validTypes := []string{"concert", "conference", "workshop", "festival"}
		errors = append(errors, fmt.Sprintf("invalid event type. Must be one of: %s", strings.Join(validTypes, ", ")))
	}

	// Validar capacidad
	if event.Capacity <= 0 {
		errors = append(errors, "event capacity must be greater than 0")
	}

	if event.Capacity > 100000 {
		errors = append(errors, "event capacity cannot exceed 100,000")
	}

	// Validar URLs si están presentes
	if event.ImageURL != nil && *event.ImageURL != "" {
		if !IsValidURL(*event.ImageURL) {
			errors = append(errors, "invalid image URL")
		}
	}

	if event.WebsiteURL != nil && *event.WebsiteURL != "" {
		if !IsValidURL(*event.WebsiteURL) {
			errors = append(errors, "invalid website URL")
		}
	}

	return len(errors) == 0, errors
}

// ValidateOrder valida entidad Order
func (dv *DomainValidator) ValidateOrder(order *entities.Order) (bool, []string) {
	var errors []string

	// Validar estado
	if order.Status == "" {
		errors = append(errors, "order status is required")
	} else if !IsValidOrderStatus(order.Status) {
		validStatuses := []string{"pending", "confirmed", "completed", "cancelled"}
		errors = append(errors, fmt.Sprintf("invalid order status. Must be one of: %s", strings.Join(validStatuses, ", ")))
	}

	// Validar total
	if order.TotalAmount < 0 {
		errors = append(errors, "order total cannot be negative")
	}

	// Validar moneda
	if order.Currency == "" {
		errors = append(errors, "currency is required")
	} else if !IsValidCurrencyCode(order.Currency) {
		errors = append(errors, "invalid currency code")
	}

	// Validar método de pago si está presente
	if order.PaymentMethod != nil && *order.PaymentMethod != "" {
		if !IsValidPaymentMethod(*order.PaymentMethod) {
			errors = append(errors, "invalid payment method")
		}
	}

	// Validar estado de pago si está presente
	if order.PaymentStatus != nil && *order.PaymentStatus != "" {
		if !IsValidPaymentStatus(*order.PaymentStatus) {
			errors = append(errors, "invalid payment status")
		}
	}

	return len(errors) == 0, errors
}

// ValidateCustomer valida entidad Customer
func (dv *DomainValidator) ValidateCustomer(customer *entities.Customer) (bool, []string) {
	var errors []string

	// Validar tipo
	if customer.Type == "" {
		errors = append(errors, "customer type is required")
	} else if !IsValidCustomerType(customer.Type) {
		validTypes := []string{"registered", "guest", "corporate"}
		errors = append(errors, fmt.Sprintf("invalid customer type. Must be one of: %s", strings.Join(validTypes, ", ")))
	}

	// Validar email si está presente
	if customer.Email != nil && *customer.Email != "" {
		if !IsValidEmail(*customer.Email) {
			errors = append(errors, "invalid email")
		}
	}

	// Validar teléfono si está presente
	if customer.Phone != nil && *customer.Phone != "" {
		if !IsValidPhone(*customer.Phone) {
			errors = append(errors, "invalid phone number")
		}
	}

	// Validar nombres si están presentes
	if customer.FirstName != nil && *customer.FirstName != "" {
		if !IsValidName(*customer.FirstName) {
			errors = append(errors, "invalid first name")
		}
	}

	if customer.LastName != nil && *customer.LastName != "" {
		if !IsValidName(*customer.LastName) {
			errors = append(errors, "invalid last name")
		}
	}

	return len(errors) == 0, errors
}

// ValidateValueObject valida value objects
func (dv *DomainValidator) ValidateValueObject(vo interface{}) (bool, []string) {
	var errors []string

	switch v := vo.(type) {
	case valueobjects.Email:
		if !IsValidEmail(v.String()) {
			errors = append(errors, "invalid email")
		}
	case valueobjects.Phone:
		if !IsValidPhone(v.String()) {
			errors = append(errors, "invalid phone number")
		}
	case valueobjects.UUID:
		if !IsValidUUID(v.String()) {
			errors = append(errors, "invalid UUID")
		}
	default:
		errors = append(errors, "unsupported value object type")
	}

	return len(errors) == 0, errors
}

// ValidateEntityState valida estado de la entidad
func (dv *DomainValidator) ValidateEntityState(entity interface{}) (bool, []string) {
	var errors []string

	// Validaciones comunes para todas las entidades
	switch e := entity.(type) {
	case *entities.User:
		if !e.IsActive && e.DeletedAt == nil {
			errors = append(errors, "inactive user must have deletion timestamp")
		}
	case *entities.Ticket:
		if e.Status == "used" && e.UsedAt == nil {
			errors = append(errors, "used ticket must have usage timestamp")
		}
	case *entities.Event:
		if e.Status == "completed" && e.EndDate.After(time.Now()) {
			errors = append(errors, "completed event cannot have future end date")
		}
	case *entities.Order:
		if e.Status == "completed" && e.CompletedAt == nil {
			errors = append(errors, "completed order must have completion timestamp")
		}
	}

	return len(errors) == 0, errors
}
